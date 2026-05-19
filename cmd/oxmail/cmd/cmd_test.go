package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	apiURL = server.URL
	return server
}

func TestDomainListJSON(t *testing.T) {
	expected := DomainListResponse{
		Data: []Domain{
			{Name: "example.com", CreatedAt: "2024-01-01T00:00:00Z"},
		},
	}

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domains" || r.Method != "GET" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer server.Close()

	resp, err := apiRequest("GET", "/api/domains", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got DomainListResponse
	if err := json.Unmarshal(resp, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(got.Data) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(got.Data))
	}
	if got.Data[0].Name != "example.com" {
		t.Errorf("expected domain name 'example.com', got %q", got.Data[0].Name)
	}
}

func TestDomainAdd(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domains" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test.com" {
			t.Errorf("expected domain 'test.com', got %q", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"name": "test.com"})
	})
	defer server.Close()

	resp, err := apiRequest("POST", "/api/domains", map[string]string{"name": "test.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}
}

func TestDomainRm(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domains/test.com" || r.Method != "DELETE" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	_, err := apiRequest("DELETE", "/api/domains/test.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIErrorHandling(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIErrorResponse{
			Error: APIError{Code: "not_found", Message: "Domain not found"},
		})
	})
	defer server.Close()

	_, err := apiRequest("GET", "/api/domains/missing.com", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "Domain not found (code: not_found)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestUserAdd(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["email"] != "user@test.com" {
			t.Errorf("expected email 'user@test.com', got %q", body["email"])
		}
		if body["password"] != "secret123" {
			t.Errorf("expected password 'secret123', got %q", body["password"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"email": "user@test.com"})
	})
	defer server.Close()

	resp, err := apiRequest("POST", "/api/users", map[string]string{
		"email":    "user@test.com",
		"password": "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}
}

func TestAliasAdd(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/aliases" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["source"] != "info@test.com" {
			t.Errorf("expected source 'info@test.com', got %q", body["source"])
		}
		if body["destination"] != "admin@test.com" {
			t.Errorf("expected destination 'admin@test.com', got %q", body["destination"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "1"})
	})
	defer server.Close()

	resp, err := apiRequest("POST", "/api/aliases", map[string]string{
		"source":      "info@test.com",
		"destination": "admin@test.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}
}

func TestHealthStatus(t *testing.T) {
	expected := HealthResponse{
		Data: []ServiceHealth{
			{Name: "postfix", Status: "healthy"},
			{Name: "dovecot", Status: "healthy"},
			{Name: "rspamd", Status: "unhealthy"},
		},
	}

	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" || r.Method != "GET" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer server.Close()

	resp, err := apiRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got HealthResponse
	if err := json.Unmarshal(resp, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(got.Data) != 3 {
		t.Fatalf("expected 3 services, got %d", len(got.Data))
	}
	if got.Data[2].Status != "unhealthy" {
		t.Errorf("expected rspamd unhealthy, got %q", got.Data[2].Status)
	}
}

func TestSendTest(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mail/send" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["from"] != "sender@test.com" {
			t.Errorf("expected from 'sender@test.com', got %q", body["from"])
		}
		if body["to"] != "receiver@test.com" {
			t.Errorf("expected to 'receiver@test.com', got %q", body["to"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
	})
	defer server.Close()

	resp, err := apiRequest("POST", "/api/mail/send", map[string]string{
		"from":    "sender@test.com",
		"to":      "receiver@test.com",
		"subject": "Oxmail Test Email",
		"body":    "This is a test email sent from the Oxmail CLI.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	commands := rootCmd.Commands()
	expectedNames := []string{"alias", "domain", "logs", "send-test", "status", "user"}

	found := make(map[string]bool)
	for _, cmd := range commands {
		found[cmd.Name()] = true
	}

	for _, name := range expectedNames {
		if !found[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}
