package mail

import (
	"fmt"

	"github.com/microcosm-cc/bluemonday"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// IMAPBridge defines the interface for IMAP operations against Dovecot.
type IMAPBridge interface {
	FetchInbox(user, password string, page, limit int) ([]domain.MailMessage, int, error)
	FetchFolderMessages(user, password, folder string, page, limit int) ([]domain.MailMessage, int, error)
	FetchMessage(user, password string, uid uint32) (*domain.MailMessage, error)
	DeleteMessage(user, password string, uid uint32) error
	MarkRead(user, password string, uid uint32, read bool) error
	SearchMessages(user, password string, query string) ([]domain.MailMessage, error)
	ListFolders(user, password string) ([]domain.MailFolder, error)
	CreateFolder(user, password, folderName string) error
	DeleteFolder(user, password, folderName string) error
	RenameFolder(user, password, oldName, newName string) error
	MoveMessage(user, password string, uid uint32, fromFolder, toFolder string) error
}

// imapClient abstracts the IMAP connection for testability.
type imapClient interface {
	Login(user, password string) error
	FetchMessages(page, limit int) ([]domain.MailMessage, int, error)
	FetchFolderMessages(folder string, page, limit int) ([]domain.MailMessage, int, error)
	FetchMessage(uid uint32) (*domain.MailMessage, error)
	DeleteMessage(uid uint32) error
	MarkRead(uid uint32, read bool) error
	SearchMessages(query string) ([]domain.MailMessage, error)
	ListFolders() ([]domain.MailFolder, error)
	CreateFolder(name string) error
	DeleteFolder(name string) error
	RenameFolder(oldName, newName string) error
	MoveMessage(uid uint32, fromFolder, toFolder string) error
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

// ListFolders retrieves all IMAP folders with unread counts.
func (b *DovecotBridge) ListFolders(user, password string) ([]domain.MailFolder, error) {
	client, err := b.connect(user, password)
	if err != nil {
		return nil, err
	}
	defer client.Logout()

	folders, err := client.ListFolders()
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	return folders, nil
}

// FetchFolderMessages retrieves messages from a specific IMAP folder.
func (b *DovecotBridge) FetchFolderMessages(user, password, folder string, page, limit int) ([]domain.MailMessage, int, error) {
	client, err := b.connect(user, password)
	if err != nil {
		return nil, 0, err
	}
	defer client.Logout()

	msgs, total, err := client.FetchFolderMessages(folder, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch folder %s: %w", folder, err)
	}
	return msgs, total, nil
}

// CreateFolder creates a new IMAP folder.
func (b *DovecotBridge) CreateFolder(user, password, folderName string) error {
	client, err := b.connect(user, password)
	if err != nil {
		return err
	}
	defer client.Logout()

	if err := client.CreateFolder(folderName); err != nil {
		return fmt.Errorf("create folder %s: %w", folderName, err)
	}
	return nil
}

// DeleteFolder removes an IMAP folder.
func (b *DovecotBridge) DeleteFolder(user, password, folderName string) error {
	client, err := b.connect(user, password)
	if err != nil {
		return err
	}
	defer client.Logout()

	if err := client.DeleteFolder(folderName); err != nil {
		return fmt.Errorf("delete folder %s: %w", folderName, err)
	}
	return nil
}

// RenameFolder renames an IMAP folder.
func (b *DovecotBridge) RenameFolder(user, password, oldName, newName string) error {
	client, err := b.connect(user, password)
	if err != nil {
		return err
	}
	defer client.Logout()

	if err := client.RenameFolder(oldName, newName); err != nil {
		return fmt.Errorf("rename folder %s to %s: %w", oldName, newName, err)
	}
	return nil
}

// MoveMessage moves a message from one folder to another.
func (b *DovecotBridge) MoveMessage(user, password string, uid uint32, fromFolder, toFolder string) error {
	client, err := b.connect(user, password)
	if err != nil {
		return err
	}
	defer client.Logout()

	if err := client.MoveMessage(uid, fromFolder, toFolder); err != nil {
		return fmt.Errorf("move message %d: %w", uid, err)
	}
	return nil
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
