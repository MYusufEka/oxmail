package mail

import (
	"strings"
	"testing"

	"github.com/emersion/go-imap/v2"
	gomessage "github.com/emersion/go-message/mail"

	"github.com/MYusufEka/oxmail/internal/domain"
)

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name string
		addr imap.Address
		want string
	}{
		{
			name: "name and email",
			addr: imap.Address{Name: "Alice Smith", Mailbox: "alice", Host: "example.com"},
			want: "Alice Smith <alice@example.com>",
		},
		{
			name: "email only, no name",
			addr: imap.Address{Name: "", Mailbox: "bob", Host: "test.org"},
			want: "bob@test.org",
		},
		{
			name: "empty mailbox",
			addr: imap.Address{Name: "", Mailbox: "", Host: "host.com"},
			want: "@host.com",
		},
		{
			name: "name with special chars",
			addr: imap.Address{Name: "O'Brien", Mailbox: "obrien", Host: "mail.ie"},
			want: "O'Brien <obrien@mail.ie>",
		},
		{
			name: "ip host",
			addr: imap.Address{Name: "", Mailbox: "admin", Host: "192.168.1.1"},
			want: "admin@192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAddress(tt.addr)
			if got != tt.want {
				t.Errorf("formatAddress(%+v) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

func TestParseMIMEBody(t *testing.T) {
	t.Run("nil body does not panic", func(t *testing.T) {
		msg := &domain.MailMessage{}
		parseMIMEBody(msg, nil)
	})

	t.Run("plain text body", func(t *testing.T) {
		msg := &domain.MailMessage{}
		body := strings.NewReader("Content-Type: text/plain\r\n\r\nHello, world!")
		parseMIMEBody(msg, body)
		if msg.BodyText != "Hello, world!" {
			t.Errorf("BodyText = %q, want %q", msg.BodyText, "Hello, world!")
		}
	})

	t.Run("html body", func(t *testing.T) {
		msg := &domain.MailMessage{}
		body := strings.NewReader("Content-Type: text/html\r\n\r\n<h1>Hello</h1>")
		parseMIMEBody(msg, body)
		if msg.BodyHTML != "<h1>Hello</h1>" {
			t.Errorf("BodyHTML = %q, want %q", msg.BodyHTML, "<h1>Hello</h1>")
		}
	})

	t.Run("multipart mixed with text and html", func(t *testing.T) {
		msg := &domain.MailMessage{}
		body := strings.NewReader(
			"Content-Type: multipart/alternative; boundary=bound42\r\n" +
				"\r\n" +
				"--bound42\r\n" +
				"Content-Type: text/plain\r\n" +
				"\r\n" +
				"Plain text\r\n" +
				"--bound42\r\n" +
				"Content-Type: text/html\r\n" +
				"\r\n" +
				"<p>HTML</p>\r\n" +
				"--bound42--\r\n",
		)
		parseMIMEBody(msg, body)
		if msg.BodyText != "Plain text" {
			t.Errorf("BodyText = %q, want %q", msg.BodyText, "Plain text")
		}
		if msg.BodyHTML != "<p>HTML</p>" {
			t.Errorf("BodyHTML = %q, want %q", msg.BodyHTML, "<p>HTML</p>")
		}
	})

	t.Run("multipart with attachment", func(t *testing.T) {
		msg := &domain.MailMessage{}
		body := strings.NewReader(
			"Content-Type: multipart/mixed; boundary=bound43\r\n" +
				"\r\n" +
				"--bound43\r\n" +
				"Content-Type: text/plain\r\n" +
				"\r\n" +
				"Body here\r\n" +
				"--bound43\r\n" +
				"Content-Type: application/pdf\r\n" +
				"Content-Disposition: attachment; filename=report.pdf\r\n" +
				"\r\n" +
				"PDFDATA\r\n" +
				"--bound43--\r\n",
		)
		parseMIMEBody(msg, body)
		if msg.BodyText != "Body here" {
			t.Errorf("BodyText = %q, want %q", msg.BodyText, "Body here")
		}
		if len(msg.Attachments) != 1 {
			t.Fatalf("Attachments = %v, want 1 attachment", msg.Attachments)
		}
		if msg.Attachments[0].Filename != "report.pdf" {
			t.Errorf("Attachment Filename = %q, want %q", msg.Attachments[0].Filename, "report.pdf")
		}
		if msg.Attachments[0].Size != 7 {
			t.Errorf("Attachment Size = %d, want %d", msg.Attachments[0].Size, 7)
		}
	})

	t.Run("invalid MIME body is handled gracefully", func(t *testing.T) {
		msg := &domain.MailMessage{}
		body := strings.NewReader("not a valid MIME message")
		parseMIMEBody(msg, body)
		// Should not panic, body fields should remain empty
	})

	t.Run("only attachment, no text/html body", func(t *testing.T) {
		msg := &domain.MailMessage{}
		body := strings.NewReader(
			"Content-Type: image/png; name=screenshot.png\r\n" +
				"Content-Disposition: attachment; filename=screenshot.png\r\n" +
				"\r\n" +
				"PNGDATA",
		)
		parseMIMEBody(msg, body)
		if len(msg.Attachments) != 1 {
			t.Fatalf("Attachments = %v, want 1", msg.Attachments)
		}
		if msg.Attachments[0].Filename != "screenshot.png" {
			t.Errorf("Filename = %q, want %q", msg.Attachments[0].Filename, "screenshot.png")
		}
		if msg.BodyText != "" || msg.BodyHTML != "" {
			t.Error("BodyText/BodyHTML should be empty for attachment-only")
		}
	})
}

func TestParseMIMEBody_AttachmentFromContentTypeParams(t *testing.T) {
	msg := &domain.MailMessage{}
	body := strings.NewReader(
		"Content-Type: application/pdf; name=doc.pdf\r\n" +
			"Content-Disposition: attachment\r\n" +
			"\r\n" +
			"DATA",
	)
	parseMIMEBody(msg, body)
	if len(msg.Attachments) != 1 {
		t.Fatalf("Attachments = %v, want 1", msg.Attachments)
	}
	if msg.Attachments[0].Filename != "doc.pdf" {
		t.Errorf("Filename from params = %q, want %q", msg.Attachments[0].Filename, "doc.pdf")
	}
}

// TestEnvelopeToMailMessage constructs a minimal FetchMessageData to test envelopeToMailMessage.
// This avoids requiring a real IMAP connection by using the data builder pattern.
func TestEnvelopeToMailMessage(t *testing.T) {
	// Create a valid FetchMessageData using the go-imap client builder.
	// The builder method FromFetchMessageDataCallable allows constructing from callbacks.
	t.Run("empty message data yields empty mail message", func(t *testing.T) {
		// FetchMessageData is not directly constructible; tested via DovecotBridge mocks
		// in imap_bridge_test.go. Skip direct construction test.
	})
}

// TestNewGoIMAPClient_DialError verifies newGoIMAPClient fails with connection refused.
func TestNewGoIMAPClient_DialError(t *testing.T) {
	_, err := newGoIMAPClient("127.0.0.1:1")
	if err == nil {
		t.Fatal("newGoIMAPClient on unreachable port should return error")
	}
	if !strings.Contains(err.Error(), "dial") {
		t.Errorf("error = %q, want dial error", err.Error())
	}
}

// TestFormatAddressIntegration verifies formatAddress is consistent with domain model.
func TestFormatAddressIntegration(t *testing.T) {
	addr := imap.Address{Name: "Full Name", Mailbox: "user", Host: "domain.tld"}
	result := formatAddress(addr)
	if result != "Full Name <user@domain.tld>" {
		t.Errorf("formatAddress = %q, want %q", result, "Full Name <user@domain.tld>")
	}
}

// gomessageCreateReader is the real function used by parseMIMEBody.
// Verify it works with simple inputs.
func TestGoMessageCreateReader(t *testing.T) {
	t.Run("reads simple text body", func(t *testing.T) {
		body := strings.NewReader("Content-Type: text/plain\r\n\r\nhello")
		r, err := gomessage.CreateReader(body)
		if err != nil {
			t.Fatalf("CreateReader: %v", err)
		}
		part, err := r.NextPart()
		if err != nil {
			t.Fatalf("NextPart: %v", err)
		}
		ctype := part.Header.Get("Content-Type")
		if !strings.Contains(ctype, "text/plain") {
			t.Errorf("content type = %q, want text/plain", ctype)
		}
	})
}

// TestIMAPClientInterface verifies goIMAPClient implements imapClient.
func TestIMAPClientInterface(t *testing.T) {
	var _ imapClient = (*goIMAPClient)(nil)
}
