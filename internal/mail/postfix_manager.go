package mail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// PostfixConfig holds paths for Postfix config files.
type PostfixConfig struct {
	DomainsPath string
	AliasesPath string
}

// PostfixManager applies generated configs to a running Postfix instance.
type PostfixManager struct {
	config     PostfixConfig
	executor   CommandExecutor
	retryDelay time.Duration
	maxRetries int
}

// NewPostfixManager creates a PostfixManager with the given config and executor.
func NewPostfixManager(cfg PostfixConfig, executor CommandExecutor) *PostfixManager {
	return &PostfixManager{
		config:     cfg,
		executor:   executor,
		retryDelay: 1 * time.Second,
		maxRetries: 3,
	}
}

// ApplyDomainConfig writes the virtual_domains file and reloads Postfix.
// virtual_mailbox_domains is a plain file — no postmap needed.
func (m *PostfixManager) ApplyDomainConfig(domains []domain.Domain) error {
	var builder strings.Builder
	for _, d := range domains {
		if !d.Active {
			continue
		}
		builder.WriteString(d.Name)
		builder.WriteByte('\n')
	}

	dir := filepath.Dir(m.config.DomainsPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := atomicWrite(m.config.DomainsPath, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write virtual_domains: %w", err)
	}

	if err := m.Postmap(m.config.DomainsPath); err != nil {
		return err
	}

	return m.Reload()
}

// ApplyAliasConfig writes the virtual_aliases file and reloads Postfix.
// virtual_alias_maps is a plain file — no postmap needed.
func (m *PostfixManager) ApplyAliasConfig(aliases []domain.Alias) error {
	destsBySource := make(map[string][]string)
	sourceOrder := make([]string, 0)
	for _, alias := range aliases {
		if !alias.Active {
			continue
		}
		if _, seen := destsBySource[alias.SourceAddress]; !seen {
			sourceOrder = append(sourceOrder, alias.SourceAddress)
		}
		destsBySource[alias.SourceAddress] = append(destsBySource[alias.SourceAddress], alias.DestinationAddress)
	}

	var builder strings.Builder
	for _, src := range sourceOrder {
		builder.WriteString(src)
		builder.WriteByte(' ')
		builder.WriteString(strings.Join(destsBySource[src], ","))
		builder.WriteByte('\n')
	}

	dir := filepath.Dir(m.config.AliasesPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := atomicWrite(m.config.AliasesPath, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write virtual_aliases: %w", err)
	}

	if err := m.Postmap(m.config.AliasesPath); err != nil {
		return err
	}

	return m.Reload()
}

// Postmap runs `postmap hash:<file>` to rebuild the lookup table.
func (m *PostfixManager) Postmap(file string) error {
	if err := m.executor.Run("postmap", "hash:"+file); err != nil {
		return fmt.Errorf("postmap %s: %w", file, err)
	}
	return nil
}

// Reload sends a reload signal to Postfix. Retries up to maxRetries times on failure.
func (m *PostfixManager) Reload() error {
	var lastErr error
	for attempt := 0; attempt < m.maxRetries; attempt++ {
		if err := m.executor.Run("postfix", "reload"); err != nil {
			lastErr = err
			if attempt < m.maxRetries-1 {
				time.Sleep(m.retryDelay)
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("postfix reload failed after %d attempts: %w", m.maxRetries, lastErr)
}
