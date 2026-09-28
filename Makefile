# Load .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Variables
DATABASE_URL       := postgres://$(DB_USERNAME):$(DB_PASSWORD)@psql_bp:5432/$(DB_DATABASE)?sslmode=disable
PROD_DATABASE_URL  := postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):5432/$(DB_DATABASE)?sslmode=$(DB_SSLMODE)
COMPOSE            := docker compose
COMPOSE_PROD       := docker compose -f docker-compose.prod.yml

# Atlas
MIGRATIONS_DIR := file:///src/migrations
SCHEMA_DIR     := file:///src/internal/database/sql/schema
DEV_DB_URL     := postgres://$(DB_USERNAME):$(DB_PASSWORD)@psql_bp:5432/atlas_dev?sslmode=disable
ATLAS          := $(COMPOSE) --profile tools run --rm atlas
ATLAS_PROD     := $(COMPOSE_PROD) --profile tools run --rm atlas

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

dev-build: ## Force rebuild development environment
	@$(COMPOSE) up --build -d

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
	@$(MAKE) migrate-up
	@echo "Database reset complete"

db-backup: ## Backup development database
	@mkdir -p backups
	@$(COMPOSE) exec -T psql_bp pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created"

db-schema-dump: ## Dump current DB schema to db/schema/_current.sql
	@mkdir -p db/schema
	@$(COMPOSE) exec -T psql_bp pg_dump \
		--schema-only --no-owner --no-privileges \
		-U $(DB_USERNAME) $(DB_DATABASE) > db/schema/_current.sql
	@echo "Schema dumped to db/schema/_current.sql"

# Migrations (Development) — Atlas
migrate-diff: ## Generate migration from db/schema changes (usage: make migrate-diff NAME=add_notes)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-diff NAME=add_notes"; \
		exit 1; \
	fi
	@echo "Generating migration: $(NAME)"
	@$(ATLAS) migrate diff $(NAME) \
		--dir $(MIGRATIONS_DIR) \
		--to $(SCHEMA_DIR) \
		--dev-url "$(DEV_DB_URL)"
	@echo "Migration generated in migrations/"

migrate-up: ## Apply migrations to development
	@echo "Applying migrations..."
	@$(ATLAS) migrate apply \
		--dir $(MIGRATIONS_DIR) \
		--url "$(DATABASE_URL)" \
		--revisions-schema public

migrate-down: ## Rollback last migration in development
	@echo "Rolling back last migration..."
	@$(ATLAS) migrate down \
		--dir $(MIGRATIONS_DIR) \
		--url "$(DATABASE_URL)" \
		--revisions-schema public \
		--to-version -1

migrate-version: ## Show current migration status
	@$(ATLAS) migrate status \
		--dir $(MIGRATIONS_DIR) \
		--url "$(DATABASE_URL)"\
		--revisions-schema public

migrate-validate: ## Validate migrations (atlas.sum integrity)
	@$(ATLAS) migrate validate \
		--dir $(MIGRATIONS_DIR) \
		--dev-url $(DEV_DB_URL)

migrate-lint: ## Lint latest migration
	@$(ATLAS) migrate lint \
		--dir $(MIGRATIONS_DIR) \
		--dev-url $(DEV_DB_URL) \
		--latest 1

migrate-hash: ## Regenerate atlas.sum
	@$(ATLAS) migrate hash --dir $(MIGRATIONS_DIR)

migrate-apply-version: ## Apply to specific version (usage: make migrate-apply-version VERSION=20240101120000)
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-apply-version VERSION=20240101120000"; \
		exit 1; \
	fi
	@$(ATLAS) migrate apply \
		--dir $(MIGRATIONS_DIR) \
		--url "$(DATABASE_URL)" \
		--to-version $(VERSION)

# Migrations (Production) — Atlas
prod-migrate-up: ## Apply migrations to production
	@echo "Applying migrations to production..."
	@$(ATLAS_PROD) migrate apply \
		--dir $(MIGRATIONS_DIR) \
		--url "$(PROD_DATABASE_URL)"

prod-migrate-down: ## Rollback last migration in production
	@echo "Rolling back production migration..."
	@$(ATLAS_PROD) migrate down \
		--dir $(MIGRATIONS_DIR) \
		--url "$(PROD_DATABASE_URL)" \
		--to-version -1

prod-migrate-version: ## Show production migration status
	@$(ATLAS_PROD) migrate status \
		--dir $(MIGRATIONS_DIR) \
		--url "$(PROD_DATABASE_URL)"

prod-migrate-validate: ## Validate migrations before prod deploy
	@$(ATLAS_PROD) migrate validate \
		--dir $(MIGRATIONS_DIR) \
		--dev-url "$(DEV_DB_URL)"

# Production database
prod-db-shell: ## Open production database shell
	@$(COMPOSE_PROD) exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE)

prod-db-backup: ## Backup production database
	@mkdir -p backups
	@$(COMPOSE_PROD) exec -T psql_bp pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) > backups/prod_backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Production backup created"

# Code generation and testing
sqlc: ## Generate SQL code from queries
	@echo "Generating Go code from SQL queries..."
	@docker run --rm -v "$(PWD):/src" -w /src sqlc/sqlc:1.27.0 generate
	@echo "SQL code generation complete - check internal/database/"

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
	@$(MAKE) prod-db-backup
	@echo "Step 4: Validating migrations"
	@$(MAKE) prod-migrate-validate
	@echo "Step 5: Running migrations"
	@$(MAKE) prod-migrate-up
	@echo "Step 6: Starting services"
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

.PHONY: build run watch dev dev-build dev-down dev-logs \
	prod prod-down prod-logs \
	db-shell db-reset db-backup db-schema-dump \
	migrate-diff migrate-up migrate-down migrate-version \
	migrate-hash migrate-validate migrate-lint migrate-hash migrate-apply-version \
	prod-migrate-up prod-migrate-down prod-migrate-version prod-migrate-validate \
	prod-db-shell prod-db-backup \
	sqlc test itest deploy clean logs help