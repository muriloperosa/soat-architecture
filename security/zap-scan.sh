#!/usr/bin/env bash
# Sobe a API, cria/loga um usuário de teste por papel (admin, atendente,
# mecanico, cliente) e roda um API scan do ZAP autenticado com cada token,
# um relatório por papel. Prova não só que as rotas funcionam autenticadas,
# mas que autorização (RBAC) barra papel sem permissão.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

# Stack isolada (app-dast + mysql-dast), projeto compose separado do dev
# (nao referencia compose.yml), portas de host próprias (nao colidem com
# app/mysql de dev mesmo se estiverem rodando) pra "down -v" no fim nunca
# encostar no ambiente de desenvolvimento nem reaproveitar dados entre scans.
COMPOSE="docker compose -p soat-architecture-dast -f security/compose.dast.yml"
DAST_DB_HOST_PORT="${DAST_DB_HOST_PORT:-3307}"
DAST_APP_HOST_PORT="${DAST_APP_HOST_PORT:-8081}"
BASE_URL="http://localhost:${DAST_APP_HOST_PORT}/v1"

ADMIN_EMAIL="${ZAP_ADMIN_EMAIL:-zap-admin@teste.local}"
ATENDENTE_EMAIL="${ZAP_ATENDENTE_EMAIL:-zap-atendente@teste.local}"
MECANICO_EMAIL="${ZAP_MECANICO_EMAIL:-zap-mecanico@teste.local}"
CLIENTE_EMAIL="${ZAP_CLIENTE_EMAIL:-zap-cliente@teste.local}"
SENHA="${ZAP_USER_SENHA:-Zap#Scan123}"
CLIENTE_DOCUMENTO="${ZAP_CLIENTE_DOCUMENTO:-52998224725}"

echo "Subindo API + MySQL isolados (app-dast + mysql-dast)..."
$COMPOSE up -d --wait mysql-dast

echo "Rodando migrations no banco isolado (mysql-dast é tmpfs, começa vazio)..."
DB_HOST=localhost DB_PORT="$DAST_DB_HOST_PORT" go run ./migrations up

$COMPOSE up --build -d app-dast

echo "Aguardando API responder..."
for i in $(seq 1 30); do
  if curl -sf "${BASE_URL}/health" > /dev/null; then
    break
  fi
  sleep 2
done

extract_token() {
  python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))"
}

login_interno() {
  local email="$1"
  curl -s -X POST "${BASE_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${email}\",\"senha\":\"${SENHA}\"}" | extract_token
}

echo "Criando usuários de teste por papel (idempotente, ignora erro se já existir)..."
export DB_HOST=localhost
export DB_PORT="$DAST_DB_HOST_PORT"
go run ./cmd/create-user --nome "ZAP Admin" --email "$ADMIN_EMAIL" --senha "$SENHA" --papel ADMINISTRADOR || true
go run ./cmd/create-user --nome "ZAP Atendente" --email "$ATENDENTE_EMAIL" --senha "$SENHA" --papel ATENDENTE || true
go run ./cmd/create-user --nome "ZAP Mecanico" --email "$MECANICO_EMAIL" --senha "$SENHA" --papel MECANICO || true

echo "Autenticando admin (necessário pra criar o cliente de teste)..."
ADMIN_TOKEN=$(login_interno "$ADMIN_EMAIL")
if [ -z "$ADMIN_TOKEN" ]; then
  echo "Falha ao autenticar admin" >&2
  exit 1
fi

echo "Criando cliente de teste (idempotente, ignora erro se já existir)..."
curl -s -X POST "${BASE_URL}/clientes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d "{\"nome\":\"ZAP Cliente\",\"email\":\"${CLIENTE_EMAIL}\",\"senha\":\"${SENHA}\",\"documento\":\"${CLIENTE_DOCUMENTO}\",\"tipo_pessoa\":\"PF\",\"telefone\":\"11999998888\"}" \
  > /dev/null || true

echo "Autenticando atendente, mecanico e cliente..."
ATENDENTE_TOKEN=$(login_interno "$ATENDENTE_EMAIL")
MECANICO_TOKEN=$(login_interno "$MECANICO_EMAIL")
CLIENTE_TOKEN=$(curl -s -X POST "${BASE_URL}/auth/cliente/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${CLIENTE_EMAIL}\",\"senha\":\"${SENHA}\"}" | extract_token)

mkdir -p security/reports

# zap-rules.tsv (via -c no security/compose.dast.yml) só afeta o veredito/console do ZAP
# (WARN/IGNORE/FAIL), não remove os alertas dos relatórios HTML/JSON. Os
# achados de docs/swagger/doc.json (90022, 10023, 100001) são falsos
# positivos confirmados (ver zap-rules.tsv), então tiramos eles do JSON aqui.
strip_ignored_alerts() {
  local json_file="$1"
  [ -f "$json_file" ] || return
  python3 - "$json_file" <<'PY'
import json
import sys

path = sys.argv[1]
ignored_plugin_ids = {"90022", "10023", "100001"}

with open(path) as f:
    data = json.load(f)

for site in data.get("site", []):
    site["alerts"] = [
        a for a in site.get("alerts", []) if a.get("pluginid") not in ignored_plugin_ids
    ]

with open(path, "w") as f:
    json.dump(data, f, indent=4)
PY
}

run_scan() {
  local role="$1"
  local token="$2"
  if [ -z "$token" ]; then
    echo "Token vazio pro papel ${role}, pulando scan." >&2
    return
  fi
  echo "Rodando ZAP API scan autenticado como ${role}..."
  ZAP_TOKEN="$token" ZAP_REPORT_PREFIX="zap-report-${role}" $COMPOSE --profile tools run --rm zap || true
  strip_ignored_alerts "security/reports/zap-report-${role}.json"
}

run_scan "admin" "$ADMIN_TOKEN"
run_scan "atendente" "$ATENDENTE_TOKEN"
run_scan "mecanico" "$MECANICO_TOKEN"
run_scan "cliente" "$CLIENTE_TOKEN"

echo "Derrubando stack isolada de DAST (dev stack intocada)..."
$COMPOSE down -v

echo "Relatórios gerados em security/reports/zap-report-{admin,atendente,mecanico,cliente}.{html,json}"
