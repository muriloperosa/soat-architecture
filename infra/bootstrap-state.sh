#!/usr/bin/env bash
# Cria, se ainda não existir, o bucket S3 do state do Terraform e imprime o
# nome dele. Fica fora do Terraform porque o backend precisa do bucket antes
# do init. Usado pelo make tf-init e pelo job infra do CD.
#
# Nome padrão: oficina-tfstate-<conta>, único por conta do Learner Lab.
# TF_STATE_BUCKET sobrescreve.
set -euo pipefail

region="${AWS_REGION:-us-east-1}"
account="$(aws sts get-caller-identity --query Account --output text)"
bucket="${TF_STATE_BUCKET:-oficina-tfstate-${account}}"

if ! aws s3api head-bucket --bucket "$bucket" --region "$region" >/dev/null 2>&1; then
  echo "Criando o bucket de state ${bucket} em ${region}" >&2
  # us-east-1 recusa LocationConstraint explícito.
  if [[ "$region" == "us-east-1" ]]; then
    aws s3api create-bucket --bucket "$bucket" --region "$region" >/dev/null
  else
    aws s3api create-bucket --bucket "$bucket" --region "$region" \
      --create-bucket-configuration "LocationConstraint=${region}" >/dev/null
  fi
  aws s3api wait bucket-exists --bucket "$bucket" --region "$region"
  aws s3api put-bucket-versioning --bucket "$bucket" --region "$region" \
    --versioning-configuration Status=Enabled
  aws s3api put-public-access-block --bucket "$bucket" --region "$region" \
    --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
fi

echo "$bucket"
