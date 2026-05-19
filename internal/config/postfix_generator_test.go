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

func TestPostfixDomainsGenerator_Generate(t *testing.T) {
	t.Run("generates file with one domain per line", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "postfix", "virtual_domains")
		gen := config.NewPostfixDomainsGenerator(outputPath)

		domains := []domain.Domain{
			{ID: 1, Name: "example.com", Active: true},
			{ID: 2, Name: "mail.org", Active: true},
		}

		err := gen.Generate(domains)
		require.NoError(t, err)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, "example.com\nmail.org\n", string(content))
	})

	t.Run("skips inactive domains", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "postfix", "virtual_domains")
		gen := config.NewPostfixDomainsGenerator(outputPath)

		domains := []domain.Domain{
			{ID: 1, Name: "active.com", Active: true},
			{ID: 2, Name: "inactive.com", Active: false},
			{ID: 3, Name: "also-active.net", Active: true},
		}

		err := gen.Generate(domains)
		require.NoError(t, err)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, "active.com\nalso-active.net\n", string(content))
	})

	t.Run("generates empty file when no domains", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "postfix", "virtual_domains")
		gen := config.NewPostfixDomainsGenerator(outputPath)

		err := gen.Generate(nil)
		require.NoError(t, err)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, "", string(content))
	})

	t.Run("creates output directory if missing", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "deep", "nested", "virtual_domains")
		gen := config.NewPostfixDomainsGenerator(outputPath)

		err := gen.Generate([]domain.Domain{
			{ID: 1, Name: "test.com", Active: true},
		})
		require.NoError(t, err)

		_, err = os.Stat(outputPath)
		assert.NoError(t, err)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "virtual_domains")
		gen := config.NewPostfixDomainsGenerator(outputPath)

		err := os.WriteFile(outputPath, []byte("old-content\n"), 0o644)
		require.NoError(t, err)

		err = gen.Generate([]domain.Domain{
			{ID: 1, Name: "new.com", Active: true},
		})
		require.NoError(t, err)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, "new.com\n", string(content))
	})
}
