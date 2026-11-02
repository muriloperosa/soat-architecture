resource "aws_ses_email_identity" "sender" {
  email = var.ses_sender_email
}