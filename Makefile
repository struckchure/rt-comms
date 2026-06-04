.PHONY: help run build test db-up db-down db-logs migrate-up migrate-down migrate-version dev

# Load .env when present (PORT, DATABASE_URL, STUN_URL)
ifneq (,$(wildcard .env))
include .env
export
endif

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/postgres

# golang-migrate expects postgresql:// and local dev needs sslmode=disable
MIGRATE_DATABASE_URL := $(subst postgres://,postgresql://,$(DATABASE_URL))
ifeq ($(findstring sslmode=,$(MIGRATE_DATABASE_URL)),)
MIGRATE_DATABASE_URL := $(MIGRATE_DATABASE_URL)?sslmode=disable
endif

MIGRATIONS_PATH := migrations
MIGRATE := go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

help:
	@echo "Targets:"
	@echo "  make db-up            Start Postgres (docker compose)"
	@echo "  make db-down          Stop Postgres"
	@echo "  make migrate-up       Apply migrations"
	@echo "  make migrate-down     Roll back one migration"
	@echo "  make migrate-version  Show current migration version"
	@echo "  make run              Run the API server"
	@echo "  make dev              db-up + migrate-up + run"
	@echo "  make build            Build server binary to bin/server"
	@echo "  make test             Run tests with race detector"

db-up:
	docker compose up -d --wait

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres

migrate-up: db-up
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(MIGRATE_DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(MIGRATE_DATABASE_URL)" down 1

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(MIGRATE_DATABASE_URL)" version

run:
	go run ./cmd/server

dev: migrate-up run

build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./... -race
