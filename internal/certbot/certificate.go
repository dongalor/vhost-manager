package certbot

import (
	"fmt"
	"time"
)

// CertificateInfo contains information about an SSL certificate
type CertificateInfo struct {
	Domain     string
	ExpiryDate time.Time
	DaysLeft   int
	Issuer     string
	CertPath   string
	KeyPath    string
	Renewed    bool
}

// IsExpiringSoon returns true if the certificate expires in less than the specified days
func (c *CertificateInfo) IsExpiringSoon(days int) bool {
	return c.DaysLeft < days
}

// IsExpired returns true if the certificate has expired
func (c *CertificateInfo) IsExpired() bool {
	return c.DaysLeft <= 0
}

// Status returns a human-readable status string
func (c *CertificateInfo) Status() string {
	if c.IsExpired() {
		return "✗ Expired"
	}
	if c.IsExpiringSoon(30) {
		return "⚠ Expiring Soon"
	}
	return "✓ Valid"
}

// String returns a formatted string representation of the certificate info
func (c *CertificateInfo) String() string {
	return fmt.Sprintf(`Domain: %s
Issuer: %s
Expiry: %s (%d days remaining)
Status: %s
Certificate: %s
Private Key: %s`,
		c.Domain,
		c.Issuer,
		c.ExpiryDate.Format("2006-01-02 15:04:05"),
		c.DaysLeft,
		c.Status(),
		c.CertPath,
		c.KeyPath,
	)
}
