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

ask() {
    local prompt="$1" default="$2" answer
    read -r -p "$prompt [$default]: " answer
    echo "${answer:-$default}"
}

if [ ! -f .env ]; then
    log "Criando .env a partir do .env.example"
    cp .env.example .env

    default_app_port="$(grep -m1 '^APP_HOST_PORT=' .env.example | cut -d= -f2)"
    default_db_port="$(grep -m1 '^DB_HOST_PORT=' .env.example | cut -d= -f2)"

    app_port="$(ask 'Porta da API no host' "$default_app_port")"
    db_port="$(ask 'Porta do MySQL no host' "$default_db_port")"

    sed -i.bak "s/^APP_HOST_PORT=.*/APP_HOST_PORT=${app_port}/" .env
    sed -i.bak "s/^DB_HOST_PORT=.*/DB_HOST_PORT=${db_port}/" .env
    rm -f .env.bak
fi

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

app_port="$(grep -m1 '^APP_HOST_PORT=' .env | cut -d= -f2)"

log "Ambiente pronto!"
echo "$CREDENCIAIS"
echo "Use as credenciais acima para testar a aplicação."
echo "Health check: http://localhost:${app_port}/v1/health"
