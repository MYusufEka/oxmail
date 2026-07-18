package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserAddCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/users" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		newTestResponse(w, http.StatusCreated, map[string]string{"email": "user@test.com"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "add", "user@test.com", "--password", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "user@test.com") {
		t.Errorf("expected email in stderr, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestUserAddCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusCreated, map[string]string{"email": "user@test.com"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "add", "user@test.com", "--password", "secret123", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "user@test.com") {
		t.Errorf("expected JSON output in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestUserAddCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("user", "add")
	requireCmdErr(t, err, "user add requires exactly 1 arg")
}

func TestUserAddCmd_PasswordFlagExists(t *testing.T) {
	flag := userAddCmd.Flags().Lookup("password")
	if flag == nil {
		t.Fatal("expected --password flag on user add")
	}
}

func TestUserListCmd_Success(t *testing.T) {
	resetFlags()
	userDomainFilter = ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, UserListResponse{
			Data: []User{
				{Email: "alice@test.com", Domain: "test.com", CreatedAt: "2024-01-01T00:00:00Z"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "alice@test.com") {
		t.Errorf("expected user in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestUserListCmd_DomainFilter(t *testing.T) {
	resetFlags()
	userDomainFilter = ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "domain=test.com") {
			t.Errorf("expected domain filter in query, got: %s", r.URL.RawQuery)
		}
		newTestResponse(w, http.StatusOK, UserListResponse{
			Data: []User{
				{Email: "alice@test.com", Domain: "test.com"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "list", "--domain", "test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "alice@test.com") {
		t.Errorf("expected filtered user in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestUserListCmd_Empty(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, UserListResponse{Data: []User{}})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, stderr, "No users found")
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestUserListCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("user", "list", "extra")
	requireCmdErr(t, err, "user list accepts no args")
}

func TestUserRmCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/users/user@test.com" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "rm", "user@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "user@test.com") {
		t.Errorf("expected email in stderr, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestUserRmCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("user", "rm", "user@test.com", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
	_ = stdout
}

func TestUserRmCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("user", "rm")
	requireCmdErr(t, err, "user rm requires exactly 1 arg")
}

func TestUserCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("user", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Manage mail users")
	assertContains(t, output, "add")
	assertContains(t, output, "list")
	assertContains(t, output, "rm")
}

func TestUserListCmd_DomainFlagExists(t *testing.T) {
	flag := userListCmd.Flags().Lookup("domain")
	if flag == nil {
		t.Fatal("expected --domain flag on user list")
	}
}
