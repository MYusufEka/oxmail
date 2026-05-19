package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// PostfixAliasesGenerator generates the virtual_alias_maps file for Postfix.
type PostfixAliasesGenerator struct {
	outputPath string
}

// NewPostfixAliasesGenerator creates a new generator with the given output path.
func NewPostfixAliasesGenerator(outputPath string) *PostfixAliasesGenerator {
	return &PostfixAliasesGenerator{outputPath: outputPath}
}

// Generate writes the virtual_alias_maps file from the given aliases.
// Format: one line per alias, "source destination" separated by space.
// Only active aliases are included.
func (g *PostfixAliasesGenerator) Generate(aliases []domain.Alias) error {
	dir := filepath.Dir(g.outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	var builder strings.Builder
	for _, alias := range aliases {
		if !alias.Active {
			continue
		}
		builder.WriteString(alias.SourceAddress)
		builder.WriteByte(' ')
		builder.WriteString(alias.DestinationAddress)
		builder.WriteByte('\n')
	}

	if err := os.WriteFile(g.outputPath, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write virtual_aliases file: %w", err)
	}

	return nil
}
