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
    SES_FROM_ADDRESS = var.ses_from_address
  })
}
