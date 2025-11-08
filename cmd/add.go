package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dongalor/vhost-manager/internal/certbot"
	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a new virtual host",
	Long: `Add a new virtual host for the specified domain.
This will create a configuration file in sites-available and create
a symlink in sites-enabled.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Add specific flags for the add command
	addCmd.Flags().String("document-root", "", "document root directory (default: /var/www/<domain>)")
	addCmd.Flags().Int("port", 80, "port number for the virtual host")
	addCmd.Flags().String("server-name", "", "server name (default: <domain>)")
	addCmd.Flags().Bool("redirect-https", true, "redirect HTTP to HTTPS")
	addCmd.Flags().String("template", "default", "nginx template to use")
	addCmd.Flags().Bool("cert", false, "install SSL certificate with certbot after creating vhost")
	addCmd.Flags().String("email", "", "email address for Let's Encrypt notifications (used with --cert)")
	addCmd.Flags().Bool("staging", false, "use Let's Encrypt staging environment (used with --cert)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	domain := args[0]
	
	// Validate domain name
	if !isValidDomain(domain) {
		return fmt.Errorf("invalid domain name: %s", domain)
	}

	// Get configuration
	nginxPath := viper.GetString("nginx.path")
	dryRun := viper.GetBool("dry-run")
	documentRoot, _ := cmd.Flags().GetString("document-root")
	port, _ := cmd.Flags().GetInt("port")
	serverName, _ := cmd.Flags().GetString("server-name")
	redirectHTTPS, _ := cmd.Flags().GetBool("redirect-https")
	template, _ := cmd.Flags().GetString("template")

	// Set defaults
	if documentRoot == "" {
		documentRoot = fmt.Sprintf("/var/www/%s", domain)
	}
	if serverName == "" {
		serverName = domain
	}

	// Create nginx manager
	manager := nginx.NewManager(nginxPath)

	// Check if virtual host already exists
	exists, err := manager.VirtualHostExists(domain)
	if err != nil {
		return fmt.Errorf("failed to check if virtual host exists: %w", err)
	}
	if exists {
		return fmt.Errorf("virtual host for domain '%s' already exists", domain)
	}

	// Create virtual host configuration
	config := nginx.VirtualHostConfig{
		Domain:        domain,
		ServerName:    serverName,
		DocumentRoot:  documentRoot,
		Port:          port,
		RedirectHTTPS: redirectHTTPS,
		Template:      template,
	}

	if dryRun {
		fmt.Printf("DRY RUN: Would create virtual host for domain '%s'\n", domain)
		fmt.Printf("  Document root: %s\n", documentRoot)
		fmt.Printf("  Port: %d\n", port)
		fmt.Printf("  Server name: %s\n", serverName)
		fmt.Printf("  Redirect to HTTPS: %t\n", redirectHTTPS)
		fmt.Printf("  Template: %s\n", template)
		return nil
	}

	// Create the virtual host
	if err := manager.CreateVirtualHost(config); err != nil {
		return fmt.Errorf("failed to create virtual host: %w", err)
	}

	fmt.Printf("Successfully created virtual host for domain '%s'\n", domain)
	fmt.Printf("Configuration file: %s\n", filepath.Join(nginxPath, "sites-available", domain))
	fmt.Printf("Symlink created: %s\n", filepath.Join(nginxPath, "sites-enabled", domain))
	fmt.Printf("Document root: %s\n", documentRoot)

	// Test nginx configuration
	if err := manager.TestConfiguration(); err != nil {
		fmt.Printf("Warning: nginx configuration test failed: %v\n", err)
		fmt.Println("Please check the configuration and run 'nginx -t' manually")
		return nil
	}

	fmt.Println("nginx configuration test passed")

	// Check if SSL certificate should be installed
	installCert, _ := cmd.Flags().GetBool("cert")
	if installCert {
		email, _ := cmd.Flags().GetString("email")
		staging, _ := cmd.Flags().GetBool("staging")

		// Check for email in config if not provided
		if email == "" {
			email = viper.GetString("certbot.email")
		}

		// Check if certbot is installed
		if err := certbot.CheckInstalled(); err != nil {
			fmt.Printf("\nWarning: Cannot install SSL certificate: %v\n", err)
			fmt.Println("Virtual host created successfully, but SSL certificate was not installed.")
			fmt.Printf("You can install it later with: vhost-manager cert install %s\n", domain)
			return nil
		}

		fmt.Println("\nInstalling SSL certificate...")

		certbotMgr, err := certbot.NewManager(email, staging, false)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize certbot: %v\n", err)
			fmt.Printf("You can install SSL later with: vhost-manager cert install %s\n", domain)
			return nil
		}

		// Install certificate
		if err := certbotMgr.InstallCertificate(domain, true); err != nil {
			fmt.Printf("Warning: Failed to install certificate: %v\n", err)
			fmt.Printf("You can try again with: vhost-manager cert install %s\n", domain)
			return nil
		}

		// Update nginx configuration to enable SSL
		fmt.Println("Updating nginx configuration to enable SSL...")
		if err := manager.EnableSSL(domain); err != nil {
			fmt.Printf("Warning: Failed to enable SSL in nginx configuration: %v\n", err)
			return nil
		}

		// Test nginx configuration again
		if err := manager.TestConfiguration(); err != nil {
			fmt.Printf("Warning: nginx configuration test failed after enabling SSL: %v\n", err)
			fmt.Println("Configuration backup available at: " + manager.GetConfigPath(domain) + ".backup")
			return nil
		}

		// Reload nginx
		fmt.Println("Reloading nginx...")
		if err := manager.ReloadNginx(); err != nil {
			fmt.Printf("Warning: failed to reload nginx: %v\n", err)
			fmt.Println("Please reload nginx manually with: sudo systemctl reload nginx")
		} else {
			fmt.Println("✓ nginx reloaded successfully")
		}

		fmt.Printf("\n✓ Virtual host created with SSL certificate for %s\n", domain)
		fmt.Printf("Your site is now available at https://%s\n", domain)
	} else {
		fmt.Println("Run 'sudo systemctl reload nginx' to apply changes")
		fmt.Printf("To add SSL certificate later, run: vhost-manager cert install %s\n", domain)
	}

	return nil
}

func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}
	
	// Basic domain validation
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	
	// Check each part
	for _, part := range parts {
		if part == "" || len(part) > 63 {
			return false
		}
		// Check for valid characters (simplified)
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || 
				 (char >= 'A' && char <= 'Z') || 
				 (char >= '0' && char <= '9') || 
				 char == '-') {
				return false
			}
		}
	}
	
	return true
}