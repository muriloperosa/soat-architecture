# Cluster EKS do lab. Não cria VPC, NAT Gateway nem role IAM.
# A Academy não permite iam:CreateRole e o NAT cobra mesmo com o cluster ocioso.
#
# Ao fim da sessão do Learner Lab, rode `terraform destroy` neste diretório
# antes de encerrar o lab. O control plane continua cobrando por hora depois
# que a sessão expira se o cluster ficar de pé.

data "aws_iam_role" "lab" {
  name = "LabRole"
}

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
}

resource "aws_eks_cluster" "this" {
  name     = var.cluster_name
  role_arn = data.aws_iam_role.lab.arn
  version  = var.kubernetes_version

  vpc_config {
    subnet_ids              = local.subnet_ids
    endpoint_public_access  = true
    endpoint_private_access = false
  }

  tags = {
    Project = "soat-oficina"
  }
}

resource "aws_eks_node_group" "this" {
  cluster_name    = aws_eks_cluster.this.name
  node_group_name = "${var.cluster_name}-nodes"
  node_role_arn   = data.aws_iam_role.lab.arn
  subnet_ids      = local.subnet_ids

  instance_types = [var.node_instance_type]
  ami_type       = "AL2023_x86_64_STANDARD"
  capacity_type  = "ON_DEMAND"

  scaling_config {
    desired_size = var.node_desired_size
    min_size     = var.node_min_size
    max_size     = var.node_max_size
  }

  tags = {
    Project = "soat-oficina"
  }
}
