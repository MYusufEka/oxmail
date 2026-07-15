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
// Only active aliases are included.
func (g *PostfixAliasesGenerator) Generate(aliases []domain.Alias) error {
	dir := filepath.Dir(g.outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	destsBySource := make(map[string][]string)
	sourceOrder := make([]string, 0)
	for _, alias := range aliases {
		if !alias.Active {
			continue
		}
		if _, seen := destsBySource[alias.SourceAddress]; !seen {
			sourceOrder = append(sourceOrder, alias.SourceAddress)
		}
		destsBySource[alias.SourceAddress] = append(destsBySource[alias.SourceAddress], alias.DestinationAddress)
	}

	var builder strings.Builder
	for _, src := range sourceOrder {
		builder.WriteString(src)
		builder.WriteByte(' ')
		builder.WriteString(strings.Join(destsBySource[src], ","))
		builder.WriteByte('\n')
	}

	if err := os.WriteFile(g.outputPath, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write virtual_aliases file: %w", err)
	}

	return nil
}
