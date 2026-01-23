.PHONY: help build run test clean migrate-up migrate-down docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate-up   - Run database migrations up"
	@echo "  make migrate-down - Run database migrations down"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"

build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

run:
	@echo "Running application..."
	go run cmd/api/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/

migrate-up:
	@echo "Running migrations up..."
	@for file in migrations/*up.sql; do \
		echo "Executing $$file"; \
		psql $(DB_URL) -f $$file; \
	done

migrate-down:
	@echo "Running migrations down..."
	@for file in migrations/*down.sql; do \
		echo "Executing $$file"; \
		psql $(DB_URL) -f $$file; \
	done

docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down
