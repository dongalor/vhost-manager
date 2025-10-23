.PHONY: build install clean test help

# Variables
BINARY_NAME=vhost-manager
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Build complete for all platforms"

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run .

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  install      - Install the binary to $(INSTALL_DIR)"
	@echo "  uninstall    - Remove the binary from $(INSTALL_DIR)"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  run          - Run the application"
	@echo "  help         - Show this help message"