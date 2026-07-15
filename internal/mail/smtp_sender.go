package mail

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MYusufEka/oxmail/internal/domain"
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
func (s *SMTPSender) Send(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
	if from == "" {
		return "", fmt.Errorf("from address is required")
	}
	if len(to) == 0 {
		return "", fmt.Errorf("at least one recipient is required")
	}

	if err := validateAttachments(attachments); err != nil {
		return "", err
	}

	domain := extractDomain(from)
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)

	msg := buildMIMEMessage(from, to, cc, subject, bodyText, bodyHTML, messageID, attachments)

	allRecipients := make([]string, 0, len(to)+len(cc))
	allRecipients = append(allRecipients, to...)
	allRecipients = append(allRecipients, cc...)

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	client, err := smtp.Dial(addr)
	if err != nil {
		return "", fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         s.config.Host,
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			_ = err
		}
	}

	if s.config.Username != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return "", fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return "", fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	for _, rcpt := range allRecipients {
		if err := client.Rcpt(rcpt); err != nil {
			return "", fmt.Errorf("smtp RCPT TO %s: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return "", fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("smtp close data: %w", err)
	}

	return messageID, client.Quit()
}

func buildMIMEMessage(from string, to []string, cc []string, subject, bodyText, bodyHTML, messageID string, attachments []domain.SendMailAttachment) string {
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

	hasBody := bodyHTML != "" || bodyText != ""
	hasAttachments := len(attachments) > 0

	if hasAttachments && hasBody {
		outerBoundary := fmt.Sprintf("oxmail-mixed-%s", uuid.New().String()[:8])
		innerBoundary := fmt.Sprintf("oxmail-alt-%s", uuid.New().String()[:8])

		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", outerBoundary))
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
		writeBodyPart(&msg, bodyText, bodyHTML, innerBoundary)

		for _, att := range attachments {
			msg.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.MimeType))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", mime.QEncoding.Encode("utf-8", att.Filename)))
			msg.WriteString("\r\n")
			msg.WriteString(wrapBase64(att.Content))
			msg.WriteString("\r\n")
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", outerBoundary))
	} else if hasAttachments {
		outerBoundary := fmt.Sprintf("oxmail-mixed-%s", uuid.New().String()[:8])
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", outerBoundary))
		msg.WriteString("\r\n")

		for _, att := range attachments {
			msg.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.MimeType))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", mime.QEncoding.Encode("utf-8", att.Filename)))
			msg.WriteString("\r\n")
			msg.WriteString(wrapBase64(att.Content))
			msg.WriteString("\r\n")
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", outerBoundary))
	} else {
		if bodyText != "" && bodyHTML != "" {
			boundary := fmt.Sprintf("oxmail-alt-%s", uuid.New().String()[:8])
			writeBodyPart(&msg, bodyText, bodyHTML, boundary)
		} else {
			writeBodyPart(&msg, bodyText, bodyHTML, "")
		}
	}

	return msg.String()
}

func writeBodyPart(msg *strings.Builder, bodyText, bodyHTML, boundary string) {
	if boundary != "" && bodyHTML != "" && bodyText != "" {
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
}

func validateAttachments(attachments []domain.SendMailAttachment) error {
	maxSize := 10 * 1024 * 1024 // 10MB default
	if envSize := os.Getenv("OXMAIL_MAX_ATTACHMENT_SIZE"); envSize != "" {
		if parsed, err := strconv.Atoi(envSize); err == nil && parsed > 0 {
			maxSize = parsed
		}
	}

	var totalSize int
	for _, att := range attachments {
		decoded, err := base64.StdEncoding.DecodeString(att.Content)
		if err != nil {
			return fmt.Errorf("invalid base64 content for attachment %q", att.Filename)
		}
		totalSize += len(decoded)
	}

	if totalSize > maxSize {
		return fmt.Errorf("total attachment size %d bytes exceeds limit of %d bytes", totalSize, maxSize)
	}

	return nil
}

func wrapBase64(content string) string {
	const lineLen = 76
	var sb strings.Builder
	for i := 0; i < len(content); i += lineLen {
		end := i + lineLen
		if end > len(content) {
			end = len(content)
		}
		sb.WriteString(content[i:end])
		sb.WriteString("\r\n")
	}
	return sb.String()
}

func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return "localhost"
}
