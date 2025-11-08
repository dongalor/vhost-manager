# vhost-manager

A command-line tool for managing nginx virtual hosts. Easily add, remove, and list virtual hosts with proper nginx configuration templates.

## Features

- **Add virtual hosts**: Create nginx configuration files and symlinks
- **Remove virtual hosts**: Clean up virtual host configurations
- **Enable/Disable virtual hosts**: Activate or deactivate sites without deleting configurations
- **List virtual hosts**: View all available and enabled virtual hosts
- **SSL certificate management**: Automated Let's Encrypt certificate installation and renewal with certbot
- **Multiple templates**: Choose from different nginx configuration templates
- **Configuration validation**: Test nginx configuration after changes
- **Nginx reload**: Gracefully reload nginx to apply changes
- **Dry run mode**: Preview changes without making them
- **JSON output**: Machine-readable output for scripting

## Installation

### From source

```bash
# Clone the repository
git clone https://github.com/dongalor/vhost-manager.git
cd vhost-manager

# Build and install
make install
```

### Manual installation

```bash
# Build the binary
go build -o vhost-manager .

# Install to system
sudo cp vhost-manager /usr/local/bin/
sudo chmod +x /usr/local/bin/vhost-manager
```

## Usage

### Add a virtual host

```bash
# Basic usage
vhost-manager add example.com

# With custom options
vhost-manager add example.com \
  --document-root /var/www/example.com \
  --port 80 \
  --redirect-https \
  --template redirect-https
```

### Remove a virtual host

```bash
# Remove symlink only
vhost-manager remove example.com

# Remove symlink and configuration file
vhost-manager remove example.com --delete-config

# Force removal without confirmation
vhost-manager remove example.com --force
```

### List virtual hosts

```bash
# List all virtual hosts
vhost-manager list

# List only enabled virtual hosts
vhost-manager list --enabled-only

# List only available virtual hosts
vhost-manager list --available-only

# JSON output
vhost-manager list --format json
```

### Enable/Disable virtual hosts

```bash
# Enable a virtual host (create symlink in sites-enabled)
vhost-manager enable example.com

# Disable a virtual host (remove symlink, keep config)
vhost-manager disable example.com
```

### SSL Certificate Management

#### Install SSL certificate

```bash
# Install certificate for an existing virtual host
vhost-manager cert install example.com

# Install with specific email
vhost-manager cert install example.com --email admin@example.com

# Use staging environment for testing
vhost-manager cert install example.com --staging

# Create virtual host with SSL certificate in one command
vhost-manager add example.com --cert --email admin@example.com
```

#### Renew certificates

```bash
# Renew certificate for a specific domain
vhost-manager cert renew example.com

# Renew all certificates
vhost-manager cert renew --all
```

#### Check certificate status

```bash
# Check status of a specific certificate
vhost-manager cert status example.com

# List all certificates
vhost-manager cert status
```

#### Remove certificate

```bash
# Remove a certificate
vhost-manager cert remove example.com

# Skip confirmation
vhost-manager cert remove example.com -y
```

### Reload nginx

```bash
# Reload nginx configuration
vhost-manager reload
```

### Global options

```bash
# Dry run mode
vhost-manager add example.com --dry-run

# Custom nginx path
vhost-manager add example.com --nginx-path /etc/nginx

# Configuration file
vhost-manager --config /path/to/config.yaml add example.com
```

## Configuration

Create a configuration file at `~/.vhost-manager.yaml`:

```yaml
nginx:
  path: /etc/nginx

certbot:
  email: admin@example.com  # Email for Let's Encrypt notifications

dry-run: false
```

## Templates

The tool includes several nginx configuration templates:

- **default**: Standard configuration with security headers and optimizations
- **redirect-https**: HTTP to HTTPS redirect with SSL configuration placeholders (commented)
- **ssl**: Full SSL configuration with Let's Encrypt certificates (auto-applied by cert install)
- **basic**: Minimal configuration for simple sites

## Examples

### Setting up a new website with SSL

```bash
# Option 1: One-command setup (recommended)
sudo vhost-manager add mywebsite.com --cert --email admin@mywebsite.com

# Option 2: Step-by-step setup
# 1. Add the virtual host
sudo vhost-manager add mywebsite.com

# 2. Create the document root
sudo mkdir -p /var/www/mywebsite.com
sudo chown -R $USER:$USER /var/www/mywebsite.com

# 3. Add your website files
echo "<h1>Welcome to mywebsite.com</h1>" > /var/www/mywebsite.com/index.html

# 4. Install SSL certificate
sudo vhost-manager cert install mywebsite.com --email admin@mywebsite.com

# The tool automatically:
# - Obtains certificate from Let's Encrypt via certbot
# - Updates nginx configuration to enable SSL
# - Tests nginx configuration
# - Reloads nginx

# Your site is now live at https://mywebsite.com
```

### Setting up a website without SSL

```bash
# 1. Add the virtual host
sudo vhost-manager add mywebsite.com

# 2. Create the document root and add files
sudo mkdir -p /var/www/mywebsite.com
echo "<h1>Welcome</h1>" > /var/www/mywebsite.com/index.html

# 3. Reload nginx
sudo vhost-manager reload

# Site is now live at http://mywebsite.com
```

### Removing a website

```bash
# Remove the virtual host
vhost-manager remove mywebsite.com --delete-config

# Remove the document root (optional)
sudo rm -rf /var/www/mywebsite.com

# Reload nginx
sudo systemctl reload nginx
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Building

```bash
# Download dependencies
make deps

# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

### Project structure

```
vhost-manager/
├── cmd/                        # CLI commands
│   ├── root.go                # Root command and global flags
│   ├── add.go                 # Add command
│   ├── remove.go              # Remove command
│   ├── list.go                # List command
│   ├── enable.go              # Enable command
│   ├── disable.go             # Disable command
│   ├── cert.go                # Certificate management commands
│   ├── reload.go              # Nginx reload command
│   └── version.go             # Version command
├── internal/
│   ├── nginx/                 # Nginx management package
│   │   ├── manager.go         # Core nginx operations
│   │   └── templates.go       # Nginx configuration templates
│   └── certbot/               # Certbot/SSL management package
│       ├── manager.go         # Certbot operations
│       └── certificate.go     # Certificate info types
├── main.go                    # Application entry point
├── go.mod                     # Go module file
├── Makefile                   # Build automation
└── README.md                  # This file
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Troubleshooting

### Permission denied errors

The tool needs to write to `/etc/nginx/sites-available/` and `/etc/nginx/sites-enabled/`. Make sure you have the necessary permissions:

```bash
# Run with sudo
sudo vhost-manager add example.com

# Or add your user to the nginx group (if applicable)
sudo usermod -a -G nginx $USER
```

### Nginx configuration test fails

If nginx configuration test fails after adding a virtual host:

```bash
# Check nginx configuration
sudo nginx -t

# Check the generated configuration file
cat /etc/nginx/sites-available/example.com

# Check for syntax errors
sudo nginx -T
```

### Symlink already exists

If you get an error about symlink already existing:

```bash
# Check if the symlink exists
ls -la /etc/nginx/sites-enabled/example.com

# Remove it manually if needed
sudo rm /etc/nginx/sites-enabled/example.com

# Then try again
vhost-manager add example.com
```

### Certbot not found

If you get an error that certbot is not installed:

```bash
# Install certbot on Ubuntu/Debian
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx

# Install certbot on CentOS/RHEL
sudo yum install certbot python3-certbot-nginx

# Install certbot on other systems
# Visit https://certbot.eff.org/ for instructions
```

### SSL certificate installation fails

If certificate installation fails:

```bash
# Ensure domain points to your server
dig +short example.com

# Check if port 80/443 are accessible
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443

# Try using staging environment first for testing
vhost-manager cert install example.com --staging

# Check certbot logs
sudo tail -f /var/log/letsencrypt/letsencrypt.log
```

### Certificate renewal

Set up automatic renewal with a cron job:

```bash
# Add to crontab
sudo crontab -e

# Add this line to run renewal check twice daily
0 0,12 * * * vhost-manager cert renew --all

# Or use certbot's built-in timer
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```