package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// DKIMGenerator produces OpenDKIM signing and key table configuration files.
type DKIMGenerator struct {
	keyBasePath string
}

// NewDKIMGenerator creates a DKIMGenerator with the given base path for private keys.
func NewDKIMGenerator(keyBasePath string) *DKIMGenerator {
	return &DKIMGenerator{keyBasePath: keyBasePath}
}

// SigningTable generates the OpenDKIM signing table content.
// Format: *@domain selector._domainkey.domain
func (g *DKIMGenerator) SigningTable(keys []domain.DKIMKey) string {
	if len(keys) == 0 {
		return ""
	}

	var lines []string
	for _, key := range keys {
		line := fmt.Sprintf("*@%s %s._domainkey.%s", key.Domain, key.Selector, key.Domain)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}

// KeyTable generates the OpenDKIM key table content.
// Format: selector._domainkey.domain domain:selector:/path/to/key.private
func (g *DKIMGenerator) KeyTable(keys []domain.DKIMKey) string {
	if len(keys) == 0 {
		return ""
	}

	var lines []string
	for _, key := range keys {
		keyPath := g.keyBasePath + "/" + key.Domain + "/" + key.Selector + ".private"
		line := fmt.Sprintf("%s._domainkey.%s %s:%s:%s",
			key.Selector, key.Domain,
			key.Domain, key.Selector, keyPath,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}

// WriteFiles writes the signing table and key table to the given output directory.
// Returns the paths to the written files.
func (g *DKIMGenerator) WriteFiles(outputDir string, keys []domain.DKIMKey) (signingPath, keyTablePath string, err error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("create output directory: %w", err)
	}

	signingPath = filepath.Join(outputDir, "signing.table")
	keyTablePath = filepath.Join(outputDir, "key.table")

	signingContent := g.SigningTable(keys)
	if err := os.WriteFile(signingPath, []byte(signingContent), 0644); err != nil {
		return "", "", fmt.Errorf("write signing table: %w", err)
	}

	keyContent := g.KeyTable(keys)
	if err := os.WriteFile(keyTablePath, []byte(keyContent), 0644); err != nil {
		return "", "", fmt.Errorf("write key table: %w", err)
	}

	return signingPath, keyTablePath, nil
}
