package mail

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SMTPSenderConfig holds configuration for the SMTP sender.
type SMTPSenderConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

// SMTPSender sends emails via an SMTP submission port.
type SMTPSender struct {
	config SMTPSenderConfig
}

// NewSMTPSender creates a new SMTPSender with the given config.
func NewSMTPSender(config SMTPSenderConfig) *SMTPSender {
	return &SMTPSender{config: config}
}

// Send submits an email via SMTP and returns the generated Message-ID.
func (s *SMTPSender) Send(from string, to []string, cc []string, subject, bodyText, bodyHTML string) (string, error) {
	if from == "" {
		return "", fmt.Errorf("from address is required")
	}
	if len(to) == 0 {
		return "", fmt.Errorf("at least one recipient is required")
	}

	domain := extractDomain(from)
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)

	msg := buildMIMEMessage(from, to, cc, subject, bodyText, bodyHTML, messageID)

	allRecipients := make([]string, 0, len(to)+len(cc))
	allRecipients = append(allRecipients, to...)
	allRecipients = append(allRecipients, cc...)

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	err := smtp.SendMail(addr, auth, from, allRecipients, []byte(msg))
	if err != nil {
		return "", fmt.Errorf("smtp send failed: %w", err)
	}

	return messageID, nil
}

func buildMIMEMessage(from string, to []string, cc []string, subject, bodyText, bodyHTML, messageID string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	if len(cc) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z)))
	msg.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if bodyHTML != "" && bodyText != "" {
		boundary := fmt.Sprintf("oxmail-%s", uuid.New().String()[:8])
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		msg.WriteString("\r\n")
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(bodyText)
		msg.WriteString("\r\n")
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(bodyHTML)
		msg.WriteString("\r\n")
		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if bodyHTML != "" {
		msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(bodyHTML)
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(bodyText)
	}

	return msg.String()
}

func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return "localhost"
}
