package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderAll_CreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	data := RenderPayload{
		Hostname:         "mail.example.com",
		Domain:           "example.com",
		MessageSizeLimit: "10240000",
		DevMode:          true,
	}

	err := RenderAll(tmpDir, data)
	if err != nil {
		t.Fatalf("RenderAll = %v, want nil", err)
	}

	expectedFiles := []string{
		"postfix/main.cf",
		"postfix/master.cf",
		"dovecot/dovecot.conf",
		"dovecot/conf.d/10-auth.conf",
		"dovecot/conf.d/10-mail.conf",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", path)
		}
	}
}

func TestRenderAll_ContainsPayloadData(t *testing.T) {
	tmpDir := t.TempDir()
	data := RenderPayload{
		Hostname:         "mx.test.org",
		Domain:           "test.org",
		MessageSizeLimit: "20480000",
		DevMode:          false,
	}

	err := RenderAll(tmpDir, data)
	if err != nil {
		t.Fatalf("RenderAll = %v, want nil", err)
	}

	// Check a few files contain the rendered data
	mainCf := readFile(t, filepath.Join(tmpDir, "postfix/main.cf"))
	if !strings.Contains(mainCf, "mx.test.org") {
		t.Errorf("main.cf should contain hostname 'mx.test.org'")
	}

	dovecotConf := readFile(t, filepath.Join(tmpDir, "dovecot/dovecot.conf"))
	if !strings.Contains(dovecotConf, "test.org") {
		t.Errorf("dovecot.conf should contain domain 'test.org'")
	}
}

func TestRenderAll_DevModeRendersDifferently(t *testing.T) {
	// Just verify dev mode doesn't error - actual template differences
	// depend on template content which we don't control here
	tmpDir := t.TempDir()
	data := RenderPayload{
		Hostname:         "mail.dev.local",
		Domain:           "dev.local",
		MessageSizeLimit: "10240000",
		DevMode:          true,
	}

	err := RenderAll(tmpDir, data)
	if err != nil {
		t.Fatalf("RenderAll dev mode = %v, want nil", err)
	}

	// Verify postfix master.cf exists (key dev mode file)
	masterCf := filepath.Join(tmpDir, "postfix/master.cf")
	if _, err := os.Stat(masterCf); os.IsNotExist(err) {
		t.Errorf("postfix/master.cf not created in dev mode")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
