package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/config"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/MYusufEka/oxmail/internal/mail"
)

// DomainListResponse is the envelope for paginated domain lists.
type DomainListResponse struct {
	Domains    []domain.Domain   `json:"domains"`
	Pagination domain.Pagination `json:"pagination"`
}

// DomainsHandler handles HTTP requests for domain management.
type DomainsHandler struct {
	service        *domain.DomainService
	generator      *config.PostfixDomainsGenerator
	postfixManager *mail.PostfixManager
	router         *chi.Mux
}

// NewDomainsHandler creates a new DomainsHandler with routes configured.
func NewDomainsHandler(service *domain.DomainService, configPath string, postfixManager *mail.PostfixManager) *DomainsHandler {
	h := &DomainsHandler{
		service:        service,
		generator:      config.NewPostfixDomainsGenerator(configPath),
		postfixManager: postfixManager,
		router:         chi.NewRouter(),
	}

	h.router.Route("/api/domains", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{name}", h.handleGet)
		r.Delete("/{name}", h.handleDelete)
	})

	return h
}

// Router returns the chi router for testing.
func (h *DomainsHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts domain routes onto an existing router.
func (h *DomainsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/domains", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{name}", h.handleGet)
		r.Delete("/{name}", h.handleDelete)
	})
}

func (h *DomainsHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	created, err := h.service.Create(r.Context(), req.Name)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.regenerateConfig(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *DomainsHandler) handleList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	domains, total, err := h.service.List(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list domains")
		return
	}

	if domains == nil {
		domains = []domain.Domain{}
	}

	resp := DomainListResponse{
		Domains: domains,
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

func (h *DomainsHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	d, err := h.service.Get(r.Context(), name)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

func (h *DomainsHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := h.service.Delete(r.Context(), name); err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.regenerateConfig(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *DomainsHandler) handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrDomainExists) {
		writeError(w, http.StatusConflict, "domain_exists", "domain already exists")
		return
	}
	if errors.Is(err, domain.ErrDomainNotFound) {
		writeError(w, http.StatusNotFound, "domain_not_found", "domain not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidDomain) {
		writeError(w, http.StatusBadRequest, "invalid_domain", "invalid domain name")
		return
	}

	writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
}

func (h *DomainsHandler) regenerateConfig(r *http.Request) {
	domains, _, err := h.service.List(r.Context(), 1, 10000)
	if err != nil {
		return
	}
	h.generator.Generate(domains)
	if h.postfixManager != nil {
		h.postfixManager.ApplyDomainConfig(domains)
	}
}
