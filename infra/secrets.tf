resource "random_password" "db" {
  length  = 24
  special = false
}

resource "random_password" "jwt" {
  length  = 64
  special = false
}

resource "aws_secretsmanager_secret" "app" {
  name                    = var.app_secret_name
  recovery_window_in_days = 0

  tags = local.tags
}

resource "aws_secretsmanager_secret_version" "app" {
  secret_id = aws_secretsmanager_secret.app.id
  secret_string = jsonencode({
    DB_HOST          = aws_db_instance.this.address
    DB_USER          = aws_db_instance.this.username
    DB_PASSWORD      = random_password.db.result
    JWT_SECRET       = random_password.jwt.result
  })
}

resource "aws_secretsmanager_secret" "smtp" {
  name                    = var.smtp_secret_name
  recovery_window_in_days = 0

  tags = local.tags
}

resource "aws_secretsmanager_secret_version" "smtp" {
  secret_id = aws_secretsmanager_secret.smtp.id

  secret_string_wo = jsonencode({
    SMTP_HOST        = var.smtp_host
    SMTP_PORT        = tostring(var.smtp_port)
    SMTP_USERNAME    = var.smtp_username
    SMTP_PASSWORD    = var.smtp_password
    SMTP_FROM_ADDRESS = var.smtp_from_address
  })

  secret_string_wo_version = var.smtp_secret_version
}
