# Load .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Variables
DATABASE_URL := postgres://$(DB_USERNAME):$(DB_PASSWORD)@psql_bp:5432/$(DB_DATABASE)?sslmode=disable
PROD_DATABASE_URL := postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):5432/$(DB_DATABASE)?sslmode=$(DB_SSLMODE)
COMPOSE := docker compose
COMPOSE_PROD := docker compose -f docker-compose.prod.yml
MIGRATE := $(COMPOSE) --profile tools run --rm golang-migrate

# Build and run
build: ## Build the application
	@go build -o bin/api ./cmd/api

run: ## Run application locally
	@go run ./cmd/api

watch: ## Run with hot reload
	@docker compose up -d psql_bp
	@air

# Development
dev: ## Start development environment
	@$(COMPOSE) up --build

dev-down: ## Stop development environment
	@$(COMPOSE) down

dev-logs: ## View development logs
	@$(COMPOSE) logs -f api

# Production
prod: ## Start production environment
	@$(COMPOSE_PROD) up -d --build

prod-down: ## Stop production environment
	@$(COMPOSE_PROD) down

prod-logs: ## View production logs
	@$(COMPOSE_PROD) logs -f api

# Database
db-shell: ## Open development database shell
	@$(COMPOSE) exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE)

db-reset: ## Reset development database (drops all data)
	@echo "Resetting database..."
	@$(COMPOSE) exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MIGRATE) -path /migrations -database "$(DATABASE_URL)" up
	@echo "Database reset complete"

db-backup: ## Backup development database
	@mkdir -p backups
	@$(COMPOSE) exec -T psql_bp pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created"

# Migrations (Development)
migrate-create: ## Create new migration
	@read -p "Migration name: " name; \
	$(MIGRATE) create -ext sql -dir /migrations -seq $$name

migrate-up: ## Apply migrations to development
	@echo "Applying migrations..."
	@$(MIGRATE) -path /migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration in development
	@echo "Rolling back last migration..."
	@$(MIGRATE) -path /migrations -database "$(DATABASE_URL)" down 1

migrate-version: ## Show current migration version
	@$(MIGRATE) -path /migrations -database "$(DATABASE_URL)" version

# Migrations (Production)
prod-migrate-up: ## Apply migrations to production
	@echo "Applying migrations to production..."
	@$(COMPOSE_PROD) run --rm golang-migrate -path /migrations -database "$(PROD_DATABASE_URL)" up

prod-migrate-down: ## Rollback last migration in production
	@echo "Rolling back production migration..."
	@$(COMPOSE_PROD) run --rm golang-migrate -path /migrations -database "$(PROD_DATABASE_URL)" down 1

prod-migrate-version: ## Show production migration version
	@$(COMPOSE_PROD) run --rm golang-migrate -path /migrations -database "$(PROD_DATABASE_URL)" version

# Production database
prod-db-shell: ## Open production database shell
	@$(COMPOSE_PROD) exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE)

prod-db-backup: ## Backup production database
	@mkdir -p backups
	@$(COMPOSE_PROD) exec -T psql_bp pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) > backups/prod_backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Production backup created"

# Code generation and testing
sqlc: ## Generate SQL code
	@docker run --rm -v "$(PWD):/src" -w /src sqlc/sqlc:1.27.0 generate

test: ## Run tests
	@go test ./...

itest: ## Run integration tests
	@go test ./internal/database -v

# Deployment
deploy: ## Deploy to production
	@echo "Step 1: Pulling latest code"
	@git pull origin main
	@echo "Step 2: Building images"
	@$(COMPOSE_PROD) build
	@echo "Step 3: Backing up database"
	@make prod-db-backup
	@echo "Step 4: Running migrations"
	@$(COMPOSE_PROD) run --rm golang-migrate -path /migrations -database "$(PROD_DATABASE_URL)" up
	@echo "Step 5: Starting services"
	@$(COMPOSE_PROD) up -d
	@echo "Deployment complete"

# Utility
clean: ## Clean build artifacts
	@rm -rf bin/

logs: ## View application logs
	@$(COMPOSE) logs -f api

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

.PHONY: build run watch dev dev-down dev-logs prod prod-down prod-logs db-shell db-reset db-backup migrate-create migrate-up migrate-down migrate-version prod-migrate-up prod-migrate-down prod-migrate-version prod-db-shell prod-db-backup sqlc test itest deploy clean logs help