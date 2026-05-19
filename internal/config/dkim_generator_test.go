package config

import (
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDKIMGenerator_SigningTable(t *testing.T) {
	gen := NewDKIMGenerator("/etc/dkim")

	keys := []domain.DKIMKey{
		{Domain: "example.com", Selector: "default"},
		{Domain: "other.org", Selector: "mail"},
	}

	t.Run("generates signing table entries", func(t *testing.T) {
		result := gen.SigningTable(keys)
		assert.Contains(t, result, "*@example.com default._domainkey.example.com")
		assert.Contains(t, result, "*@other.org mail._domainkey.other.org")
	})

	t.Run("returns empty string for no keys", func(t *testing.T) {
		result := gen.SigningTable(nil)
		assert.Empty(t, result)
	})
}

func TestDKIMGenerator_KeyTable(t *testing.T) {
	gen := NewDKIMGenerator("/etc/dkim")

	keys := []domain.DKIMKey{
		{Domain: "example.com", Selector: "default"},
		{Domain: "other.org", Selector: "mail"},
	}

	t.Run("generates key table entries", func(t *testing.T) {
		result := gen.KeyTable(keys)
		assert.Contains(t, result, "default._domainkey.example.com example.com:default:/etc/dkim/example.com/default.private")
		assert.Contains(t, result, "mail._domainkey.other.org other.org:mail:/etc/dkim/other.org/mail.private")
	})

	t.Run("returns empty string for no keys", func(t *testing.T) {
		result := gen.KeyTable(nil)
		assert.Empty(t, result)
	})
}

func TestDKIMGenerator_WriteFiles(t *testing.T) {
	tempDir := t.TempDir()
	gen := NewDKIMGenerator("/etc/dkim")

	keys := []domain.DKIMKey{
		{Domain: "example.com", Selector: "default"},
	}

	t.Run("writes signing and key table files", func(t *testing.T) {
		signingPath, keyPath, err := gen.WriteFiles(tempDir, keys)
		require.NoError(t, err)

		assert.FileExists(t, signingPath)
		assert.FileExists(t, keyPath)
	})
}
