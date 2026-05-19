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
	messages    map[uint32]*domain.MailMessage
	loginErr    error
	fetchErr    error
	deleteErr   error
	markReadErr error
	searchErr   error
}

func (m *mockIMAPClient) Login(user, password string) error {
	return m.loginErr
}

func (m *mockIMAPClient) FetchMessages(page, limit int) ([]domain.MailMessage, int, error) {
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

func (m *mockIMAPClient) Logout() error {
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
