package api_test

import (
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetDNSResolver_LookupMX(t *testing.T) {
	resolver := &api.NetDNSResolver{}

	t.Run("resolves known domain MX records", func(t *testing.T) {
		hosts, err := resolver.LookupMX("gmail.com")
		if err != nil {
			t.Skip("DNS lookup failed (no network?):", err)
		}
		require.NotEmpty(t, hosts)
		for _, h := range hosts {
			assert.Contains(t, h, "gmail")
		}
	})

	t.Run("returns error for invalid domain", func(t *testing.T) {
		_, err := resolver.LookupMX("nonexistent-domain-xyz12345.test")
		assert.Error(t, err)
	})
}

func TestNetDNSResolver_LookupTXT(t *testing.T) {
	resolver := &api.NetDNSResolver{}

	t.Run("resolves known domain TXT records", func(t *testing.T) {
		records, err := resolver.LookupTXT("gmail.com")
		if err != nil {
			t.Skip("DNS lookup failed (no network?):", err)
		}
		require.NotEmpty(t, records)
		hasSPF := false
		for _, r := range records {
			if len(r) > 5 && r[:6] == "v=spf1" {
				hasSPF = true
				break
			}
		}
		assert.True(t, hasSPF, "gmail.com should have SPF records")
	})

	t.Run("returns error for invalid domain", func(t *testing.T) {
		_, err := resolver.LookupTXT("nonexistent-domain-xyz12345.test")
		assert.Error(t, err)
	})
}
