# Infraestrutura AWS

O Terraform de [`infra/`](../infra/) sobe tudo o que o deploy precisa numa conta do **AWS Academy Learner Lab**, região `us-east-1`. Quem roda o `apply` é o próprio CD: cada tag `MAJOR.MINOR.PATCH` na `main` executa o job `infra` ([`.github/workflows/cd.yml`](../.github/workflows/cd.yml)) antes de publicar a imagem e fazer o rollout. Não é preciso rodar `apply` à mão antes de criar a tag.

Não há NAT Gateway. A VPC e as subnets públicas já existem (a default da conta, salvo `vpc_id` e `subnet_ids`). Os nós recebem IP público por causa de `map-public-ip-on-launch` e alcançam o ECR sem NAT.

## Restrições do Learner Lab

- **Não dá para criar role IAM nem provedor OIDC.** O cluster e os nós usam a `LabRole`, que já existe em toda conta do lab. O GitHub Actions usa as credenciais temporárias da sessão do lab, guardadas como secrets do environment `aws-deploy` ([`AWS_SETUP.md`](AWS_SETUP.md)).
- **As credenciais expiram com a sessão** (cerca de 4 horas). A cada sessão nova, atualize os três secrets no GitHub antes de criar a tag.
- **Quem aplica vira admin do cluster.** Pipeline e quem roda o Terraform localmente usam a mesma role da sessão (`voclabs`), então não existe access entry extra.
- **O control plane continua cobrando com a sessão encerrada.** O lab desliga as instâncias EC2 no fim da sessão, mas o EKS, o RDS e o Load Balancer seguem consumindo o crédito.
- **SES em sandbox.** Remetente e destinatário precisam ser identidades verificadas.

## O que é criado

- Cluster EKS `oficina`, Kubernetes 1.34, endpoint público, autenticação `API_AND_CONFIG_MAP`, com a `LabRole` como role do cluster.
- Node group `oficina-nodes`: `t3.small`, AMI `AL2023_x86_64_STANDARD`, `desired = 1`, `min = 1`, `max = 2`, com a `LabRole` como role dos nós. O hop limit do IMDS fica em 2 para o pod conseguir usar essa role (ECR e SES). Não há access key no cluster.
- Addon `metrics-server`. Sem ele o HPA fica com CPU `<unknown>`.
- Repositório ECR `oficina-api`, com tags imutáveis e scan a cada push. `force_delete` deixa o `destroy` apagar as imagens.
- RDS MySQL 8, instância `db.t3.micro`, 20 GiB, uma zona, sem backup e sem snapshot final. O database é `mecanica` e o usuário é `oficina`. A instância não é pública. O security group só aceita a porta 3306 vinda do security group do cluster.
- Identidade de e-mail do SES para `ses_sender_email` (o e-mail do grupo). O `apply` dispara um e-mail de verificação da AWS para esse endereço; o envio só funciona depois do clique no link.
- Secret `oficina/api` no Secrets Manager, com `DB_HOST` (hostname, sem `:3306`), `DB_USER`, `DB_PASSWORD`, `JWT_SECRET` e `SES_FROM_ADDRESS` (o mesmo e-mail da identidade do SES). Com `SES_FROM_ADDRESS` preenchido, o CD sobe a API com `EMAIL_PROVIDER=ses`. As senhas são só letras e números. `recovery_window_in_days = 0` apaga o secret no destroy; sem isso o nome fica reservado e o próximo apply falha.

## State no S3

O backend é S3 com lock por arquivo (`use_lockfile`, Terraform 1.10 ou mais novo). O bloco `backend "s3" {}` em [`versions.tf`](../infra/versions.tf) é parcial: bucket, key e região vêm do `-backend-config`.

- Bucket: `oficina-tfstate-<conta>`, ou o valor da variable `TF_STATE_BUCKET` do GitHub (e da variável de ambiente de mesmo nome no `make`).
- Key: `oficina/terraform.tfstate`.

O bucket fica fora do Terraform, porque o backend precisa dele antes do `init`. O script [`infra/bootstrap-state.sh`](../infra/bootstrap-state.sh) cria o bucket se ele não existir, com versionamento e bloqueio de acesso público. O CD e o `make tf-init` chamam esse script, então o primeiro run numa conta nova cria o bucket sozinho.

O state contém a senha do banco e o `JWT_SECRET`. Ele nunca vai para o Git (`*.tfstate` está no `.gitignore`).

## Rodar localmente

Normalmente não é preciso: o CD aplica a cada tag. Rode à mão para inspecionar um `plan`, para corrigir algo fora do pipeline ou para o `destroy`.

1. Copie as credenciais em **AWS Details → AWS CLI** do Learner Lab para o profile `oficina-lab` em `~/.aws/credentials` (as três linhas, incluindo `aws_session_token`). Não grave a access key no repositório.
2. Confira a conta:

   ```bash
   export AWS_PROFILE=oficina-lab
   export AWS_REGION=us-east-1
   aws sts get-caller-identity
   ```

3. Na raiz do repositório:

   ```bash
   make tf-plan
   make tf-apply
   ```

O `tf-init` cria o bucket se for preciso e aponta para o mesmo state do CD. O lock no S3 impede um `apply` local e um do pipeline ao mesmo tempo.

Depois do apply:

```bash
aws eks update-kubeconfig --region us-east-1 --name oficina --profile oficina-lab
kubectl get nodes
aws secretsmanager get-secret-value --secret-id oficina/api --query SecretString --output text --profile oficina-lab
aws sesv2 get-email-identity --email-identity soatoficinapos@gmail.com --query VerifiedForSendingStatus --profile oficina-lab
```

## Derrubar

O ambiente ligado o tempo todo consome o crédito do lab em poucos dias. Suba perto da demonstração e derrube depois.

Antes do destroy, ainda com `kubectl` apontando para o cluster, apague o que o Service criou fora do Terraform:

```bash
kubectl delete namespace oficina
```

Espere o Load Balancer sumir e só então, na raiz do repositório:

```bash
make tf-destroy CONFIRM=yes
```

O destroy apaga também o ECR e a identidade do SES. O bucket do state fica: a próxima tag recria o ambiente a partir dele.
