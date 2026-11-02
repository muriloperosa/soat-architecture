output "cluster_name" {
  description = "Nome do cluster EKS."
  value       = aws_eks_cluster.this.name
}

output "cluster_endpoint" {
  description = "Endpoint público da API do cluster."
  value       = aws_eks_cluster.this.endpoint
}

output "update_kubeconfig" {
  description = "Comando para apontar o kubectl para este cluster."
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${aws_eks_cluster.this.name}"
}

output "rds_address" {
  description = "Hostname do RDS, sem porta. É o DB_HOST do secret."
  value       = aws_db_instance.this.address
}

output "rds_username" {
  description = "Usuário master do RDS."
  value       = aws_db_instance.this.username
}

output "app_secret_name" {
  description = "Secret lido pelo CD."
  value       = aws_secretsmanager_secret.app.name
}

output "vpc_id" {
  description = "VPC existente usada pelo cluster e pelo RDS."
  value       = local.vpc_id
}

output "subnet_ids" {
  description = "Subnets públicas existentes usadas pelo cluster, pelos nós e pelo RDS."
  value       = local.subnet_ids
}
