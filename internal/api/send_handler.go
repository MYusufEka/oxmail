package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
)

// Sender abstracts email sending for testability.
type Sender interface {
	Send(from string, to []string, cc []string, subject, bodyText, bodyHTML string) (string, error)
}

// SendHandler handles the POST /api/mail/send endpoint.
type SendHandler struct {
	sender    Sender
	logger    *slog.Logger
	rateLimit int
	mu        sync.Mutex
	sends     map[string][]time.Time
}

// NewSendHandler creates a new SendHandler with default rate limit of 50/hour.
func NewSendHandler(sender Sender) *SendHandler {
	return &SendHandler{
		sender:    sender,
		logger:    slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		rateLimit: 50,
		sends:     make(map[string][]time.Time),
	}
}

// RegisterRoutes registers the send mail route.
func (h *SendHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/mail/send", h.handleSend)
}

func (h *SendHandler) handleSend(w http.ResponseWriter, r *http.Request) {
	authUser := r.URL.Query().Get("auth_user")
	if authUser == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "auth_user query parameter is required")
		return
	}

	var req domain.SendMailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	if len(req.To) == 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "at least one recipient is required")
		return
	}

	if req.From != authUser {
		writeError(w, http.StatusForbidden, "SENDER_MISMATCH", "from address must match authenticated user")
		return
	}

	if h.isRateLimited(authUser) {
		writeError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "rate limit exceeded, try again later")
		return
	}

	messageID, err := h.sender.Send(req.From, req.To, req.CC, req.Subject, req.BodyText, req.BodyHTML)
	if err != nil {
		h.logger.Error("failed to send email", "error", err, "from", req.From)
		writeError(w, http.StatusInternalServerError, "SEND_FAILED", "failed to send email")
		return
	}

	h.recordSend(authUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(domain.SendMailResponse{
		MessageID: messageID,
		Status:    "sent",
	})
}

func (h *SendHandler) isRateLimited(user string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Hour)

	timestamps := h.sends[user]
	valid := timestamps[:0]
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			valid = append(valid, ts)
		}
	}
	h.sends[user] = valid

	return len(valid) >= h.rateLimit
}

func (h *SendHandler) recordSend(user string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sends[user] = append(h.sends[user], time.Now())
}


