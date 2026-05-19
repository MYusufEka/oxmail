package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDNSResolver implements DNSResolver for testing.
type mockDNSResolver struct {
	mxRecords  []string
	txtRecords []string
	err        error
}

func (m *mockDNSResolver) LookupMX(host string) ([]string, error) {
	return m.mxRecords, m.err
}

func (m *mockDNSResolver) LookupTXT(host string) ([]string, error) {
	return m.txtRecords, m.err
}

func newTestDNSHandler(dkimSvc *domain.DKIMService, resolver DNSResolver) *DNSHandler {
	return &DNSHandler{
		domain:   "example.com",
		publicIP: "203.0.113.1",
		dkimSvc:  dkimSvc,
		resolver: resolver,
	}
}

func setupDNSRouter(handler *DNSHandler) *chi.Mux {
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func TestDNSRecords_ReturnsAllRequiredRecords(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	_, err := dkimSvc.Generate("example.com", "default")
	require.NoError(t, err)

	handler := newTestDNSHandler(dkimSvc, &mockDNSResolver{})
	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Records []domain.DNSRecord `json:"records"`
	}
	err = json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	// Should have MX, SPF, DKIM, DMARC, rDNS = 5 records
	assert.Len(t, resp.Records, 5)

	recordTypes := make(map[string]bool)
	for _, r := range resp.Records {
		recordTypes[r.Type] = true
	}
	assert.True(t, recordTypes["MX"], "should have MX record")
	assert.True(t, recordTypes["TXT"], "should have TXT records (SPF/DKIM/DMARC)")
	assert.True(t, recordTypes["PTR"], "should have PTR record")
}

func TestDNSRecords_MXRecord(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	_, err := dkimSvc.Generate("example.com", "default")
	require.NoError(t, err)

	handler := newTestDNSHandler(dkimSvc, &mockDNSResolver{})
	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp struct {
		Records []domain.DNSRecord `json:"records"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	var mx *domain.DNSRecord
	for i := range resp.Records {
		if resp.Records[i].Type == "MX" {
			mx = &resp.Records[i]
			break
		}
	}
	require.NotNil(t, mx)
	assert.Equal(t, "example.com", mx.Domain)
	assert.Equal(t, "example.com", mx.Name)
	assert.Equal(t, "mail.example.com", mx.Value)
	assert.Equal(t, 10, mx.Priority)
}

func TestDNSRecords_SPFRecord(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	_, err := dkimSvc.Generate("example.com", "default")
	require.NoError(t, err)

	handler := newTestDNSHandler(dkimSvc, &mockDNSResolver{})
	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp struct {
		Records []domain.DNSRecord `json:"records"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	var spf *domain.DNSRecord
	for i := range resp.Records {
		if resp.Records[i].Name == "example.com" && resp.Records[i].Type == "TXT" {
			spf = &resp.Records[i]
			break
		}
	}
	require.NotNil(t, spf)
	assert.Equal(t, `v=spf1 ip4:203.0.113.1 -all`, spf.Value)
}

func TestDNSRecords_DMARCRecord(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	_, err := dkimSvc.Generate("example.com", "default")
	require.NoError(t, err)

	handler := newTestDNSHandler(dkimSvc, &mockDNSResolver{})
	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp struct {
		Records []domain.DNSRecord `json:"records"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	var dmarc *domain.DNSRecord
	for i := range resp.Records {
		if resp.Records[i].Name == "_dmarc.example.com" {
			dmarc = &resp.Records[i]
			break
		}
	}
	require.NotNil(t, dmarc)
	assert.Equal(t, "TXT", dmarc.Type)
	assert.Equal(t, `v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com`, dmarc.Value)
}

func TestDNSRecords_DKIMRecord(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	key, err := dkimSvc.Generate("example.com", "default")
	require.NoError(t, err)

	handler := newTestDNSHandler(dkimSvc, &mockDNSResolver{})
	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp struct {
		Records []domain.DNSRecord `json:"records"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	var dkim *domain.DNSRecord
	for i := range resp.Records {
		if resp.Records[i].Name == "default._domainkey.example.com" {
			dkim = &resp.Records[i]
			break
		}
	}
	require.NotNil(t, dkim)
	assert.Equal(t, "TXT", dkim.Type)
	assert.Equal(t, key.DNSRecord, dkim.Value)
}

func TestDNSRecords_NoDKIMKey(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	// No DKIM key generated

	handler := newTestDNSHandler(dkimSvc, &mockDNSResolver{})
	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Records []domain.DNSRecord `json:"records"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	// Should still have 4 records (MX, SPF, DMARC, rDNS) without DKIM
	assert.Len(t, resp.Records, 4)
}

func TestDNSCheck_AllValid(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())
	key, err := dkimSvc.Generate("example.com", "default")
	require.NoError(t, err)

	resolver := &mockDNSResolver{
		mxRecords:  []string{"mail.example.com."},
		txtRecords: []string{`v=spf1 ip4:203.0.113.1 -all`},
	}

	handler := &DNSHandler{
		domain:   "example.com",
		publicIP: "203.0.113.1",
		dkimSvc:  dkimSvc,
		resolver: resolver,
	}
	_ = key

	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/check", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Results []domain.DNSCheckResult `json:"results"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	// Should check MX, SPF, DKIM, DMARC
	assert.GreaterOrEqual(t, len(resp.Results), 4)
}

func TestDNSCheck_MXInvalid(t *testing.T) {
	dkimSvc := domain.NewDKIMService(t.TempDir())

	resolver := &mockDNSResolver{
		mxRecords:  []string{"wrong.example.com."},
		txtRecords: []string{},
	}

	handler := &DNSHandler{
		domain:   "example.com",
		publicIP: "203.0.113.1",
		dkimSvc:  dkimSvc,
		resolver: resolver,
	}

	router := setupDNSRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/dns/check", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Results []domain.DNSCheckResult `json:"results"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	var mxResult *domain.DNSCheckResult
	for i := range resp.Results {
		if resp.Results[i].Record == "MX" {
			mxResult = &resp.Results[i]
			break
		}
	}
	require.NotNil(t, mxResult)
	assert.False(t, mxResult.Valid)
}
