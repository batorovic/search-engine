.PHONY: build run test clean docker-up docker-down migrate swagger sqlc

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=search-engine

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) .

# Run the application locally
run:
	$(GOCMD) run .

# Run tests
test:
	$(GOTEST) -v -cover ./...

# Run tests with coverage report
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Start Docker services
docker-up:
	docker-compose up -d postgres redis

# Stop Docker services
docker-down:
	docker-compose down

# Build and run with Docker
docker-build:
	docker-compose build api

# Run everything with Docker
docker-run:
	docker-compose up

# Run database migrations
migrate:
	@echo "Running migrations..."
	docker exec -i postgres psql -U postgres -d search_engine < migrations/001_init.sql

# Generate SQLC code
sqlc:
	sqlc generate

# Generate Swagger docs
swagger:
	swag init -g main.go -o docs

# Install development tools
tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Lint code
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GOCMD) fmt ./...
