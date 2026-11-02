# Preparação do AWS Learner Lab para o CD

Roteiro para deixar o GitHub pronto para o workflow [`cd.yml`](../.github/workflows/cd.yml) numa conta do AWS Academy Learner Lab. A infraestrutura (EKS, RDS, ECR, SES e o secret da aplicação) é criada pelo próprio CD, pelo Terraform de [`infra/`](../infra/) ([`TERRAFORM.md`](TERRAFORM.md)). Aqui só entram as credenciais.

O Learner Lab não deixa criar provedor OIDC nem role IAM. Por isso o pipeline não usa OIDC: ele usa as credenciais temporárias da sessão do lab, guardadas como secrets do environment `aws-deploy`.

Cada integrante do grupo tem a própria conta de Learner Lab. Escolham **uma** conta para o ambiente e usem sempre as credenciais dela no GitHub: o state, o cluster e o bucket ficam nessa conta, e credenciais de outra conta não enxergam nada disso.

Pré-requisitos:

- Sessão do Learner Lab iniciada (**Start Lab**, indicador verde)
- GitHub CLI autenticada (`gh auth login`), com permissão de admin no repositório

## 1. Environment `aws-deploy` (uma vez)

```bash
export REPO=muriloperosa/soat-architecture
gh api -X PUT "repos/${REPO}/environments/aws-deploy"
```

Recomendado: em **Settings → Environments → aws-deploy → Deployment branches and tags**, escolha **Selected** e adicione a regra de tag `[0-9]*.[0-9]*.[0-9]*`. Só os runs de tag passam a ler os secrets.

## 2. Credenciais da sessão (a cada sessão do lab)

No Learner Lab, abra **AWS Details → AWS CLI → Show**. Copie os três valores e grave no environment:

```bash
gh secret set AWS_ACCESS_KEY_ID     --repo "$REPO" --env aws-deploy --body '<aws_access_key_id>'
gh secret set AWS_SECRET_ACCESS_KEY --repo "$REPO" --env aws-deploy --body '<aws_secret_access_key>'
gh secret set AWS_SESSION_TOKEN     --repo "$REPO" --env aws-deploy --body '<aws_session_token>'
```

As credenciais expiram quando a sessão acaba (cerca de 4 horas). Um run com credencial vencida falha no primeiro passo da AWS com `ExpiredToken` ou `InvalidClientTokenId`: atualize os três secrets e rode de novo (**Actions → CD → Run workflow**, escolhendo a tag).

## 3. Variables opcionais

Todas têm padrão no workflow:

| Variable | Padrão |
| --- | --- |
| `AWS_REGION` | `us-east-1` |
| `EKS_CLUSTER_NAME` | `oficina` |
| `ECR_REPOSITORY` | `oficina-api` |
| `APP_SECRET_ID` | `oficina/api` |
| `TF_STATE_BUCKET` | `oficina-tfstate-<conta>` |

## 4. Primeiro deploy

```bash
git switch main && git pull
git tag 1.0.0
git push origin 1.0.0
```

Acompanhe em **Actions → CD (Continuous Deployment)**. O primeiro run cria o bucket do state e todo o ambiente, e leva por volta de 20 minutos (cluster EKS e RDS).

## 5. Verificar o e-mail do SES

O primeiro `apply` cria a identidade do SES e a AWS manda um e-mail de verificação para o endereço do grupo (`ses_sender_email` em [`variables.tf`](../infra/variables.tf)). Clique no link. Até lá, a API sobe normalmente, mas o envio das notificações falha e só aparece no log do pod.

Em sandbox, o destinatário também precisa estar verificado. Para a demonstração, use como e-mail do cliente o próprio e-mail do grupo ou verifique o destinatário:

```bash
aws sesv2 create-email-identity --email-identity <destinatario> --region us-east-1
```

## Conta nova do lab

Se o lab for reiniciado (**Reset**) ou trocado, a conta volta vazia. Atualize os secrets (passo 2) e crie uma tag: o CD recria o bucket do state e o ambiente do zero. A verificação do SES precisa ser feita de novo.
