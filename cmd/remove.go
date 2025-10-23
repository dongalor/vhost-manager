package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove <domain>",
	Short: "Remove a virtual host",
	Long: `Remove a virtual host for the specified domain.
This will remove the symlink from sites-enabled and optionally
delete the configuration file from sites-available.`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Add specific flags for the remove command
	removeCmd.Flags().Bool("delete-config", false, "also delete the configuration file from sites-available")
	removeCmd.Flags().Bool("force", false, "force removal without confirmation")
}

func runRemove(cmd *cobra.Command, args []string) error {
	domain := args[0]
	deleteConfig, _ := cmd.Flags().GetBool("delete-config")
	force, _ := cmd.Flags().GetBool("force")
	dryRun := viper.GetBool("dry-run")
	nginxPath := viper.GetString("nginx.path")

	// Create nginx manager
	manager := nginx.NewManager(nginxPath)

	// Check if virtual host exists
	exists, err := manager.VirtualHostExists(domain)
	if err != nil {
		return fmt.Errorf("failed to check if virtual host exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("virtual host for domain '%s' does not exist", domain)
	}

	if dryRun {
		fmt.Printf("DRY RUN: Would remove virtual host for domain '%s'\n", domain)
		if deleteConfig {
			fmt.Printf("  Would delete configuration file: %s\n", filepath.Join(nginxPath, "sites-available", domain))
		}
		fmt.Printf("  Would remove symlink: %s\n", filepath.Join(nginxPath, "sites-enabled", domain))
		return nil
	}

	// Confirm removal unless forced
	if !force {
		fmt.Printf("Are you sure you want to remove virtual host for domain '%s'? (y/N): ", domain)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Remove the virtual host
	if err := manager.RemoveVirtualHost(domain, deleteConfig); err != nil {
		return fmt.Errorf("failed to remove virtual host: %w", err)
	}

	fmt.Printf("Successfully removed virtual host for domain '%s'\n", domain)
	if deleteConfig {
		fmt.Printf("Configuration file deleted: %s\n", filepath.Join(nginxPath, "sites-available", domain))
	}
	fmt.Printf("Symlink removed: %s\n", filepath.Join(nginxPath, "sites-enabled", domain))

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