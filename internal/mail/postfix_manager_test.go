package mail

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommandExecutor records commands and returns configurable results.
type mockCommandExecutor struct {
	commands [][]string
	err      error
}

func (m *mockCommandExecutor) Run(name string, args ...string) error {
	cmd := append([]string{name}, args...)
	m.commands = append(m.commands, cmd)
	return m.err
}

func (m *mockCommandExecutor) RunWithOutput(name string, args ...string) (string, error) {
	cmd := append([]string{name}, args...)
	m.commands = append(m.commands, cmd)
	return "", m.err
}

// failingCommandExecutor fails on the Nth call.
type failingCommandExecutor struct {
	commands  [][]string
	failUntil int
	callCount int
}

func (f *failingCommandExecutor) Run(name string, args ...string) error {
	cmd := append([]string{name}, args...)
	f.commands = append(f.commands, cmd)
	f.callCount++
	if f.callCount <= f.failUntil {
		return fmt.Errorf("command failed: %s", name)
	}
	return nil
}

func (f *failingCommandExecutor) RunWithOutput(name string, args ...string) (string, error) {
	cmd := append([]string{name}, args...)
	f.commands = append(f.commands, cmd)
	f.callCount++
	if f.callCount <= f.failUntil {
		return "", fmt.Errorf("command failed: %s", name)
	}
	return "", nil
}

func TestPostfixManager_ApplyDomainConfig(t *testing.T) {
	tmpDir := t.TempDir()
	domainsPath := filepath.Join(tmpDir, "virtual_domains")

	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{
		DomainsPath: domainsPath,
		AliasesPath: filepath.Join(tmpDir, "virtual_aliases"),
	}, executor)

	domains := []domain.Domain{
		{ID: 1, Name: "example.com", Active: true},
		{ID: 2, Name: "test.org", Active: true},
		{ID: 3, Name: "inactive.net", Active: false},
	}

	err := manager.ApplyDomainConfig(domains)
	require.NoError(t, err)

	// Verify file was written with only active domains.
	content, err := os.ReadFile(domainsPath)
	require.NoError(t, err)
	assert.Equal(t, "example.com\ntest.org\n", string(content))

	// Verify postmap was called.
	require.Len(t, executor.commands, 2)
	assert.Equal(t, []string{"postmap", "hash:" + domainsPath}, executor.commands[0])

	// Verify reload was called.
	assert.Equal(t, []string{"postfix", "reload"}, executor.commands[1])
}

func TestPostfixManager_ApplyDomainConfig_EmptyDomains(t *testing.T) {
	tmpDir := t.TempDir()
	domainsPath := filepath.Join(tmpDir, "virtual_domains")

	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{
		DomainsPath: domainsPath,
		AliasesPath: filepath.Join(tmpDir, "virtual_aliases"),
	}, executor)

	err := manager.ApplyDomainConfig([]domain.Domain{})
	require.NoError(t, err)

	content, err := os.ReadFile(domainsPath)
	require.NoError(t, err)
	assert.Equal(t, "", string(content))
}

func TestPostfixManager_ApplyAliasConfig(t *testing.T) {
	tmpDir := t.TempDir()
	aliasesPath := filepath.Join(tmpDir, "virtual_aliases")

	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: filepath.Join(tmpDir, "virtual_domains"),
		AliasesPath: aliasesPath,}, executor)

	aliases := []domain.Alias{
		{ID: 1, SourceAddress: "info@example.com", DestinationAddress: "admin@example.com", Active: true},
		{ID: 2, SourceAddress: "old@test.org", DestinationAddress: "new@test.org", Active: true},
		{ID: 3, SourceAddress: "disabled@test.org", DestinationAddress: "x@test.org", Active: false},
	}

	err := manager.ApplyAliasConfig(aliases)
	require.NoError(t, err)

	content, err := os.ReadFile(aliasesPath)
	require.NoError(t, err)
	assert.Equal(t, "info@example.com admin@example.com\nold@test.org new@test.org\n", string(content))

	// Verify postmap + reload.
	require.Len(t, executor.commands, 2)
	assert.Equal(t, []string{"postmap", "hash:" + aliasesPath}, executor.commands[0])
	assert.Equal(t, []string{"postfix", "reload"}, executor.commands[1])
}

func TestPostfixManager_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	domainsPath := filepath.Join(tmpDir, "virtual_domains")

	// Pre-create the file with old content.
	err := os.WriteFile(domainsPath, []byte("old.com\n"), 0o644)
	require.NoError(t, err)

	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: domainsPath,
		AliasesPath: filepath.Join(tmpDir, "virtual_aliases"),}, executor)

	domains := []domain.Domain{
		{ID: 1, Name: "new.com", Active: true},
	}

	err = manager.ApplyDomainConfig(domains)
	require.NoError(t, err)

	// Old content should be replaced atomically.
	content, err := os.ReadFile(domainsPath)
	require.NoError(t, err)
	assert.Equal(t, "new.com\n", string(content))

	// No temp file should remain.
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "*.tmp"))
	assert.Empty(t, matches)
}

func TestPostfixManager_Postmap(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "virtual_domains")

	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: filePath,
		AliasesPath: filepath.Join(tmpDir, "virtual_aliases"),}, executor)

	err := manager.Postmap(filePath)
	require.NoError(t, err)

	require.Len(t, executor.commands, 1)
	assert.Equal(t, []string{"postmap", "hash:" + filePath}, executor.commands[0])
}

func TestPostfixManager_Reload(t *testing.T) {
	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: "/etc/postfix/virtual_domains",
		AliasesPath: "/etc/postfix/virtual_aliases",}, executor)

	err := manager.Reload()
	require.NoError(t, err)

	require.Len(t, executor.commands, 1)
	assert.Equal(t, []string{"postfix", "reload"}, executor.commands[0])
}

func TestPostfixManager_ReloadRetry(t *testing.T) {
	// Fails first 2 times, succeeds on 3rd.
	executor := &failingCommandExecutor{failUntil: 2}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: "/etc/postfix/virtual_domains",
		AliasesPath: "/etc/postfix/virtual_aliases",}, executor)
	manager.retryDelay = 0 // No delay in tests.

	err := manager.Reload()
	require.NoError(t, err)

	// Should have been called 3 times (2 failures + 1 success).
	assert.Equal(t, 3, executor.callCount)
}

func TestPostfixManager_ReloadRetryExhausted(t *testing.T) {
	// Fails all 3 attempts.
	executor := &failingCommandExecutor{failUntil: 3}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: "/etc/postfix/virtual_domains",
		AliasesPath: "/etc/postfix/virtual_aliases",}, executor)
	manager.retryDelay = 0

	err := manager.Reload()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postfix reload failed after 3 attempts")
	assert.Equal(t, 3, executor.callCount)
}

func TestPostfixManager_PostmapError(t *testing.T) {
	executor := &mockCommandExecutor{err: fmt.Errorf("postmap binary not found")}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: "/etc/postfix/virtual_domains",
		AliasesPath: "/etc/postfix/virtual_aliases",}, executor)

	err := manager.Postmap("/etc/postfix/virtual_domains")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postmap")
}

func TestPostfixManager_ApplyDomainConfig_PostmapFails(t *testing.T) {
	tmpDir := t.TempDir()
	domainsPath := filepath.Join(tmpDir, "virtual_domains")

	executor := &mockCommandExecutor{err: fmt.Errorf("exec error")}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: domainsPath,
		AliasesPath: filepath.Join(tmpDir, "virtual_aliases"),}, executor)

	domains := []domain.Domain{
		{ID: 1, Name: "example.com", Active: true},
	}

	err := manager.ApplyDomainConfig(domains)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exec error")

	// File should still be written even if postmap fails.
	content, err := os.ReadFile(domainsPath)
	require.NoError(t, err)
	assert.Equal(t, "example.com\n", string(content))
}

func TestPostfixManager_ConfigDirectoryCreated(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "sub", "dir", "virtual_domains")

	executor := &mockCommandExecutor{}
	manager := NewPostfixManager(PostfixConfig{DomainsPath: nestedPath,
		AliasesPath: filepath.Join(tmpDir, "sub", "dir", "virtual_aliases"),}, executor)

	domains := []domain.Domain{
		{ID: 1, Name: "example.com", Active: true},
	}

	err := manager.ApplyDomainConfig(domains)
	require.NoError(t, err)

	content, err := os.ReadFile(nestedPath)
	require.NoError(t, err)
	assert.Equal(t, "example.com\n", string(content))
}
