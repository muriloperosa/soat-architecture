# 0008. Deploy no EKS por tag semver, com a CI como etapa obrigatória e credenciais por OIDC

## Status
Aceita

## Contexto
A CI (`.github/workflows/ci.yml`) valida pull requests e a `main`, mas não publica nada. A Fase 2 pede um pipeline que também construa e publique a imagem, aplique os manifestos, rode as migrations e faça o deploy no cluster Kubernetes. O cluster é um EKS numa conta AWS própria do projeto.

Precisávamos decidir:

- o que dispara um deploy e como garantir que só código validado chega ao cluster;
- como a versão da imagem é identificada;
- como o GitHub Actions se autentica na AWS sem guardar chaves de acesso;
- de onde vêm os segredos da aplicação (`DB_PASSWORD`, `JWT_SECRET`), que hoje ficam num `secrets.env` local.

## Decisão
- **Gatilho:** push de uma tag `MAJOR.MINOR.PATCH` (ex.: `1.0.1`). Um job `guard` recusa a tag se o commit dela não estiver na `main`, porque o filtro de tags do GitHub não olha a branch. Merge na `main` não faz deploy. O `workflow_dispatch` numa tag existente serve para redeploy e rollback.
- **CI verde antes de tocar a AWS, sem rodar de novo:** o ruleset da `main` já exige todos os checks de CI no PR, com a branch atualizada e sem bypass. O job `guard` consulta a API do GitHub e exige que o check **CI gate** do commit da tag tenha terminado com sucesso, esperando até 30 minutos se ainda estiver rodando. Se a CI falhou ou não existe, nenhuma imagem é publicada e o cluster não muda.
- **Versão da imagem:** a própria tag (`oficina-api:1.0.1`), num repositório ECR com tags imutáveis. Um redeploy da mesma versão reaproveita a imagem que já está no ECR.
- **Credenciais AWS por OIDC:** os jobs `image` e `deploy` pedem um token OIDC ao GitHub e assumem a role `oficina-github-deploy`, com credenciais temporárias. A role só aceita o environment `aws-deploy` deste repositório e só pode publicar no ECR do projeto, gerar o kubeconfig do cluster e ler o secret da aplicação. O acesso dentro do Kubernetes vem de um access entry do EKS. Provedor OIDC, role, ECR e access entry são criados uma vez por um administrador, conforme `docs/CI_CD.md`.
- **Segredos da aplicação:** ficam no AWS Secrets Manager (`oficina/api`, com `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `JWT_SECRET`). O job de deploy lê o secret, mascara os valores no log e gera um `secrets.env` temporário que o Kustomize transforma no Secret do Kubernetes. O arquivo é apagado no fim do job.
- **Deploy no cluster:** o job chama `make k8s-deploy K8S_OVERLAY=eks`, o mesmo alvo usado manualmente, que aplica configuração, roda o Job de migration e faz o rollout (ADR 0006).
- **Terraform no pipeline, desligado por enquanto:** o job `infra` (`terraform apply` de `infra/` com role e environment próprios, `aws-infra`) existe no `cd.yml`, mas está desligado (`if: false`) e aparece como pulado em todo run. Ele roda entre a publicação da imagem e o deploy, e o deploy roda com ele pulado. Enquanto isso, cluster, RDS e rede vêm do Terraform de `infra/` aplicado fora do pipeline. Endpoint e usuário do banco substituem os placeholders do overlay EKS só dentro do runner.

Alternativas descartadas:

- **Chaves de acesso de um usuário IAM em secrets do GitHub:** credencial de longa duração, que precisa ser rotacionada à mão e vale para qualquer job que consiga ler o secret.
- **Rodar a CI inteira de novo no CD (`workflow_call`):** com o ruleset, repetiria checks que já passaram no mesmo conteúdo e somaria 15–25 minutos a cada release.
- **Deploy a cada merge na `main`:** cada merge viraria release, sem controle de quando algo vai para o ar e sem um número de versão legível.
- **External Secrets Operator no cluster:** possível com conta própria (Pod Identity), mas adiciona um componente a operar no cluster. Ler o Secrets Manager no pipeline atende hoje; o operador continua como evolução se os segredos passarem a mudar sem deploy.
- **Tag da imagem pelo SHA do commit:** continua sendo o padrão do deploy manual pelo Makefile, mas a tag semver mostra direto qual release está rodando e é o que aparece no ECR e no resumo do run.

## Consequências
Positivas:
- Uma versão só chega ao cluster depois de passar por todos os checks da CI, e o release leva minutos em vez de repetir a CI inteira.
- Nenhuma chave da AWS e nenhum segredo da aplicação ficam no repositório ou nos secrets do GitHub. As credenciais do pipeline duram no máximo uma hora.
- Um job de pull request não consegue assumir a role, porque o trust exige o environment `aws-deploy`.
- O que está no ar é identificado pela tag (`kubectl get deploy -o wide` mostra `oficina-api:1.0.1`), e rollback é rodar o deploy da tag anterior.
- O pipeline e o deploy manual usam o mesmo alvo do Makefile, então a ordem config → migration → rollout é uma só.

Negativas:
- A preparação da conta (provedor OIDC, role, ECR, access entry, secret) é manual, feita uma vez com a AWS CLI.
- A role de deploy é admin dentro do cluster, porque o primeiro deploy cria o namespace. Dá para restringir ao namespace `oficina` se o namespace passar a ser criado pelo Terraform.
- Commits da `main` que alteram só Markdown não disparam a CI no push (filtro `paths-ignore`). Uma tag num desses commits falha no `guard` até alguém rodar a CI manualmente na `main`.
- Uma vulnerabilidade divulgada depois do merge não é pega no deploy, porque o SCA não roda de novo. Ela aparece na próxima execução da CI.
- Tags imutáveis no ECR impedem republicar uma versão: uma tag criada no commit errado e já publicada exige um novo número de versão.
- EKS não tem free tier: o control plane cobra por hora mesmo sem tráfego. Com o crédito de US$ 110 da conta, o ambiente completo ligado o mês inteiro não fecha o mês; o cluster deve subir perto das demonstrações.
