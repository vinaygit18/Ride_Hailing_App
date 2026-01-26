.PHONY: help docker-up docker-down migrate-up migrate-down migrate-create run build test-unit test-integration test-coverage lint clean env setup seed

# Variables
BINARY_NAME=gocomet-api
MIGRATION_DIR=./migrations
DB_URL=postgresql://postgres:postgres@localhost:5432/gocomet?sslmode=disable
GOBIN=$(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

docker-up: ## Start PostgreSQL and Redis containers
	@echo "Starting infrastructure..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Infrastructure is ready!"

docker-down: ## Stop and remove containers
	@echo "Stopping infrastructure..."
	docker-compose down
	@echo "Infrastructure stopped!"

docker-clean: ## Stop containers and remove volumes
	@echo "Cleaning infrastructure..."
	docker-compose down -v
	@echo "Infrastructure cleaned!"

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "Migrations completed!"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down
	@echo "Rollback completed!"

migrate-create: ## Create a new migration file (usage: make migrate-create name=create_users)
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make migrate-create name=create_users"; exit 1; fi
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(name)

deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies downloaded!"

run: ## Run the application
	@echo "Starting application..."
	go run cmd/api/main.go

build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Build complete: bin/$(BINARY_NAME)"

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -short ./tests/unit/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

test-all: ## Run all tests
	@echo "Running all tests..."
	go test -v ./...

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted!"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

seed: ## Seed PostgreSQL and Redis with sample data
	@echo "Seeding database..."
	go run scripts/seed_data.go

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .
	@echo "Docker image built!"

docker-run: ## Run application in Docker
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env --network gocomet_gocomet-network $(BINARY_NAME):latest

env: ## Create .env file from .env.example if it doesn't exist
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo ".env file created!"; \
	else \
		echo ".env file already exists."; \
	fi

setup: env docker-up deps migrate-up seed ## Complete setup (env + infrastructure + dependencies + migrations + seed data)
	@echo ""
	@echo "=========================================="
	@echo "  Setup complete!"
	@echo "=========================================="
	@echo ""
	@echo "Run 'make run' to start the application."
	@echo ""
	@echo "Then open:"
	@echo "  - Rider UI:  http://localhost:8080/rider"
	@echo "  - Driver UI: http://localhost:8080/driver"
	@echo ""

dev: ## Run in development mode with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air
