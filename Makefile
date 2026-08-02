# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main.exe cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-up:
	@docker compose up --build

# Shutdown DB container
docker-down:
	@docker compose down

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# # Live Reload
# watch:
# 	@if command -v air >/dev/null 2>&1; then \
#         air; \
#         echo "Watching..."; \
#     else \
#         echo "Installing air..."; \
#         go install github.com/air-verse/air@latest; \
#         air; \
#         echo "Watching..."; \
#     fi
# # @powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
# 		air; \
# 		Write-Output 'Watching...'; \
# 	} else { \
# 		Write-Output 'Installing air...'; \
# 		go install github.com/air-verse/air@latest; \
# 		air; \
# 		Write-Output 'Watching...'; \
# 	}"

# # Generate Go code from SQL queries using sqlc
# sqlc-gen:
# 	@echo "Generating Go code from SQL queries..."
# 	@sqlc generate

# .PHONY: all build run test clean watch docker-run docker-down itest

# Live Reload
watch:
	@if command -v air >/dev/null 2>&1; then \
        air; \
        echo "Watching..."; \
    else \
        echo "Installing air..."; \
        go install github.com/air-verse/air@latest; \
        air; \
        echo "Watching..."; \
    fi

# Generate Go code from SQL queries using sqlc (Docker-based)
sqlc-gen:
	@echo "Generating Go code from SQL queries..."
	@docker run --rm -v "$$(pwd):/src" -w /src sqlc/sqlc:1.27.0 generate

# Database migrations (requires migrations folder)
migrate-up:
	@docker compose --profile tools run --rm golang-migrate \
		-path /migrations \
		-database "postgres://${DB_USERNAME}:${DB_PASSWORD}@psql_bp:5432/${DB_DATABASE}?sslmode=disable" \
		up

migrate-down:
	@docker compose --profile tools run --rm golang-migrate \
		-path /migrations \
		-database "postgres://${DB_USERNAME}:${DB_PASSWORD}@psql_bp:5432/${DB_DATABASE}?sslmode=disable" \
		down 1

migrate-create:
	@read -p "Enter migration name: " name; \
	docker compose --profile tools run --rm golang-migrate \
		create -ext sql -dir /migrations $$name

# Quick setup for new developers
setup:
	@echo "Setting up development environment..."
	@docker compose --profile tools run --rm sqlc
	@echo "Building containers..."
	@docker compose build
	@echo "Setup complete! Run 'make docker-up' to start"

# Start fresh (reset everything)
fresh: docker-down
	@docker compose down -v
	@docker compose --profile tools run --rm sqlc
	@docker compose up --build

.PHONY: all build run test clean watch docker-up docker-down itest sqlc-gen migrate-up migrate-down migrate-create setup fresh


get-win-ip:
	@cat /etc/resolv.conf | grep nameserver | awk '{print $2}'  
