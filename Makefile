.PHONY: help run build test coverage lint up down db-up db-down tidy vendor setup dev debug

GREEN := $(shell tput setaf 2)
RESET := $(shell tput sgr0)

help:
	@echo "Available commands:"
	@echo "  $(GREEN)help$(RESET):     See this help"
	@echo "  $(GREEN)run$(RESET):      Run the API locally (go run)"
	@echo "  $(GREEN)build$(RESET):    Build the API binary into bin/api"
	@echo "  $(GREEN)test$(RESET):     Run all tests with coverage"
	@echo "  $(GREEN)coverage$(RESET): Run tests and print total coverage + open HTML report"
	@echo "  $(GREEN)lint$(RESET):     Run go vet"
	@echo "  $(GREEN)up$(RESET):       Start API + MySQL via docker compose"
	@echo "  $(GREEN)down$(RESET):     Stop the docker compose stack"
	@echo "  $(GREEN)db-up$(RESET):    Start only the MySQL container"
	@echo "  $(GREEN)db-down$(RESET):  Stop only the MySQL container"
	@echo "  $(GREEN)tidy$(RESET):     Run go mod tidy"
	@echo "  $(GREEN)vendor$(RESET):   Run go mod vendor (after tidy)"
	@echo "  $(GREEN)setup$(RESET):    Download deps and install dev tools (swag, migrate, air, delve)"
	@echo "  $(GREEN)dev$(RESET):      Run the API with hot reload (air)"
	@echo "  $(GREEN)debug$(RESET):    Run the API under delve, headless, listening on :2345"

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v -cover

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html

lint:
	go vet ./...

up:
	docker compose up --build -d

down:
	docker compose down

db-up:
	docker compose up -d mysql

db-down:
	docker compose stop mysql

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

dev:
	air -c .air.toml

debug:
	dlv debug ./cmd/api --headless --listen=:2345 --api-version=2 --accept-multiclient
