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
# @powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
		air; \
		Write-Output 'Watching...'; \
	} else { \
		Write-Output 'Installing air...'; \
		go install github.com/air-verse/air@latest; \
		air; \
		Write-Output 'Watching...'; \
	}"

# Generate Go code from SQL queries using sqlc
sqlc-gen:
	@echo "Generating Go code from SQL queries..."
	@sqlc generate

.PHONY: all build run test clean watch docker-run docker-down itest

get-win-ip:
	@cat /etc/resolv.conf | grep nameserver | awk '{print $2}'  
