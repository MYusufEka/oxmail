package mail

import (
	"os"
	"testing"
)

func TestNewOutboundConfig(t *testing.T) {
	// Save and restore env
	savedDomain := os.Getenv("OXMAIL_DOMAIN")
	savedIP := os.Getenv("OXMAIL_PUBLIC_IP")
	savedRate := os.Getenv("OXMAIL_OUTBOUND_RATE_LIMIT")
	defer func() {
		os.Setenv("OXMAIL_DOMAIN", savedDomain)
		os.Setenv("OXMAIL_PUBLIC_IP", savedIP)
		os.Setenv("OXMAIL_OUTBOUND_RATE_LIMIT", savedRate)
	}()

	t.Run("default values", func(t *testing.T) {
		os.Unsetenv("OXMAIL_DOMAIN")
		os.Unsetenv("OXMAIL_PUBLIC_IP")
		os.Unsetenv("OXMAIL_OUTBOUND_RATE_LIMIT")

		cfg := NewOutboundConfig()
		if cfg.Domain != "local.test" {
			t.Errorf("Domain = %q, want %q", cfg.Domain, "local.test")
		}
		if cfg.Hostname != "mail.local.test" {
			t.Errorf("Hostname = %q, want %q", cfg.Hostname, "mail.local.test")
		}
		if cfg.PublicIP != "" {
			t.Errorf("PublicIP = %q, want empty", cfg.PublicIP)
		}
		if cfg.RateLimit != 100 {
			t.Errorf("RateLimit = %d, want %d", cfg.RateLimit, 100)
		}
		if !cfg.EnableTLS {
			t.Error("EnableTLS = false, want true")
		}
	})

	t.Run("custom domain", func(t *testing.T) {
		os.Setenv("OXMAIL_DOMAIN", "example.com")
		os.Unsetenv("OXMAIL_PUBLIC_IP")
		os.Unsetenv("OXMAIL_OUTBOUND_RATE_LIMIT")

		cfg := NewOutboundConfig()
		if cfg.Domain != "example.com" {
			t.Errorf("Domain = %q, want %q", cfg.Domain, "example.com")
		}
		if cfg.Hostname != "mail.example.com" {
			t.Errorf("Hostname = %q, want %q", cfg.Hostname, "mail.example.com")
		}
	})

	t.Run("custom public IP", func(t *testing.T) {
		os.Setenv("OXMAIL_DOMAIN", "local.test")
		os.Setenv("OXMAIL_PUBLIC_IP", "203.0.113.42")
		os.Unsetenv("OXMAIL_OUTBOUND_RATE_LIMIT")

		cfg := NewOutboundConfig()
		if cfg.PublicIP != "203.0.113.42" {
			t.Errorf("PublicIP = %q, want %q", cfg.PublicIP, "203.0.113.42")
		}
	})

	t.Run("custom rate limit", func(t *testing.T) {
		os.Setenv("OXMAIL_DOMAIN", "local.test")
		os.Unsetenv("OXMAIL_PUBLIC_IP")
		os.Setenv("OXMAIL_OUTBOUND_RATE_LIMIT", "50")

		cfg := NewOutboundConfig()
		if cfg.RateLimit != 50 {
			t.Errorf("RateLimit = %d, want %d", cfg.RateLimit, 50)
		}
	})

	t.Run("invalid rate limit falls back to default", func(t *testing.T) {
		os.Setenv("OXMAIL_OUTBOUND_RATE_LIMIT", "not-a-number")

		cfg := NewOutboundConfig()
		if cfg.RateLimit != 100 {
			t.Errorf("RateLimit = %d, want %d", cfg.RateLimit, 100)
		}
	})

	t.Run("zero rate limit falls back to default", func(t *testing.T) {
		os.Setenv("OXMAIL_OUTBOUND_RATE_LIMIT", "0")

		cfg := NewOutboundConfig()
		if cfg.RateLimit != 100 {
			t.Errorf("RateLimit = %d, want %d", cfg.RateLimit, 100)
		}
	})
}

func TestPostfixOutboundParams(t *testing.T) {
	cfg := &OutboundConfig{
		Domain:    "example.com",
		Hostname:  "mail.example.com",
		PublicIP:  "203.0.113.42",
		RateLimit: 100,
		EnableTLS: true,
	}

	params := cfg.PostfixOutboundParams()
	checks := map[string]string{
		"myhostname":                   "mail.example.com",
		"mydomain":                     "example.com",
		"myorigin":                     "$mydomain",
		"smtp_helo_name":               "mail.example.com",
		"smtp_tls_security_level":      "may",
		"smtp_tls_loglevel":            "1",
		"smtpd_tls_security_level":     "may",
		"smtp_destination_rate_delay":  "36s",
	}

	for key, want := range checks {
		t.Run(key, func(t *testing.T) {
			got, ok := params[key]
			if !ok {
				t.Errorf("params[%q] missing", key)
				return
			}
			if got != want {
				t.Errorf("params[%q] = %q, want %q", key, got, want)
			}
		})
	}

	if _, ok := params["relayhost"]; ok {
		t.Error("relayhost should not be present when RelayHost is empty")
	}
}

func TestPostfixOutboundParams_WithRelayHost(t *testing.T) {
	cfg := &OutboundConfig{
		Domain:    "example.com",
		Hostname:  "mail.example.com",
		RateLimit: 100,
		RelayHost: "[smtp.example.com]:587",
	}

	params := cfg.PostfixOutboundParams()
	if params["relayhost"] != "[smtp.example.com]:587" {
		t.Errorf("relayhost = %q, want %q", params["relayhost"], "[smtp.example.com]:587")
	}
}

func TestCalculateRateDelay(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"100 per hour", 100, "36s"},
		{"200 per hour", 200, "18s"},
		{"3600 per hour", 3600, "1s"},
		{"7200 per hour", 7200, "0s"},
		{"1 per hour", 1, "3600s"},
		{"0 per hour", 0, "0s"},
		{"negative", -5, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRateDelay(tt.input)
			if got != tt.want {
				t.Errorf("calculateRateDelay(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
