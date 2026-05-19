package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/MYusufEka/oxmail/internal/mail"
)

// InboxResponse is the envelope for paginated inbox messages.
type InboxResponse struct {
	Messages   []domain.MailMessage `json:"messages"`
	Pagination domain.Pagination   `json:"pagination"`
}

// SearchResponse is the envelope for search results.
type SearchResponse struct {
	Messages []domain.MailMessage `json:"messages"`
}

// MarkReadRequest is the payload for marking a message read/unread.
type MarkReadRequest struct {
	Read bool `json:"read"`
}

// MailHandler handles HTTP requests for webmail operations.
type MailHandler struct {
	bridge mail.IMAPBridge
}

// NewMailHandler creates a new MailHandler with the given IMAP bridge.
func NewMailHandler(bridge mail.IMAPBridge) *MailHandler {
	return &MailHandler{bridge: bridge}
}

// RegisterRoutes mounts mail routes onto an existing router.
func (h *MailHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/mail", func(r chi.Router) {
		r.Get("/inbox", h.handleInbox)
		r.Get("/messages/{id}", h.handleGetMessage)
		r.Delete("/messages/{id}", h.handleDeleteMessage)
		r.Patch("/messages/{id}", h.handlePatchMessage)
		r.Get("/search", h.handleSearch)
	})
}

func (h *MailHandler) handleInbox(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	messages, total, err := h.bridge.FetchInbox(user, password, page, limit)
	if err != nil {
		h.handleBridgeError(w, err)
		return
	}

	if messages == nil {
		messages = []domain.MailMessage{}
	}

	resp := InboxResponse{
		Messages: messages,
		Pagination: domain.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *MailHandler) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	uid, err := h.parseMessageID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "message ID must be a number")
		return
	}

	msg, err := h.bridge.FetchMessage(user, password, uid)
	if err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
}

func (h *MailHandler) handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	uid, err := h.parseMessageID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "message ID must be a number")
		return
	}

	if err := h.bridge.DeleteMessage(user, password, uid); err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *MailHandler) handlePatchMessage(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	uid, err := h.parseMessageID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "message ID must be a number")
		return
	}

	var req MarkReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if err := h.bridge.MarkRead(user, password, uid, req.Read); err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *MailHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "missing_query", "q parameter is required")
		return
	}

	messages, err := h.bridge.SearchMessages(user, password, query)
	if err != nil {
		h.handleBridgeError(w, err)
		return
	}

	if messages == nil {
		messages = []domain.MailMessage{}
	}

	resp := SearchResponse{Messages: messages}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *MailHandler) extractCredentials(r *http.Request) (user, password string, ok bool) {
	user = r.URL.Query().Get("user")
	if user == "" {
		return "", "", false
	}
	password = r.URL.Query().Get("password")
	if password == "" {
		// Fall back to Authorization header (Basic auth)
		password = r.Header.Get("X-Mail-Password")
	}
	return user, password, true
}

func (h *MailHandler) parseMessageID(r *http.Request) (uint32, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}

func (h *MailHandler) handleBridgeError(w http.ResponseWriter, err error) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "not found") {
		writeError(w, http.StatusNotFound, "not_found", "message not found")
		return
	}
	if strings.Contains(errMsg, "invalid credentials") || strings.Contains(errMsg, "login") {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", "failed to process request")
}
