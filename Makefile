APP_NAME = ticktockbox
BUILD_DIR = bin
MAIN_PATH = cmd/server/main.go
DOCKER_IMAGE = ticktockbox:latest

GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod

LDFLAGS = -ldflags "-X main.version=$(shell git describe --tags --always --dirty) -s -w"
BUILD_FLAGS = -v $(LDFLAGS)

.PHONY: all build clean test deps run docker-build docker-run help

all: clean deps test build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(APP_NAME)"

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build completed"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR) coverage.out coverage.html release/
	@echo "Clean completed"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

run: build
	@echo "Starting $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	@echo "Starting $(APP_NAME) in development mode..."
	$(GOCMD) run $(MAIN_PATH)

deps-up:
	@echo "Starting dependencies..."
	docker compose up -d
	@echo "Dependencies started"

deps-down:
	@echo "Stopping dependencies..."
	docker compose down
	@echo "Dependencies stopped"

deps-restart: deps-down deps-up

deps-logs:
	docker compose logs -f

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run:
	@echo "Starting full stack with Docker Compose..."
	docker compose -f docker-compose.full.yml up -d
	@echo "Full stack started"

docker-stop:
	@echo "Stopping Docker Compose stack..."
	docker compose -f docker-compose.full.yml down
	@echo "Docker stack stopped"

docker-logs:
	docker compose -f docker-compose.full.yml logs -f

lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "Linting completed"; \
	else \
		echo "golangci-lint not installed. Run 'make install-tools' first"; \
	fi

install-tools:
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air..."; \
		$(GOCMD) install github.com/cosmtrek/air@latest; \
	fi
	@echo "Development tools installed"

watch:
	@echo "Starting hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Run 'make install-tools' first"; \
	fi

db-reset:
	@echo "Resetting database..."
	docker compose restart questdb
	@sleep 5
	@echo "Database reset completed"

health:
	@echo "Checking application health..."
	@curl -f http://localhost:3000/api/messages >/dev/null 2>&1 && echo "Application is healthy" || echo "Application not responding"

status:
	@echo "=== TickTockBox Status ==="
	@echo "Application:"
	@curl -s http://localhost:3000/api/messages >/dev/null 2>&1 && echo "Running" || echo "Not running"
	@echo "QuestDB:"
	@curl -s http://localhost:9000 >/dev/null 2>&1 && echo "Running" || echo "Not running"
	@echo "RabbitMQ:"
	@curl -s http://localhost:15672 >/dev/null 2>&1 && echo "Running" || echo "Not running"

test-message:
	@echo "Creating test message..."
	@curl -X POST http://localhost:3000/api/messages \
		-H "Content-Type: application/json" \
		-d '{"message": "Test message from Makefile", "expire_at": "'$(shell date -u -d '+1 minute' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+1M +%Y-%m-%dT%H:%M:%SZ)'"}' \
		2>/dev/null | jq . 2>/dev/null || echo "Message sent (jq not available for formatting)"

get-messages:
	@echo "Getting all messages..."
	@curl -s http://localhost:3000/api/messages | jq . 2>/dev/null || curl -s http://localhost:3000/api/messages

test-websocket:
	@echo "Testing WebSocket connection..."
	@echo "Connect to: ws://localhost:3000/ws"
	@echo "Use a WebSocket client or browser console to test real-time updates"

release: clean build-all
	@echo "Creating release archive..."
	@mkdir -p release
	@tar -czf release/$(APP_NAME)-$(shell git describe --tags --always).tar.gz -C $(BUILD_DIR) .
	@echo "Release archive created: release/$(APP_NAME)-$(shell git describe --tags --always).tar.gz"

help:
	@echo "TickTockBox Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the application"
	@echo "  build-all      Build for multiple platforms (Linux, macOS, Windows)"
	@echo "  clean          Clean build artifacts and temporary files"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev            Run in development mode (no build required)"
	@echo "  run            Build and run the application"
	@echo "  watch          Start hot reload for development (requires air)"
	@echo "  fmt            Format Go code"
	@echo "  lint           Lint code (requires golangci-lint)"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps           Download and tidy Go dependencies"
	@echo "  deps-up        Start QuestDB and RabbitMQ with Docker"
	@echo "  deps-down      Stop dependency containers"
	@echo "  deps-restart   Restart dependency containers"
	@echo "  deps-logs      View dependency container logs"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run full stack with Docker Compose"
	@echo "  docker-stop    Stop Docker Compose stack"
	@echo "  docker-logs    View Docker container logs"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-reset       Reset QuestDB database"
	@echo ""
	@echo "Testing Commands:"
	@echo "  health         Check application health"
	@echo "  status         Show all services status"
	@echo "  test-message   Create a test message via API"
	@echo "  get-messages   Get all messages via API"
	@echo "  test-websocket Show WebSocket connection info"
	@echo ""
	@echo "  help           Show this help message" 