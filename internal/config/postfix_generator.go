package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// PostfixDomainsGenerator generates the virtual_domains file for Postfix.
type PostfixDomainsGenerator struct {
	outputPath string
}

// NewPostfixDomainsGenerator creates a new generator with the given output path.
func NewPostfixDomainsGenerator(outputPath string) *PostfixDomainsGenerator {
	return &PostfixDomainsGenerator{outputPath: outputPath}
}

// Generate writes the virtual_domains file from the given domains.
// Format: one domain per line. Only active domains are included.
func (g *PostfixDomainsGenerator) Generate(domains []domain.Domain) error {
	dir := filepath.Dir(g.outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	var builder strings.Builder
	for _, d := range domains {
		if !d.Active {
			continue
		}
		builder.WriteString(d.Name)
		builder.WriteByte('\n')
	}

	if err := os.WriteFile(g.outputPath, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write virtual_domains file: %w", err)
	}

	return nil
}
