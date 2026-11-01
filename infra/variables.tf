variable "aws_region" {
  description = "Região da conta. O CD usa us-east-1 quando a variable AWS_REGION não existe."
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "Nome do cluster EKS. O CD usa oficina quando EKS_CLUSTER_NAME não existe."
  type        = string
  default     = "oficina"
}

variable "kubernetes_version" {
  description = "Versão do Kubernetes em suporte padrão. Troque se o apply recusar esta versão."
  type        = string
  default     = "1.34"
}

variable "vpc_id" {
  description = "VPC já existente. Vazio usa a VPC default da conta. Nenhuma VPC nova é criada."
  type        = string
  default     = null
}

variable "subnet_ids" {
  description = "Subnets públicas em pelo menos duas zonas. Vazio descobre as subnets públicas da VPC."
  type        = list(string)
  default     = null
}

variable "eks_availability_zones" {
  description = "Zonas aceitas pelo control plane do EKS. us-east-1e fica de fora."
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1f"]
}

variable "node_instance_type" {
  description = "Tipo da instância do node group."
  type        = string
  default     = "t3.small"
}

variable "node_desired_size" {
  description = "Quantidade desejada de nós."
  type        = number
  default     = 1
}

variable "node_min_size" {
  description = "Mínimo de nós do node group."
  type        = number
  default     = 1
}

variable "node_max_size" {
  description = "Máximo de nós. 2 cabe no HPA da API (até 6 réplicas) junto com os DaemonSets."
  type        = number
  default     = 2
}

variable "github_deploy_role_name" {
  description = "Role criada à mão no AWS_SETUP.md. O CD assume essa role para o kubectl."
  type        = string
  default     = "oficina-github-deploy"
}

variable "app_secret_name" {
  description = "Secret do Secrets Manager lido pelo CD. O padrão do workflow é oficina/api."
  type        = string
  default     = "oficina/api"
}

variable "db_name" {
  description = "Database criado no RDS. É o DB_NAME do ConfigMap base."
  type        = string
  default     = "mecanica"
}

variable "db_username" {
  description = "Usuário master do RDS. Só letras, números e _. O CD rejeita outro formato."
  type        = string
  default     = "oficina"
}

variable "db_instance_class" {
  description = "Classe da instância RDS."
  type        = string
  default     = "db.t3.micro"
}

variable "ses_from_address" {
  description = "Remetente verificado no SES. Vazio faz o CD subir a API com EMAIL_PROVIDER=log."
  type        = string
  default     = ""
}

variable "db_allocated_storage" {
  description = "Armazenamento inicial do RDS, em GiB."
  type        = number
  default     = 20
}
