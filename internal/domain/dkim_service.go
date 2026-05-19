package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DKIMService manages DKIM key generation, storage, and retrieval.
type DKIMService struct {
	keyBasePath string
	mu          sync.RWMutex
	keys        map[string]*DKIMKey // keyed by "domain/selector"
}

// NewDKIMService creates a new DKIMService with the given base path for key storage.
func NewDKIMService(keyBasePath string) *DKIMService {
	return &DKIMService{
		keyBasePath: keyBasePath,
		keys:        make(map[string]*DKIMKey),
	}
}

// Generate creates a new RSA 2048-bit DKIM key pair for the given domain and selector.
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

	dkimKey := &DKIMKey{
		Domain:    domain,
		Selector:  selector,
		PublicKey: publicKeyPEM,
		DNSRecord: dnsRecord,
		CreatedAt: time.Now().UTC(),
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

	if !exists {
		return nil, fmt.Errorf("DKIM key not found for %s/%s", domain, selector)
	}

	return key, nil
}

// Delete removes the DKIM key for the given domain and selector.
func (s *DKIMService) Delete(domain, selector string) error {
	s.mu.RLock()
	_, exists := s.keys[keyID(domain, selector)]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("DKIM key not found for %s/%s", domain, selector)
	}

	keyPath := filepath.Join(s.keyBasePath, domain, selector+".private")
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove private key file: %w", err)
	}

	s.mu.Lock()
	delete(s.keys, keyID(domain, selector))
	s.mu.Unlock()

	return nil
}

// Rotate deletes the existing key and generates a new one.
func (s *DKIMService) Rotate(domain, selector string) (*DKIMKey, error) {
	s.mu.RLock()
	_, exists := s.keys[keyID(domain, selector)]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("DKIM key not found for %s/%s: cannot rotate non-existent key", domain, selector)
	}

	if err := s.Delete(domain, selector); err != nil {
		return nil, fmt.Errorf("delete old key during rotation: %w", err)
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
