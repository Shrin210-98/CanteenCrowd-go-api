# Load .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Build
build:
	@go build -o bin/api ./cmd/api

# Run locally
run:
	@go run ./cmd/api

# Local hot reload (DB in Docker, API on host)
watch:
	@docker compose up -d psql_bp
	@air

# Docker development (full stack with hot reload)
dev:
	@docker compose -f docker-compose.dev.yml up --build

dev-detach:
	@docker compose -f docker-compose.dev.yml up --build -d

dev-down:
	@docker compose -f docker-compose.dev.yml down

dev-logs:
	@docker compose -f docker-compose.dev.yml logs -f api

# Docker production
prod:
	@docker compose up --build -d

prod-down:
	@docker compose down

# Rebuild only API service
api:
	@docker compose up -d --build api

# Database
db-shell:
	@docker compose exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE)

# Development: Reset DB and apply schemas directly
db-reset:
	@docker compose exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@for file in internal/database/sql/schemas/*.sql; do \
		docker compose exec -T psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE) < $$file; \
	done

# Migrations
migrate-create:
	@read -p "Enter migration name: " name; \
	docker compose --profile tools run --rm golang-migrate \
		create -ext sql -dir /migrations $$name

migrate-up:
	@docker compose --profile tools run --rm golang-migrate -path /migrations -database "postgres://$(DB_USERNAME):$(DB_PASSWORD)@psql_bp:5432/$(DB_DATABASE)?sslmode=disable" up

migrate-down:
	@docker compose --profile tools run --rm golang-migrate -path /migrations -database "postgres://$(DB_USERNAME):$(DB_PASSWORD)@psql_bp:5432/$(DB_DATABASE)?sslmode=disable" down 1

# SQL generation
sqlc:
	@docker run --rm -v "$(PWD):/src" -w /src sqlc/sqlc:1.27.0 generate

# Testing
test:
	@go test ./...

itest:
	@go test ./internal/database -v

# Clean
clean:
	@rm -rf bin/

# Logs
logs:
	@docker compose logs -f api

.PHONY: build run watch dev dev-detach dev-down dev-logs prod prod-down api db-shell db-reset migrate-create migrate-up migrate-down sqlc test itest clean logs