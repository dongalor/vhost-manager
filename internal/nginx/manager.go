package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Manager handles nginx virtual host operations
type Manager struct {
	nginxPath string
}

// VirtualHostConfig represents the configuration for a virtual host
type VirtualHostConfig struct {
	Domain        string
	ServerName    string
	DocumentRoot  string
	Port          int
	RedirectHTTPS bool
	Template      string
}

// NewManager creates a new nginx manager
func NewManager(nginxPath string) *Manager {
	return &Manager{
		nginxPath: nginxPath,
	}
}

// VirtualHostExists checks if a virtual host exists in sites-available
func (m *Manager) VirtualHostExists(domain string) (bool, error) {
	configPath := filepath.Join(m.nginxPath, "sites-available", domain)
	_, err := os.Stat(configPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateVirtualHost creates a new virtual host
func (m *Manager) CreateVirtualHost(config VirtualHostConfig) error {
	// Create sites-available directory if it doesn't exist
	sitesAvailableDir := filepath.Join(m.nginxPath, "sites-available")
	if err := os.MkdirAll(sitesAvailableDir, 0755); err != nil {
		return fmt.Errorf("failed to create sites-available directory: %w", err)
	}

	// Create sites-enabled directory if it doesn't exist
	sitesEnabledDir := filepath.Join(m.nginxPath, "sites-enabled")
	if err := os.MkdirAll(sitesEnabledDir, 0755); err != nil {
		return fmt.Errorf("failed to create sites-enabled directory: %w", err)
	}

	// Generate nginx configuration
	configContent, err := m.generateConfig(config)
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Write configuration file
	configPath := filepath.Join(sitesAvailableDir, config.Domain)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	// Create symlink in sites-enabled
	symlinkPath := filepath.Join(sitesEnabledDir, config.Domain)
	if err := os.Symlink(configPath, symlinkPath); err != nil {
		// Clean up the config file if symlink creation fails
		os.Remove(configPath)
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// RemoveVirtualHost removes a virtual host
func (m *Manager) RemoveVirtualHost(domain string, deleteConfig bool) error {
	// Remove symlink from sites-enabled
	symlinkPath := filepath.Join(m.nginxPath, "sites-enabled", domain)
	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	// Optionally delete configuration file from sites-available
	if deleteConfig {
		configPath := filepath.Join(m.nginxPath, "sites-available", domain)
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove configuration file: %w", err)
		}
	}

	return nil
}

// ListAvailableVirtualHosts lists all virtual hosts in sites-available
func (m *Manager) ListAvailableVirtualHosts() ([]string, error) {
	sitesAvailableDir := filepath.Join(m.nginxPath, "sites-available")
	entries, err := os.ReadDir(sitesAvailableDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var domains []string
	for _, entry := range entries {
		if !entry.IsDir() {
			domains = append(domains, entry.Name())
		}
	}

	return domains, nil
}

// ListEnabledVirtualHosts lists all virtual hosts in sites-enabled
func (m *Manager) ListEnabledVirtualHosts() ([]string, error) {
	sitesEnabledDir := filepath.Join(m.nginxPath, "sites-enabled")
	entries, err := os.ReadDir(sitesEnabledDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var domains []string
	for _, entry := range entries {
		if !entry.IsDir() {
			// Check if it's a symlink
			symlinkPath := filepath.Join(sitesEnabledDir, entry.Name())
			if _, err := os.Readlink(symlinkPath); err == nil {
				domains = append(domains, entry.Name())
			}
		}
	}

	return domains, nil
}

// TestConfiguration tests the nginx configuration
func (m *Manager) TestConfiguration() error {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx configuration test failed: %s", string(output))
	}
	return nil
}

// generateConfig generates nginx configuration content
func (m *Manager) generateConfig(config VirtualHostConfig) (string, error) {
	tmpl := getTemplate(config.Template)
	
	t, err := template.New("nginx").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// getTemplate returns the nginx template based on the template name
func getTemplate(templateName string) string {
	switch templateName {
	case "default":
		return defaultTemplate
	case "redirect-https":
		return redirectHTTPSTemplate
	case "basic":
		return basicTemplate
	default:
		return defaultTemplate
	}
}