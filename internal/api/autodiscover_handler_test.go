package api

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutodiscoverHandler_OutlookXML(t *testing.T) {
	os.Setenv("OXMAIL_DOMAIN", "example.com")
	defer os.Unsetenv("OXMAIL_DOMAIN")

	h := NewAutodiscoverHandler()
	req := httptest.NewRequest(http.MethodGet, "/autodiscover/autodiscover.xml", nil)
	rec := httptest.NewRecorder()

	h.HandleOutlookXML(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/xml; charset=utf-8", rec.Header().Get("Content-Type"))

	body := rec.Body.String()
	assert.Contains(t, body, "mail.example.com")
	assert.Contains(t, body, "<Port>143</Port>")
	assert.Contains(t, body, "<Port>587</Port>")
	assert.Contains(t, body, "<Port>465</Port>")
	assert.Contains(t, body, "Autodiscover")

	// Validate XML well-formedness
	var stub interface{}
	err := xml.Unmarshal([]byte(body), &stub)
	require.NoError(t, err, "response must be valid XML")
}

func TestAutodiscoverHandler_AppleJSON(t *testing.T) {
	os.Setenv("OXMAIL_DOMAIN", "example.com")
	defer os.Unsetenv("OXMAIL_DOMAIN")

	h := NewAutodiscoverHandler()
	req := httptest.NewRequest(http.MethodGet, "/autodiscover/autodiscover.json?email=alice@example.com", nil)
	rec := httptest.NewRecorder()

	h.HandleAppleJSON(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	body := rec.Body.String()
	assert.Contains(t, body, "mail.example.com")
	assert.Contains(t, body, "143")
	assert.Contains(t, body, "587")
	assert.Contains(t, body, "465")
}

func TestAutodiscoverHandler_MozillaXML(t *testing.T) {
	os.Setenv("OXMAIL_DOMAIN", "example.com")
	defer os.Unsetenv("OXMAIL_DOMAIN")

	h := NewAutodiscoverHandler()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/autoconfig/mail/config-v1.1.xml", nil)
	rec := httptest.NewRecorder()

	h.HandleMozillaXML(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/xml; charset=utf-8", rec.Header().Get("Content-Type"))

	body := rec.Body.String()
	assert.Contains(t, body, "mail.example.com")
	assert.Contains(t, body, "<port>143</port>")
	assert.Contains(t, body, "<port>587</port>")
	assert.Contains(t, body, "<port>465</port>")
	assert.Contains(t, body, "clientConfig")
	assert.Contains(t, body, "STARTTLS")
	assert.Contains(t, body, "SSL")

	// Validate XML well-formedness
	var stub interface{}
	err := xml.Unmarshal([]byte(body), &stub)
	require.NoError(t, err, "response must be valid XML")
}

func TestAutodiscoverHandler_DefaultDomain(t *testing.T) {
	os.Unsetenv("OXMAIL_DOMAIN")

	h := NewAutodiscoverHandler()
	req := httptest.NewRequest(http.MethodGet, "/autodiscover/autodiscover.xml", nil)
	rec := httptest.NewRecorder()

	h.HandleOutlookXML(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "mail.local.test")
}

func TestAutodiscoverHandler_RoutesRegistered(t *testing.T) {
	// Verify routes are reachable through the router
	h := NewAutodiscoverHandler()
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/autodiscover/autodiscover.xml"},
		{http.MethodGet, "/autodiscover/autodiscover.json"},
		{http.MethodGet, "/.well-known/autoconfig/mail/config-v1.1.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
