SHELL := /bin/bash
GO ?= go

COMPOSE_FILE := ../../docker-compose.yml
MIGRATION_DIR := ./migrations

.PHONY: db-up db-down db-logs migrate-up migrate-down migrate-force run gen test

db-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-logs:
	docker compose -f $(COMPOSE_FILE) logs -f postgres

migrate-up:
	$(GO) run ./cmd/migrate -direction up

migrate-down:
	$(GO) run ./cmd/migrate -direction down -steps 1

migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required"; exit 1; fi
	$(GO) run ./cmd/migrate -direction force -version $(VERSION)

gen:
	sqlc generate
	$(GO) run github.com/99designs/gqlgen@v0.17.0 generate

run:
	$(GO) run ./cmd/api

test:
	$(GO) test ./...
