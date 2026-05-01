.PHONY: help build run docker-up docker-down docker-logs test clean

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application locally"
	@echo "  make docker-up    - Start all Docker containers"
	@echo "  make docker-down  - Stop all Docker containers"
	@echo "  make docker-logs  - View Docker container logs"
	@echo "  make docker-build - Rebuild and restart Docker containers"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"

build:
	go build -o fragments cmd/fragments/main.go

run:
	go run cmd/fragments/main.go

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose up -d --build

test:
	go test ./... -v

clean:
	rm -f fragments
