# Variables
BINARY_NAME=vtcpd-cli
BUILD_DIR=build
MAIN_FILE=cmd/vtcpd-cli/main.go
MAIN_FILE_TESTING=cmd/vtcpd-cli/testing/main.go

# Commands
.PHONY: all build clean test run

# Default
all: clean build

# Build project
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Build project in testing mode
build-testing:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE_TESTING)	

# Clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Tests launching
test:
	@echo "Running tests..."
	@go test ./...

# Program lounching
run: build
	@echo "Running..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Installing dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download

# Checking code formatting
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Linter check
lint:
	@echo "Linting..."
	@golangci-lint run

# Creating a configuration file
config:
	@echo "Creating config file..."
	@mkdir -p $(BUILD_DIR)
	@cp conf.yaml $(BUILD_DIR)/conf.yaml

# Complete assembly with all checks
all: fmt lint test build config
