package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func getMailDomain() string {
	d := os.Getenv("OXMAIL_DOMAIN")
	if d == "" {
		d = "local.test"
	}
	return d
}

func getMailHostname() string {
	return fmt.Sprintf("mail.%s", getMailDomain())
}

// AutodiscoverHandler holds handler methods for mail auto-config endpoints.
type AutodiscoverHandler struct{}

func NewAutodiscoverHandler() *AutodiscoverHandler {
	return &AutodiscoverHandler{}
}

// HandleOutlookXML responds to GET /autodiscover/autodiscover.xml
// Microsoft Outlook / Exchange ActiveSync autodiscover format.
func (h *AutodiscoverHandler) HandleOutlookXML(w http.ResponseWriter, r *http.Request) {
	domain := getMailDomain()
	hostname := getMailHostname()

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// MS Outlook Autodiscover response schema
	fmt.Fprintf(w, `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006">
  <Response xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a">
    <User>
      <DisplayName>%%DISPLAY_NAME%%</DisplayName>
    </User>
    <Account>
      <AccountType>email</AccountType>
      <Action>settings</Action>
      <Protocol>
        <Type>IMAP</Type>
        <Server>%[1]s</Server>
        <Port>143</Port>
        <DomainRequired>on</DomainRequired>
        <LoginName>%%EMAILADDRESS%%</LoginName>
        <SPA>off</SPA>
        <SSL>off</SSL>
        <AuthRequired>on</AuthRequired>
        <UsePOPAuth>off</UsePOPAuth>
        <SMTPLast>off</SMTPLast>
      </Protocol>
      <Protocol>
        <Type>SMTP</Type>
        <Server>%[1]s</Server>
        <Port>587</Port>
        <DomainRequired>on</DomainRequired>
        <LoginName>%%EMAILADDRESS%%</LoginName>
        <SPA>off</SPA>
        <SSL>off</SSL>
        <AuthRequired>on</AuthRequired>
        <UsePOPAuth>off</UsePOPAuth>
        <SMTPLast>off</SMTPLast>
      </Protocol>
      <Protocol>
        <Type>SMTP</Type>
        <Server>%[1]s</Server>
        <Port>465</Port>
        <DomainRequired>on</DomainRequired>
        <LoginName>%%EMAILADDRESS%%</LoginName>
        <SPA>off</SPA>
        <SSL>on</SSL>
        <AuthRequired>on</AuthRequired>
        <UsePOPAuth>off</UsePOPAuth>
        <SMTPLast>off</SMTPLast>
      </Protocol>
    </Account>
  </Response>
</Autodiscover>`, hostname, domain)
}

// HandleAppleJSON responds to GET /autodiscover/autodiscover.json
// Apple Mail / Thunderbird autodiscover JSON format.
func (h *AutodiscoverHandler) HandleAppleJSON(w http.ResponseWriter, r *http.Request) {
	hostname := getMailHostname()

	resp := map[string]interface{}{
		"account": map[string]interface{}{
			"emailAddress": r.URL.Query().Get("email"),
			"imap": map[string]interface{}{
				"server":      hostname,
				"port":        143,
				"tls":         true,
				"username":    "%EMAILADDRESS%",
				"authentication": "password-cleartext",
			},
			"smtp": []map[string]interface{}{
				{
					"server":         hostname,
					"port":           587,
					"tls":            true,
					"username":       "%EMAILADDRESS%",
					"authentication": "password-cleartext",
				},
				{
					"server":         hostname,
					"port":           465,
					"ssl":            true,
					"username":       "%EMAILADDRESS%",
					"authentication": "password-cleartext",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleMozillaXML responds to GET /.well-known/autoconfig/mail/config-v1.1.xml
// Mozilla Thunderbird autoconfig format.
func (h *AutodiscoverHandler) HandleMozillaXML(w http.ResponseWriter, r *http.Request) {
	domain := getMailDomain()
	hostname := getMailHostname()

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<clientConfig version="1.1">
  <emailProvider id="%[2]s">
    <domain>%[2]s</domain>
    <displayName>%[2]s Mail</displayName>
    <displayShortName>%[2]s</displayShortName>
    <incomingServer type="imap">
      <hostname>%[1]s</hostname>
      <port>143</port>
      <socketType>STARTTLS</socketType>
      <authentication>password-cleartext</authentication>
      <username>%%EMAILADDRESS%%</username>
    </incomingServer>
    <outgoingServer type="smtp">
      <hostname>%[1]s</hostname>
      <port>587</port>
      <socketType>STARTTLS</socketType>
      <authentication>password-cleartext</authentication>
      <username>%%EMAILADDRESS%%</username>
    </outgoingServer>
    <outgoingServer type="smtp">
      <hostname>%[1]s</hostname>
      <port>465</port>
      <socketType>SSL</socketType>
      <authentication>password-cleartext</authentication>
      <username>%%EMAILADDRESS%%</username>
    </outgoingServer>
  </emailProvider>
</clientConfig>`, hostname, domain)
}

// RegisterRoutes registers autodiscover routes on the given router.
func (h *AutodiscoverHandler) RegisterRoutes(r chi.Router) {
	r.Get("/autodiscover/autodiscover.xml", h.HandleOutlookXML)
	r.Get("/autodiscover/autodiscover.json", h.HandleAppleJSON)
	r.Get("/.well-known/autoconfig/mail/config-v1.1.xml", h.HandleMozillaXML)
}
