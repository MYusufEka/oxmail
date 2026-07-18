package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendTestCmd_Success(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/mail/send" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		newTestResponse(w, http.StatusOK, map[string]string{"status": "sent"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("send-test", "from@test.com", "to@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "from@test.com") || !strings.Contains(stderr, "to@test.com") {
		t.Errorf("expected addresses in stderr, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestSendTestCmd_JSON(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusOK, map[string]string{"status": "sent"})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	stdout, stderr, err := executeCommandC("send-test", "from@test.com", "to@test.com", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "sent") {
		t.Errorf("expected JSON response in stdout, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %s", stderr)
	}
}

func TestSendTestCmd_WrongArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("send-test", "from@test.com")
	requireCmdErr(t, err, "send-test requires exactly 2 args")
}

func TestSendTestCmd_NoArgs(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("send-test")
	requireCmdErr(t, err, "send-test requires exactly 2 args")
}

func TestSendTestCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("send-test", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Send a test email")
}

func TestSendTestCmd_APIError(t *testing.T) {
	resetFlags()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newTestResponse(w, http.StatusBadRequest, APIErrorResponse{
			Error: APIError{Code: "send_failed", Message: "Failed to send email"},
		})
	}))
	defer server.Close()
	apiURL = server.URL
	defer resetAPIURL()

	_, err := apiRequest("POST", "/api/mail/send", map[string]string{
		"from": "from@test.com",
		"to":   "to@test.com",
	})
	if err == nil {
		t.Fatal("expected error from API")
	}
	if !strings.Contains(err.Error(), "Failed to send email") {
		t.Errorf("unexpected error: %v", err)
	}
}
