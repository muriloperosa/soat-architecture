# Conta própria do projeto (não é o Learner Lab). Não cria VPC, NAT Gateway,
# repositório ECR nem a role oficina-github-deploy: esses dois últimos ficam
# de pé no destroy porque a imagem e o OIDC sobrevivem ao cluster.
#
# O control plane cobra cerca de US$ 0,10/h mesmo parado. Antes do
# terraform destroy, apague o namespace oficina. O Service LoadBalancer
# cria um balanceador que o Terraform não gerencia; se ele sobrar, continua
# cobrando e trava a remoção do security group.

data "aws_vpc" "default" {
  count   = var.vpc_id == null ? 1 : 0
  default = true
}

locals {
  vpc_id = var.vpc_id != null ? var.vpc_id : data.aws_vpc.default[0].id
}

data "aws_subnets" "public" {
  count = var.subnet_ids == null ? 1 : 0

  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }

  filter {
    name   = "map-public-ip-on-launch"
    values = ["true"]
  }
}

locals {
  subnet_ids = var.subnet_ids != null ? var.subnet_ids : data.aws_subnets.public[0].ids

  tags = {
    Project = "soat-oficina"
  }
}

data "aws_iam_role" "github_deploy" {
  name = var.github_deploy_role_name
}

data "aws_rds_engine_version" "mysql" {
  engine  = "mysql"
  version = "8.0"
  latest  = true
}
