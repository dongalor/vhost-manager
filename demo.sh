#!/bin/bash

# vhost-manager demonstration script

set -e

echo "=== vhost-manager Demo ==="
echo ""

# Create a temporary nginx directory for demo
DEMO_DIR="/tmp/vhost-manager-demo"
NGINX_DIR="$DEMO_DIR/nginx"

echo "Creating demo environment in $DEMO_DIR..."
mkdir -p "$NGINX_DIR/sites-available"
mkdir -p "$NGINX_DIR/sites-enabled"

# Build the application
echo "Building vhost-manager..."
go build -o vhost-manager .

echo ""
echo "=== Demo 1: Add a virtual host (dry run) ==="
./vhost-manager add vhost.example.com --nginx-path "$NGINX_DIR" --dry-run

echo ""
echo "=== Demo 2: Add a virtual host (actual) ==="
./vhost-manager add vhost.example.com --nginx-path "$NGINX_DIR"

echo ""
echo "=== Demo 3: List virtual hosts ==="
./vhost-manager list --nginx-path "$NGINX_DIR"

echo ""
echo "=== Demo 4: Add another virtual host with custom options ==="
./vhost-manager add example.com \
  --nginx-path "$NGINX_DIR" \
  --document-root "/var/www/example.com" \
  --port 8080 \
  --template "redirect-https"

echo ""
echo "=== Demo 5: List virtual hosts (JSON format) ==="
./vhost-manager list --nginx-path "$NGINX_DIR" --format json

echo ""
echo "=== Demo 6: Show generated nginx configuration ==="
echo "Configuration for vhost.example.com:"
echo "----------------------------------------"
cat "$NGINX_DIR/sites-available/vhost.example.com"

echo ""
echo "=== Demo 7: Remove a virtual host ==="
./vhost-manager remove example.com --nginx-path "$NGINX_DIR"

echo ""
echo "=== Demo 8: Final list ==="
./vhost-manager list --nginx-path "$NGINX_DIR"

echo ""
echo "=== Demo completed! ==="
echo "Generated files:"
echo "- sites-available: $(ls -la "$NGINX_DIR/sites-available")"
echo "- sites-enabled: $(ls -la "$NGINX_DIR/sites-enabled")"

# Cleanup
echo ""
echo "Cleaning up demo environment..."
rm -rf "$DEMO_DIR"
rm -f vhost-manager

echo "Demo completed successfully!"
