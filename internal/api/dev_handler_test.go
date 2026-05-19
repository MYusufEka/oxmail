package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDevSender struct {
	lastFrom    string
	lastTo      []string
	lastSubject string
	lastBody    string
	returnID    string
	returnErr   error
}

func (m *mockDevSender) Send(from string, to []string, cc []string, subject, bodyText, bodyHTML string) (string, error) {
	m.lastFrom = from
	m.lastTo = to
	m.lastSubject = subject
	m.lastBody = bodyText
	return m.returnID, m.returnErr
}

func TestDevHandler_SendTest_Success(t *testing.T) {
	sender := &mockDevSender{returnID: "test-msg-123"}
	handler := NewDevHandler(sender)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/send-test?from=alice@local.test&to=bob@local.test&subject=Hello&body=Test+body", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "alice@local.test", sender.lastFrom)
	assert.Equal(t, []string{"bob@local.test"}, sender.lastTo)
	assert.Equal(t, "Hello", sender.lastSubject)
	assert.Equal(t, "Test body", sender.lastBody)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "test-msg-123", resp["messageId"])
	assert.Equal(t, "sent", resp["status"])
}

func TestDevHandler_SendTest_DefaultSubjectAndBody(t *testing.T) {
	sender := &mockDevSender{returnID: "msg-456"}
	handler := NewDevHandler(sender)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/send-test?from=alice@local.test&to=bob@local.test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Test email from Oxmail dev mode", sender.lastSubject)
	assert.Equal(t, "This is a test email sent via the dev endpoint.", sender.lastBody)
}

func TestDevHandler_SendTest_MissingFrom(t *testing.T) {
	sender := &mockDevSender{}
	handler := NewDevHandler(sender)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/send-test?to=bob@local.test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDevHandler_SendTest_MissingTo(t *testing.T) {
	sender := &mockDevSender{}
	handler := NewDevHandler(sender)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/send-test?from=alice@local.test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDevHandler_SendTest_SenderError(t *testing.T) {
	sender := &mockDevSender{returnErr: assert.AnError}
	handler := NewDevHandler(sender)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/send-test?from=alice@local.test&to=bob@local.test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestIsDevMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"dev mode enabled", "dev", true},
		{"production mode", "production", false},
		{"empty string", "", false},
		{"other value", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("OXMAIL_MODE")
			} else {
				os.Setenv("OXMAIL_MODE", tt.envValue)
			}
			defer os.Unsetenv("OXMAIL_MODE")

			assert.Equal(t, tt.expected, IsDevMode())
		})
	}
}
