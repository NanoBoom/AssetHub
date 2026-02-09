.PHONY: help build run dev test lint clean db-create swag-init swag-fmt docs
.PHONY: docker-build docker-up docker-down docker-logs docker-ps docker-clean

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload (development mode)"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make db-create      - Create database"
	@echo "  make swag-init      - Generate Swagger documentation"
	@echo "  make swag-fmt       - Format Swagger annotations"
	@echo "  make docs           - Generate Swagger docs (alias for swag-init)"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start all services with docker-compose"
	@echo "  make docker-down    - Stop all services"
	@echo "  make docker-logs    - View API logs"
	@echo "  make docker-ps      - List running containers"
	@echo "  make docker-clean   - Remove containers and volumes"

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

# ========================================
# Docker Commands
# ========================================
docker-build:
	docker buildx build --platform linux/amd64 -t assethub:latest .

docker-up:
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services started:"
	@echo "  API:      http://localhost:8080"
	@echo "  Swagger:  http://localhost:8080/swagger/index.html"
	@echo "  MinIO:    http://localhost:9001 (minioadmin/minioadmin)"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

docker-ps:
	docker-compose ps

docker-clean:
	docker-compose down -v
	docker system prune -f
	@echo "All containers, volumes, and images removed"

