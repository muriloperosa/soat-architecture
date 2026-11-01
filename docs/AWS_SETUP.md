# Preparação da conta AWS para o CD (via CLI)

Roteiro único para deixar a conta AWS pronta para o workflow [`cd.yml`](../.github/workflows/cd.yml). O contexto de cada recurso está em [`CI_CD.md`](CI_CD.md#parte-2-deploy).

Pré-requisitos:

- AWS CLI v2 autenticada como administrador da conta (`aws configure` ou `aws sso login`)
- GitHub CLI autenticada (`gh auth login`)
- Rodar todos os passos no mesmo terminal, porque as variáveis do passo 0 são reaproveitadas

## 0. Variáveis

```bash
export AWS_REGION=us-east-1
export ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
export REPO=muriloperosa/soat-architecture
export ROLE_NAME=oficina-github-deploy
export ECR_REPO=oficina-api
export CLUSTER=oficina
export APP_SECRET=oficina/api
# Prefixo do claim `sub` do token OIDC. Com "immutable subject" ligado no repo,
# o GitHub inclui os IDs numéricos: repo:<owner>@<id>/<repo>@<id>.
export SUB_PREFIX=$(gh api "repos/${REPO}/actions/oidc/customization/sub" --jq '.sub_claim_prefix // empty')
export SUB_PREFIX=${SUB_PREFIX:-repo:${REPO}}
echo "Conta: $ACCOUNT"
echo "Sub:   ${SUB_PREFIX}:environment:aws-deploy"
```

## 1. Provedor OIDC do GitHub (uma vez por conta)

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com
```

> Se aparecer `EntityAlreadyExists`, o provedor já existe. Siga para o passo 2.

## 2. Role de deploy

### 2.1 Trust policy (quem pode assumir a role)

Só jobs do environment `aws-deploy` deste repositório. O `sub` precisa ser idêntico ao que o GitHub manda no token (a condição é `StringEquals`); por isso vem de `SUB_PREFIX`, e não do nome do repo escrito à mão:

```bash
cat > trust.json <<JSON
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "Federated": "arn:aws:iam::${ACCOUNT}:oidc-provider/token.actions.githubusercontent.com" },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
        "token.actions.githubusercontent.com:sub": "${SUB_PREFIX}:environment:aws-deploy"
      }
    }
  }]
}
JSON

aws iam create-role --role-name "$ROLE_NAME" \
  --assume-role-policy-document file://trust.json \
  --max-session-duration 3600 \
  --query Role.Arn --output text
```

A saída é o ARN da role. Ele vai para o GitHub no passo 6.

### 2.2 Permissões (o que a role pode fazer)

```bash
cat > policy.json <<JSON
{
  "Version": "2012-10-17",
  "Statement": [
    { "Sid": "EcrLogin", "Effect": "Allow", "Action": "ecr:GetAuthorizationToken", "Resource": "*" },
    {
      "Sid": "EcrPushPull",
      "Effect": "Allow",
      "Action": [
        "ecr:BatchCheckLayerAvailability", "ecr:BatchGetImage", "ecr:CompleteLayerUpload",
        "ecr:DescribeImages", "ecr:GetDownloadUrlForLayer", "ecr:InitiateLayerUpload",
        "ecr:PutImage", "ecr:UploadLayerPart"
      ],
      "Resource": "arn:aws:ecr:${AWS_REGION}:${ACCOUNT}:repository/${ECR_REPO}"
    },
    {
      "Sid": "EksKubeconfig",
      "Effect": "Allow",
      "Action": "eks:DescribeCluster",
      "Resource": "arn:aws:eks:${AWS_REGION}:${ACCOUNT}:cluster/${CLUSTER}"
    },
    {
      "Sid": "AppSecret",
      "Effect": "Allow",
      "Action": "secretsmanager:GetSecretValue",
      "Resource": "arn:aws:secretsmanager:${AWS_REGION}:${ACCOUNT}:secret:${APP_SECRET}-*"
    }
  ]
}
JSON

aws iam put-role-policy --role-name "$ROLE_NAME" \
  --policy-name oficina-deploy \
  --policy-document file://policy.json
```

## 3. Repositório ECR

Tags imutáveis: uma versão publicada (ex.: `1.0.1`) nunca é sobrescrita.

```bash
aws ecr create-repository --repository-name "$ECR_REPO" \
  --region "$AWS_REGION" \
  --image-tag-mutability IMMUTABLE \
  --image-scanning-configuration scanOnPush=true
```

## 4. Secret da aplicação

Os valores do banco vêm do RDS. `DB_PASSWORD` só pode ter letras e números, porque entra na URL do Job de migration ([ADR 0006](adr/0006-migrations-em-job-kubernetes.md)).

```bash
JWT=$(openssl rand -hex 32)

aws secretsmanager create-secret --name "$APP_SECRET" --region "$AWS_REGION" \
  --secret-string "{
    \"DB_HOST\": \"<endpoint do RDS>\",
    \"DB_USER\": \"<usuário do RDS>\",
    \"DB_PASSWORD\": \"<senha só com letras e números>\",
    \"JWT_SECRET\": \"${JWT}\"
  }"
```

Para trocar valores depois:

```bash
aws secretsmanager put-secret-value --secret-id "$APP_SECRET" --secret-string '{...}'
```

## 5. Acesso da role ao cluster EKS (quando o cluster existir)

O cluster precisa estar com autenticação `API` ou `API_AND_CONFIG_MAP`:

```bash
aws eks describe-cluster --name "$CLUSTER" \
  --query cluster.accessConfig.authenticationMode --output text

# Se a saída for CONFIG_MAP:
aws eks update-cluster-config --name "$CLUSTER" \
  --access-config authenticationMode=API_AND_CONFIG_MAP
```

Crie o access entry. A role recebe admin do cluster porque o primeiro deploy cria o namespace `oficina`:

```bash
aws eks create-access-entry --cluster-name "$CLUSTER" \
  --principal-arn "arn:aws:iam::${ACCOUNT}:role/${ROLE_NAME}"

aws eks associate-access-policy --cluster-name "$CLUSTER" \
  --principal-arn "arn:aws:iam::${ACCOUNT}:role/${ROLE_NAME}" \
  --policy-arn arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy \
  --access-scope type=cluster
```

O security group do RDS também precisa aceitar conexões na porta 3306 vindas dos nós do EKS, e o banco `mecanica` precisa existir.

## 6. GitHub: variable e environment

```bash
gh variable set AWS_DEPLOY_ROLE_ARN --repo "$REPO" \
  --body "arn:aws:iam::${ACCOUNT}:role/${ROLE_NAME}"

gh api -X PUT "repos/${REPO}/environments/aws-deploy"
```

Recomendado: em **Settings → Environments → aws-deploy → Deployment branches and tags**, escolha **Selected** e adicione a regra de tag `[0-9]*.[0-9]*.[0-9]*`.

Nenhum secret precisa ser criado no GitHub: a autenticação é por OIDC.

## 7. Conferência

```bash
aws iam get-role --role-name "$ROLE_NAME" --query Role.Arn --output text
aws ecr describe-repositories --repository-names "$ECR_REPO" --query 'repositories[0].repositoryUri' --output text
aws secretsmanager describe-secret --secret-id "$APP_SECRET" --query ARN --output text
aws eks list-access-entries --cluster-name "$CLUSTER"
gh variable list --repo "$REPO"
```

## 8. Primeiro deploy

```bash
git switch main && git pull
git tag 1.0.0
git push origin 1.0.0
```

Acompanhe em **Actions → CD (Continuous Deployment)**.

## Limpeza

```bash
rm -f trust.json policy.json
```
