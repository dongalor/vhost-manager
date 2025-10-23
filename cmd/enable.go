package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// enableCmd represents the enable command
var enableCmd = &cobra.Command{
	Use:   "enable <domain>",
	Short: "Enable a virtual host",
	Long: `Enable a virtual host by creating a symlink in sites-enabled.
The configuration file must already exist in sites-available.`,
	Args: cobra.ExactArgs(1),
	RunE: runEnable,
}

func init() {
	rootCmd.AddCommand(enableCmd)
}

func runEnable(cmd *cobra.Command, args []string) error {
	domain := args[0]
	dryRun := viper.GetBool("dry-run")
	nginxPath := viper.GetString("nginx.path")

	// Create nginx manager
	manager := nginx.NewManager(nginxPath)

	// Check if config file exists in sites-available
	exists, err := manager.VirtualHostExists(domain)
	if err != nil {
		return fmt.Errorf("failed to check if virtual host exists: %w", err)
	}
	if !exists {
		fmt.Printf("Configuration file for domain '%s' does not exist in sites-available\n", domain)
		fmt.Printf("Use 'vhost-manager add %s' to create a new virtual host\n", domain)
		return nil
	}

	// Check if already enabled
	enabled, err := manager.IsVirtualHostEnabled(domain)
	if err != nil {
		return fmt.Errorf("failed to check if virtual host is enabled: %w", err)
	}
	if enabled {
		fmt.Printf("Virtual host for domain '%s' is already enabled\n", domain)
		return nil
	}

	if dryRun {
		fmt.Printf("DRY RUN: Would enable virtual host for domain '%s'\n", domain)
		fmt.Printf("  Would create symlink: %s -> %s\n",
			filepath.Join(nginxPath, "sites-enabled", domain),
			filepath.Join(nginxPath, "sites-available", domain))
		return nil
	}

	// Enable the virtual host
	if err := manager.EnableVirtualHost(domain); err != nil {
		return fmt.Errorf("failed to enable virtual host: %w", err)
	}

	fmt.Printf("Successfully enabled virtual host for domain '%s'\n", domain)
	fmt.Printf("Symlink created: %s -> %s\n",
		filepath.Join(nginxPath, "sites-enabled", domain),
		filepath.Join(nginxPath, "sites-available", domain))

	// Test nginx configuration
	if err := manager.TestConfiguration(); err != nil {
		fmt.Printf("Warning: nginx configuration test failed: %v\n", err)
		fmt.Println("Please check the configuration and run 'nginx -t' manually")
	} else {
		fmt.Println("nginx configuration test passed")
		fmt.Println("Run 'sudo systemctl reload nginx' to apply changes")
	}

	return nil
}
