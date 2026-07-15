package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDKIMService_Generate(t *testing.T) {
	tempDir := t.TempDir()
	service := domain.NewDKIMService(nil, tempDir)

	t.Run("generates valid RSA 2048-bit key pair", func(t *testing.T) {
		result, err := service.Generate("example.com", "default")
		require.NoError(t, err)

		assert.Equal(t, "example.com", result.Domain)
		assert.Equal(t, "default", result.Selector)
		assert.NotEmpty(t, result.PublicKey)
		assert.Contains(t, result.DNSRecord, "v=DKIM1; k=rsa; p=")
		assert.False(t, result.CreatedAt.IsZero())
	})

	t.Run("stores private key on filesystem", func(t *testing.T) {
		_, err := service.Generate("fs-test.com", "default")
		require.NoError(t, err)

		keyPath := filepath.Join(tempDir, "fs-test.com", "default.private")
		_, err = os.Stat(keyPath)
		assert.NoError(t, err, "private key file should exist")

		content, err := os.ReadFile(keyPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "BEGIN RSA PRIVATE KEY")
	})

	t.Run("private key file has restricted permissions", func(t *testing.T) {
		if os.Getenv("GOOS") == "windows" || filepath.Separator == '\\' {
			t.Skip("file permission test not applicable on Windows")
		}

		_, err := service.Generate("perm-test.com", "default")
		require.NoError(t, err)

		keyPath := filepath.Join(tempDir, "perm-test.com", "default.private")
		info, err := os.Stat(keyPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("returns error for empty domain", func(t *testing.T) {
		_, err := service.Generate("", "default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "domain")
	})

	t.Run("returns error for empty selector", func(t *testing.T) {
		_, err := service.Generate("example.com", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "selector")
	})

	t.Run("overwrites existing key for same domain/selector", func(t *testing.T) {
		first, err := service.Generate("overwrite.com", "default")
		require.NoError(t, err)

		second, err := service.Generate("overwrite.com", "default")
		require.NoError(t, err)

		assert.NotEqual(t, first.PublicKey, second.PublicKey)
	})
}

func TestDKIMService_Get(t *testing.T) {
	tempDir := t.TempDir()
	service := domain.NewDKIMService(nil, tempDir)

	t.Run("returns generated key info", func(t *testing.T) {
		generated, err := service.Generate("get-test.com", "default")
		require.NoError(t, err)

		result, err := service.Get("get-test.com", "default")
		require.NoError(t, err)

		assert.Equal(t, generated.Domain, result.Domain)
		assert.Equal(t, generated.Selector, result.Selector)
		assert.Equal(t, generated.PublicKey, result.PublicKey)
		assert.Equal(t, generated.DNSRecord, result.DNSRecord)
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := service.Get("nonexistent.com", "default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDKIMService_Delete(t *testing.T) {
	tempDir := t.TempDir()
	service := domain.NewDKIMService(nil, tempDir)

	t.Run("deletes existing key", func(t *testing.T) {
		_, err := service.Generate("delete-test.com", "default")
		require.NoError(t, err)

		err = service.Delete("delete-test.com", "default")
		require.NoError(t, err)

		_, err = service.Get("delete-test.com", "default")
		assert.Error(t, err)

		keyPath := filepath.Join(tempDir, "delete-test.com", "default.private")
		_, err = os.Stat(keyPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		err := service.Delete("nonexistent.com", "default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDKIMService_Rotate(t *testing.T) {
	tempDir := t.TempDir()
	service := domain.NewDKIMService(nil, tempDir)

	t.Run("generates new key replacing old one", func(t *testing.T) {
		original, err := service.Generate("rotate-test.com", "default")
		require.NoError(t, err)

		rotated, err := service.Rotate("rotate-test.com", "default")
		require.NoError(t, err)

		assert.Equal(t, "rotate-test.com", rotated.Domain)
		assert.Equal(t, "default", rotated.Selector)
		assert.NotEqual(t, original.PublicKey, rotated.PublicKey)
		assert.Contains(t, rotated.DNSRecord, "v=DKIM1; k=rsa; p=")
	})

	t.Run("returns error when no existing key to rotate", func(t *testing.T) {
		_, err := service.Rotate("no-key.com", "default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
