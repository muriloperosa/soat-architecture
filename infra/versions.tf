terraform {
  # use_lockfile do backend S3 exige 1.10 ou mais novo.
  required_version = ">= 1.11.0"

  # Backend parcial: bucket, key e region vêm do -backend-config do
  # make tf-init e do CD. O bucket é criado por bootstrap-state.sh.
  backend "s3" {}

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
