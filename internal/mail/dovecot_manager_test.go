package mail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dovecotMockExecutor records commands and can simulate failures.
type dovecotMockExecutor struct {
	commands []dovecotExecCommand
	failOn   map[string]error
}

type dovecotExecCommand struct {
	name string
	args []string
}

func newDovecotMockExecutor() *dovecotMockExecutor {
	return &dovecotMockExecutor{
		failOn: make(map[string]error),
	}
}

func (m *dovecotMockExecutor) Run(name string, args ...string) error {
	m.commands = append(m.commands, dovecotExecCommand{name: name, args: args})
	key := name + " " + strings.Join(args, " ")
	if err, ok := m.failOn[key]; ok {
		return err
	}
	return nil
}

func dovecotTestUsers() []domain.User {
	return []domain.User{
		{
			ID:           1,
			Email:        "alice@example.com",
			PasswordHash: "$2b$12$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012",
			DomainID:     1,
			Active:       true,
		},
		{
			ID:           2,
			Email:        "bob@example.com",
			PasswordHash: "$2b$12$zyxwvutsrqponmlkjihgfeZYXWVUTSRQPONMLKJIHGFEDCBA987",
			DomainID:     1,
			Active:       true,
		},
	}
}

func TestApplyUserConfig_WritesPassdbFile(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(configDir, mailRoot, executor)

	users := dovecotTestUsers()
	err := manager.ApplyUserConfig(users)
	require.NoError(t, err)

	passdbPath := filepath.Join(configDir, "passdb")
	content, err := os.ReadFile(passdbPath)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "alice@example.com:{BLF-CRYPT}$2b$12$")
	assert.Contains(t, lines[1], "bob@example.com:{BLF-CRYPT}$2b$12$")
}

func TestApplyUserConfig_WritesUserdbFile(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(configDir, mailRoot, executor)

	users := dovecotTestUsers()
	err := manager.ApplyUserConfig(users)
	require.NoError(t, err)

	userdbPath := filepath.Join(configDir, "userdb")
	content, err := os.ReadFile(userdbPath)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "alice@example.com::5000:5000::/var/mail/vhosts/example.com/alice", lines[0])
	assert.Equal(t, "bob@example.com::5000:5000::/var/mail/vhosts/example.com/bob", lines[1])
}

func TestApplyUserConfig_CallsReloadAndFlush(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(configDir, mailRoot, executor)

	err := manager.ApplyUserConfig(dovecotTestUsers())
	require.NoError(t, err)

	require.Len(t, executor.commands, 2)
	assert.Equal(t, "doveadm", executor.commands[0].name)
	assert.Equal(t, []string{"reload"}, executor.commands[0].args)
	assert.Equal(t, "doveadm", executor.commands[1].name)
	assert.Equal(t, []string{"auth", "cache", "flush"}, executor.commands[1].args)
}

func TestApplyUserConfig_AtomicWrite_NoTmpFileRemains(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(configDir, mailRoot, executor)

	err := manager.ApplyUserConfig(dovecotTestUsers())
	require.NoError(t, err)

	// No .tmp files should remain
	entries, err := os.ReadDir(configDir)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, strings.HasSuffix(entry.Name(), ".tmp"), "temp file should not remain: %s", entry.Name())
	}
}

func TestApplyUserConfig_EmptyUsers(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(configDir, mailRoot, executor)

	err := manager.ApplyUserConfig([]domain.User{})
	require.NoError(t, err)

	passdbContent, err := os.ReadFile(filepath.Join(configDir, "passdb"))
	require.NoError(t, err)
	assert.Empty(t, string(passdbContent))

	userdbContent, err := os.ReadFile(filepath.Join(configDir, "userdb"))
	require.NoError(t, err)
	assert.Empty(t, string(userdbContent))
}

func TestApplyUserConfig_ReloadFailure_ReturnsError(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	executor.failOn["doveadm reload"] = fmt.Errorf("connection refused")
	manager := NewDovecotManager(configDir, mailRoot, executor)

	err := manager.ApplyUserConfig(dovecotTestUsers())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reload dovecot")
}

func TestApplyUserConfig_FlushFailure_ReturnsError(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	executor.failOn["doveadm auth cache flush"] = fmt.Errorf("timeout")
	manager := NewDovecotManager(configDir, mailRoot, executor)

	err := manager.ApplyUserConfig(dovecotTestUsers())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flush auth cache")
}

func TestCreateMaildir_CreatesCorrectStructure(t *testing.T) {
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(t.TempDir(), mailRoot, executor)

	err := manager.CreateMaildir("alice@example.com", "example.com")
	require.NoError(t, err)

	expectedBase := filepath.Join(mailRoot, "example.com", "alice", "Maildir")
	for _, sub := range []string{"cur", "new", "tmp"} {
		dir := filepath.Join(expectedBase, sub)
		info, err := os.Stat(dir)
		require.NoError(t, err, "directory %s should exist", sub)
		assert.True(t, info.IsDir())
	}
}

func TestCreateMaildir_Idempotent(t *testing.T) {
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(t.TempDir(), mailRoot, executor)

	// Call twice — should not error
	err := manager.CreateMaildir("alice@example.com", "example.com")
	require.NoError(t, err)

	err = manager.CreateMaildir("alice@example.com", "example.com")
	require.NoError(t, err)
}

func TestCreateMaildir_MultipleDomainsAndUsers(t *testing.T) {
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(t.TempDir(), mailRoot, executor)

	require.NoError(t, manager.CreateMaildir("alice@example.com", "example.com"))
	require.NoError(t, manager.CreateMaildir("bob@other.org", "other.org"))

	assert.DirExists(t, filepath.Join(mailRoot, "example.com", "alice", "Maildir", "cur"))
	assert.DirExists(t, filepath.Join(mailRoot, "other.org", "bob", "Maildir", "cur"))
}

func TestReload_ExecutesDoveadmReload(t *testing.T) {
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(t.TempDir(), t.TempDir(), executor)

	err := manager.Reload()
	require.NoError(t, err)

	require.Len(t, executor.commands, 1)
	assert.Equal(t, "doveadm", executor.commands[0].name)
	assert.Equal(t, []string{"reload"}, executor.commands[0].args)
}

func TestFlushAuthCache_ExecutesDoveadmFlush(t *testing.T) {
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(t.TempDir(), t.TempDir(), executor)

	err := manager.FlushAuthCache()
	require.NoError(t, err)

	require.Len(t, executor.commands, 1)
	assert.Equal(t, "doveadm", executor.commands[0].name)
	assert.Equal(t, []string{"auth", "cache", "flush"}, executor.commands[0].args)
}

func TestReload_PropagatesExecutorError(t *testing.T) {
	executor := newDovecotMockExecutor()
	executor.failOn["doveadm reload"] = fmt.Errorf("not running")
	manager := NewDovecotManager(t.TempDir(), t.TempDir(), executor)

	err := manager.Reload()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestFlushAuthCache_PropagatesExecutorError(t *testing.T) {
	executor := newDovecotMockExecutor()
	executor.failOn["doveadm auth cache flush"] = fmt.Errorf("socket error")
	manager := NewDovecotManager(t.TempDir(), t.TempDir(), executor)

	err := manager.FlushAuthCache()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "socket error")
}

func TestApplyUserConfig_OverwritesPreviousFiles(t *testing.T) {
	configDir := t.TempDir()
	mailRoot := t.TempDir()
	executor := newDovecotMockExecutor()
	manager := NewDovecotManager(configDir, mailRoot, executor)

	// First apply with two users
	err := manager.ApplyUserConfig(dovecotTestUsers())
	require.NoError(t, err)

	// Second apply with one user (simulating delete)
	singleUser := []domain.User{dovecotTestUsers()[0]}
	err = manager.ApplyUserConfig(singleUser)
	require.NoError(t, err)

	passdbContent, err := os.ReadFile(filepath.Join(configDir, "passdb"))
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(passdbContent)), "\n")
	assert.Len(t, lines, 1)
	assert.Contains(t, lines[0], "alice@example.com")
}
