.PHONY: help run build test test-integration coverage lint up down db-up db-down tidy vendor setup dev debug mocks swagger hooks-install hooks-uninstall create-user sonar-up sonar-down sonar-scan sec-sca sec-sast sec-dast

GREEN := $(shell tput setaf 2)
RESET := $(shell tput sgr0)

GOBIN := $(shell go env GOPATH)/bin

ifneq (,$(wildcard ./.env))
include .env
export
endif

help:
	@echo "Available commands:"
	@echo "  $(GREEN)help$(RESET):     See this help"
	@echo "  $(GREEN)run$(RESET):      Run the API locally (go run)"
	@echo "  $(GREEN)build$(RESET):    Build the API binary into bin/api"
	@echo "  $(GREEN)test$(RESET):     Run all tests with coverage"
	@echo "  $(GREEN)test-integration$(RESET): Run HTTP integration tests against a real MySQL (testcontainers, needs Docker)"
	@echo "  $(GREEN)coverage$(RESET): Run tests and print total coverage + open HTML report"
	@echo "  $(GREEN)lint$(RESET):     Run go vet"
	@echo "  $(GREEN)up$(RESET):       Start API + MySQL via docker compose"
	@echo "  $(GREEN)down$(RESET):     Stop the docker compose stack"
	@echo "  $(GREEN)db-up$(RESET):    Start only the MySQL container"
	@echo "  $(GREEN)db-down$(RESET):  Stop only the MySQL container"
	@echo "  $(GREEN)tidy$(RESET):     Run go mod tidy"
	@echo "  $(GREEN)vendor$(RESET):   Run go mod vendor (after tidy)"
	@echo "  $(GREEN)setup$(RESET):    Download deps and install dev tools (swag, migrate, air, delve, mockery)"
	@echo "  $(GREEN)dev$(RESET):      Run the API with hot reload (air)"
	@echo "  $(GREEN)debug$(RESET):    Run the API under delve, headless, listening on :2345"
	@echo "  $(GREEN)mocks$(RESET):    Generate mocks for domain/application interfaces (mockery)"
	@echo "  $(GREEN)swagger$(RESET):  Generate Swagger docs into docs/swagger (swag)"
	@echo "  $(GREEN)hooks-install$(RESET):   Install the pre-push Git hook (mocks, lint, test, swagger)"
	@echo "  $(GREEN)hooks-uninstall$(RESET): Uninstall the pre-push Git hook"
	@echo "  $(GREEN)migrate-up$(RESET):      Run database migrations up (migrate)"
	@echo "  $(GREEN)migrate-down$(RESET):    Run database migrations down (migrate)"
	@echo "  $(GREEN)migrate-version$(RESET): Run database migrations version (migrate)"
	@echo "  $(GREEN)db-setup$(RESET):         Start MySQL and run migrations up"
	@echo "  $(GREEN)create-user$(RESET):   Create an internal user (NOME, EMAIL, SENHA, PAPEL optional=admin)"
	@echo "  $(GREEN)sonar-up$(RESET):      Start SonarQube server via docker compose"
	@echo "  $(GREEN)sonar-down$(RESET):    Stop the SonarQube container"
	@echo "  $(GREEN)sonar-scan$(RESET):    Run coverage + sonar-scanner (needs SONAR_TOKEN)"
	@echo "  $(GREEN)sec-sca$(RESET):       Run govulncheck (SCA) into security/reports/govulncheck.json"
	@echo "  $(GREEN)sec-sast$(RESET):      Run gosec (SAST) into security/reports/gosec.json"
	@echo "  $(GREEN)sec-dast$(RESET):      Run authenticated OWASP ZAP API scan (isolated app-dast+mysql-dast stack) into security/reports/zap-report-{admin,atendente,mecanico,cliente}.*"

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v -cover

test-integration:
	go test -tags integration ./test/integration/... -v

coverage:
	@mkdir -p test/reports
	go test ./... -coverprofile=test/reports/coverage.out
	go tool cover -func=test/reports/coverage.out | tail -1
	go tool cover -html=test/reports/coverage.out -o test/reports/coverage.html

lint:
	go vet ./...

up:
	docker compose up --build -d

down:
	docker compose down

db-up:
	docker compose up -d --wait mysql

db-down:
	docker compose stop mysql
	
migrate-up:
	go run ./migrations up

migrate-down:
	go run ./migrations down

migrate-version:
	go run ./migrations version

migrate-force:
	go run ./migrations force $(VERSION)

db-setup: db-up migrate-up

create-user:
	go run ./cmd/create-user --nome "$(NOME)" --email "$(EMAIL)" --senha "$(SENHA)" --papel "$(if $(PAPEL),$(PAPEL),ADMINISTRADOR)"

db-reset:
	docker compose down -v
	docker compose up -d --wait mysql
	$(MAKE) migrate-up

tidy:
	go mod tidy

vendor: tidy
	go mod vendor

setup:
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest
	go install github.com/go-delve/delve/cmd/dlv@latest
	go install github.com/vektra/mockery/v2@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

mocks:
	$(GOBIN)/mockery

swagger:
	$(GOBIN)/swag init -g cmd/api/main.go -o docs/swagger

dev:
	$(GOBIN)/air -c .air.toml

debug:
	$(GOBIN)/dlv debug ./cmd/api --headless --listen=:2345 --api-version=2 --accept-multiclient

hooks-install:
	@echo "Installing Git hooks..."
	@cp .dev/hooks/stubs/pre-push.stub .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "$(GREEN)Git hooks installed successfully!$(RESET)"

hooks-uninstall:
	@echo "Uninstalling Git hooks..."
	@rm -f .git/hooks/pre-push
	@echo "$(GREEN)Git hooks uninstalled successfully!$(RESET)"

sonar-up:
	docker compose -f compose.yml -f compose.tools.yml up -d --wait sonarqube

sonar-down:
	docker compose -f compose.yml -f compose.tools.yml stop sonarqube

sonar-scan: coverage
	@if [ -z "$(SONAR_TOKEN)" ]; then \
		echo "SONAR_TOKEN not set. Add SONAR_TOKEN=<seu_token> ao .env"; \
		exit 1; \
	fi
	docker compose -f compose.yml -f compose.tools.yml --profile tools run --rm sonar-scanner

sec-sca:
	@mkdir -p security/reports
	$(GOBIN)/govulncheck -json ./... > security/reports/govulncheck.json
	@echo "$(GREEN)Relatório gerado em security/reports/govulncheck.json$(RESET)"

sec-sast:
	@mkdir -p security/reports
	$(GOBIN)/gosec -fmt=json -out=security/reports/gosec.json ./... || true
	@echo "$(GREEN)Relatório gerado em security/reports/gosec.json$(RESET)"

sec-dast:
	@bash security/zap-scan.sh
