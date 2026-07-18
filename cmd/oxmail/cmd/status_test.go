package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		newTestResponse(w, http.StatusOK, HealthResponse{
			Data: []ServiceHealth{
				{Name: "postfix", Status: "healthy"},
				{Name: "dovecot", Status: "healthy"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr := captureOutput(func() {
		statusCmd.Run(statusCmd, []string{})
	})
	if !strings.Contains(stdout, "postfix") || !strings.Contains(stdout, "dovecot") {
		t.Errorf("expected services in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestStatusCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, HealthResponse{
			Data: []ServiceHealth{
				{Name: "postfix", Status: "healthy"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	jsonOutput = true
	stdout, stderr := captureOutput(func() {
		statusCmd.Run(statusCmd, []string{})
	})
	jsonOutput = false
	if !strings.Contains(stdout, "postfix") || !strings.Contains(stdout, "healthy") {
		t.Errorf("expected JSON health in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestStatusCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("status", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Show service health status")
}

func TestStatusCmd_WrongArgs(t *testing.T) {
	resetFlags()
	err := statusCmd.Args(statusCmd, []string{"extra"})
	requireCmdErr(t, err, "status accepts no args")
}

func TestStatusCmd_UnhealthyService(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, HealthResponse{
			Data: []ServiceHealth{
				{Name: "postfix", Status: "healthy"},
				{Name: "rspamd", Status: "unhealthy"},
				{Name: "dovecot", Status: "degraded"},
			},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr := captureOutput(func() {
		statusCmd.Run(statusCmd, []string{})
	})
	if !strings.Contains(stdout, "unhealthy") && !strings.Contains(stderr, "unhealthy") {
		t.Errorf("expected unhealthy status in output, got stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "degraded") && !strings.Contains(stderr, "degraded") {
		t.Errorf("expected degraded status in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestStatusCmd_APIError(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusServiceUnavailable, APIErrorResponse{
			Error: APIError{Code: "unavailable", Message: "Health check failed"},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	_, err := apiRequest("GET", "/api/health", nil)
	if err == nil {
		t.Fatal("expected error from API")
	}
	if !strings.Contains(err.Error(), "Health check failed") {
		t.Errorf("unexpected error: %v", err)
	}
}
