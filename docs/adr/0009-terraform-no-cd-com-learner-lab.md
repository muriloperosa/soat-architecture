# 0009. Terraform no CD, state no S3 e AWS Academy Learner Lab

## Status
Aceita. Substitui, na [ADR 0008](0008-deploy-por-tag-no-eks.md), as decisões sobre credenciais AWS por OIDC e sobre o Terraform fora do pipeline.

## Contexto
O ambiente saiu da conta própria do projeto e foi para o AWS Academy Learner Lab. O lab não deixa criar provedor OIDC nem role IAM, e as credenciais da sessão expiram em cerca de 4 horas. A role `oficina-github-deploy` e as roles do cluster e dos nós, criadas na conta antiga, não existem no lab.

Até aqui, o release tinha dois passos manuais: rodar `make tf-apply` com o state local e depois criar a tag. O state ficava na máquina de quem aplicava, e o job `infra` do CD estava desligado (`if: false`) por falta de backend remoto.

## Decisão
- **Terraform no CD, a cada tag:** o job `infra` roda `terraform apply` de `infra/` depois do `guard` e antes do build da imagem. A tag é o único passo manual do release.
- **State no S3:** backend `s3` parcial, com bucket `oficina-tfstate-<conta>`, key `oficina/terraform.tfstate` e lock por arquivo (`use_lockfile`). O bucket é criado por `infra/bootstrap-state.sh`, chamado pelo CD e pelo `make tf-init`, porque o backend precisa dele antes do `init`.
- **LabRole:** cluster e nós usam a `LabRole` que já existe na conta, lida por `data "aws_iam_role"`. Nenhuma role IAM é criada.
- **Credenciais da sessão do lab:** `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` e `AWS_SESSION_TOKEN` como secrets do environment `aws-deploy`. Pipeline e Terraform local usam a mesma role (`voclabs`), que vira admin do cluster ao criá-lo; não há access entry.
- **ECR no Terraform:** o repositório `oficina-api` passa a ser criado pelo Terraform, com `force_delete`, porque a conta do lab pode ser reiniciada e não há mais preparação manual da conta.
- **E-mail fora do Terraform:** a infraestrutura não cria nem SES nem SMTP. O Learner Lab negou `ses:VerifyEmailIdentity`, então SMTP via secret separado no Secrets Manager é o caminho usado no ambiente. O adapter SES permanece disponível para configuração manual. Sem provedor configurado, o CD usa `EMAIL_PROVIDER=log`.

Alternativas descartadas:

- **Aplicar a cada merge na `main` e criar a tag automaticamente:** tiraria do time o controle de quando sai uma versão e de qual número ela recebe.
- **State versionado no repositório:** contém a senha do banco e o `JWT_SECRET`.
- **Bucket do state criado pelo próprio Terraform:** o backend precisa do bucket antes do `init`; seria um segundo Terraform só para isso.
- **Usuário IAM com access key fixa:** o lab não deixa criar usuário IAM.

## Consequências
Positivas:
- O release é uma tag: infraestrutura, imagem e rollout saem no mesmo run, sempre do mesmo commit.
- O state é compartilhado entre o pipeline e quem roda o Terraform localmente, com lock.
- Uma conta de lab nova sobe do zero com uma tag, depois de atualizar os secrets.

Negativas:
- As credenciais vencem com a sessão do lab. Cada sessão exige atualizar três secrets no GitHub antes da tag.
- As credenciais do lab dão acesso amplo à conta, não o mínimo que o pipeline precisa. Quem tem acesso ao environment `aws-deploy` tem acesso à conta do lab.
- Todo run de tag, inclusive rollback, passa pelo `terraform apply`, o que soma cerca de um minuto quando nada mudou.
- O control plane do EKS, o RDS e o Load Balancer continuam cobrando com a sessão do lab encerrada.
