package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAliasAddCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/aliases" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		newTestResponse(w, http.StatusCreated, map[string]string{"id": "1"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("alias", "add", "info@test.com", "admin@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "info@test.com") || !strings.Contains(stderr, "admin@test.com") {
		t.Errorf("expected alias info in stderr, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestAliasAddCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusCreated, map[string]string{"id": "1"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("alias", "add", "info@test.com", "admin@test.com", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"id"`) {
		t.Errorf("expected JSON output in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestAliasAddCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("alias", "add", "info@test.com")
	requireCmdErr(t, err, "alias add requires exactly 2 args")
}

func TestAliasAddCmd_NoArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("alias", "add")
	requireCmdErr(t, err, "alias add requires exactly 2 args")
}

func TestAliasListCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, AliasListResponse{
			Data: []Alias{
				{ID: "1", Source: "info@test.com", Destination: "admin@test.com", CreatedAt: "2024-01-01T00:00:00Z"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("alias", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "info@test.com") || !strings.Contains(stdout, "admin@test.com") {
		t.Errorf("expected aliases in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestAliasListCmd_Empty(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, AliasListResponse{Data: []Alias{}})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("alias", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, stderr, "No aliases found")
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestAliasListCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("alias", "list", "extra")
	requireCmdErr(t, err, "alias list accepts no args")
}

func TestAliasRmCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/aliases/1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("alias", "rm", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "1") {
		t.Errorf("expected alias ID in stderr, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestAliasRmCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("alias", "rm", "1", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
	_ = stdout
}

func TestAliasRmCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("alias", "rm")
	requireCmdErr(t, err, "alias rm requires exactly 1 arg")
}

func TestAliasCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("alias", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Manage mail aliases")
	assertContains(t, output, "add")
	assertContains(t, output, "list")
	assertContains(t, output, "rm")
}
