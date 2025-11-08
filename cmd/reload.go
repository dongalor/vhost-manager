package cmd

import (
	"fmt"

	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reloadCmd represents the reload command
var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload nginx configuration",
	Long: `Reload the nginx configuration to apply changes without downtime.
This runs 'nginx -s reload' to gracefully reload the nginx service.`,
	RunE: runReload,
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}

func runReload(cmd *cobra.Command, args []string) error {
	nginxPath := viper.GetString("nginx.path")
	dryRun := viper.GetBool("dry-run")

	manager := nginx.NewManager(nginxPath)

	if dryRun {
		fmt.Println("[DRY RUN] Would reload nginx")
		return nil
	}

	// Test configuration before reloading
	fmt.Println("Testing nginx configuration...")
	if err := manager.TestConfiguration(); err != nil {
		return fmt.Errorf("nginx configuration test failed: %w\nPlease fix the configuration before reloading", err)
	}

	fmt.Println("✓ Configuration test passed")
	fmt.Println("Reloading nginx...")

	if err := manager.ReloadNginx(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	fmt.Println("✓ nginx reloaded successfully")
	return nil
}
