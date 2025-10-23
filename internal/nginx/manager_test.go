package nginx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_CreateVirtualHost(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	config := VirtualHostConfig{
		Domain:        "test.example.com",
		ServerName:    "test.example.com",
		DocumentRoot:  "/var/www/test.example.com",
		Port:          80,
		RedirectHTTPS: true,
		Template:      "basic",
	}

	// Test creating virtual host
	err := manager.CreateVirtualHost(config)
	if err != nil {
		t.Fatalf("Failed to create virtual host: %v", err)
	}

	// Check if configuration file was created
	configPath := filepath.Join(tempDir, "sites-available", config.Domain)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Configuration file was not created: %s", configPath)
	}

	// Check if symlink was created
	symlinkPath := filepath.Join(tempDir, "sites-enabled", config.Domain)
	if _, err := os.Stat(symlinkPath); os.IsNotExist(err) {
		t.Errorf("Symlink was not created: %s", symlinkPath)
	}

	// Check if symlink points to the correct file
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Errorf("Failed to read symlink: %v", err)
	}
	if target != configPath {
		t.Errorf("Symlink target is incorrect. Expected: %s, Got: %s", configPath, target)
	}
}

func TestManager_VirtualHostExists(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Test non-existent virtual host
	exists, err := manager.VirtualHostExists("nonexistent.com")
	if err != nil {
		t.Fatalf("Failed to check virtual host existence: %v", err)
	}
	if exists {
		t.Error("Virtual host should not exist")
	}

	// Create a virtual host
	config := VirtualHostConfig{
		Domain:       "test.example.com",
		ServerName:   "test.example.com",
		DocumentRoot: "/var/www/test.example.com",
		Port:         80,
		Template:     "basic",
	}

	err = manager.CreateVirtualHost(config)
	if err != nil {
		t.Fatalf("Failed to create virtual host: %v", err)
	}

	// Test existing virtual host
	exists, err = manager.VirtualHostExists("test.example.com")
	if err != nil {
		t.Fatalf("Failed to check virtual host existence: %v", err)
	}
	if !exists {
		t.Error("Virtual host should exist")
	}
}

func TestManager_RemoveVirtualHost(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create a virtual host first
	config := VirtualHostConfig{
		Domain:       "test.example.com",
		ServerName:   "test.example.com",
		DocumentRoot: "/var/www/test.example.com",
		Port:         80,
		Template:     "basic",
	}

	err := manager.CreateVirtualHost(config)
	if err != nil {
		t.Fatalf("Failed to create virtual host: %v", err)
	}

	// Test removing virtual host (symlink only)
	err = manager.RemoveVirtualHost("test.example.com", false)
	if err != nil {
		t.Fatalf("Failed to remove virtual host: %v", err)
	}

	// Check if symlink was removed
	symlinkPath := filepath.Join(tempDir, "sites-enabled", config.Domain)
	if _, err := os.Stat(symlinkPath); !os.IsNotExist(err) {
		t.Error("Symlink should have been removed")
	}

	// Check if configuration file still exists
	configPath := filepath.Join(tempDir, "sites-available", config.Domain)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Configuration file should still exist")
	}

	// Test removing with delete-config
	err = manager.RemoveVirtualHost("test.example.com", true)
	if err != nil {
		t.Fatalf("Failed to remove virtual host with config: %v", err)
	}

	// Check if configuration file was removed
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Configuration file should have been removed")
	}
}

func TestManager_ListVirtualHosts(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create some virtual hosts
	domains := []string{"test1.com", "test2.com", "test3.com"}
	for _, domain := range domains {
		config := VirtualHostConfig{
			Domain:       domain,
			ServerName:   domain,
			DocumentRoot: "/var/www/" + domain,
			Port:         80,
			Template:     "basic",
		}

		err := manager.CreateVirtualHost(config)
		if err != nil {
			t.Fatalf("Failed to create virtual host %s: %v", domain, err)
		}
	}

	// Test listing available virtual hosts
	available, err := manager.ListAvailableVirtualHosts()
	if err != nil {
		t.Fatalf("Failed to list available virtual hosts: %v", err)
	}

	if len(available) != len(domains) {
		t.Errorf("Expected %d available virtual hosts, got %d", len(domains), len(available))
	}

	// Test listing enabled virtual hosts
	enabled, err := manager.ListEnabledVirtualHosts()
	if err != nil {
		t.Fatalf("Failed to list enabled virtual hosts: %v", err)
	}

	if len(enabled) != len(domains) {
		t.Errorf("Expected %d enabled virtual hosts, got %d", len(domains), len(enabled))
	}
}