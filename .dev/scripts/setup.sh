#!/usr/bin/env bash
#
# Prepara o ambiente do projeto: builda e sobe a stack via
# docker compose, roda migrations e testes dentro de containers (não exige
# Go instalado localmente), popula o banco com os usuários de login e com
# dados fictícios, e ao final deixa a API rodando.
#
# Uso: .dev/scripts/setup.sh

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

COMPOSE_TOOLS="docker compose -f compose.yml -f compose.tools.yml --profile tools"

log() {
    printf '\n==> %s\n' "$1"
}

log "Limpando containers e volumes de execuções anteriores"
docker compose down -v

log "Buildando as imagens (app e toolbox)"
docker compose build app
$COMPOSE_TOOLS build toolbox

log "Subindo o MySQL"
docker compose up -d --wait mysql

log "Rodando as migrations"
$COMPOSE_TOOLS run --rm toolbox go run ./migrations up

log "Rodando os testes unitários"
$COMPOSE_TOOLS run --rm toolbox go test ./... -cover

log "Rodando os testes de integração"
$COMPOSE_TOOLS run --rm toolbox go test -tags integration ./test/integration/...

log "Populando o banco com usuários de login e dados fictícios"
CREDENCIAIS="$($COMPOSE_TOOLS run --rm toolbox go run ./cmd/seed)"

log "Subindo a aplicação"
docker compose up -d --build app

log "Ambiente pronto!"
echo "$CREDENCIAIS"
echo "Use as credenciais acima para testar a aplicação."
