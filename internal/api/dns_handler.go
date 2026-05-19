package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
)

// DNSResolver abstracts DNS lookups for testability.
type DNSResolver interface {
	LookupMX(host string) ([]string, error)
	LookupTXT(host string) ([]string, error)
}

// DNSHandler handles DNS records and verification endpoints.
type DNSHandler struct {
	domain   string
	publicIP string
	dkimSvc  *domain.DKIMService
	resolver DNSResolver
}

// NewDNSHandler creates a DNSHandler from environment configuration.
func NewDNSHandler(domainName, publicIP string, dkimSvc *domain.DKIMService, resolver DNSResolver) *DNSHandler {
	return &DNSHandler{
		domain:   domainName,
		publicIP: publicIP,
		dkimSvc:  dkimSvc,
		resolver: resolver,
	}
}

// RegisterRoutes mounts DNS endpoints on the given router.
func (h *DNSHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/dns/records", h.handleRecords)
	r.Get("/api/dns/check", h.handleCheck)
}

func (h *DNSHandler) handleRecords(w http.ResponseWriter, r *http.Request) {
	records := h.buildRecords()

	writeJSON(w, http.StatusOK, map[string][]domain.DNSRecord{
		"records": records,
	})
}

func (h *DNSHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	results := h.checkRecords()

	writeJSON(w, http.StatusOK, map[string][]domain.DNSCheckResult{
		"results": results,
	})
}

func (h *DNSHandler) buildRecords() []domain.DNSRecord {
	records := []domain.DNSRecord{
		{
			Domain:   h.domain,
			Type:     "MX",
			Name:     h.domain,
			Value:    fmt.Sprintf("mail.%s", h.domain),
			Priority: 10,
		},
		{
			Domain: h.domain,
			Type:   "TXT",
			Name:   h.domain,
			Value:  fmt.Sprintf("v=spf1 ip4:%s -all", h.publicIP),
		},
		{
			Domain: h.domain,
			Type:   "TXT",
			Name:   fmt.Sprintf("_dmarc.%s", h.domain),
			Value:  fmt.Sprintf("v=DMARC1; p=quarantine; rua=mailto:postmaster@%s", h.domain),
		},
		{
			Domain: h.domain,
			Type:   "PTR",
			Name:   h.publicIP,
			Value:  fmt.Sprintf("mail.%s", h.domain),
		},
	}

	dkimKey, err := h.dkimSvc.Get(h.domain, "default")
	if err == nil {
		records = append(records, domain.DNSRecord{
			Domain: h.domain,
			Type:   "TXT",
			Name:   fmt.Sprintf("default._domainkey.%s", h.domain),
			Value:  dkimKey.DNSRecord,
		})
	}

	return records
}

func (h *DNSHandler) checkRecords() []domain.DNSCheckResult {
	var results []domain.DNSCheckResult

	// Check MX
	expectedMX := fmt.Sprintf("mail.%s", h.domain)
	mxRecords, _ := h.resolver.LookupMX(h.domain)
	mxActual := strings.Join(mxRecords, ", ")
	mxValid := false
	for _, mx := range mxRecords {
		cleaned := strings.TrimSuffix(mx, ".")
		if cleaned == expectedMX {
			mxValid = true
			break
		}
	}
	results = append(results, domain.DNSCheckResult{
		Domain:   h.domain,
		Record:   "MX",
		Expected: expectedMX,
		Actual:   mxActual,
		Valid:    mxValid,
	})

	// Check SPF
	expectedSPF := fmt.Sprintf("v=spf1 ip4:%s -all", h.publicIP)
	spfRecords, _ := h.resolver.LookupTXT(h.domain)
	spfActual := ""
	spfValid := false
	for _, txt := range spfRecords {
		if strings.HasPrefix(txt, "v=spf1") {
			spfActual = txt
			if txt == expectedSPF {
				spfValid = true
			}
			break
		}
	}
	results = append(results, domain.DNSCheckResult{
		Domain:   h.domain,
		Record:   "SPF",
		Expected: expectedSPF,
		Actual:   spfActual,
		Valid:    spfValid,
	})

	// Check DKIM
	dkimKey, err := h.dkimSvc.Get(h.domain, "default")
	dkimHost := fmt.Sprintf("default._domainkey.%s", h.domain)
	expectedDKIM := ""
	if err == nil {
		expectedDKIM = dkimKey.DNSRecord
	}
	dkimRecords, _ := h.resolver.LookupTXT(dkimHost)
	dkimActual := ""
	dkimValid := false
	for _, txt := range dkimRecords {
		if strings.HasPrefix(txt, "v=DKIM1") {
			dkimActual = txt
			if txt == expectedDKIM {
				dkimValid = true
			}
			break
		}
	}
	results = append(results, domain.DNSCheckResult{
		Domain:   h.domain,
		Record:   "DKIM",
		Expected: expectedDKIM,
		Actual:   dkimActual,
		Valid:    dkimValid,
	})

	// Check DMARC
	dmarcHost := fmt.Sprintf("_dmarc.%s", h.domain)
	expectedDMARC := fmt.Sprintf("v=DMARC1; p=quarantine; rua=mailto:postmaster@%s", h.domain)
	dmarcRecords, _ := h.resolver.LookupTXT(dmarcHost)
	dmarcActual := ""
	dmarcValid := false
	for _, txt := range dmarcRecords {
		if strings.HasPrefix(txt, "v=DMARC1") {
			dmarcActual = txt
			if txt == expectedDMARC {
				dmarcValid = true
			}
			break
		}
	}
	results = append(results, domain.DNSCheckResult{
		Domain:   h.domain,
		Record:   "DMARC",
		Expected: expectedDMARC,
		Actual:   dmarcActual,
		Valid:    dmarcValid,
	})

	return results
}
