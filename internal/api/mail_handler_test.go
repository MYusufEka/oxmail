package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIMAPBridge implements mail.IMAPBridge for handler tests.
type mockIMAPBridge struct {
	messages    map[uint32]*domain.MailMessage
	fetchErr    error
	deleteErr   error
	markReadErr error
	searchErr   error
}

func (m *mockIMAPBridge) FetchInbox(user, password string, page, limit int) ([]domain.MailMessage, int, error) {
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

func (m *mockIMAPBridge) FetchMessage(user, password string, uid uint32) (*domain.MailMessage, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	msg, ok := m.messages[uid]
	if !ok {
		return nil, errors.New("message not found")
	}
	return msg, nil
}

func (m *mockIMAPBridge) DeleteMessage(user, password string, uid uint32) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.messages[uid]; !ok {
		return errors.New("message not found")
	}
	delete(m.messages, uid)
	return nil
}

func (m *mockIMAPBridge) MarkRead(user, password string, uid uint32, read bool) error {
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

func (m *mockIMAPBridge) SearchMessages(user, password string, query string) ([]domain.MailMessage, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var results []domain.MailMessage
	for _, msg := range m.messages {
		if strings.Contains(msg.Subject, query) || strings.Contains(msg.BodyText, query) {
			results = append(results, *msg)
		}
	}
	return results, nil
}

func newTestMailMessages() map[uint32]*domain.MailMessage {
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
		},
	}
}

func setupMailHandler() (*MailHandler, *chi.Mux) {
	mock := &mockIMAPBridge{messages: newTestMailMessages()}
	handler := NewMailHandler(mock)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	return handler, router
}

func TestMailHandler_GetInbox(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/inbox?user=alice@local.test&password=secret&page=1&limit=50", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp InboxResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Pagination.Total)
	assert.Len(t, resp.Messages, 2)
}

func TestMailHandler_GetInbox_DefaultPagination(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/inbox?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp InboxResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Pagination.Page)
	assert.Equal(t, 50, resp.Pagination.Limit)
}

func TestMailHandler_GetInbox_MissingUser(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/inbox", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMailHandler_GetMessage(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/messages/1?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var msg domain.MailMessage
	err := json.NewDecoder(rec.Body).Decode(&msg)
	require.NoError(t, err)
	assert.Equal(t, "Hello Alice", msg.Subject)
	assert.Equal(t, "bob@local.test", msg.From)
}

func TestMailHandler_GetMessage_NotFound(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/messages/999?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMailHandler_GetMessage_InvalidID(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/messages/abc?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMailHandler_DeleteMessage(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/mail/messages/1?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMailHandler_DeleteMessage_NotFound(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/mail/messages/999?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMailHandler_PatchMessage_MarkRead(t *testing.T) {
	_, router := setupMailHandler()

	body := strings.NewReader(`{"read": true}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/mail/messages/1?user=alice@local.test&password=secret", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMailHandler_PatchMessage_MarkUnread(t *testing.T) {
	_, router := setupMailHandler()

	body := strings.NewReader(`{"read": false}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/mail/messages/2?user=alice@local.test&password=secret", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMailHandler_PatchMessage_InvalidBody(t *testing.T) {
	_, router := setupMailHandler()

	body := strings.NewReader(`invalid json`)
	req := httptest.NewRequest(http.MethodPatch, "/api/mail/messages/1?user=alice@local.test&password=secret", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMailHandler_Search(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/search?user=alice@local.test&password=secret&q=Hello", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp SearchResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Messages)
}

func TestMailHandler_Search_MissingQuery(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/search?user=alice@local.test&password=secret", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMailHandler_Search_MissingUser(t *testing.T) {
	_, router := setupMailHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/mail/search?q=Hello", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
