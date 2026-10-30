output "cluster_name" {
  description = "Nome do cluster EKS."
  value       = aws_eks_cluster.this.name
}

output "cluster_endpoint" {
  description = "Endpoint público da API do cluster."
  value       = aws_eks_cluster.this.endpoint
}

output "vpc_id" {
  description = "VPC existente usada pelo cluster. Nenhum VPC novo é criado."
  value       = local.vpc_id
}

output "subnet_ids" {
  description = "Subnets públicas existentes usadas pelo control plane e pelo node group."
  value       = local.subnet_ids
}

output "update_kubeconfig" {
  description = "Comando para apontar o kubectl para este cluster."
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${aws_eks_cluster.this.name}"
}
