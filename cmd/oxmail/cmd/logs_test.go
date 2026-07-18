package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogsCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/logs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"time":"2024-01-01T00:00:00Z","message":"test log"}]`))
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("logs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "test log") {
		t.Errorf("expected log in stdout, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestLogsCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"message":"json log"}]`))
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("logs", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "json log") {
		t.Errorf("expected JSON log in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestLogsCmd_ServiceFilter(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "service=postfix") {
			t.Errorf("expected service filter in query, got: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"service":"postfix","message":"postfix log"}]`))
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("logs", "--service", "postfix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "postfix log") {
		t.Errorf("expected filtered log in stdout, got: %s", stdout)
	}
	_ = stderr
}

func TestLogsCmd_ServiceFlagExists(t *testing.T) {
	flag := logsCmd.Flags().Lookup("service")
	if flag == nil {
		t.Fatal("expected --service flag on logs")
	}
}

func TestLogsCmd_FollowFlagExists(t *testing.T) {
	flag := logsCmd.Flags().Lookup("follow")
	if flag == nil {
		t.Fatal("expected --follow flag on logs")
	}
	shortFlag := logsCmd.Flags().ShorthandLookup("f")
	if shortFlag == nil {
		t.Fatal("expected -f shorthand for --follow")
	}
}

func TestLogsCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("logs", "extra")
	requireCmdErr(t, err, "logs accepts no args")
}

func TestLogsCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("logs", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "View or tail service logs")
	assertContains(t, output, "--follow")
	assertContains(t, output, "--service")
}

func TestLogsCmd_APIError(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"code":"server_error","message":"Logs unavailable"}}`))
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	_, err := apiRequest("GET", "/api/logs", nil)
	if err == nil {
		t.Fatal("expected error from API")
	}
	if !strings.Contains(err.Error(), "Logs unavailable") {
		t.Errorf("unexpected error: %v", err)
	}
}
