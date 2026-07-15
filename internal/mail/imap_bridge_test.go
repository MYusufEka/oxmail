package mail

import (
	"errors"
	"testing"
	"time"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIMAPClient implements imapClient for testing.
type mockIMAPClient struct {
	messages       map[uint32]*domain.MailMessage
	folders        []domain.MailFolder
	loginErr       error
	fetchErr       error
	deleteErr      error
	markReadErr    error
	searchErr      error
	listErr        error
	createErr      error
	deleteFldrErr  error
	renameErr      error
	moveErr        error
	createdFolders []string
	renamedFrom    string
	renamedTo      string
	movedUID       uint32
	movedFrom      string
	movedTo        string
}

func (m *mockIMAPClient) Login(user, password string) error {
	return m.loginErr
}

func (m *mockIMAPClient) FetchMessages(page, limit int) ([]domain.MailMessage, int, error) {
	return m.FetchFolderMessages("INBOX", page, limit)
}

func (m *mockIMAPClient) FetchFolderMessages(folder string, page, limit int) ([]domain.MailMessage, int, error) {
	if m.fetchErr != nil {
		return nil, 0, m.fetchErr
	}
	msgs := make([]domain.MailMessage, 0, len(m.messages))
	for _, msg := range m.messages {
		msgs = append(msgs, *msg)
	}
	total := len(msgs)
	start := (page - 1) * limit
	if start >= total {
		return []domain.MailMessage{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return msgs[start:end], total, nil
}

func (m *mockIMAPClient) FetchMessage(uid uint32) (*domain.MailMessage, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	msg, ok := m.messages[uid]
	if !ok {
		return nil, errors.New("message not found")
	}
	return msg, nil
}

func (m *mockIMAPClient) DeleteMessage(uid uint32) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.messages, uid)
	return nil
}

func (m *mockIMAPClient) MarkRead(uid uint32, read bool) error {
	if m.markReadErr != nil {
		return m.markReadErr
	}
	msg, ok := m.messages[uid]
	if !ok {
		return errors.New("message not found")
	}
	msg.Read = read
	return nil
}

func (m *mockIMAPClient) SearchMessages(query string) ([]domain.MailMessage, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var results []domain.MailMessage
	for _, msg := range m.messages {
		results = append(results, *msg)
	}
	return results, nil
}

func (m *mockIMAPClient) ListFolders() ([]domain.MailFolder, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.folders != nil {
		return m.folders, nil
	}
	return []domain.MailFolder{
		{Name: "INBOX", Delimiter: "/", Unread: 2},
		{Name: "Sent", Delimiter: "/", Unread: 0},
		{Name: "Drafts", Delimiter: "/", Unread: 0},
		{Name: "Trash", Delimiter: "/", Unread: 0},
	}, nil
}

func (m *mockIMAPClient) Logout() error {
	return nil
}

func (m *mockIMAPClient) CreateFolder(name string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createdFolders = append(m.createdFolders, name)
	return nil
}

func (m *mockIMAPClient) DeleteFolder(name string) error {
	if m.deleteFldrErr != nil {
		return m.deleteFldrErr
	}
	filtered := m.folders[:0]
	for _, f := range m.folders {
		if f.Name != name {
			filtered = append(filtered, f)
		}
	}
	m.folders = filtered
	return nil
}

func (m *mockIMAPClient) RenameFolder(oldName, newName string) error {
	if m.renameErr != nil {
		return m.renameErr
	}
	m.renamedFrom = oldName
	m.renamedTo = newName
	return nil
}

func (m *mockIMAPClient) MoveMessage(uid uint32, fromFolder, toFolder string) error {
	if m.moveErr != nil {
		return m.moveErr
	}
	m.movedUID = uid
	m.movedFrom = fromFolder
	m.movedTo = toFolder
	return nil
}

func newTestMessages() map[uint32]*domain.MailMessage {
	return map[uint32]*domain.MailMessage{
		1: {
			ID:         1,
			From:       "bob@local.test",
			To:         []string{"alice@local.test"},
			Subject:    "Hello Alice",
			BodyText:   "Hi there!",
			BodyHTML:   "<p>Hi there!</p>",
			Read:       false,
			ReceivedAt: time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
			MessageID:  "<msg1@local.test>",
		},
		2: {
			ID:         2,
			From:       "carol@local.test",
			To:         []string{"alice@local.test"},
			Subject:    "Meeting tomorrow",
			BodyText:   "Let's meet at 10am",
			BodyHTML:   "<p>Let's meet at 10am</p>",
			Read:       true,
			ReceivedAt: time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC),
			MessageID:  "<msg2@local.test>",
			InReplyTo:  "<msg1@local.test>",
		},
		3: {
			ID:         3,
			From:       "dave@local.test",
			To:         []string{"alice@local.test"},
			Subject:    "Report attached",
			BodyText:   "See attachment",
			BodyHTML:   "<p>See attachment</p><script>alert('xss')</script>",
			Read:       false,
			ReceivedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
			MessageID:  "<msg3@local.test>",
			Attachments: []domain.Attachment{
				{Filename: "report.pdf", Size: 1024},
			},
		},
	}
}

func TestDovecotBridge_FetchInbox(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msgs, total, err := bridge.FetchInbox("alice@local.test", "password", 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, msgs, 3)
}

func TestDovecotBridge_FetchInbox_Pagination(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msgs, total, err := bridge.FetchInbox("alice@local.test", "password", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, msgs, 2)

	msgs, total, err = bridge.FetchInbox("alice@local.test", "password", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, msgs, 1)
}

func TestDovecotBridge_FetchInbox_LoginError(t *testing.T) {
	mock := &mockIMAPClient{
		messages: newTestMessages(),
		loginErr: errors.New("invalid credentials"),
	}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	_, _, err := bridge.FetchInbox("alice@local.test", "wrong", 1, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestDovecotBridge_FetchMessage(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msg, err := bridge.FetchMessage("alice@local.test", "password", 1)
	require.NoError(t, err)
	assert.Equal(t, "Hello Alice", msg.Subject)
	assert.Equal(t, "bob@local.test", msg.From)
}

func TestDovecotBridge_FetchMessage_NotFound(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	_, err := bridge.FetchMessage("alice@local.test", "password", 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDovecotBridge_FetchMessage_SanitizesHTML(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msg, err := bridge.FetchMessage("alice@local.test", "password", 3)
	require.NoError(t, err)
	assert.NotContains(t, msg.BodyHTML, "<script>")
	assert.Contains(t, msg.BodyHTML, "<p>See attachment</p>")
}

func TestDovecotBridge_DeleteMessage(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.DeleteMessage("alice@local.test", "password", 1)
	require.NoError(t, err)

	_, err = bridge.FetchMessage("alice@local.test", "password", 1)
	require.Error(t, err)
}

func TestDovecotBridge_MarkRead(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.MarkRead("alice@local.test", "password", 1, true)
	require.NoError(t, err)

	msg, err := bridge.FetchMessage("alice@local.test", "password", 1)
	require.NoError(t, err)
	assert.True(t, msg.Read)
}

func TestDovecotBridge_MarkUnread(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.MarkRead("alice@local.test", "password", 2, false)
	require.NoError(t, err)

	msg, err := bridge.FetchMessage("alice@local.test", "password", 2)
	require.NoError(t, err)
	assert.False(t, msg.Read)
}

func TestDovecotBridge_SearchMessages(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msgs, err := bridge.SearchMessages("alice@local.test", "password", "Hello")
	require.NoError(t, err)
	assert.NotEmpty(t, msgs)
}

func TestSearchMessages_ReturnsBodyText(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msgs, err := bridge.SearchMessages("alice@local.test", "password", "meeting")
	require.NoError(t, err)
	require.NotEmpty(t, msgs)
	for _, msg := range msgs {
		assert.NotEmpty(t, msg.BodyText, "search result should have BodyText populated")
		assert.LessOrEqual(t, len(msg.BodyText), 500, "BodyText should be truncated to 500 chars")
	}
}

func TestDovecotBridge_ConnectionError(t *testing.T) {
	bridge := &DovecotBridge{
		address: "dovecot:143",
		newClient: func(addr string) (imapClient, error) {
			return nil, errors.New("connection refused")
		},
	}

	_, _, err := bridge.FetchInbox("alice@local.test", "password", 1, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestDovecotBridge_ListFolders(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	folders, err := bridge.ListFolders("alice@local.test", "password")
	require.NoError(t, err)
	assert.Len(t, folders, 4)
	assert.Equal(t, "INBOX", folders[0].Name)
	assert.Equal(t, 2, folders[0].Unread)
	assert.Equal(t, "Sent", folders[1].Name)
	assert.Equal(t, 0, folders[1].Unread)
}

func TestDovecotBridge_ListFolders_Error(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages(), listErr: errors.New("list failed")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	_, err := bridge.ListFolders("alice@local.test", "password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list failed")
}

func TestDovecotBridge_FetchFolderMessages(t *testing.T) {
	mock := &mockIMAPClient{messages: newTestMessages()}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	msgs, total, err := bridge.FetchFolderMessages("alice@local.test", "password", "Sent", 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, msgs, 3)
}

func TestDovecotBridge_FetchFolderMessages_LoginError(t *testing.T) {
	mock := &mockIMAPClient{
		messages: newTestMessages(),
		loginErr: errors.New("invalid credentials"),
	}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	_, _, err := bridge.FetchFolderMessages("alice@local.test", "wrong", "Sent", 1, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestDovecotBridge_CreateFolder(t *testing.T) {
	mock := &mockIMAPClient{}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.CreateFolder("alice@local.test", "password", "Projects")
	require.NoError(t, err)
	assert.Equal(t, []string{"Projects"}, mock.createdFolders)
}

func TestDovecotBridge_CreateFolder_Error(t *testing.T) {
	mock := &mockIMAPClient{createErr: errors.New("mailbox exists")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.CreateFolder("alice@local.test", "password", "INBOX")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mailbox exists")
}

func TestDovecotBridge_CreateFolder_LoginError(t *testing.T) {
	mock := &mockIMAPClient{loginErr: errors.New("invalid credentials")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.CreateFolder("alice@local.test", "wrong", "Projects")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestDovecotBridge_DeleteFolder(t *testing.T) {
	mock := &mockIMAPClient{
		folders: []domain.MailFolder{
			{Name: "INBOX"}, {Name: "Projects"},
		},
	}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.DeleteFolder("alice@local.test", "password", "Projects")
	require.NoError(t, err)
	assert.Len(t, mock.folders, 1)
	assert.Equal(t, "INBOX", mock.folders[0].Name)
}

func TestDovecotBridge_DeleteFolder_Error(t *testing.T) {
	mock := &mockIMAPClient{deleteFldrErr: errors.New("folder not found")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.DeleteFolder("alice@local.test", "password", "Missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "folder not found")
}

func TestDovecotBridge_RenameFolder(t *testing.T) {
	mock := &mockIMAPClient{}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.RenameFolder("alice@local.test", "password", "OldName", "NewName")
	require.NoError(t, err)
	assert.Equal(t, "OldName", mock.renamedFrom)
	assert.Equal(t, "NewName", mock.renamedTo)
}

func TestDovecotBridge_RenameFolder_Error(t *testing.T) {
	mock := &mockIMAPClient{renameErr: errors.New("rename failed")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.RenameFolder("alice@local.test", "password", "OldName", "NewName")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rename failed")
}

func TestDovecotBridge_MoveMessage(t *testing.T) {
	mock := &mockIMAPClient{}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.MoveMessage("alice@local.test", "password", 42, "INBOX", "Archive")
	require.NoError(t, err)
	assert.Equal(t, uint32(42), mock.movedUID)
	assert.Equal(t, "INBOX", mock.movedFrom)
	assert.Equal(t, "Archive", mock.movedTo)
}

func TestDovecotBridge_MoveMessage_Error(t *testing.T) {
	mock := &mockIMAPClient{moveErr: errors.New("copy failed")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.MoveMessage("alice@local.test", "password", 42, "INBOX", "Archive")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "copy failed")
}

func TestDovecotBridge_MoveMessage_LoginError(t *testing.T) {
	mock := &mockIMAPClient{loginErr: errors.New("invalid credentials")}
	bridge := &DovecotBridge{
		address:   "dovecot:143",
		newClient: func(addr string) (imapClient, error) { return mock, nil },
	}

	err := bridge.MoveMessage("alice@local.test", "wrong", 42, "INBOX", "Archive")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}
