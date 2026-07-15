package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
)

// MailConfigApplier abstracts Dovecot config application for testability.
type MailConfigApplier interface {
	ApplyUserConfig(users []domain.User) error
	CreateMaildir(email string, domainName string) error
	MaildirSize(email string, domainName string) (int64, error)
}

// UserLister provides listing all users for config regeneration.
type UserLister interface {
	List(ctx context.Context, params domain.UserListParams) ([]domain.User, int, error)
}

// DomainNameResolver resolves a domain ID to its name.
type DomainNameResolver interface {
	GetByID(ctx context.Context, id int64) (string, error)
}

// UserListResponse is the envelope for paginated user lists.
type UserListResponse struct {
	Data       []domain.User    `json:"data"`
	Pagination domain.Pagination `json:"pagination"`
}

// UsersHandler handles HTTP requests for user/mailbox management.
type UsersHandler struct {
	service        *domain.UserService
	mailConfig     MailConfigApplier
	userLister     UserLister
	domainResolver DomainNameResolver
	router         *chi.Mux
}

// NewUsersHandler creates a new UsersHandler with routes configured.
func NewUsersHandler(service *domain.UserService, mailConfig MailConfigApplier) *UsersHandler {
	h := &UsersHandler{
		service:    service,
		mailConfig: mailConfig,
		userLister: service,
		router:     chi.NewRouter(),
	}

	h.router.Route("/api/users", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Patch("/{id}", h.handleUpdate)
		r.Delete("/{id}", h.handleDelete)
	})

	return h
}

// WithDomainResolver sets the domain name resolver (required for domain-scoped routes).
func (h *UsersHandler) WithDomainResolver(resolver DomainNameResolver) *UsersHandler {
	h.domainResolver = resolver
	return h
}

// Router returns the chi router for testing.
func (h *UsersHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts user routes onto an existing router.
func (h *UsersHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/users", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Patch("/{id}", h.handleUpdate)
		r.Delete("/{id}", h.handleDelete)
	})
}

func (h *UsersHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "email and password are required")
		return
	}

	user, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	if h.mailConfig != nil {
		domainName := extractDomainFromEmail(user.Email)
		if err := h.mailConfig.CreateMaildir(user.Email, domainName); err != nil {
			slog.Error("failed to create maildir", "email", user.Email, "error", err)
		}
		if err := h.applyAllUsersConfig(r.Context()); err != nil {
			slog.Error("failed to apply dovecot config after create", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UsersHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	user, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UsersHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	if h.mailConfig != nil {
		domainName := extractDomainFromEmail(user.Email)
		size, _ := h.mailConfig.MaildirSize(user.Email, domainName)
		user.StorageUsed = size
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UsersHandler) handleList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	domainFilter := r.URL.Query().Get("domain")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	users, total, err := h.service.List(r.Context(), domain.UserListParams{
		Domain: domainFilter,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}

	if users == nil {
		users = []domain.User{}
	}

	if h.mailConfig != nil {
		for i := range users {
			domainName := extractDomainFromEmail(users[i].Email)
			size, _ := h.mailConfig.MaildirSize(users[i].Email, domainName)
			users[i].StorageUsed = size
		}
	}

	resp := UserListResponse{
		Data: users,
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

func (h *UsersHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete user")
		return
	}

	if h.mailConfig != nil {
		if err := h.applyAllUsersConfig(r.Context()); err != nil {
			slog.Error("failed to apply dovecot config after delete", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *UsersHandler) handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrDomainNotFound) {
		writeError(w, http.StatusBadRequest, "domain_not_found", "domain does not exist")
		return
	}
	if errors.Is(err, domain.ErrUserExists) {
		writeError(w, http.StatusConflict, "user_exists", "user already exists")
		return
	}
	if errors.Is(err, domain.ErrInvalidEmail) {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email address")
		return
	}

	writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
}

// applyAllUsersConfig fetches all users and regenerates Dovecot passdb/userdb.
func (h *UsersHandler) applyAllUsersConfig(ctx context.Context) error {
	users, _, err := h.userLister.List(ctx, domain.UserListParams{Page: 1, Limit: 10000})
	if err != nil {
		return err
	}
	return h.mailConfig.ApplyUserConfig(users)
}

func (h *UsersHandler) handleListByDomain(w http.ResponseWriter, r *http.Request) {
	domainIDStr := chi.URLParam(r, "domainID")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid domain ID")
		return
	}

	if h.domainResolver == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "domain resolver not configured")
		return
	}

	domainName, err := h.domainResolver.GetByID(r.Context(), domainID)
	if err != nil {
		if errors.Is(err, domain.ErrDomainNotFound) {
			writeError(w, http.StatusNotFound, "domain_not_found", "domain not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to resolve domain")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	users, total, err := h.service.List(r.Context(), domain.UserListParams{
		Domain: domainName,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}

	if users == nil {
		users = []domain.User{}
	}

	if h.mailConfig != nil {
		for i := range users {
			size, _ := h.mailConfig.MaildirSize(users[i].Email, domainName)
			users[i].StorageUsed = size
		}
	}

	resp := UserListResponse{
		Data: users,
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

func extractDomainFromEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
