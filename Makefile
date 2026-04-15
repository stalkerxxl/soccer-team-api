# Load environment variables from .env when the file exists.
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

APP_NAME := soccer-team-api
GO_CMD := go
DOCKER_COMPOSE := docker compose

MIGRATIONS_DIR := ./migrations
DOCKER_MIGRATE_CMD := $(DOCKER_COMPOSE) run --rm migrate

.DEFAULT_GOAL := help

.PHONY: help build run up down logs restart migrate-up migrate-down migrate-status migrate-create test clean

## build: Build the API binary locally
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	@$(GO_CMD) build -o bin/api ./cmd/api

## run: Run the API locally
run:
	@echo "Running $(APP_NAME)..."
	@$(GO_CMD) run ./cmd/api

## up: Build images and start PostgreSQL, migrations, and API in detached mode
up:
	@echo "Building and starting containers..."
	@$(DOCKER_COMPOSE) up --build -d

## down: Stop and remove local containers
down:
	@echo "Stopping containers..."
	@$(DOCKER_COMPOSE) down

## logs: Follow container logs
logs:
	@$(DOCKER_COMPOSE) logs -f postgres migrate api

## restart: Restart local infrastructure
restart: down up

## migrate-up: Apply all up migrations inside the migration container
migrate-up:
	@echo "Running migrations up..."
	@$(DOCKER_MIGRATE_CMD) up

## migrate-down: Roll back the last migration inside the migration container
migrate-down:
	@echo "Rolling back the last migration..."
	@$(DOCKER_MIGRATE_CMD) down

## migrate-status: Show migration status inside the migration container
migrate-status:
	@echo "Checking migration status..."
	@$(DOCKER_MIGRATE_CMD) status

## migrate-create: Create a new SQL migration (usage: make migrate-create name=create_users)
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: please specify migration name. Usage: make migrate-create name=create_users"; \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	@timestamp=$$(date -u +%Y%m%d%H%M%S); \
	file="$(MIGRATIONS_DIR)/$${timestamp}_$(name).sql"; \
	printf '%s\n' \
		'-- +goose Up' \
		'-- +goose StatementBegin' \
		'' \
		'-- +goose StatementEnd' \
		'' \
		'-- +goose Down' \
		'-- +goose StatementBegin' \
		'' \
		'-- +goose StatementEnd' > "$$file"; \
	echo "Created migration $$file"

## test: Run all tests
test:
	@echo "Running tests..."
	@$(GO_CMD) test -v -race ./...

## clean: Remove local build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin
	@$(GO_CMD) clean

## help: Show available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'match($$0, /^## ([^:]+): (.*)$$/, a) {printf "  \033[36m%-18s\033[0m %s\n", a[1], a[2]}' Makefile
