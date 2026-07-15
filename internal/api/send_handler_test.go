package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSMTPSender implements the Sender interface for testing.
type mockSMTPSender struct {
	sendFunc func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error)
}

func (m *mockSMTPSender) Send(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
	return m.sendFunc(from, to, cc, subject, bodyText, bodyHTML, attachments)
}

func newTestSendHandler(sender Sender) (*SendHandler, *chi.Mux) {
	handler := NewSendHandler(sender, nil)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	return handler, router
}

func TestSendHandler_Success(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "<abc-123@example.com>", nil
		},
	}

	_, router := newTestSendHandler(sender)

	body := domain.SendMailRequest{
		From:     "user@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "Hello",
		BodyText: "Hi there",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user=user@example.com", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.SendMailResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "<abc-123@example.com>", resp.MessageID)
	assert.Equal(t, "sent", resp.Status)
}

func TestSendHandler_SenderMismatch(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "", nil
		},
	}

	_, router := newTestSendHandler(sender)

	body := domain.SendMailRequest{
		From:     "attacker@example.com",
		To:       []string{"victim@example.com"},
		Subject:  "Spoofed",
		BodyText: "bad",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user=user@example.com", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "SENDER_MISMATCH", resp.Error.Code)
}

func TestSendHandler_MissingAuthUser(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "", nil
		},
	}

	_, router := newTestSendHandler(sender)

	body := domain.SendMailRequest{
		From:     "user@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyText: "body",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/mail/send", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSendHandler_InvalidJSON(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "", nil
		},
	}

	_, router := newTestSendHandler(sender)

	req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user=user@example.com", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendHandler_EmptyTo(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "", nil
		},
	}

	_, router := newTestSendHandler(sender)

	body := domain.SendMailRequest{
		From:     "user@example.com",
		To:       []string{},
		Subject:  "Test",
		BodyText: "body",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user=user@example.com", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendHandler_RateLimit(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "<msg@example.com>", nil
		},
	}

	handler := NewSendHandler(sender, nil)
	handler.rateLimit = 3 // Low limit for testing
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	body := domain.SendMailRequest{
		From:     "user@example.com",
		To:       []string{"to@example.com"},
		Subject:  "Test",
		BodyText: "body",
	}
	payload, _ := json.Marshal(body)

	// Send 3 emails successfully
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user=user@example.com", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d should succeed", i+1)
	}

	// 4th should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user=user@example.com", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	var resp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", resp.Error.Code)
}

func TestSendHandler_RateLimit_DifferentUsers(t *testing.T) {
	sender := &mockSMTPSender{
		sendFunc: func(from string, to []string, cc []string, subject, bodyText, bodyHTML string, attachments []domain.SendMailAttachment) (string, error) {
			return "<msg@example.com>", nil
		},
	}

	handler := NewSendHandler(sender, nil)
	handler.rateLimit = 2
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	sendAs := func(user string) int {
		body := domain.SendMailRequest{
			From:     user,
			To:       []string{"to@example.com"},
			Subject:  "Test",
			BodyText: "body",
		}
		payload, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/mail/send?auth_user="+user, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Code
	}

	// User A sends 2
	assert.Equal(t, http.StatusOK, sendAs("a@example.com"))
	assert.Equal(t, http.StatusOK, sendAs("a@example.com"))
	// User A rate limited
	assert.Equal(t, http.StatusTooManyRequests, sendAs("a@example.com"))
	// User B still fine
	assert.Equal(t, http.StatusOK, sendAs("b@example.com"))
}
