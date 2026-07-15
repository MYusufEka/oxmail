package mail

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
)

// SpamGlobalSieveScript files mail flagged by Rspamd (X-Spam-Flag / X-Spam-Score
// headers set by the Postfix milter) into Junk before per-user scripts run.
const SpamGlobalSieveScript = `require ["fileinto", "mailbox", "relational", "comparator-i;ascii-numeric"];
if anyof (
  header :contains "X-Spam-Flag" "YES",
  header :value "ge" :comparator "i;ascii-numeric" "X-Spam-Score" "5"
) {
  fileinto "Junk";
  stop;
}
`

// SieveManager manages Sieve scripts for Dovecot.
// Writes/reads/compiles scripts inside the dovecot container via docker exec.
type SieveManager struct {
	scriptDir string // /var/lib/sieve/scripts
	globalDir string // /var/lib/sieve/global
	executor  CommandExecutor
}

// NewSieveManager creates a SieveManager with paths and executor.
func NewSieveManager(scriptDir, globalDir string, executor CommandExecutor) *SieveManager {
	return &SieveManager{
		scriptDir: scriptDir,
		globalDir: globalDir,
		executor:  executor,
	}
}

// scriptPath returns the full .sieve path and its parent directory for an email.
func (m *SieveManager) scriptPath(email string) (sievePath, dir string, err error) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid email: %s", email)
	}
	domain := parts[1]
	localPart := parts[0]
	dir = filepath.Join(m.scriptDir, domain)
	sievePath = filepath.Join(dir, localPart+".sieve")
	return sievePath, dir, nil
}

// SetScript writes a sieve script for the given email and compiles it.
// If the script content is empty, the script is removed instead.
func (m *SieveManager) SetScript(email, scriptContent string) error {
	path, dir, err := m.scriptPath(email)
	if err != nil {
		return fmt.Errorf("sieve set: %w", err)
	}

	// Create directory inside dovecot container
	if err := m.executor.Run("mkdir", "-p", dir); err != nil {
		return fmt.Errorf("sieve mkdir: %w", err)
	}

	// Write file via heredoc to avoid shell escaping issues.
	// Quoted delimiter prevents ALL shell expansion — content taken literally.
	delim := fmt.Sprintf("SIEVE_SCRIPT_%d", rand.Int63())
	cmd := fmt.Sprintf("cat > '%s' << '%s'\n%s\n%s", path, delim, scriptContent, delim)
	if err := m.executor.Run("sh", "-c", cmd); err != nil {
		return fmt.Errorf("sieve write: %w", err)
	}

	// Compile via sievec inside container
	if err := m.executor.Run("sievec", path); err != nil {
		return fmt.Errorf("sieve compile: %w", err)
	}

	return nil
}

// GetScript reads the sieve script content for the given email.
// Returns empty string if no script exists (no error).
func (m *SieveManager) GetScript(email string) (string, error) {
	path, _, err := m.scriptPath(email)
	if err != nil {
		return "", fmt.Errorf("sieve get: %w", err)
	}

	// Check if file exists first
	if err := m.executor.Run("test", "-f", path); err != nil {
		return "", nil // script doesn't exist, return empty
	}

	out, err := m.executor.RunWithOutput("cat", path)
	if err != nil {
		return "", fmt.Errorf("sieve read: %w", err)
	}

	return out, nil
}

// DeleteScript removes the sieve script and its compiled binary.
func (m *SieveManager) DeleteScript(email string) error {
	path, _, err := m.scriptPath(email)
	if err != nil {
		return fmt.Errorf("sieve delete: %w", err)
	}

	binPath := path + ".svbin"

	// Remove both files; ignore errors if they don't exist
	m.executor.Run("rm", "-f", path)
	m.executor.Run("rm", "-f", binPath)

	return nil
}

// globalScriptPath returns the full .sieve path for a named global script.
func (m *SieveManager) globalScriptPath(name string) string {
	return filepath.Join(m.globalDir, name+".sieve")
}

// SetGlobalScript writes a global sieve script (applied to every user via
// sieve_before/sieve_after in the Dovecot plugin config) and compiles it.
func (m *SieveManager) SetGlobalScript(name, scriptContent string) error {
	path := m.globalScriptPath(name)

	// Create the global scripts directory inside the dovecot container
	if err := m.executor.Run("mkdir", "-p", m.globalDir); err != nil {
		return fmt.Errorf("sieve global mkdir: %w", err)
	}

	// Write file via heredoc to avoid shell escaping issues.
	delim := fmt.Sprintf("SIEVE_SCRIPT_%d", rand.Int63())
	cmd := fmt.Sprintf("cat > '%s' << '%s'\n%s\n%s", path, delim, scriptContent, delim)
	if err := m.executor.Run("sh", "-c", cmd); err != nil {
		return fmt.Errorf("sieve global write: %w", err)
	}

	// Compile via sievec inside container
	if err := m.executor.Run("sievec", path); err != nil {
		return fmt.Errorf("sieve global compile: %w", err)
	}

	return nil
}

// vacationScriptPath returns the .sieve path for a user's vacation script.
func (m *SieveManager) vacationScriptPath(email string) (sievePath, dir string, err error) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid email: %s", email)
	}
	dir = filepath.Join(m.scriptDir, parts[1])
	sievePath = filepath.Join(dir, parts[0]+".vacation.sieve")
	return sievePath, dir, nil
}

// SetVacation writes a vacation auto-reply sieve script for the given email and compiles it.
func (m *SieveManager) SetVacation(email, subject, body string, enabled bool) error {
	path, dir, err := m.vacationScriptPath(email)
	if err != nil {
		return fmt.Errorf("sieve vacation set: %w", err)
	}

	if !enabled {
		return m.DeleteVacation(email)
	}

	if err := m.executor.Run("mkdir", "-p", dir); err != nil {
		return fmt.Errorf("sieve vacation mkdir: %w", err)
	}

	script := fmt.Sprintf(`require ["vacation"];
vacation :days 1 :subject "%s" "%s";
`, subject, body)

	delim := fmt.Sprintf("SIEVE_VACATION_%d", rand.Int63())
	cmd := fmt.Sprintf("cat > '%s' << '%s'\n%s\n%s", path, delim, script, delim)
	if err := m.executor.Run("sh", "-c", cmd); err != nil {
		return fmt.Errorf("sieve vacation write: %w", err)
	}

	if err := m.executor.Run("sievec", path); err != nil {
		return fmt.Errorf("sieve vacation compile: %w", err)
	}

	return nil
}

// GetVacation reads the vacation script for the given email.
// Returns enabled=false and empty fields if no vacation script exists.
func (m *SieveManager) GetVacation(email string) (subject, body string, enabled bool, err error) {
	path, _, err := m.vacationScriptPath(email)
	if err != nil {
		return "", "", false, fmt.Errorf("sieve vacation get: %w", err)
	}

	if err := m.executor.Run("test", "-f", path); err != nil {
		return "", "", false, nil
	}

	out, err := m.executor.RunWithOutput("cat", path)
	if err != nil {
		return "", "", false, fmt.Errorf("sieve vacation read: %w", err)
	}

	return out, "", true, nil
}

// DeleteVacation removes the vacation sieve script and its compiled binary.
func (m *SieveManager) DeleteVacation(email string) error {
	path, _, err := m.vacationScriptPath(email)
	if err != nil {
		return fmt.Errorf("sieve vacation delete: %w", err)
	}

	m.executor.Run("rm", "-f", path)
	m.executor.Run("rm", "-f", path+".svbin")

	return nil
}

// ActivateScript runs sievec to ensure the script is compiled and active.
// Returns error if the script doesn't exist.
func (m *SieveManager) ActivateScript(email string) error {
	path, _, err := m.scriptPath(email)
	if err != nil {
		return fmt.Errorf("sieve activate: %w", err)
	}

	if err := m.executor.Run("test", "-f", path); err != nil {
		return fmt.Errorf("sieve activate: script not found for %s", email)
	}

	if err := m.executor.Run("sievec", path); err != nil {
		return fmt.Errorf("sieve activate compile: %w", err)
	}

	return nil
}
