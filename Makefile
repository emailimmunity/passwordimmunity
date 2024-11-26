.PHONY: all build test clean docker-build docker-run

# Default target
all: build

# Build the application
build:
	go build -o bin/passwordimmunity ./src

# Run tests
test:
	go test -v ./tests/...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f *.out

# Run application
run: build
	./bin/passwordimmunity

# Build docker image
docker-build:
	docker build -t passwordimmunity .

# Run docker container
docker-run:
	docker run -p 8000:8000 passwordimmunity

# Run with docker-compose
docker-compose-up:
	docker-compose up -d

# Stop docker-compose services
docker-compose-down:
	docker-compose down

# Generate test coverage report
coverage:
	go test -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Install development dependencies
deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database migrations
migrate-up:
	@echo "Database migrations will be implemented in a separate PR"

migrate-down:
	@echo "Database migrations will be implemented in a separate PR"

# Help target
help:
	@echo "Available targets:"
	@echo "  build            - Build the application"
	@echo "  test             - Run tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  run              - Run the application"
	@echo "  docker-build     - Build docker image"
	@echo "  docker-run       - Run docker container"
	@echo "  docker-compose-up - Start all services with docker-compose"
	@echo "  docker-compose-down - Stop all services"
	@echo "  coverage         - Generate test coverage report"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linter"
	@echo "  deps             - Install development dependencies"
	@echo "  migrate-up       - Run database migrations up"
	@echo "  migrate-down     - Run database migrations down"
