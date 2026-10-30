variable "aws_region" {
  description = "Região do lab. O Learner Lab da Academy usa us-east-1."
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "Nome do cluster EKS."
  type        = string
  default     = "oficina"
}

variable "kubernetes_version" {
  description = "Versão do Kubernetes aceita pela conta do lab. Troque se o apply recusar esta versão."
  type        = string
  default     = "1.33"
}

variable "vpc_id" {
  description = "VPC pública já existente do lab. Vazio usa a VPC default da conta."
  type        = string
  default     = null
}

variable "subnet_ids" {
  description = "Subnets públicas já existentes, em pelo menos duas zonas. Vazio descobre as subnets públicas da VPC."
  type        = list(string)
  default     = null
}

variable "node_instance_type" {
  description = "Tipo da instância do node group. t3.small cabe no crédito do lab."
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
  description = "Máximo de nós do node group."
  type        = number
  default     = 2
}
