package mail

import (
	"fmt"

	"github.com/microcosm-cc/bluemonday"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// IMAPBridge defines the interface for IMAP operations against Dovecot.
type IMAPBridge interface {
	FetchInbox(user, password string, page, limit int) ([]domain.MailMessage, int, error)
	FetchMessage(user, password string, uid uint32) (*domain.MailMessage, error)
	DeleteMessage(user, password string, uid uint32) error
	MarkRead(user, password string, uid uint32, read bool) error
	SearchMessages(user, password string, query string) ([]domain.MailMessage, error)
}

// imapClient abstracts the IMAP connection for testability.
type imapClient interface {
	Login(user, password string) error
	FetchMessages(page, limit int) ([]domain.MailMessage, int, error)
	FetchMessage(uid uint32) (*domain.MailMessage, error)
	DeleteMessage(uid uint32) error
	MarkRead(uid uint32, read bool) error
	SearchMessages(query string) ([]domain.MailMessage, error)
	Logout() error
}

// DovecotBridge implements IMAPBridge by connecting to Dovecot's IMAP server.
type DovecotBridge struct {
	address   string
	sanitizer *bluemonday.Policy
	newClient func(addr string) (imapClient, error)
}

// NewDovecotBridge creates a new DovecotBridge connecting to the given address.
func NewDovecotBridge(address string) *DovecotBridge {
	policy := bluemonday.UGCPolicy()
	return &DovecotBridge{
		address:   address,
		sanitizer: policy,
		newClient: newRealIMAPClient,
	}
}

// FetchInbox retrieves a paginated list of messages from the user's INBOX.
func (b *DovecotBridge) FetchInbox(user, password string, page, limit int) ([]domain.MailMessage, int, error) {
	client, err := b.connect(user, password)
	if err != nil {
		return nil, 0, err
	}
	defer client.Logout()

	msgs, total, err := client.FetchMessages(page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch inbox: %w", err)
	}
	return msgs, total, nil
}

// FetchMessage retrieves a single message by UID with sanitized HTML body.
func (b *DovecotBridge) FetchMessage(user, password string, uid uint32) (*domain.MailMessage, error) {
	client, err := b.connect(user, password)
	if err != nil {
		return nil, err
	}
	defer client.Logout()

	msg, err := client.FetchMessage(uid)
	if err != nil {
		return nil, fmt.Errorf("fetch message %d: %w", uid, err)
	}

	if msg.BodyHTML != "" {
		sanitizer := b.sanitizer
		if sanitizer == nil {
			sanitizer = bluemonday.UGCPolicy()
		}
		msg.BodyHTML = sanitizer.Sanitize(msg.BodyHTML)
	}

	return msg, nil
}

// DeleteMessage removes a message by UID (flags as \Deleted and expunges).
func (b *DovecotBridge) DeleteMessage(user, password string, uid uint32) error {
	client, err := b.connect(user, password)
	if err != nil {
		return err
	}
	defer client.Logout()

	if err := client.DeleteMessage(uid); err != nil {
		return fmt.Errorf("delete message %d: %w", uid, err)
	}
	return nil
}

// MarkRead sets or clears the \Seen flag on a message.
func (b *DovecotBridge) MarkRead(user, password string, uid uint32, read bool) error {
	client, err := b.connect(user, password)
	if err != nil {
		return err
	}
	defer client.Logout()

	if err := client.MarkRead(uid, read); err != nil {
		return fmt.Errorf("mark read %d: %w", uid, err)
	}
	return nil
}

// SearchMessages performs an IMAP SEARCH for messages matching the query.
func (b *DovecotBridge) SearchMessages(user, password string, query string) ([]domain.MailMessage, error) {
	client, err := b.connect(user, password)
	if err != nil {
		return nil, err
	}
	defer client.Logout()

	msgs, err := client.SearchMessages(query)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	return msgs, nil
}

func (b *DovecotBridge) connect(user, password string) (imapClient, error) {
	client, err := b.newClient(b.address)
	if err != nil {
		return nil, fmt.Errorf("connect to IMAP: %w", err)
	}

	if err := client.Login(user, password); err != nil {
		client.Logout()
		return nil, fmt.Errorf("IMAP login: %w", err)
	}

	return client, nil
}

// newRealIMAPClient creates a real IMAP client connection.
// This is the production implementation that connects to Dovecot.
func newRealIMAPClient(addr string) (imapClient, error) {
	return newGoIMAPClient(addr)
}
