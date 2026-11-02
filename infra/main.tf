# Conta do AWS Academy Learner Lab. O lab não deixa criar role IAM nem
# provedor OIDC: cluster e nós usam a LabRole, e o GitHub Actions usa as
# credenciais temporárias da sessão do lab. Não cria VPC nem NAT Gateway.
#
# O control plane cobra cerca de US$ 0,10/h mesmo com a sessão do lab parada.
# Antes do terraform destroy, apague o namespace oficina. O Service
# LoadBalancer cria um balanceador que o Terraform não gerencia; se ele sobrar,
# continua cobrando e trava a remoção do security group.

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

# us-east-1e existe na VPC default, mas o EKS recusa o control plane nessa zona.
data "aws_subnets" "eks" {
  count = var.subnet_ids == null ? 1 : 0

  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }

  filter {
    name   = "map-public-ip-on-launch"
    values = ["true"]
  }

  filter {
    name   = "availability-zone"
    values = var.eks_availability_zones
  }
}

locals {
  subnet_ids = var.subnet_ids != null ? var.subnet_ids : data.aws_subnets.public[0].ids

  tags = {
    Project = "soat-oficina"
  }
}

# Já existe em toda conta do Learner Lab. Confia em eks.amazonaws.com e
# ec2.amazonaws.com, então serve de role do cluster e dos nós.
data "aws_iam_role" "lab" {
  name = var.lab_role_name
}

data "aws_rds_engine_version" "mysql" {
  engine  = "mysql"
  version = "8.0"
  latest  = true
}
