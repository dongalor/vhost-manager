package certbot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles certbot operations for SSL certificate management
type Manager struct {
	certbotPath string
	email       string
	staging     bool
	dryRun      bool
}

// NewManager creates a new certbot manager instance
func NewManager(email string, staging bool, dryRun bool) (*Manager, error) {
	certbotPath, err := exec.LookPath("certbot")
	if err != nil {
		return nil, fmt.Errorf("certbot not found: %w (install with: sudo apt-get install certbot python3-certbot-nginx)", err)
	}

	return &Manager{
		certbotPath: certbotPath,
		email:       email,
		staging:     staging,
		dryRun:      dryRun,
	}, nil
}

// CheckInstalled verifies that certbot is installed and accessible
func CheckInstalled() error {
	_, err := exec.LookPath("certbot")
	if err != nil {
		return fmt.Errorf("certbot not found: install with 'sudo apt-get install certbot python3-certbot-nginx' or visit https://certbot.eff.org/")
	}
	return nil
}

// InstallCertificate obtains and installs an SSL certificate for the specified domain
func (m *Manager) InstallCertificate(domain string, useNginxPlugin bool) error {
	if m.dryRun {
		fmt.Printf("[DRY RUN] Would run certbot to install certificate for %s\n", domain)
		return nil
	}

	args := []string{
		"certonly",
		"-d", domain,
		"--non-interactive",
		"--agree-tos",
	}

	// Add email if provided
	if m.email != "" {
		args = append(args, "--email", m.email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	// Use staging environment if requested
	if m.staging {
		args = append(args, "--staging")
	}

	// Choose plugin: nginx or webroot
	if useNginxPlugin {
		args = append(args, "--nginx")
	} else {
		args = append(args, "--webroot", "-w", fmt.Sprintf("/var/www/%s", domain))
	}

	cmd := exec.Command(m.certbotPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install certificate: %w", err)
	}

	fmt.Printf("✓ Certificate successfully installed for %s\n", domain)
	return nil
}

// RenewCertificate renews the certificate for a specific domain
func (m *Manager) RenewCertificate(domain string) error {
	if m.dryRun {
		fmt.Printf("[DRY RUN] Would renew certificate for %s\n", domain)
		return nil
	}

	args := []string{
		"renew",
		"--cert-name", domain,
		"--non-interactive",
	}

	if m.staging {
		args = append(args, "--staging")
	}

	cmd := exec.Command(m.certbotPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to renew certificate: %w", err)
	}

	fmt.Printf("✓ Certificate renewed for %s\n", domain)
	return nil
}

// RenewAll renews all certificates managed by certbot
func (m *Manager) RenewAll() error {
	if m.dryRun {
		fmt.Printf("[DRY RUN] Would renew all certificates\n")
		return nil
	}

	args := []string{
		"renew",
		"--non-interactive",
	}

	if m.staging {
		args = append(args, "--staging")
	}

	cmd := exec.Command(m.certbotPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to renew certificates: %w", err)
	}

	fmt.Println("✓ All certificates renewed successfully")
	return nil
}

// RemoveCertificate deletes the certificate for the specified domain
func (m *Manager) RemoveCertificate(domain string) error {
	if m.dryRun {
		fmt.Printf("[DRY RUN] Would remove certificate for %s\n", domain)
		return nil
	}

	args := []string{
		"delete",
		"--cert-name", domain,
		"--non-interactive",
	}

	cmd := exec.Command(m.certbotPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove certificate: %w", err)
	}

	fmt.Printf("✓ Certificate removed for %s\n", domain)
	return nil
}

// GetCertificateInfo retrieves information about a certificate
func (m *Manager) GetCertificateInfo(domain string) (*CertificateInfo, error) {
	certPath := filepath.Join("/etc/letsencrypt/live", domain, "fullchain.pem")

	// Check if certificate exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate not found for domain: %s", domain)
	}

	// Use openssl to get certificate information
	cmd := exec.Command("openssl", "x509", "-in", certPath, "-noout", "-enddate", "-issuer")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate info: %w", err)
	}

	info := &CertificateInfo{
		Domain:   domain,
		CertPath: certPath,
		KeyPath:  filepath.Join("/etc/letsencrypt/live", domain, "privkey.pem"),
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "notAfter=") {
			dateStr := strings.TrimPrefix(line, "notAfter=")
			expiryDate, err := time.Parse("Jan 2 15:04:05 2006 MST", dateStr)
			if err != nil {
				// Try alternative format
				expiryDate, err = time.Parse(time.RFC3339, dateStr)
			}
			if err == nil {
				info.ExpiryDate = expiryDate
				info.DaysLeft = int(time.Until(expiryDate).Hours() / 24)
			}
		} else if strings.HasPrefix(line, "issuer=") {
			info.Issuer = strings.TrimPrefix(line, "issuer=")
		}
	}

	return info, nil
}

// ListCertificates returns a list of all certificates managed by certbot
func (m *Manager) ListCertificates() ([]*CertificateInfo, error) {
	cmd := exec.Command(m.certbotPath, "certificates")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %w", err)
	}

	var certificates []*CertificateInfo
	lines := strings.Split(string(output), "\n")

	var currentDomain string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for certificate name lines
		if strings.HasPrefix(line, "Certificate Name:") {
			currentDomain = strings.TrimSpace(strings.TrimPrefix(line, "Certificate Name:"))
		} else if strings.HasPrefix(line, "Domains:") && currentDomain != "" {
			// Get detailed info for this certificate
			info, err := m.GetCertificateInfo(currentDomain)
			if err == nil {
				certificates = append(certificates, info)
			}
			currentDomain = ""
		}
	}

	return certificates, nil
}

// GetCertificatePaths returns the paths to the certificate files for a domain
func GetCertificatePaths(domain string) (certPath, keyPath, optionsPath, dhparamPath string) {
	baseDir := filepath.Join("/etc/letsencrypt/live", domain)
	return filepath.Join(baseDir, "fullchain.pem"),
		filepath.Join(baseDir, "privkey.pem"),
		"/etc/letsencrypt/options-ssl-nginx.conf",
		"/etc/letsencrypt/ssl-dhparams.pem"
}

// CertificateExists checks if a certificate exists for the given domain
func CertificateExists(domain string) bool {
	certPath := filepath.Join("/etc/letsencrypt/live", domain, "fullchain.pem")
	_, err := os.Stat(certPath)
	return err == nil
}
