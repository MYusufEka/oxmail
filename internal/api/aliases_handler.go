package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
)

// AliasHandler handles HTTP requests for alias management.
type AliasHandler struct {
	service *domain.AliasService
	router  *chi.Mux
}

// NewAliasHandler creates a new AliasHandler with routes configured.
func NewAliasHandler(service *domain.AliasService) *AliasHandler {
	h := &AliasHandler{
		service: service,
		router:  chi.NewRouter(),
	}

	h.router.Route("/api/aliases", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Delete("/{id}", h.handleDelete)
	})

	return h
}

// Router returns the chi router for testing.
func (h *AliasHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts alias routes onto an existing router.
func (h *AliasHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/aliases", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Delete("/{id}", h.handleDelete)
	})
}

func (h *AliasHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.SourceAddress == "" || req.DestinationAddress == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "sourceAddress and destinationAddress are required")
		return
	}

	alias, err := h.service.Create(req.SourceAddress, req.DestinationAddress)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alias)
}

func (h *AliasHandler) handleList(w http.ResponseWriter, r *http.Request) {
	aliases, err := h.service.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list aliases")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(aliases)
}

func (h *AliasHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid alias ID")
		return
	}

	alias, err := h.service.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", "alias not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get alias")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alias)
}

func (h *AliasHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid alias ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", "alias not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete alias")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *AliasHandler) handleServiceError(w http.ResponseWriter, err error) {
	msg := err.Error()

	if errors.Is(err, domain.ErrDomainNotFound) {
		writeError(w, http.StatusNotFound, "domain_not_found", msg)
		return
	}
	if strings.Contains(msg, "already exists") {
		writeError(w, http.StatusConflict, "alias_exists", msg)
		return
	}
	if strings.Contains(msg, "circular") {
		writeError(w, http.StatusBadRequest, "circular_alias", msg)
		return
	}
	if strings.Contains(msg, "invalid") {
		writeError(w, http.StatusBadRequest, "invalid_request", msg)
		return
	}

	writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
}

func parseIDParam(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.ErrorResponse{
		Error: domain.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
