package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dongalor/vhost-manager/internal/certbot"
	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// certCmd represents the cert command
var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage SSL certificates with certbot",
	Long: `Manage SSL certificates for your virtual hosts using Let's Encrypt and certbot.
This command provides subcommands to install, renew, check status, and remove certificates.`,
}

// certInstallCmd represents the cert install command
var certInstallCmd = &cobra.Command{
	Use:   "install <domain>",
	Short: "Install an SSL certificate for a domain",
	Long: `Install an SSL certificate for the specified domain using Let's Encrypt.
This will obtain a certificate via certbot and update the nginx configuration to use SSL.`,
	Args: cobra.ExactArgs(1),
	RunE: runCertInstall,
}

// certRenewCmd represents the cert renew command
var certRenewCmd = &cobra.Command{
	Use:   "renew [domain]",
	Short: "Renew SSL certificate(s)",
	Long: `Renew SSL certificates. If a domain is specified, renew only that certificate.
If no domain is specified with --all flag, renew all certificates.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCertRenew,
}

// certStatusCmd represents the cert status command
var certStatusCmd = &cobra.Command{
	Use:   "status [domain]",
	Short: "Check SSL certificate status",
	Long: `Check the status of SSL certificates. If a domain is specified, show details for that certificate.
If no domain is specified, list all certificates.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCertStatus,
}

// certRemoveCmd represents the cert remove command
var certRemoveCmd = &cobra.Command{
	Use:   "remove <domain>",
	Short: "Remove an SSL certificate",
	Long: `Remove the SSL certificate for the specified domain.
This will delete the certificate from Let's Encrypt storage.`,
	Args: cobra.ExactArgs(1),
	RunE: runCertRemove,
}

func init() {
	rootCmd.AddCommand(certCmd)

	// Add subcommands
	certCmd.AddCommand(certInstallCmd)
	certCmd.AddCommand(certRenewCmd)
	certCmd.AddCommand(certStatusCmd)
	certCmd.AddCommand(certRemoveCmd)

	// Flags for cert install
	certInstallCmd.Flags().String("email", "", "email address for Let's Encrypt notifications")
	certInstallCmd.Flags().Bool("staging", false, "use Let's Encrypt staging environment for testing")
	certInstallCmd.Flags().Bool("nginx-plugin", true, "use certbot nginx plugin (vs webroot)")

	// Flags for cert renew
	certRenewCmd.Flags().Bool("all", false, "renew all certificates")
	certRenewCmd.Flags().Bool("staging", false, "use Let's Encrypt staging environment")

	// Flags for cert remove
	certRemoveCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
}

func runCertInstall(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Get flags
	email, _ := cmd.Flags().GetString("email")
	staging, _ := cmd.Flags().GetBool("staging")
	nginxPlugin, _ := cmd.Flags().GetBool("nginx-plugin")
	dryRun := viper.GetBool("dry-run")
	nginxPath := viper.GetString("nginx.path")

	// Check for email in config if not provided
	if email == "" {
		email = viper.GetString("certbot.email")
	}

	// Check if certbot is installed
	if err := certbot.CheckInstalled(); err != nil {
		return err
	}

	// Create managers
	nginxMgr := nginx.NewManager(nginxPath)
	certbotMgr, err := certbot.NewManager(email, staging, dryRun)
	if err != nil {
		return err
	}

	// Check if virtual host exists
	exists, err := nginxMgr.VirtualHostExists(domain)
	if err != nil {
		return fmt.Errorf("failed to check if virtual host exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("virtual host for domain '%s' does not exist. Create it first with 'vhost-manager add %s'", domain, domain)
	}

	// Check if certificate already exists
	if certbot.CertificateExists(domain) {
		fmt.Printf("Certificate already exists for %s\n", domain)
		fmt.Println("Use 'vhost-manager cert renew' to renew it")
		return nil
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would install SSL certificate for domain '%s'\n", domain)
		if email != "" {
			fmt.Printf("  Email: %s\n", email)
		}
		fmt.Printf("  Staging: %t\n", staging)
		fmt.Printf("  Nginx plugin: %t\n", nginxPlugin)
		return nil
	}

	fmt.Printf("Installing SSL certificate for %s...\n", domain)

	// Install certificate using certbot
	if err := certbotMgr.InstallCertificate(domain, nginxPlugin); err != nil {
		return fmt.Errorf("failed to install certificate: %w", err)
	}

	// Update nginx configuration to enable SSL
	fmt.Println("Updating nginx configuration to enable SSL...")
	if err := nginxMgr.EnableSSL(domain); err != nil {
		return fmt.Errorf("failed to enable SSL in nginx configuration: %w", err)
	}

	// Test nginx configuration
	if err := nginxMgr.TestConfiguration(); err != nil {
		fmt.Printf("Warning: nginx configuration test failed: %v\n", err)
		fmt.Println("Configuration backup available at: " + nginxMgr.GetConfigPath(domain) + ".backup")
		return fmt.Errorf("nginx configuration test failed")
	}

	// Reload nginx
	fmt.Println("Reloading nginx...")
	if err := nginxMgr.ReloadNginx(); err != nil {
		fmt.Printf("Warning: failed to reload nginx: %v\n", err)
		fmt.Println("Please reload nginx manually with: sudo systemctl reload nginx")
	} else {
		fmt.Println("✓ nginx reloaded successfully")
	}

	fmt.Printf("\n✓ SSL certificate successfully installed for %s\n", domain)
	fmt.Printf("Your site is now available at https://%s\n", domain)

	return nil
}

func runCertRenew(cmd *cobra.Command, args []string) error {
	renewAll, _ := cmd.Flags().GetBool("all")
	staging, _ := cmd.Flags().GetBool("staging")
	dryRun := viper.GetBool("dry-run")
	nginxPath := viper.GetString("nginx.path")
	email := viper.GetString("certbot.email")

	// Check if certbot is installed
	if err := certbot.CheckInstalled(); err != nil {
		return err
	}

	certbotMgr, err := certbot.NewManager(email, staging, dryRun)
	if err != nil {
		return err
	}

	nginxMgr := nginx.NewManager(nginxPath)

	if renewAll {
		fmt.Println("Renewing all certificates...")
		if err := certbotMgr.RenewAll(); err != nil {
			return err
		}
	} else if len(args) > 0 {
		domain := args[0]
		fmt.Printf("Renewing certificate for %s...\n", domain)
		if err := certbotMgr.RenewCertificate(domain); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("either specify a domain or use --all flag to renew all certificates")
	}

	if !dryRun {
		// Reload nginx to apply renewed certificates
		fmt.Println("Reloading nginx...")
		if err := nginxMgr.ReloadNginx(); err != nil {
			fmt.Printf("Warning: failed to reload nginx: %v\n", err)
			fmt.Println("Please reload nginx manually with: sudo systemctl reload nginx")
		} else {
			fmt.Println("✓ nginx reloaded successfully")
		}
	}

	return nil
}

func runCertStatus(cmd *cobra.Command, args []string) error {
	email := viper.GetString("certbot.email")
	dryRun := viper.GetBool("dry-run")

	// Check if certbot is installed
	if err := certbot.CheckInstalled(); err != nil {
		return err
	}

	certbotMgr, err := certbot.NewManager(email, false, dryRun)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		// Show status for specific domain
		domain := args[0]
		info, err := certbotMgr.GetCertificateInfo(domain)
		if err != nil {
			return fmt.Errorf("failed to get certificate info: %w", err)
		}

		fmt.Println(info.String())
	} else {
		// List all certificates
		certificates, err := certbotMgr.ListCertificates()
		if err != nil {
			return fmt.Errorf("failed to list certificates: %w", err)
		}

		if len(certificates) == 0 {
			fmt.Println("No certificates found")
			return nil
		}

		fmt.Printf("Found %d certificate(s):\n\n", len(certificates))

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tEXPIRY DATE\tDAYS LEFT\tSTATUS")
		fmt.Fprintln(w, "------\t-----------\t---------\t------")

		for _, cert := range certificates {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
				cert.Domain,
				cert.ExpiryDate.Format("2006-01-02"),
				cert.DaysLeft,
				cert.Status())
		}

		w.Flush()
	}

	return nil
}

func runCertRemove(cmd *cobra.Command, args []string) error {
	domain := args[0]
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	dryRun := viper.GetBool("dry-run")
	email := viper.GetString("certbot.email")

	// Check if certbot is installed
	if err := certbot.CheckInstalled(); err != nil {
		return err
	}

	certbotMgr, err := certbot.NewManager(email, false, dryRun)
	if err != nil {
		return err
	}

	// Check if certificate exists
	if !certbot.CertificateExists(domain) {
		return fmt.Errorf("certificate for domain '%s' does not exist", domain)
	}

	// Confirm deletion
	if !skipConfirm && !dryRun {
		fmt.Printf("Are you sure you want to remove the certificate for %s? (y/N): ", domain)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would remove certificate for domain '%s'\n", domain)
		return nil
	}

	// Remove certificate
	if err := certbotMgr.RemoveCertificate(domain); err != nil {
		return err
	}

	fmt.Printf("\nNote: This only removed the certificate. The nginx configuration still has SSL enabled.\n")
	fmt.Printf("You may want to update the nginx configuration for %s to disable SSL.\n", domain)

	return nil
}
