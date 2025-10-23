#!/bin/bash

# vhost-manager installation script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="vhost-manager"
REPO_URL="https://github.com/dongalor/vhost-manager.git"
TEMP_DIR="/tmp/vhost-manager-install"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root. Please run as a regular user."
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_status "Visit https://golang.org/doc/install for installation instructions."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    
    if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
        print_error "Go version $GO_VERSION is not supported. Please install Go $REQUIRED_VERSION or later."
        exit 1
    fi
    
    if ! command_exists git; then
        print_error "Git is not installed. Please install Git."
        exit 1
    fi
    
    print_status "Prerequisites check passed."
}

# Function to build the binary
build_binary() {
    print_status "Building vhost-manager..."
    
    # Create temporary directory
    rm -rf "$TEMP_DIR"
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # Clone repository
    print_status "Cloning repository..."
    git clone "$REPO_URL" .
    
    # Download dependencies
    print_status "Downloading dependencies..."
    go mod download
    go mod tidy
    
    # Build binary
    print_status "Building binary..."
    go build -o "$BINARY_NAME" .
    
    print_status "Build completed successfully."
}

# Function to install the binary
install_binary() {
    print_status "Installing vhost-manager to $INSTALL_DIR..."
    
    # Check if install directory exists
    if [[ ! -d "$INSTALL_DIR" ]]; then
        print_status "Creating install directory: $INSTALL_DIR"
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    # Copy binary
    sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    print_status "Installation completed successfully."
}

# Function to verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command_exists "$BINARY_NAME"; then
        VERSION=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        print_status "vhost-manager installed successfully!"
        print_status "Version: $VERSION"
        print_status "Location: $(which $BINARY_NAME)"
    else
        print_error "Installation verification failed."
        exit 1
    fi
}

# Function to show usage information
show_usage() {
    print_status "Usage examples:"
    echo "  $BINARY_NAME add example.com"
    echo "  $BINARY_NAME list"
    echo "  $BINARY_NAME remove example.com"
    echo "  $BINARY_NAME --help"
    echo ""
    print_status "For more information, run: $BINARY_NAME --help"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"
}

# Main installation process
main() {
    print_status "Starting vhost-manager installation..."
    
    # Check if not running as root
    check_root
    
    # Check prerequisites
    check_prerequisites
    
    # Build binary
    build_binary
    
    # Install binary
    install_binary
    
    # Verify installation
    verify_installation
    
    # Show usage information
    show_usage
    
    # Cleanup
    cleanup
    
    print_status "Installation completed successfully!"
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"