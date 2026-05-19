package mail

import (
	"fmt"
	"os"
	"strconv"
)

// OutboundConfig holds configuration for production outbound mail delivery.
type OutboundConfig struct {
	Domain        string
	Hostname      string
	PublicIP      string
	RateLimit     int
	RelayHost     string
	EnableTLS     bool
}

// DefaultOutboundRateLimit is the default outbound rate limit per hour.
const DefaultOutboundRateLimit = 100

// NewOutboundConfig creates an OutboundConfig from environment variables.
func NewOutboundConfig() *OutboundConfig {
	domain := os.Getenv("OXMAIL_DOMAIN")
	if domain == "" {
		domain = "local.test"
	}

	publicIP := os.Getenv("OXMAIL_PUBLIC_IP")

	rateLimit := DefaultOutboundRateLimit
	if envRate := os.Getenv("OXMAIL_OUTBOUND_RATE_LIMIT"); envRate != "" {
		if parsed, err := strconv.Atoi(envRate); err == nil && parsed > 0 {
			rateLimit = parsed
		}
	}

	return &OutboundConfig{
		Domain:    domain,
		Hostname:  fmt.Sprintf("mail.%s", domain),
		PublicIP:  publicIP,
		RateLimit: rateLimit,
		EnableTLS: true,
	}
}

// PostfixOutboundParams returns Postfix main.cf parameters for production outbound delivery.
func (c *OutboundConfig) PostfixOutboundParams() map[string]string {
	params := map[string]string{
		"myhostname":                  c.Hostname,
		"mydomain":                    c.Domain,
		"myorigin":                    "$mydomain",
		"smtp_helo_name":              c.Hostname,
		"smtp_tls_security_level":     "may",
		"smtp_tls_loglevel":           "1",
		"smtpd_tls_security_level":    "may",
		"smtp_destination_rate_delay": calculateRateDelay(c.RateLimit),
	}

	if c.RelayHost != "" {
		params["relayhost"] = c.RelayHost
	}

	return params
}

// calculateRateDelay converts emails/hour to a Postfix rate delay string.
// Postfix smtp_destination_rate_delay is in seconds between messages per destination.
func calculateRateDelay(emailsPerHour int) string {
	if emailsPerHour <= 0 {
		return "0s"
	}
	delaySeconds := 3600 / emailsPerHour
	if delaySeconds < 1 {
		return "0s"
	}
	return fmt.Sprintf("%ds", delaySeconds)
}
