package domain

import (
	"bufio"
	"context"
	"database/sql"
	"net"
	"os"
	"strings"
)

// DomainCheckResult holds the result of a single domain health check.
type DomainCheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass" | "warn" | "fail"
	Detail string `json:"detail"`
}

// DomainHealthResult holds the aggregate health result for a domain.
type DomainHealthResult struct {
	Domain string              `json:"domain"`
	Status string              `json:"status"` // "healthy" | "degraded" | "unhealthy"
	Checks []DomainCheckResult `json:"checks"`
}

// DomainHealthChecker runs health checks against a domain.
type DomainHealthChecker struct {
	db               *sql.DB
	dovecotConfigDir string
}

// NewDomainHealthChecker creates a DomainHealthChecker backed by the given DB and Dovecot config dir.
func NewDomainHealthChecker(db *sql.DB, dovecotConfigDir string) *DomainHealthChecker {
	return &DomainHealthChecker{
		db:               db,
		dovecotConfigDir: dovecotConfigDir,
	}
}

// Check runs all 5 health checks for the given domain and returns an aggregate result.
func (c *DomainHealthChecker) Check(ctx context.Context, domainName string) DomainHealthResult {
	checks := []DomainCheckResult{
		c.checkMX(domainName),
		c.checkSPF(domainName),
		c.checkDMARC(domainName),
		c.checkDKIM(ctx, domainName),
		c.checkDovecot(domainName),
	}

	status := aggregateStatus(checks)

	return DomainHealthResult{
		Domain: domainName,
		Status: status,
		Checks: checks,
	}
}

func (c *DomainHealthChecker) checkMX(domainName string) DomainCheckResult {
	records, err := net.LookupMX(domainName)
	if err != nil {
		return DomainCheckResult{Name: "mx", Status: "fail", Detail: "MX lookup failed: " + err.Error()}
	}
	if len(records) == 0 {
		return DomainCheckResult{Name: "mx", Status: "fail", Detail: "no MX records found"}
	}
	return DomainCheckResult{Name: "mx", Status: "pass", Detail: records[0].Host}
}

func (c *DomainHealthChecker) checkSPF(domainName string) DomainCheckResult {
	records, err := net.LookupTXT(domainName)
	if err != nil {
		return DomainCheckResult{Name: "spf", Status: "fail", Detail: "TXT lookup failed: " + err.Error()}
	}
	for _, rec := range records {
		if strings.HasPrefix(rec, "v=spf1") {
			return DomainCheckResult{Name: "spf", Status: "pass", Detail: rec}
		}
	}
	return DomainCheckResult{Name: "spf", Status: "fail", Detail: "no SPF record found"}
}

func (c *DomainHealthChecker) checkDMARC(domainName string) DomainCheckResult {
	records, err := net.LookupTXT("_dmarc." + domainName)
	if err != nil {
		return DomainCheckResult{Name: "dmarc", Status: "fail", Detail: "DMARC lookup failed: " + err.Error()}
	}

	for _, rec := range records {
		if !strings.Contains(rec, "v=DMARC1") {
			continue
		}
		// p=none → warn; any other p= → pass
		if strings.Contains(rec, "p=none") {
			return DomainCheckResult{Name: "dmarc", Status: "warn", Detail: "DMARC policy is p=none (monitoring only)"}
		}
		return DomainCheckResult{Name: "dmarc", Status: "pass", Detail: rec}
	}

	return DomainCheckResult{Name: "dmarc", Status: "fail", Detail: "no DMARC record found"}
}

func (c *DomainHealthChecker) checkDKIM(ctx context.Context, domainName string) DomainCheckResult {
	if c.db == nil {
		return DomainCheckResult{Name: "dkim", Status: "fail", Detail: "database unavailable"}
	}

	var exists int
	err := c.db.QueryRowContext(ctx,
		"SELECT 1 FROM dkim_keys WHERE domain = ? LIMIT 1",
		domainName,
	).Scan(&exists)

	if err == sql.ErrNoRows {
		return DomainCheckResult{Name: "dkim", Status: "fail", Detail: "no DKIM key found in database"}
	}
	if err != nil {
		return DomainCheckResult{Name: "dkim", Status: "fail", Detail: "DKIM DB query failed: " + err.Error()}
	}

	return DomainCheckResult{Name: "dkim", Status: "pass", Detail: "DKIM key present"}
}

func (c *DomainHealthChecker) checkDovecot(domainName string) DomainCheckResult {
	passdbPath := c.dovecotConfigDir + "/passdb"

	f, err := os.Open(passdbPath)
	if err != nil {
		return DomainCheckResult{Name: "dovecot", Status: "warn", Detail: "passdb file not readable: " + err.Error()}
	}
	defer f.Close()

	needle := "@" + domainName
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), needle) {
			count++
		}
	}

	if count == 0 {
		return DomainCheckResult{Name: "dovecot", Status: "warn", Detail: "no accounts found in passdb for this domain"}
	}
	return DomainCheckResult{Name: "dovecot", Status: "pass", Detail: "passdb accounts found"}
}

// aggregateStatus computes overall health: any fail→unhealthy, any warn→degraded, all pass→healthy.
func aggregateStatus(checks []DomainCheckResult) string {
	hasFail := false
	hasWarn := false
	for _, c := range checks {
		switch c.Status {
		case "fail":
			hasFail = true
		case "warn":
			hasWarn = true
		}
	}
	if hasFail {
		return "unhealthy"
	}
	if hasWarn {
		return "degraded"
	}
	return "healthy"
}
