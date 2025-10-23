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
This will remove both the symlink from sites-enabled and the
configuration file from sites-available.

Use 'disable' command if you only want to deactivate the virtual host
without deleting the configuration file.`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Add specific flags for the remove command
	removeCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
}

func runRemove(cmd *cobra.Command, args []string) error {
	domain := args[0]
	skipConfirmation, _ := cmd.Flags().GetBool("yes")
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
		fmt.Printf("  Would delete configuration file: %s\n", filepath.Join(nginxPath, "sites-available", domain))
		fmt.Printf("  Would remove symlink: %s\n", filepath.Join(nginxPath, "sites-enabled", domain))
		return nil
	}

	// Confirm removal unless skipped with -y flag
	if !skipConfirmation {
		fmt.Printf("This will permanently delete the virtual host for domain '%s'\n", domain)
		fmt.Printf("Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Remove the virtual host (always delete both symlink and config)
	if err := manager.RemoveVirtualHost(domain, true); err != nil {
		return fmt.Errorf("failed to remove virtual host: %w", err)
	}

	fmt.Printf("Successfully removed virtual host for domain '%s'\n", domain)
	fmt.Printf("Configuration file deleted: %s\n", filepath.Join(nginxPath, "sites-available", domain))
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