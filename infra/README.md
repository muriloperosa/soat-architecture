# Cluster EKS no AWS Academy

Terraform deste diretório sobe só o cluster EKS do lab e um node group. Não cria VPC, subnet, NAT Gateway, RDS, load balancer nem role IAM.

## O que é criado

- Cluster EKS (`aws_eks_cluster`), com endpoint público.
- Um node group (`aws_eks_node_group`) com `t3.small`, `desired = 1`, `min = 1`, `max = 2`.

O control plane e o node group usam a role `LabRole` que já existe na conta (`data "aws_iam_role"`). A Academy não permite `iam:CreateRole`. A VPC e as subnets também já existem: o Terraform só lê a VPC default e as subnets com `map-public-ip-on-launch`. Nada disso é criado aqui, e por isso não há NAT. O NAT Gateway cobra mesmo com o cluster parado.

O nome do cluster, a versão do Kubernetes, o tipo da instância e a VPC podem ser trocados pelas variáveis em `variables.tf`. O padrão do cluster é `oficina`, na região `us-east-1`. Se a conta do lab não tiver VPC default, passe `vpc_id` e `subnet_ids` de subnets públicas em pelo menos duas zonas.

## Credenciais

A sessão do Learner Lab expira. Antes de `plan` ou `apply`, renove as credenciais de CLI no console do lab (AWS Details) e exporte de novo:

```bash
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_SESSION_TOKEN=...
export AWS_DEFAULT_REGION=us-east-1
```

Não grave esses valores no repositório.

## Subir o cluster

Dentro de `infra/`:

```bash
terraform init
terraform plan
terraform apply
```

`apply` cria o control plane, que cobra por hora enquanto existir. Não deixe o cluster de pé depois da sessão.

Quando o apply terminar, o output `update_kubeconfig` traz o comando para o kubectl. Ele tem a forma:

```bash
aws eks update-kubeconfig --region us-east-1 --name oficina
```

Use o nome e a região que o output mostrar se você tiver mudado as variáveis.

## Encerrar a sessão

Antes de parar o Learner Lab, ainda com as credenciais válidas:

```bash
terraform destroy
```

Rode isso neste diretório. Sem o destroy, o control plane continua cobrando depois que a sessão do lab acaba.
