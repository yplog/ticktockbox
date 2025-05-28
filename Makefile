# TickTockBox Makefile

# Variables
APP_NAME = ticktockbox
BUILD_DIR = bin
MAIN_PATH = cmd/server/main.go
DOCKER_IMAGE = ticktockbox:latest

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

LDFLAGS = -ldflags "-X main.version=$(shell git describe --tags --always --dirty)"
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
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build completed"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: build
	@echo "Starting $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	@echo "Starting $(APP_NAME) in development mode..."
	$(GOCMD) run $(MAIN_PATH)

deps-up:
	@echo "Starting dependencies..."
	docker-compose up -d
	@echo "Dependencies started"

deps-down:
	@echo "Stopping dependencies..."
	docker-compose down
	@echo "Dependencies stopped"

deps-restart: deps-down deps-up

deps-logs:
	docker-compose logs -f

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run:
	@echo "Starting full stack with Docker Compose..."
	docker-compose -f docker-compose.full.yml up -d
	@echo "Full stack started"

docker-stop:
	@echo "Stopping Docker Compose stack..."
	docker-compose -f docker-compose.full.yml down
	@echo "Docker stack stopped"

docker-logs:
	docker-compose -f docker-compose.full.yml logs -f

lint:
	@echo "Linting code..."
	golangci-lint run
	@echo "Linting completed"

install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/cosmtrek/air
	@echo "Development tools installed"

watch:
	@echo "Starting hot reload..."
	air

db-reset:
	@echo "Resetting database..."
	docker-compose restart questdb
	@sleep 5
	@echo "Database reset completed"

health:
	@echo "Checking application health..."
	@curl -f http://localhost:3000/health || echo "Application not responding"

status:
	@echo "=== Application Status ==="
	@echo "Application:"
	@curl -s http://localhost:3000/health | jq . || echo "  Not running"
	@echo ""
	@echo "QuestDB:"
	@curl -s http://localhost:9000/status | jq . || echo "  Not running"
	@echo ""
	@echo "RabbitMQ:"
	@curl -s -u guest:guest http://localhost:15672/api/overview | jq .management_version || echo "  Not running"

test-message:
	@echo "Creating test message..."
	@curl -X POST http://localhost:3000/api/messages \
		-H "Content-Type: application/json" \
		-d '{"message": "Test message from Makefile", "expire_at": "'$(shell date -u -d '+1 minute' +%Y-%m-%dT%H:%M:%SZ)'"}' \
		| jq .

get-messages:
	@echo "Getting all messages..."
	@curl -s http://localhost:3000/api/messages | jq .

deploy-prod:
	@echo "Deploying to production..."
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-prod $(MAIN_PATH)
	@echo "Production binary ready: $(BUILD_DIR)/$(APP_NAME)-prod"

release: clean build-all
	@echo "Creating release archive..."
	@mkdir -p release
	@tar -czf release/$(APP_NAME)-$(shell git describe --tags --always).tar.gz -C $(BUILD_DIR) .
	@echo "Release archive created: release/$(APP_NAME)-$(shell git describe --tags --always).tar.gz"

benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

profile:
	@echo "Starting application with profiling..."
	$(GOCMD) run $(MAIN_PATH) -cpuprofile=cpu.prof -memprofile=mem.prof

help:
	@echo "TickTockBox Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the application"
	@echo "  build-all      Build for multiple platforms"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev            Run in development mode"
	@echo "  run            Build and run the application"
	@echo "  watch          Start hot reload for development"
	@echo "  fmt            Format code"
	@echo "  lint           Lint code"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps           Download Go dependencies"
	@echo "  deps-up        Start QuestDB and RabbitMQ"
	@echo "  deps-down      Stop dependencies"
	@echo "  deps-restart   Restart dependencies"
	@echo "  deps-logs      View dependency logs"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run full stack with Docker"
	@echo "  docker-stop    Stop Docker stack"
	@echo "  docker-logs    View Docker logs"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-reset       Reset QuestDB database"
	@echo ""
	@echo "Testing Commands:"
	@echo "  health         Check application health"
	@echo "  status         Show all services status"
	@echo "  test-message   Create a test message"
	@echo "  get-messages   Get all messages"
	@echo ""
	@echo "Production Commands:"
	@echo "  deploy-prod    Build production binary"
	@echo "  release        Create release archive"
	@echo ""
	@echo "Quality Commands:"
	@echo "  security-scan  Run security scan"
	@echo "  benchmark      Run benchmark tests"
	@echo "  profile        Profile the application"
	@echo ""
	@echo "Documentation:"
	@echo "  docs           Generate documentation"
	@echo "  help           Show this help message" 