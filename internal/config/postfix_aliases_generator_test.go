package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MYusufEka/oxmail/internal/config"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostfixAliasesGenerator_Generate_SingleAlias(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "virtual_aliases")

	gen := config.NewPostfixAliasesGenerator(outputPath)

	aliases := []domain.Alias{
		{ID: 1, SourceAddress: "info@local.test", DestinationAddress: "alice@local.test", Active: true},
	}

	err := gen.Generate(aliases)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "info@local.test alice@local.test\n", string(content))
}

func TestPostfixAliasesGenerator_Generate_MultipleAliases(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "virtual_aliases")

	gen := config.NewPostfixAliasesGenerator(outputPath)

	aliases := []domain.Alias{
		{ID: 1, SourceAddress: "info@local.test", DestinationAddress: "alice@local.test", Active: true},
		{ID: 2, SourceAddress: "info@local.test", DestinationAddress: "bob@local.test", Active: true},
		{ID: 3, SourceAddress: "support@local.test", DestinationAddress: "carol@local.test", Active: true},
	}

	err := gen.Generate(aliases)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	expected := "info@local.test alice@local.test\ninfo@local.test bob@local.test\nsupport@local.test carol@local.test\n"
	assert.Equal(t, expected, string(content))
}

func TestPostfixAliasesGenerator_Generate_SkipsInactive(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "virtual_aliases")

	gen := config.NewPostfixAliasesGenerator(outputPath)

	aliases := []domain.Alias{
		{ID: 1, SourceAddress: "info@local.test", DestinationAddress: "alice@local.test", Active: true},
		{ID: 2, SourceAddress: "old@local.test", DestinationAddress: "nobody@local.test", Active: false},
	}

	err := gen.Generate(aliases)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "info@local.test alice@local.test\n", string(content))
}

func TestPostfixAliasesGenerator_Generate_EmptyAliases(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "virtual_aliases")

	gen := config.NewPostfixAliasesGenerator(outputPath)

	err := gen.Generate([]domain.Alias{})
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "", string(content))
}

func TestPostfixAliasesGenerator_Generate_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "subdir", "virtual_aliases")

	gen := config.NewPostfixAliasesGenerator(outputPath)

	aliases := []domain.Alias{
		{ID: 1, SourceAddress: "info@local.test", DestinationAddress: "alice@local.test", Active: true},
	}

	err := gen.Generate(aliases)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "info@local.test alice@local.test\n", string(content))
}
