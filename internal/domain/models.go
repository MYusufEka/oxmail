package domain

import "time"

// Domain represents a mail domain managed by Oxmail.
type Domain struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// User represents a mailbox user.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DomainID     int64     `json:"domainId"`
	DisplayName  string    `json:"displayName,omitempty"`
	Quota        int64     `json:"quota"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Alias represents an email forwarding alias.
type Alias struct {
	ID                 int64     `json:"id"`
	SourceAddress      string    `json:"sourceAddress"`
	DestinationAddress string    `json:"destinationAddress"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"createdAt"`
}

// DKIMKey represents a DKIM signing key for a domain.
type DKIMKey struct {
	Domain    string    `json:"domain"`
	Selector  string    `json:"selector"`
	PublicKey string    `json:"publicKey"`
	DNSRecord string    `json:"dnsRecord,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// HealthStatus represents the overall system health.
type HealthStatus struct {
	Status   string                   `json:"status"`
	Services map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents the health of a single service.
type ServiceHealth struct {
	Status    string `json:"status"`
	LatencyMs int    `json:"latencyMs,omitempty"`
}

// LogEntry represents a single log line from a service.
type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Attachment represents email attachment metadata.
type Attachment struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// MailMessage represents an email message.
type MailMessage struct {
	ID          int64        `json:"id"`
	From        string       `json:"from"`
	To          []string     `json:"to"`
	CC          []string     `json:"cc,omitempty"`
	Subject     string       `json:"subject"`
	BodyText    string       `json:"bodyText,omitempty"`
	BodyHTML    string       `json:"bodyHtml,omitempty"`
	Read        bool         `json:"read"`
	ReceivedAt  time.Time    `json:"receivedAt"`
	Attachments []Attachment `json:"attachments,omitempty"`
	ThreadID    string       `json:"threadId,omitempty"`
	MessageID   string       `json:"messageId,omitempty"`
	InReplyTo   string       `json:"inReplyTo,omitempty"`
}

// Pagination holds pagination metadata for list responses.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// DNSRecord represents a required DNS record for mail delivery.
type DNSRecord struct {
	Domain   string `json:"domain"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority int    `json:"priority,omitempty"`
}

// DNSCheckResult represents the validation result of a DNS record.
type DNSCheckResult struct {
	Domain   string `json:"domain"`
	Record   string `json:"record"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Valid    bool   `json:"valid"`
}

// LoginRequest is the payload for authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned after successful authentication.
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// CreateDomainRequest is the payload for creating a domain.
type CreateDomainRequest struct {
	Name string `json:"name"`
}

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName,omitempty"`
	Quota       int64  `json:"quota,omitempty"`
}

// CreateAliasRequest is the payload for creating an alias.
type CreateAliasRequest struct {
	SourceAddress      string `json:"sourceAddress"`
	DestinationAddress string `json:"destinationAddress"`
}

// SendMailRequest is the payload for sending an email.
type SendMailRequest struct {
	From     string   `json:"from"`
	To       []string `json:"to"`
	CC       []string `json:"cc,omitempty"`
	Subject  string   `json:"subject"`
	BodyText string   `json:"bodyText,omitempty"`
	BodyHTML string   `json:"bodyHtml,omitempty"`
}

// SendMailResponse is returned after queuing an email for delivery.
type SendMailResponse struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
