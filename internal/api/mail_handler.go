package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	bridge        mail.IMAPBridge
	queueExecutor mail.CommandExecutor
}

func NewMailHandler(bridge mail.IMAPBridge, queueExecutor mail.CommandExecutor) *MailHandler {
	return &MailHandler{bridge: bridge, queueExecutor: queueExecutor}
}

// FoldersResponse is the envelope for folder list.
type FoldersResponse struct {
	Folders []domain.MailFolder `json:"folders"`
}

// ThreadsResponse is the envelope for grouped thread results.
type ThreadsResponse struct {
	Threads []domain.MailThread `json:"threads"`
}

// RegisterRoutes mounts mail routes onto an existing router.
func (h *MailHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/mail", func(r chi.Router) {
		r.Get("/queue", h.handleQueue)
		r.Get("/inbox", h.handleInbox)
		r.Get("/search", h.handleSearch)
		r.Get("/folders", h.handleListFolders)
		r.Post("/folders", h.handleCreateFolder)
		r.Delete("/folders/{folder}", h.handleDeleteFolder)
		r.Patch("/folders/{folder}", h.handleRenameFolder)
		r.Get("/folders/{folder}/messages", h.handleFolderMessages)
		r.Get("/folders/{folder}/threads", h.handleThreads)
		r.Post("/messages/{uid}/move", h.handleMoveMessage)
		r.Get("/messages/{id}", h.handleGetMessage)
		r.Delete("/messages/{id}", h.handleDeleteMessage)
		r.Patch("/messages/{id}", h.handlePatchMessage)
		r.Patch("/messages/{id}/toggle-read", h.handlePatchMessage)
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/inbox", h.handleInbox)
			r.Get("/messages/{id}", h.handleGetMessage)
			r.Delete("/messages/{id}", h.handleDeleteMessage)
			r.Patch("/messages/{id}", h.handlePatchMessage)
			r.Patch("/messages/{id}/toggle-read", h.handlePatchMessage)
		})
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

func (h *MailHandler) handleListFolders(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	folders, err := h.bridge.ListFolders(user, password)
	if err != nil {
		h.handleBridgeError(w, err)
		return
	}

	if folders == nil {
		folders = []domain.MailFolder{}
	}

	resp := FoldersResponse{Folders: folders}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *MailHandler) handleFolderMessages(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	folder := chi.URLParam(r, "folder")
	if folder == "" {
		writeError(w, http.StatusBadRequest, "missing_folder", "folder parameter is required")
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

	messages, total, err := h.bridge.FetchFolderMessages(user, password, folder, page, limit)
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

func (h *MailHandler) handleThreads(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	folder := chi.URLParam(r, "folder")
	if folder == "" {
		writeError(w, http.StatusBadRequest, "missing_folder", "folder parameter is required")
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

	messages, _, err := h.bridge.FetchFolderMessages(user, password, folder, page, limit)
	if err != nil {
		h.handleBridgeError(w, err)
		return
	}

	if messages == nil {
		messages = []domain.MailMessage{}
	}

	threads := groupMessagesByThread(messages)

	resp := ThreadsResponse{Threads: threads}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// groupMessagesByThread groups a flat message list into threads by ThreadID.
func groupMessagesByThread(messages []domain.MailMessage) []domain.MailThread {
	type threadAccumulator struct {
		threadID   string
		subject    string
		msgList    []domain.MailMessage
		lastDate   time.Time
		participants map[string]struct{}
		unreadCount  int
	}

	groups := make(map[string]*threadAccumulator)
	// Maintain insertion order for deterministic output.
	var orderedKeys []string

	for _, msg := range messages {
		tid := msg.ThreadID
		if tid == "" {
			// Fallback: use message ID as thread ID
			tid = msg.MessageID
		}
		if tid == "" {
			// Last fallback: use subject as thread key
			tid = msg.Subject
		}
		if tid == "" {
			tid = "orphaned"
		}

		group, exists := groups[tid]
		if !exists {
			group = &threadAccumulator{
				threadID:     tid,
				subject:      msg.Subject,
				participants: make(map[string]struct{}),
			}
			groups[tid] = group
			orderedKeys = append(orderedKeys, tid)
		}

		group.msgList = append(group.msgList, msg)
		if msg.ReceivedAt.After(group.lastDate) {
			group.lastDate = msg.ReceivedAt
		}
		if msg.From != "" {
			group.participants[msg.From] = struct{}{}
		}
		for _, to := range msg.To {
			if to != "" {
				group.participants[to] = struct{}{}
			}
		}
		if !msg.Read {
			group.unreadCount++
		}
	}

	threads := make([]domain.MailThread, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		group := groups[key]
		threads = append(threads, domain.MailThread{
			ThreadID:         group.threadID,
			Subject:          group.subject,
			Messages:         group.msgList,
			LastDate:         group.lastDate,
			ParticipantCount: len(group.participants),
			UnreadCount:      group.unreadCount,
		})
	}

	return threads
}

func (h *MailHandler) extractCredentials(r *http.Request) (user, password string, ok bool) {
	user = r.URL.Query().Get("user")
	password = r.URL.Query().Get("password")

	if user == "" {
		user, password, ok = r.BasicAuth()
		if !ok || user == "" {
			return "", "", false
		}
	}

	if password == "" {
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

// CreateFolderRequest is the payload for creating a new folder.
type CreateFolderRequest struct {
	Name string `json:"name"`
}

// RenameFolderRequest is the payload for renaming a folder.
type RenameFolderRequest struct {
	NewName string `json:"new_name"`
}

// MoveMessageRequest is the payload for moving a message between folders.
type MoveMessageRequest struct {
	FromFolder string `json:"from_folder"`
	ToFolder   string `json:"to_folder"`
}

func (h *MailHandler) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	var req CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "folder name is required")
		return
	}

	if err := h.bridge.CreateFolder(user, password, req.Name); err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created", "name": req.Name})
}

func (h *MailHandler) handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	folder := chi.URLParam(r, "folder")
	if folder == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "folder name is required")
		return
	}

	if err := h.bridge.DeleteFolder(user, password, folder); err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *MailHandler) handleRenameFolder(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	folder := chi.URLParam(r, "folder")
	if folder == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "folder name is required")
		return
	}

	var req RenameFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	if req.NewName == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "new_name is required")
		return
	}

	if err := h.bridge.RenameFolder(user, password, folder, req.NewName); err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "renamed", "name": req.NewName})
}

func (h *MailHandler) handleMoveMessage(w http.ResponseWriter, r *http.Request) {
	user, password, ok := h.extractCredentials(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "missing_user", "user parameter is required")
		return
	}

	uidStr := chi.URLParam(r, "uid")
	uid64, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "message UID must be a number")
		return
	}

	var req MoveMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	if req.FromFolder == "" || req.ToFolder == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "from_folder and to_folder are required")
		return
	}

	if err := h.bridge.MoveMessage(user, password, uint32(uid64), req.FromFolder, req.ToFolder); err != nil {
		h.handleBridgeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "moved"})
}

func (h *MailHandler) handleQueue(w http.ResponseWriter, r *http.Request) {
	status, err := mail.GetQueueStatus(h.queueExecutor)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "queue_error", "failed to get queue status")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
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
