package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

// DevHandler provides dev-only endpoints for testing.
type DevHandler struct {
	sender Sender
	logger *slog.Logger
}

// NewDevHandler creates a new DevHandler.
func NewDevHandler(sender Sender) *DevHandler {
	return &DevHandler{
		sender: sender,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

// RegisterRoutes registers dev-only routes on the router.
func (h *DevHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/dev/send-test", h.handleSendTest)
}

// IsDevMode returns true when OXMAIL_MODE is set to "dev".
func IsDevMode() bool {
	return os.Getenv("OXMAIL_MODE") == "dev"
}

func (h *DevHandler) handleSendTest(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	subject := r.URL.Query().Get("subject")
	body := r.URL.Query().Get("body")

	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "from and to query parameters are required")
		return
	}

	if subject == "" {
		subject = "Test email from Oxmail dev mode"
	}

	if body == "" {
		body = "This is a test email sent via the dev endpoint."
	}

	messageID, err := h.sender.Send(from, []string{to}, nil, subject, body, "")
	if err != nil {
		h.logger.Error("dev send-test failed", "error", err, "from", from, "to", to)
		writeError(w, http.StatusInternalServerError, "SEND_FAILED", "failed to send test email")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"messageId": messageID,
		"status":    "sent",
		"from":      from,
		"to":        to,
		"subject":   subject,
	})
}
