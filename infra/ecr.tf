# No Learner Lab o repositório sai junto com o ambiente: force_delete deixa o
# destroy apagar as imagens. Uma conta de lab nova recomeça do zero.
resource "aws_ecr_repository" "api" {
  name                 = var.ecr_repository_name
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.tags
}
