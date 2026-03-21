# =============================================================================
# Makefile — Streaming Backend
# =============================================================================

BINARY_NAME=streaming-backend
MAIN_PATH=./cmd/api

# MySQL connection (must match .env / docker-compose.yml)
DB_URL=mysql://streaming:secret@tcp(localhost:3306)/streaming

.DEFAULT_GOAL := help

## run: Run the server in development mode
.PHONY: run
run:
	@echo "🚀 Starting server..."
	go run $(MAIN_PATH)/main.go

## build: Compile the server into a binary
.PHONY: build
build:
	@echo "📦 Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME).exe $(MAIN_PATH)/main.go
	@echo "✅ Built: bin/$(BINARY_NAME).exe"

## clean: Remove build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	@if exist bin rmdir /s /q bin
	@echo "✅ Clean complete"

## test: Run all tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v -race ./...

## lint: Run Go linter
.PHONY: lint
lint:
	golangci-lint run ./...

## ─── Docker ────────────────────────────────────────────────────────────
## docker-up: Start MySQL and Redis containers
.PHONY: docker-up
docker-up:
	@echo "🐳 Starting Docker containers..."
	docker-compose up -d
	@echo "✅ Containers running. Use 'docker-compose ps' to check."

## docker-down: Stop containers (keep data)
.PHONY: docker-down
docker-down:
	docker-compose down

## docker-reset: Stop containers AND delete all data
.PHONY: docker-reset
docker-reset:
	docker-compose down -v
	@echo "🗑️ All data deleted."

## ─── Database Migrations ───────────────────────────────────────────────
## migrate-up: Run all pending migrations
.PHONY: migrate-up
migrate-up:
	@echo "⬆️  Running migrations..."
	migrate -path migrations -database "$(DB_URL)" up
	@echo "✅ Migrations complete."

## migrate-down: Roll back the last migration
.PHONY: migrate-down
migrate-down:
	@echo "⬇️  Rolling back last migration..."
	migrate -path migrations -database "$(DB_URL)" down 1

## migrate-reset: Roll back ALL migrations
.PHONY: migrate-reset
migrate-reset:
	@echo "⚠️  Rolling back ALL migrations..."
	migrate -path migrations -database "$(DB_URL)" down -all

## help: Show this help message
.PHONY: help
help:
	@echo.
	@echo  Available commands:
	@echo  ──────────────────────────────────────
	@echo   make run           - Run the server
	@echo   make build         - Compile into binary
	@echo   make clean         - Remove build files
	@echo   make test          - Run all tests
	@echo   make docker-up     - Start MySQL + Redis
	@echo   make docker-down   - Stop containers
	@echo   make docker-reset  - Stop + delete all data
	@echo   make migrate-up    - Run DB migrations
	@echo   make migrate-down  - Rollback last migration
	@echo   make migrate-reset - Rollback ALL migrations
	@echo  ──────────────────────────────────────
