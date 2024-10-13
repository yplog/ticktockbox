APP_NAME := ticktockbox
BUILD_DIR := build
SRC_DIR := ./cmd/ticktockbox
PKG := ./...

all: build

build:
	@echo "Building the application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)

run: build
	@echo "Running the application..."
	@./$(BUILD_DIR)/$(APP_NAME)

test:
	@echo "Running tests..."
	@go test $(PKG)

clean:
	@echo "Cleaning the build directory..."
	@rm -rf $(BUILD_DIR)

fmt:
	@echo "Formatting the code..."
	@go fmt $(PKG)

lint:
	@echo "Linting the code..."
	@golangci-lint run

deps:
	@echo "Installing dependencies..."
	@go mod tidy

dev:
	@echo "Running the application with live reload..."
	@mkdir -p $(BUILD_DIR)/air/tmp
	@air

.PHONY: all build run test clean fmt lint deps dev