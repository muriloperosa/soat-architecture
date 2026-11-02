# Infraestrutura AWS

Terraform deste diretório sobe o que o deploy por tag espera encontrar na conta `081473213177`, região `us-east-1`. O workflow [`.github/workflows/cd.yml`](../.github/workflows/cd.yml) não cria esses recursos: o job `infra` continua desligado até existir um backend S3. Quem sobe e derruba o ambiente é este diretório, fora do pipeline.

Não há NAT Gateway. A VPC e as subnets públicas já existem (a default da conta, salvo `vpc_id` e `subnet_ids`). Os nós recebem IP público por causa de `map-public-ip-on-launch` e alcançam o ECR sem NAT.

## O que é criado

- Role `oficina-eks-cluster` (`AmazonEKSClusterPolicy`), assumida por `eks.amazonaws.com`.
- Role `oficina-eks-node` (`AmazonEKSWorkerNodePolicy`, `AmazonEKS_CNI_Policy`, `AmazonEC2ContainerRegistryReadOnly` e `ses:SendEmail`), assumida por `ec2.amazonaws.com`. Sem a policy de leitura do ECR os pods ficam em `ImagePullBackOff`. O hop limit do IMDS fica em 2 para o pod conseguir usar essa role. Não há access key no cluster.
- Cluster EKS `oficina`, Kubernetes 1.34, endpoint público, autenticação `API_AND_CONFIG_MAP`. Quem aplica o Terraform vira admin do cluster.
- Access entry da role `oficina-github-deploy` com `AmazonEKSClusterAdminPolicy`. O primeiro deploy cria o namespace `oficina`, por isso a policy é de admin do cluster.
- Node group `oficina-nodes`: `t3.small`, AMI `AL2023_x86_64_STANDARD`, `desired = 1`, `min = 1`, `max = 2`.
- Addon `metrics-server`. Sem ele o HPA fica com CPU `<unknown>`.
- RDS MySQL 8, instância `db.t3.micro`, 20 GiB, uma zona, sem backup e sem snapshot final. O database é `mecanica` e o usuário é `oficina`. A instância não é pública. O security group só aceita a porta 3306 vinda do security group do cluster.
- Secret `oficina/api` no Secrets Manager, com `DB_HOST` (hostname, sem `:3306`), `DB_USER`, `DB_PASSWORD`, `JWT_SECRET` e `SES_FROM_ADDRESS`. As senhas são só letras e números. `SES_FROM_ADDRESS` vazio deixa o CD em `EMAIL_PROVIDER=log`. `recovery_window_in_days = 0` apaga o secret no destroy; sem isso o nome fica reservado e o próximo apply falha.

## O que não é criado

- VPC, subnet e NAT Gateway.
- Repositório ECR `oficina-api`.
- Provedor OIDC do GitHub e a role `oficina-github-deploy`.

Esses três ficam de fora para sobreviver ao `terraform destroy`. O ECR guarda as imagens. A role e o OIDC são o acesso do GitHub Actions. O roteiro deles está em [`docs/AWS_SETUP.md`](../docs/AWS_SETUP.md), passos 1 a 3. Não crie o secret nem o access entry por aquele roteiro: o Terraform faz os dois. Se o secret `oficina/api` já existir, importe antes do apply:

```bash
terraform import aws_secretsmanager_secret.app oficina/api
```

Os placeholders `REPLACE_WITH_RDS_ENDPOINT` e `REPLACE_WITH_RDS_USERNAME` do overlay `k8s/overlays/eks` continuam no Git. O CD troca os dois só dentro do runner. Não os edite.

## Antes do apply

1. A role `oficina-github-deploy` já existe (`docs/AWS_SETUP.md`, passos 1 e 2). O `plan` lê essa role e falha se ela não estiver lá.
2. A conta tem VPC default com subnets públicas em pelo menos duas zonas. Se não tiver, passe `vpc_id` e `subnet_ids`.
3. Credencial da conta no profile `oficina-temp`, região `us-east-1`. Não grave a access key no repositório.

```bash
export AWS_PROFILE=oficina-temp
export AWS_REGION=us-east-1
aws sts get-caller-identity
```

A conta esperada é `081473213177`.

## Subir

Na raiz do repositório, com o profile `oficina-temp`:

```bash
make tf-plan
make tf-apply
```

O `tf-apply` pede confirmação. O control plane passa a cobrar por hora assim que o cluster existe. Estes alvos não rodam no push da tag: o job `infra` do CD continua desligado. Suba o ambiente com `make tf-apply`, espere o nó ficar `Ready` e só então envie a tag.

Todas as variables têm padrão, então o `plan` do pipeline (quando for ligado) não precisa de `-var`.

Depois do apply:

```bash
aws eks update-kubeconfig --region us-east-1 --name oficina --profile oficina-temp
kubectl get nodes
```

Os nós devem ficar `Ready`. O secret conferido sem imprimir a senha no histórico:

```bash
aws secretsmanager get-secret-value --secret-id oficina/api --query SecretString --output text --profile oficina-temp
aws eks list-access-entries --cluster-name oficina --region us-east-1
```

O state local (`terraform.tfstate`) contém a senha do banco e o `JWT_SECRET`. Ele está no `.gitignore`. Não o commite.

## Deploy da aplicação

O deploy não sai deste diretório. Com o apply concluído, a CI da `main` verde e a variable `AWS_DEPLOY_ROLE_ARN` configurada, uma tag `MAJOR.MINOR.PATCH` no commit da `main` publica a imagem e faz o rollout. Para um teste, `0.0.1` serve:

```bash
git switch main && git pull
git tag 0.0.1
git push origin 0.0.1
```

A tag precisa estar num commit que já está na `main`. O job `infra` do workflow aparece como skipped. O job `deploy` lê o secret, roda as migrations e sobe a API. Acompanhe em Actions → CD.

## Derrubar

O ambiente completo ligado o mês inteiro estoura o crédito de US$ 110. Suba perto da demonstração.

Antes do destroy, ainda com `kubectl` apontando para o cluster, apague o que o Service criou fora do Terraform:

```bash
kubectl delete namespace oficina
```

Espere o Load Balancer sumir e só então, na raiz do repositório:

```bash
make tf-destroy CONFIRM=yes
```

ECR, provedor OIDC e a role `oficina-github-deploy` permanecem.
