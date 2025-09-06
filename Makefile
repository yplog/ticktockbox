# TickTockBox Makefile

# Variables
BINARY_NAME=ticktockbox
SERVER_BINARY=./bin/$(BINARY_NAME)
SOURCE_DIR=./cmd
BUILD_DIR=./bin
DOCKER_IMAGE=ticktockbox
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Environment
ENV_FILE=.env
DB_PATH=app.db

## Development
.PHONY: dev
dev: ## Start development server
	$(GOCMD) run $(SOURCE_DIR)/server/main.go

.PHONY: vet
vet: ## Run go vet
	$(GOCMD) vet ./...

## Build
.PHONY: build
build: clean ## Build server and seed binaries
	@echo "Building TickTockBox..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(SERVER_BINARY) $(SOURCE_DIR)/server/main.go
	@echo "Build completed: $(SERVER_BINARY)"

.PHONY: build-server
build-server: ## Build only server binary
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(SERVER_BINARY) $(SOURCE_DIR)/server/main.go

.PHONY: build-linux
build-linux: ## Build for Linux
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-server-linux $(SOURCE_DIR)/server/main.go
	
.PHONY: build-windows
build-windows: ## Build for Windows
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-server.exe $(SOURCE_DIR)/server/main.go
	
.PHONY: build-all
build-all: build build-linux build-windows ## Build for all platforms

## Run
.PHONY: run
run: build-server ## Build and run server
	$(SERVER_BINARY)

.PHONY: run-server
run-server: build-server ## Build and run server
	$(SERVER_BINARY)

.PHONY: run-seed
run-seed: ## Run database seeder directly
	$(GOCMD) run $(SOURCE_DIR)/seed/main.go

## Database
.PHONY: seed
seed: ## Seed database with test data (1000 jobs)
	$(GOCMD) run $(SOURCE_DIR)/seed/main.go

.PHONY: seed-small
seed-small: ## Seed database with small test data (100 jobs)
	SEED_COUNT=100 $(GOCMD) run $(SOURCE_DIR)/seed/main.go

.PHONY: seed-large
seed-large: ## Seed database with large test data (100000 jobs)
	SEED_COUNT=100000 $(GOCMD) run $(SOURCE_DIR)/seed/main.go

.PHONY: seed-keep
seed-keep: ## Seed database without clearing existing data
	CLEAR_EXISTING=false SEED_COUNT=1000 $(GOCMD) run $(SOURCE_DIR)/seed/main.go

.PHONY: db-reset
db-reset: ## Remove database file and recreate
	@echo "Resetting database..."
	@rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal
	@echo "Database reset completed"

.PHONY: db-backup
db-backup: ## Backup database
	@echo "Backing up database..."
	@cp $(DB_PATH) $(DB_PATH).backup.$(shell date +%Y%m%d_%H%M%S)
	@echo "Database backup completed"

## Dependencies
.PHONY: deps
deps: ## Download and verify dependencies
	$(GOMOD) download
	$(GOMOD) verify

.PHONY: deps-update
deps-update: ## Update dependencies
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

.PHONY: deps-tidy
deps-tidy: ## Clean up dependencies
	$(GOMOD) tidy

## Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 8080:8080 --name $(BINARY_NAME) $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true

.PHONY: docker-up
docker-up: ## Start all services with docker-compose
	docker compose up -d

.PHONY: docker-down
docker-down: ## Stop all services with docker-compose
	docker compose down

.PHONY: docker-logs
docker-logs: ## Show docker-compose logs
	docker compose logs -f

.PHONY: docker-rabbitmq
docker-rabbitmq: ## Start RabbitMQ with docker-compose
	docker compose up -d rabbitmq

.PHONY: docker-rabbitmq-stop
docker-rabbitmq-stop: ## Stop RabbitMQ with docker-compose
	docker compose stop rabbitmq

.PHONY: docker-rabbitmq-logs
docker-rabbitmq-logs: ## Show RabbitMQ logs
	docker compose logs -f rabbitmq

.PHONY: docker-clean
docker-clean: ## Clean docker containers and volumes
	docker compose down -v
	docker system prune -f

## Environment
.PHONY: env-setup
env-setup: ## Setup environment file
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "Creating .env file from env.example..."; \
		cp env.example $(ENV_FILE); \
		echo "Please edit $(ENV_FILE) with your configuration"; \
	else \
		echo "$(ENV_FILE) already exists"; \
	fi

.PHONY: env-check
env-check: ## Check environment configuration
	@echo "Environment configuration:"
	@echo "ADDR: $${ADDR:-:8080}"
	@echo "SQLITE_PATH: $${SQLITE_PATH:-app.db}"
	@echo "RABBITMQ_URL: $${RABBITMQ_URL:-amqp://guest:guest@localhost:5672/}"
	@echo "RABBITMQ_QUEUE: $${RABBITMQ_QUEUE:-reminders.due}"

## Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f cpu.prof mem.prof
	@echo "Cleaned build artifacts"

.PHONY: clean-all
clean-all: clean ## Clean everything including database
	@rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal
	@rm -f $(DB_PATH).backup.*
	@echo "Cleaned all artifacts and database"

## Info
.PHONY: version
version: ## Show version info
	@echo "TickTockBox"
	@echo "Go version: $$($(GOCMD) version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build time: $$(date)"

# Default target when no arguments
.DEFAULT_GOAL := help
