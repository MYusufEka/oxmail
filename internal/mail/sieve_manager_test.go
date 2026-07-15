package mail

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sieveMockExecutor implements CommandExecutor for SieveManager tests.
type sieveMockExecutor struct {
	commands [][]string
	errs     map[string]error // keyed by "name arg1 arg2..."
	outputs  map[string]string
}

func newSieveMockExecutor() *sieveMockExecutor {
	return &sieveMockExecutor{
		errs:    make(map[string]error),
		outputs: make(map[string]string),
	}
}

func (m *sieveMockExecutor) Run(name string, args ...string) error {
	cmd := append([]string{name}, args...)
	m.commands = append(m.commands, cmd)
	key := name + " " + strings.Join(args, " ")
	if err, ok := m.errs[key]; ok {
		return err
	}
	return nil
}

func (m *sieveMockExecutor) RunWithOutput(name string, args ...string) (string, error) {
	cmd := append([]string{name}, args...)
	m.commands = append(m.commands, cmd)
	key := name + " " + strings.Join(args, " ")
	if err, ok := m.errs[key]; ok {
		return "", err
	}
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return "", nil
}

func TestNewSieveManager(t *testing.T) {
	m := NewSieveManager("/scripts", "/global", newSieveMockExecutor())
	assert.NotNil(t, m)
}

func TestSetScript_CreatesDirWritesAndCompiles(t *testing.T) {
	exec := newSieveMockExecutor()
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	err := mgr.SetScript("user@example.com", "require [\"fileinto\"];")
	require.NoError(t, err)

	// Should have: mkdir -p, sh -c (write), sievec
	require.GreaterOrEqual(t, len(exec.commands), 3)
	assert.Equal(t, "mkdir", exec.commands[0][0])
	assert.Equal(t, "-p", exec.commands[0][1])
	assert.Contains(t, exec.commands[0][2], "example.com")

	assert.Equal(t, "sh", exec.commands[1][0])
	assert.Equal(t, "-c", exec.commands[1][1])
	assert.Contains(t, exec.commands[1][2], "cat > '/var/lib/sieve/scripts/example.com/user.sieve'")

	assert.Equal(t, "sievec", exec.commands[2][0])
	assert.Equal(t, "/var/lib/sieve/scripts/example.com/user.sieve", exec.commands[2][1])
}

func TestSetScript_InvalidEmail(t *testing.T) {
	exec := newSieveMockExecutor()
	mgr := NewSieveManager("/scripts", "/global", exec)

	err := mgr.SetScript("notanemail", "content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email")
}

func TestSetScript_EmptyLocalPart(t *testing.T) {
	exec := newSieveMockExecutor()
	mgr := NewSieveManager("/scripts", "/global", exec)

	err := mgr.SetScript("@domain.com", "content")
	assert.Error(t, err)
}

func TestGetScript_ReturnsContent(t *testing.T) {
	exec := newSieveMockExecutor()
	exec.outputs["cat /var/lib/sieve/scripts/example.com/user.sieve"] = "require [\"fileinto\"];\n"
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	script, err := mgr.GetScript("user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "require [\"fileinto\"];\n", script)
}

func TestGetScript_NotExists_ReturnsEmpty(t *testing.T) {
	exec := newSieveMockExecutor()
	exec.errs["test -f /var/lib/sieve/scripts/example.com/user.sieve"] = fmt.Errorf("not found")
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	script, err := mgr.GetScript("user@example.com")
	require.NoError(t, err)
	assert.Empty(t, script)
}

func TestGetScript_InvalidEmail(t *testing.T) {
	mgr := NewSieveManager("/scripts", "/global", newSieveMockExecutor())
	_, err := mgr.GetScript("bademail")
	assert.Error(t, err)
}

func TestDeleteScript_RemovesFiles(t *testing.T) {
	exec := newSieveMockExecutor()
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	err := mgr.DeleteScript("user@example.com")
	require.NoError(t, err)

	// Should have rm -f for .sieve and .svbin
	require.GreaterOrEqual(t, len(exec.commands), 2)
	assert.Equal(t, "rm", exec.commands[0][0])
	assert.Equal(t, "-f", exec.commands[0][1])
	assert.Contains(t, exec.commands[0][2], "user.sieve")
	assert.Equal(t, "rm", exec.commands[1][0])
	assert.Contains(t, exec.commands[1][2], "user.sieve.svbin")
}

func TestDeleteScript_InvalidEmail(t *testing.T) {
	mgr := NewSieveManager("/scripts", "/global", newSieveMockExecutor())
	err := mgr.DeleteScript("bademail")
	assert.Error(t, err)
}

func TestActivateScript_CompilesExistingScript(t *testing.T) {
	exec := newSieveMockExecutor()
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	err := mgr.ActivateScript("user@example.com")
	require.NoError(t, err)

	// Should have: test -f, sievec
	require.GreaterOrEqual(t, len(exec.commands), 2)
	assert.Equal(t, "test", exec.commands[0][0])
	assert.Equal(t, "sievec", exec.commands[1][0])
}

func TestActivateScript_NotExists_ReturnsError(t *testing.T) {
	exec := newSieveMockExecutor()
	exec.errs["test -f /var/lib/sieve/scripts/example.com/user.sieve"] = fmt.Errorf("not found")
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	err := mgr.ActivateScript("user@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script not found")
}

func TestActivateScript_InvalidEmail(t *testing.T) {
	mgr := NewSieveManager("/scripts", "/global", newSieveMockExecutor())
	err := mgr.ActivateScript("bademail")
	assert.Error(t, err)
}

func TestSetGlobalScript_WritesToGlobalDir(t *testing.T) {
	exec := newSieveMockExecutor()
	mgr := NewSieveManager("/var/lib/sieve/scripts", "/var/lib/sieve/global", exec)

	err := mgr.SetGlobalScript("spam-global", "require [\"fileinto\"];\n")
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(exec.commands), 3)

	assert.Equal(t, "mkdir", exec.commands[0][0])
	assert.Equal(t, "-p", exec.commands[0][1])
	assert.Equal(t, "/var/lib/sieve/global", exec.commands[0][2])

	assert.Equal(t, "sh", exec.commands[1][0])
	assert.Contains(t, exec.commands[1][2], "cat > '/var/lib/sieve/global/spam-global.sieve'")

	assert.Equal(t, "sievec", exec.commands[2][0])
	assert.Equal(t, "/var/lib/sieve/global/spam-global.sieve", exec.commands[2][1])
}

func TestSpamGlobalSieveScript_ContainsFileinto(t *testing.T) {
	assert.Contains(t, SpamGlobalSieveScript, "fileinto \"Junk\"")
	assert.Contains(t, SpamGlobalSieveScript, "X-Spam-Flag")
	assert.Contains(t, SpamGlobalSieveScript, "X-Spam-Score")
	assert.Contains(t, SpamGlobalSieveScript, "stop")
}
