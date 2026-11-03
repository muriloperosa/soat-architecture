# Preparação do AWS Academy Learner Lab para o CD

Roteiro para deixar o GitHub pronto para o workflow [`cd.yml`](../.github/workflows/cd.yml) numa conta do AWS Academy Learner Lab. A infraestrutura (EKS, RDS, ECR e os secrets `oficina/api` e `oficina/smtp`) é criada pelo próprio CD, pelo Terraform de [`infra/`](../infra/) ([`TERRAFORM.md`](TERRAFORM.md)). Aqui só entram credenciais e configuração do GitHub.

## Restrições do AWS Academy

O Learner Lab é uma conta AWS temporária com permissões reduzidas. Três limitações moldam o pipeline:

| Limitação | Consequência no projeto |
| --- | --- |
| Não deixa criar provedor OIDC, role nem usuário IAM | O CD usa as credenciais temporárias da sessão (role `voclabs`), guardadas como secrets do GitHub. Cluster e nós usam a `LabRole` que já existe na conta ([ADR 0009](adr/0009-terraform-no-cd-com-learner-lab.md)) |
| As credenciais expiram com a sessão, em cerca de 4 horas | Os três secrets AWS precisam ser atualizados a cada sessão, antes de criar a tag |
| Nega `ses:VerifyEmailIdentity` | O Amazon SES, provedor de e-mail planejado, não envia sem remetente verificado. No lab, o e-mail sai por SMTP (Gmail), configurado pelo secret `SMTP_PASSWORD` ([ADR 0010](adr/0010-smtp-no-lugar-do-ses-no-learner-lab.md)) |

Cada integrante do grupo tem a própria conta de Learner Lab. Escolham **uma** conta para o ambiente e usem sempre as credenciais dela no GitHub: o state, o cluster e o bucket ficam nessa conta, e credenciais de outra conta não enxergam nada disso.

Pré-requisitos:

- Sessão do Learner Lab iniciada (**Start Lab**, indicador verde)
- GitHub CLI autenticada (`gh auth login`), com permissão de admin no repositório
- Uma conta Gmail com verificação em duas etapas e uma [senha de app](https://myaccount.google.com/apppasswords) gerada para o envio

## Resumo: o que configurar no GitHub

As credenciais sensíveis ficam no environment `aws-deploy` (**Settings → Environments → aws-deploy → Environment secrets**); `PUBLIC_BASE_URL` é uma variable do mesmo environment.

| Nome | Tipo | Quando atualizar | Valor |
| --- | --- | --- | --- |
| `AWS_ACCESS_KEY_ID` | Secret | **A cada sessão do lab** | `aws_access_key_id` de **AWS Details → AWS CLI** |
| `AWS_SECRET_ACCESS_KEY` | Secret | **A cada sessão do lab** | `aws_secret_access_key` |
| `AWS_SESSION_TOKEN` | Secret | **A cada sessão do lab** | `aws_session_token` |
| `SMTP_PASSWORD` | Secret | Uma vez, e ao trocar a senha | Senha de app do Gmail remetente, sem espaços |
| `PUBLIC_BASE_URL` | Variable | Ao mudar o endereço público | Origem da API, com `http://` ou `https://`, sem `/v1/health` |
| `SMTP_SECRET_VERSION` | Variable | Ao trocar `SMTP_PASSWORD` | Número inteiro, incrementado a cada troca (padrão `1`) |

As demais variables são opcionais (passo 4).

## 1. Environment `aws-deploy` (uma vez)

```bash
export REPO=muriloperosa/soat-architecture
gh api -X PUT "repos/${REPO}/environments/aws-deploy"
```

Recomendado: em **Settings → Environments → aws-deploy → Deployment branches and tags**, escolha **Selected** e adicione a regra de tag `[0-9]*.[0-9]*.[0-9]*`. Só os runs de tag passam a ler os secrets.

## 2. Credenciais da sessão (a cada sessão do lab)

No Learner Lab, abra **AWS Details → AWS CLI → Show**. Copie os três valores e grave no environment `aws-deploy`:

```bash
gh secret set AWS_ACCESS_KEY_ID     --repo "$REPO" --env aws-deploy --body '<aws_access_key_id>'
gh secret set AWS_SECRET_ACCESS_KEY --repo "$REPO" --env aws-deploy --body '<aws_secret_access_key>'
gh secret set AWS_SESSION_TOKEN     --repo "$REPO" --env aws-deploy --body '<aws_session_token>'
```

Secrets de repositório com esses mesmos nomes também podem ser lidos pelos jobs; secrets do environment `aws-deploy` têm precedência. Prefira o environment para limitar o uso às etapas de deploy. As credenciais expiram quando a sessão acaba (cerca de 4 horas). Um run com credencial vencida falha no primeiro passo da AWS com `ExpiredToken` ou `InvalidClientTokenId`: atualize os três secrets e rode de novo (**Actions → CD → Run workflow**, escolhendo a tag).

## 3. E-mail: senha de app do Gmail (uma vez)

O provedor de e-mail planejado é o Amazon SES, mas o lab não deixa verificar o remetente. Por isso o ambiente envia por SMTP com STARTTLS (`smtp.gmail.com:587`).

```bash
gh secret set SMTP_PASSWORD --repo "$REPO" --env aws-deploy --body '<senha_de_app_sem_espacos>'
```

Não use uma variable, pois ela não é criptografada. O Terraform cria `oficina/smtp` no Secrets Manager e grava a senha pela propriedade write-only `secret_string_wo`, que não persiste o valor no state nem no plano. O CD lê esse secret e configura `EMAIL_PROVIDER=smtp`.

Ao trocar a senha, atualize o secret e incremente `SMTP_SECRET_VERSION`. Sem isso, o Terraform não publica uma nova versão:

```bash
gh variable set SMTP_SECRET_VERSION --repo "$REPO" --body '2'
```

## 4. Links de aprovação do orçamento

O Terraform gera automaticamente `APPROVAL_TOKEN_SECRET` como 32 bytes aleatórios em hexadecimal e o armazena no Secrets Manager `oficina/api`, separado do `JWT_SECRET`. Não é necessário cadastrar esse segredo no GitHub. O workflow o lê do Secrets Manager e o injeta no Secret Kubernetes.

Cadastre somente `PUBLIC_BASE_URL` como **Environment Variable** no environment `aws-deploy`. Use a origem pública da API, sem caminho, query ou fragmento. Por exemplo: `http://<hostname-do-load-balancer>` ou `https://api.seudominio.com`; não inclua `/v1/health`. O valor deve ser alcançável pelo cliente. O segredo gerado pelo Terraform permanece estável entre deploys; rotacioná-lo invalida links ainda válidos. A validade dos links é de 24 horas. Ao alterar `PUBLIC_BASE_URL`, rode novamente o workflow CD para atualizar a API.

Se `APPROVAL_TOKEN_SECRET` já estiver cadastrado como GitHub Environment Secret de uma configuração anterior, ele deixa de ser usado e pode ser removido.

**Transporte do link:** o Service atual do EKS expõe HTTP na porta 80. Isso permite testar o fluxo pelo LoadBalancer atual, mas o token do link funciona como uma credencial e fica visível no tráfego HTTP. Para enviar links a clientes reais, publique a API atrás de HTTPS e use essa origem em `PUBLIC_BASE_URL`.

## 5. Variables opcionais

Todas têm padrão no workflow:

| Variable | Padrão |
| --- | --- |
| `AWS_REGION` | `us-east-1` |
| `EKS_CLUSTER_NAME` | `oficina` |
| `ECR_REPOSITORY` | `oficina-api` |
| `APP_SECRET_ID` | `oficina/api` |
| `TF_STATE_BUCKET` | `oficina-tfstate-<conta>` |
| `SMTP_SECRET_ID` | `oficina/smtp` |
| `SMTP_SECRET_VERSION` | `1` |
| `SMTP_HOST` | `smtp.gmail.com` |
| `SMTP_PORT` | `587` |
| `SMTP_USERNAME` | `soatoficinapos@gmail.com` |
| `SMTP_FROM_ADDRESS` | `soatoficinapos@gmail.com` |

Ao usar outra conta Gmail, altere `SMTP_USERNAME` e `SMTP_FROM_ADDRESS` junto com o `SMTP_PASSWORD`.

## 6. Primeiro deploy

```bash
git switch main && git pull
git tag 1.0.0
git push origin 1.0.0
```

Acompanhe em **Actions → CD (Continuous Deployment)**. O primeiro run cria o bucket do state e todo o ambiente, e leva por volta de 20 minutos (cluster EKS e RDS).

## 7. Conta nova do lab

Se o lab for reiniciado (**Reset**) ou trocado, a conta volta vazia. Atualize os três secrets AWS (passo 2) e crie uma tag: o CD recria o bucket do state, os secrets `oficina/api` e `oficina/smtp` e o ambiente do zero. `SMTP_PASSWORD` fica no GitHub e não precisa ser cadastrado de novo.

## Fora do AWS Academy

Numa conta AWS sem as restrições do lab, a aplicação volta ao Amazon SES sem mudança de código: `EMAIL_PROVIDER=ses` e um `SES_FROM_ADDRESS` verificado, com o pod autenticando pela role dos nós, sem access key. O pipeline, porém, foi ajustado ao lab e sempre configura SMTP. Para usar o SES nele, torne `SMTP_PASSWORD` opcional no Terraform e no CD, grave `SES_FROM_ADDRESS` no secret `oficina/api` e dê `ses:SendEmail` à role dos nós ([ADR 0010](adr/0010-smtp-no-lugar-do-ses-no-learner-lab.md)).
