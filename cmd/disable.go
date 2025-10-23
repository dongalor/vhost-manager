package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// disableCmd represents the disable command
var disableCmd = &cobra.Command{
	Use:   "disable <domain>",
	Short: "Disable a virtual host",
	Long: `Disable a virtual host by removing the symlink from sites-enabled.
The configuration file in sites-available is preserved.`,
	Args: cobra.ExactArgs(1),
	RunE: runDisable,
}

func init() {
	rootCmd.AddCommand(disableCmd)
}

func runDisable(cmd *cobra.Command, args []string) error {
	domain := args[0]
	dryRun := viper.GetBool("dry-run")
	nginxPath := viper.GetString("nginx.path")

	// Create nginx manager
	manager := nginx.NewManager(nginxPath)

	// Check if virtual host is enabled
	enabled, err := manager.IsVirtualHostEnabled(domain)
	if err != nil {
		return fmt.Errorf("failed to check if virtual host is enabled: %w", err)
	}
	if !enabled {
		fmt.Printf("Virtual host for domain '%s' is already disabled\n", domain)
		return nil
	}

	if dryRun {
		fmt.Printf("DRY RUN: Would disable virtual host for domain '%s'\n", domain)
		fmt.Printf("  Would remove symlink: %s\n", filepath.Join(nginxPath, "sites-enabled", domain))
		fmt.Printf("  Configuration file will be preserved: %s\n", filepath.Join(nginxPath, "sites-available", domain))
		return nil
	}

	// Disable the virtual host
	if err := manager.DisableVirtualHost(domain); err != nil {
		return fmt.Errorf("failed to disable virtual host: %w", err)
	}

	fmt.Printf("Successfully disabled virtual host for domain '%s'\n", domain)
	fmt.Printf("Symlink removed: %s\n", filepath.Join(nginxPath, "sites-enabled", domain))
	fmt.Printf("Configuration file preserved: %s\n", filepath.Join(nginxPath, "sites-available", domain))

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