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
	@docker compose up --build

dev-detach:
	@docker compose up --build -d

dev-down:
	@docker compose down

dev-logs:
	@docker compose logs -f api

# Production (explicitly uses prod file)
prod:
	@docker compose -f docker-compose.prod.yml up -d --build

prod-down:
	@docker compose -f docker-compose.prod.yml down

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

# Production database commands (for EC2)
prod-db-shell:
	@docker compose -f docker-compose.prod.yml exec psql_bp psql -U $(DB_USERNAME) -d $(DB_DATABASE)

prod-db-backup:
	@docker compose -f docker-compose.prod.yml exec -T psql_bp pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) > backup_$(shell date +%Y%m%d_%H%M%S).sql

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

# Setup EC2 (run once on server)
setup-ec2:
	@chmod +x scripts/setup-ec2.sh
	@./scripts/setup-ec2.sh

# Deploy to production (run on EC2)
deploy-prod:
	@git pull origin main
	@make migrate-up || true
	@make prod
	@docker image prune -f
	@docker compose -f docker-compose.prod.yml ps

.PHONY: build run watch dev dev-detach dev-down dev-logs prod prod-down prod-restart prod-logs prod-ps api db-shell db-reset migrate-create migrate-up migrate-down migrate-status prod-db-shell prod-db-backup sqlc test itest test-coverage clean logs setup-ec2 deploy-prod