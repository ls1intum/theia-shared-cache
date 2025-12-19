.PHONY: build run test clean docker-build docker-up docker-down lint fmt

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
BINARY_NAME=gradle-cache

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server

# Run the application locally
run: build
	./$(BINARY_NAME) -config configs/config.yaml

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
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

# Format code
fmt:
	$(GOFMT) -s -w .

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t gradle-cache:latest .

# Start all services with Docker Compose
docker-up:
	cd deployments && docker-compose up -d

# Stop all services
docker-down:
	cd deployments && docker-compose down

# View logs
docker-logs:
	cd deployments && docker-compose logs -f

# Rebuild and restart
docker-restart: docker-down docker-build docker-up

# Quick health check
health:
	curl -s http://localhost:8080/health | jq .

# Check metrics
metrics:
	curl -s http://localhost:8080/metrics | grep gradle_cache
