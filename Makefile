.PHONY: help build run dev test lint clean db-create swag-init swag-fmt docs

help:
	@echo "Available commands:"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make dev       - Run with hot reload (development mode)"
	@echo "  make test      - Run tests"
	@echo "  make lint      - Run linter"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make db-create - Create database"
	@echo "  make swag-init - Generate Swagger documentation"
	@echo "  make swag-fmt  - Format Swagger annotations"
	@echo "  make docs      - Generate Swagger docs (alias for swag-init)"

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

dev:
	@command -v air >/dev/null 2>&1 || { \
		echo "air not found in PATH, using GOPATH..."; \
		$$(go env GOPATH)/bin/air; \
		exit 0; \
	}; \
	air

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/ tmp/

db-create:
	docker exec dreamlife-postgres psql -U postgres -c "CREATE DATABASE assethub;" || true

swag-init:
	swag init -g cmd/api/main.go -o docs

swag-fmt:
	swag fmt -g cmd/api/main.go

docs: swag-init
	@echo "Swagger docs generated at docs/"
