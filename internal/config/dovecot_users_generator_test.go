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

func TestDovecotUsersGenerator_GeneratePassdb(t *testing.T) {
	tmpDir := t.TempDir()
	gen := config.NewDovecotUsersGenerator(tmpDir)

	users := []domain.User{
		{Email: "alice@local.test", PasswordHash: "$2b$12$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012"},
		{Email: "bob@local.test", PasswordHash: "$2b$12$xyzxyzxyzxyzxyzxyzxyzuuABCDEFGHIJKLMNOPQRSTUVWXYZ012"},
	}

	err := gen.GeneratePassdb(users)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "passdb"))
	require.NoError(t, err)

	lines := string(content)
	assert.Contains(t, lines, "alice@local.test:{BLF-CRYPT}$2b$12$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012\n")
	assert.Contains(t, lines, "bob@local.test:{BLF-CRYPT}$2b$12$xyzxyzxyzxyzxyzxyzxyzuuABCDEFGHIJKLMNOPQRSTUVWXYZ012\n")
}

func TestDovecotUsersGenerator_GenerateUserdb(t *testing.T) {
	tmpDir := t.TempDir()
	gen := config.NewDovecotUsersGenerator(tmpDir)

	users := []domain.User{
		{Email: "alice@local.test"},
		{Email: "bob@other.test"},
	}

	err := gen.GenerateUserdb(users)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "userdb"))
	require.NoError(t, err)

	lines := string(content)
	assert.Contains(t, lines, "alice@local.test::5000:5000::/var/mail/vhosts/local.test/alice\n")
	assert.Contains(t, lines, "bob@other.test::5000:5000::/var/mail/vhosts/other.test/bob\n")
}

func TestDovecotUsersGenerator_GenerateAll(t *testing.T) {
	tmpDir := t.TempDir()
	gen := config.NewDovecotUsersGenerator(tmpDir)

	users := []domain.User{
		{Email: "alice@local.test", PasswordHash: "$2b$12$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012"},
	}

	err := gen.GenerateAll(users)
	require.NoError(t, err)

	// Both files should exist
	_, err = os.Stat(filepath.Join(tmpDir, "passdb"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "userdb"))
	assert.NoError(t, err)
}

func TestDovecotUsersGenerator_EmptyUsers(t *testing.T) {
	tmpDir := t.TempDir()
	gen := config.NewDovecotUsersGenerator(tmpDir)

	err := gen.GeneratePassdb([]domain.User{})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "passdb"))
	require.NoError(t, err)
	assert.Empty(t, string(content))
}

func TestDovecotUsersGenerator_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "nested", "dovecot")
	gen := config.NewDovecotUsersGenerator(outputDir)

	users := []domain.User{
		{Email: "alice@local.test", PasswordHash: "$2b$12$hash"},
	}

	err := gen.GeneratePassdb(users)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "passdb"))
	assert.NoError(t, err)
}
