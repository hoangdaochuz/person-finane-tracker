.PHONY: help run test lint build clean docker-build docker-run docker-stop deps

# Variables
APP_NAME=finance-tracker-backend
CMD_DIR=./cmd/api
BUILD_DIR=./build
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_COMPOSE=../deploy/docker-compose.yaml

# Default target
help:
	@echo "Available targets:"
	@echo "  make run          - Run the application locally"
	@echo "  make test         - Run tests with coverage"
	@echo "  make lint         - Run linters"
	@echo "  make build        - Build the application"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run with Docker Compose"
	@echo "  make docker-stop  - Stop Docker Compose"
	@echo "  make migration-up - Run database migrations up"
	@echo "  make migration-down - Rollback database migrations"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run $(CMD_DIR)/main.go -config config.yaml

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	@go tool cover -func=coverage.out | grep total

# Run linters
lint:
	@echo "Running linters..."
	@echo "Running go fmt..."
	@go fmt ./...
	@echo "Running golangci-lint..."
	@golangci-lint run --config .golangci.yml ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "Built $(BUILD_DIR)/$(APP_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@go clean

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

# Docker run with docker-compose
docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose -f $(DOCKER_COMPOSE) up -d
	@echo "Services started. Check logs with: docker-compose -f $(DOCKER_COMPOSE) logs -f"

# Docker stop
docker-stop:
	@echo "Stopping services..."
	@docker-compose -f $(DOCKER_COMPOSE) down

# Migration up
migration-up:
	@echo "Running migrations up..."
	@echo "Note: Install golang-migrate first: https://github.com/golang-migrate/migrate"
	@migrate -path ./migrations -database "postgres://finance:finance@localhost:5432/finance_tracker?sslmode=disable" up

# Migration down
migration-down:
	@echo "Running migrations down..."
	@echo "Note: Install golang-migrate first: https://github.com/golang-migrate/migrate"
	@migrate -path ./migrations -database "postgres://finance:finance@localhost:5432/finance_tracker?sslmode=disable" down 1
