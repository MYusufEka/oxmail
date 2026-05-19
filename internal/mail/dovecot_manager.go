package mail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// DovecotManager manages Dovecot user configuration and maildir lifecycle.
type DovecotManager struct {
	configDir string
	mailRoot  string
	executor  CommandExecutor
	uid       int
	gid       int
}

// NewDovecotManager creates a DovecotManager with the given configuration.
func NewDovecotManager(configDir, mailRoot string, executor CommandExecutor) *DovecotManager {
	return &DovecotManager{
		configDir: configDir,
		mailRoot:  mailRoot,
		executor:  executor,
		uid:       5000,
		gid:       5000,
	}
}

// ApplyUserConfig writes passdb and userdb files atomically, then reloads Dovecot and flushes auth cache.
func (m *DovecotManager) ApplyUserConfig(users []domain.User) error {
	if err := m.writePassdb(users); err != nil {
		return fmt.Errorf("write passdb: %w", err)
	}
	if err := m.writeUserdb(users); err != nil {
		return fmt.Errorf("write userdb: %w", err)
	}
	if err := m.Reload(); err != nil {
		return fmt.Errorf("reload dovecot: %w", err)
	}
	if err := m.FlushAuthCache(); err != nil {
		return fmt.Errorf("flush auth cache: %w", err)
	}
	return nil
}

// CreateMaildir creates the Maildir structure for a user with correct permissions.
func (m *DovecotManager) CreateMaildir(email string, domainName string) error {
	localPart := extractLocalPart(email)
	maildirBase := filepath.Join(m.mailRoot, domainName, localPart, "Maildir")

	for _, sub := range []string{"cur", "new", "tmp"} {
		dir := filepath.Join(maildirBase, sub)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create maildir %s: %w", sub, err)
		}
	}

	return nil
}

// Reload signals Dovecot to reload its configuration.
func (m *DovecotManager) Reload() error {
	return m.executor.Run("doveadm", "reload")
}

// FlushAuthCache clears the Dovecot authentication cache.
func (m *DovecotManager) FlushAuthCache() error {
	return m.executor.Run("doveadm", "auth", "cache", "flush")
}

func (m *DovecotManager) writePassdb(users []domain.User) error {
	if err := os.MkdirAll(m.configDir, 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	var builder strings.Builder
	for _, user := range users {
		fmt.Fprintf(&builder, "%s:{BLF-CRYPT}%s\n", user.Email, user.PasswordHash)
	}

	return atomicWrite(filepath.Join(m.configDir, "passdb"), []byte(builder.String()), 0640)
}

func (m *DovecotManager) writeUserdb(users []domain.User) error {
	if err := os.MkdirAll(m.configDir, 0750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	var builder strings.Builder
	for _, user := range users {
		localPart, domainPart := splitEmail(user.Email)
		home := fmt.Sprintf("/var/mail/vhosts/%s/%s", domainPart, localPart)
		fmt.Fprintf(&builder, "%s::5000:5000::%s\n", user.Email, home)
	}

	return atomicWrite(filepath.Join(m.configDir, "userdb"), []byte(builder.String()), 0640)
}

// atomicWrite writes data to a temp file then renames it to the target path.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp to target: %w", err)
	}
	return nil
}

func extractLocalPart(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}
	return parts[0]
}

func splitEmail(email string) (local, domainPart string) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email, ""
	}
	return parts[0], parts[1]
}
