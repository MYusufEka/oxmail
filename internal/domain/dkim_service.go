package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DKIMService manages DKIM key generation, storage, and retrieval.
// Keys are persisted to both SQLite (dkim_keys table) and PEM files on disk.
type DKIMService struct {
	db          *sql.DB
	keyBasePath string
	mu          sync.RWMutex
	keys        map[string]*DKIMKey // keyed by "domain/selector"
}

// NewDKIMService creates a new DKIMService backed by SQLite and filesystem.
// Pass nil for db in tests (keys stored in-memory only).
func NewDKIMService(db *sql.DB, keyBasePath string) *DKIMService {
	s := &DKIMService{
		db:          db,
		keyBasePath: keyBasePath,
		keys:        make(map[string]*DKIMKey),
	}
	if db != nil {
		s.loadFromDB()
	}
	return s
}

// loadFromDB loads all existing DKIM keys from SQLite into memory.
func (s *DKIMService) loadFromDB() {
	if s.db == nil {
		return
	}
	rows, err := s.db.Query("SELECT domain, selector, public_key_pem, created_at FROM dkim_keys")
	if err != nil {
		return
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var key DKIMKey
		var createdAt time.Time
		if err := rows.Scan(&key.Domain, &key.Selector, &key.PublicKey, &createdAt); err != nil {
			continue
		}
		key.CreatedAt = createdAt
		key.DNSRecord = buildDNSRecord(key.PublicKey)
		s.keys[keyID(key.Domain, key.Selector)] = &key
	}
}

// Generate creates a new RSA 2048-bit DKIM key pair, persisted to SQLite + disk.
func (s *DKIMService) Generate(domain, selector string) (*DKIMKey, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain must not be empty")
	}
	if selector == "" {
		return nil, fmt.Errorf("selector must not be empty")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}

	if err := s.storePrivateKey(domain, selector, privateKey); err != nil {
		return nil, fmt.Errorf("store private key: %w", err)
	}

	publicKeyPEM, err := marshalPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}

	dnsRecord := buildDNSRecord(publicKeyPEM)
	now := time.Now().UTC()

	dkimKey := &DKIMKey{
		Domain:    domain,
		Selector:  selector,
		PublicKey: publicKeyPEM,
		DNSRecord: dnsRecord,
		CreatedAt: now,
	}

	// Persist to SQLite
	if s.db != nil {
		_, err = s.db.Exec(
			"INSERT OR REPLACE INTO dkim_keys (domain, selector, public_key_pem, created_at) VALUES (?, ?, ?, ?)",
			domain, selector, publicKeyPEM, now,
		)
		if err != nil {
			return nil, fmt.Errorf("persist DKIM key: %w", err)
		}
	}

	s.mu.Lock()
	s.keys[keyID(domain, selector)] = dkimKey
	s.mu.Unlock()

	return dkimKey, nil
}

// Get retrieves the DKIM key info for the given domain and selector.
func (s *DKIMService) Get(domain, selector string) (*DKIMKey, error) {
	s.mu.RLock()
	key, exists := s.keys[keyID(domain, selector)]
	s.mu.RUnlock()

	if exists {
		return key, nil
	}

	if s.db == nil {
		return nil, fmt.Errorf("DKIM key not found for %s/%s", domain, selector)
	}

	var dbKey DKIMKey
	var createdAt time.Time
	err := s.db.QueryRow(
		"SELECT domain, selector, public_key_pem, created_at FROM dkim_keys WHERE domain = ? AND selector = ?",
		domain, selector,
	).Scan(&dbKey.Domain, &dbKey.Selector, &dbKey.PublicKey, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("DKIM key not found for %s/%s", domain, selector)
	}

	dbKey.CreatedAt = createdAt
	dbKey.DNSRecord = buildDNSRecord(dbKey.PublicKey)

	s.mu.Lock()
	s.keys[keyID(domain, selector)] = &dbKey
	s.mu.Unlock()

	return &dbKey, nil
}

// Delete removes the DKIM key for the given domain and selector.
func (s *DKIMService) Delete(domain, selector string) error {
	s.mu.Lock()
	_, exists := s.keys[keyID(domain, selector)]
	delete(s.keys, keyID(domain, selector))
	s.mu.Unlock()

	if s.db == nil {
		if !exists {
			return fmt.Errorf("DKIM key not found for %s/%s", domain, selector)
		}
		keyPath := filepath.Join(s.keyBasePath, domain, selector+".private")
		if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove private key file: %w", err)
		}
		return nil
	}

	_, err := s.db.Exec("DELETE FROM dkim_keys WHERE domain = ? AND selector = ?", domain, selector)
	if err != nil {
		return fmt.Errorf("delete DKIM key from DB: %w", err)
	}

	// Remove private key file
	keyPath := filepath.Join(s.keyBasePath, domain, selector+".private")
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove private key file: %w", err)
	}

	return nil
}

func (s *DKIMService) Rotate(domain, selector string) (*DKIMKey, error) {
	s.mu.RLock()
	_, exists := s.keys[keyID(domain, selector)]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("DKIM key not found for %s/%s", domain, selector)
	}

	if err := s.Delete(domain, selector); err != nil {
		return nil, fmt.Errorf("rotate delete: %w", err)
	}

	return s.Generate(domain, selector)
}

func (s *DKIMService) storePrivateKey(domain, selector string, key *rsa.PrivateKey) error {
	dirPath := filepath.Join(s.keyBasePath, domain)
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	keyPath := filepath.Join(dirPath, selector+".private")

	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	pemData := pem.EncodeToMemory(pemBlock)

	if err := os.WriteFile(keyPath, pemData, 0600); err != nil {
		return fmt.Errorf("write private key file: %w", err)
	}

	return nil
}

func marshalPublicKey(pub *rsa.PublicKey) (string, error) {
	derBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("marshal PKIX public key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(derBytes), nil
}

func buildDNSRecord(publicKeyBase64 string) string {
	return fmt.Sprintf("v=DKIM1; k=rsa; p=%s", publicKeyBase64)
}

func keyID(domain, selector string) string {
	return domain + "/" + selector
}
