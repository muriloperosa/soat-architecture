resource "aws_db_subnet_group" "this" {
  name       = "${var.cluster_name}-db"
  subnet_ids = local.subnet_ids

  tags = local.tags
}

resource "aws_security_group" "rds" {
  name        = "${var.cluster_name}-rds"
  description = "MySQL do RDS aceita conexao so do security group do cluster EKS"
  vpc_id      = local.vpc_id

  tags = local.tags
}

resource "aws_security_group_rule" "rds_from_eks" {
  type                     = "ingress"
  description              = "MySQL a partir dos nos do EKS"
  from_port                = 3306
  to_port                  = 3306
  protocol                 = "tcp"
  security_group_id        = aws_security_group.rds.id
  source_security_group_id = aws_eks_cluster.this.vpc_config[0].cluster_security_group_id
}

resource "aws_db_instance" "this" {
  identifier     = "${var.cluster_name}-mysql"
  engine         = "mysql"
  engine_version = data.aws_rds_engine_version.mysql.version
  instance_class = var.db_instance_class

  allocated_storage = var.db_allocated_storage
  storage_type      = "gp3"
  storage_encrypted = true

  db_name  = var.db_name
  username = var.db_username
  password = random_password.db.result

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  multi_az               = false

  backup_retention_period = 0
  skip_final_snapshot     = true
  deletion_protection     = false

  tags = local.tags
}
