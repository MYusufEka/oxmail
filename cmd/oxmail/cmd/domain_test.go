package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDomainAddCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/domains" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		newTestResponse(w, http.StatusCreated, map[string]string{"name": "test.com"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "add", "test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Domain") || !strings.Contains(stderr, "test.com") {
		t.Errorf("expected success message in stderr, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestDomainAddCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusCreated, map[string]string{"name": "test.com"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "add", "test.com", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "test.com") {
		t.Errorf("expected JSON output in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestDomainAddCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("domain", "add")
	requireCmdErr(t, err, "domain add requires exactly 1 arg")
}

func TestDomainAddCmd_ExtraArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("domain", "add", "a.com", "b.com")
	requireCmdErr(t, err, "domain add requires exactly 1 arg")
}

func TestDomainListCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, DomainListResponse{
			Data: []Domain{
				{Name: "example.com", CreatedAt: "2024-01-01T00:00:00Z"},
				{Name: "test.org", CreatedAt: "2024-02-01T00:00:00Z"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "example.com") || !strings.Contains(stdout, "test.org") {
		t.Errorf("expected domains in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestDomainListCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, DomainListResponse{
			Data: []Domain{
				{Name: "example.com", CreatedAt: "2024-01-01T00:00:00Z"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "example.com") {
		t.Errorf("expected JSON output in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestDomainListCmd_Empty(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, DomainListResponse{Data: []Domain{}})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, stderr, "No domains found")
	if stdout != "" {
		t.Errorf("expected empty stdout for empty list, got: %s", stdout)
	}
}

func TestDomainRmCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/domains/test.com" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "rm", "test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "test.com") {
		t.Errorf("expected domain name in output, got stderr=%q", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestDomainRmCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("domain", "rm", "test.com", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
	_ = stdout
}

func TestDomainRmCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("domain", "rm")
	requireCmdErr(t, err, "domain rm requires exactly 1 arg")
}

func TestDomainCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("domain", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Manage mail domains")
	assertContains(t, output, "add")
	assertContains(t, output, "list")
	assertContains(t, output, "rm")
}

func TestDomainListCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("domain", "list", "extra")
	requireCmdErr(t, err, "domain list accepts no args")
}
