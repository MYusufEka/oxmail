package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// DovecotUsersGenerator generates passdb and userdb files for Dovecot.
type DovecotUsersGenerator struct {
	outputDir string
}

// NewDovecotUsersGenerator creates a generator that writes to the given directory.
func NewDovecotUsersGenerator(outputDir string) *DovecotUsersGenerator {
	return &DovecotUsersGenerator{outputDir: outputDir}
}

// GenerateAll writes both passdb and userdb files.
func (g *DovecotUsersGenerator) GenerateAll(users []domain.User) error {
	if err := g.GeneratePassdb(users); err != nil {
		return fmt.Errorf("generate passdb: %w", err)
	}
	if err := g.GenerateUserdb(users); err != nil {
		return fmt.Errorf("generate userdb: %w", err)
	}
	return nil
}

// GeneratePassdb writes the Dovecot passdb file.
// Format: user@domain:{BLF-CRYPT}$2b$12$...
func (g *DovecotUsersGenerator) GeneratePassdb(users []domain.User) error {
	if err := os.MkdirAll(g.outputDir, 0750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	var builder strings.Builder
	for _, user := range users {
		fmt.Fprintf(&builder, "%s:{BLF-CRYPT}%s\n", user.Email, user.PasswordHash)
	}

	path := filepath.Join(g.outputDir, "passdb")
	if err := os.WriteFile(path, []byte(builder.String()), 0640); err != nil {
		return fmt.Errorf("write passdb: %w", err)
	}

	return nil
}

// GenerateUserdb writes the Dovecot userdb file.
// Format: user@domain::5000:5000::/var/mail/vhosts/domain/user
func (g *DovecotUsersGenerator) GenerateUserdb(users []domain.User) error {
	if err := os.MkdirAll(g.outputDir, 0750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	var builder strings.Builder
	for _, user := range users {
		localPart, domainPart := splitEmail(user.Email)
		home := fmt.Sprintf("/var/mail/vhosts/%s/%s", domainPart, localPart)
		fmt.Fprintf(&builder, "%s::5000:5000::%s\n", user.Email, home)
	}

	path := filepath.Join(g.outputDir, "userdb")
	if err := os.WriteFile(path, []byte(builder.String()), 0640); err != nil {
		return fmt.Errorf("write userdb: %w", err)
	}

	return nil
}

func splitEmail(email string) (local, domainPart string) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email, ""
	}
	return parts[0], parts[1]
}
