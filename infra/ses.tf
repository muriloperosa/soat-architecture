# O SES do Learner Lab fica em sandbox. O apply manda um e-mail de verificação
# para ses_sender_email, e o envio só funciona depois do clique no link. Em
# sandbox o destinatário também precisa ser uma identidade verificada.
resource "aws_ses_email_identity" "sender" {
  email = var.ses_sender_email
}
