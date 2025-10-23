package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dongalor/vhost-manager/internal/nginx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all virtual hosts",
	Long: `List all virtual hosts configured in nginx.
Shows both available and enabled virtual hosts.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Add specific flags for the list command
	listCmd.Flags().Bool("enabled-only", false, "show only enabled virtual hosts")
	listCmd.Flags().Bool("available-only", false, "show only available virtual hosts")
	listCmd.Flags().String("format", "table", "output format (table, json)")
}

func runList(cmd *cobra.Command, args []string) error {
	enabledOnly, _ := cmd.Flags().GetBool("enabled-only")
	availableOnly, _ := cmd.Flags().GetBool("available-only")
	format, _ := cmd.Flags().GetString("format")
	nginxPath := viper.GetString("nginx.path")

	// Create nginx manager
	manager := nginx.NewManager(nginxPath)

	// Get virtual hosts
	available, err := manager.ListAvailableVirtualHosts()
	if err != nil {
		return fmt.Errorf("failed to list available virtual hosts: %w", err)
	}

	enabled, err := manager.ListEnabledVirtualHosts()
	if err != nil {
		return fmt.Errorf("failed to list enabled virtual hosts: %w", err)
	}

	// Filter based on flags
	var domains []string
	if enabledOnly {
		domains = enabled
	} else if availableOnly {
		domains = available
	} else {
		// Combine and deduplicate
		domainMap := make(map[string]bool)
		for _, domain := range available {
			domainMap[domain] = true
		}
		for _, domain := range enabled {
			domainMap[domain] = true
		}
		for domain := range domainMap {
			domains = append(domains, domain)
		}
	}

	sort.Strings(domains)

	if format == "json" {
		return listJSON(domains, available, enabled)
	}

	return listTable(domains, available, enabled)
}

func listTable(domains, available, enabled []string) error {
	if len(domains) == 0 {
		fmt.Println("No virtual hosts found")
		return nil
	}

	fmt.Printf("%-30s %-10s %-10s\n", "DOMAIN", "AVAILABLE", "ENABLED")
	fmt.Println(strings.Repeat("-", 50))

	for _, domain := range domains {
		availableStatus := "No"
		enabledStatus := "No"

		if contains(available, domain) {
			availableStatus = "Yes"
		}
		if contains(enabled, domain) {
			enabledStatus = "Yes"
		}

		fmt.Printf("%-30s %-10s %-10s\n", domain, availableStatus, enabledStatus)
	}

	return nil
}

func listJSON(domains, available, enabled []string) error {
	fmt.Println("{")
	fmt.Println("  \"virtual_hosts\": [")
	
	for i, domain := range domains {
		availableStatus := contains(available, domain)
		enabledStatus := contains(enabled, domain)
		
		comma := ","
		if i == len(domains)-1 {
			comma = ""
		}
		
		fmt.Printf("    {\n")
		fmt.Printf("      \"domain\": \"%s\",\n", domain)
		fmt.Printf("      \"available\": %t,\n", availableStatus)
		fmt.Printf("      \"enabled\": %t\n", enabledStatus)
		fmt.Printf("    }%s\n", comma)
	}
	
	fmt.Println("  ]")
	fmt.Println("}")
	
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}