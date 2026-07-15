package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
)

// BouncesHandler handles HTTP requests for bounce record CRUD.
type BouncesHandler struct {
	service *domain.BounceService
	router  *chi.Mux
}

// NewBouncesHandler creates a new BouncesHandler with routes configured.
func NewBouncesHandler(service *domain.BounceService) *BouncesHandler {
	h := &BouncesHandler{
		service: service,
		router:  chi.NewRouter(),
	}

	h.router.Route("/api/bounces", func(r chi.Router) {
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Delete("/{id}", h.handleDelete)
	})

	return h
}

// Router returns the chi router for testing.
func (h *BouncesHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts bounce routes onto an existing router.
func (h *BouncesHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/bounces", func(r chi.Router) {
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Delete("/{id}", h.handleDelete)
	})
}

func (h *BouncesHandler) handleList(w http.ResponseWriter, r *http.Request) {
	recipient := r.URL.Query().Get("recipient")

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "invalid_param", "limit must be a non-negative integer")
			return
		}
		limit = parsed
	}

	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "invalid_param", "offset must be a non-negative integer")
			return
		}
		offset = parsed
	}

	bounces, err := h.service.ListBounces(r.Context(), domain.BounceFilter{
		Recipient: recipient,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list bounces")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"bounces": bounces})
}

func (h *BouncesHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid bounce ID")
		return
	}

	bounce, err := h.service.GetBounce(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBounceNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "bounce not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get bounce")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bounce)
}

func (h *BouncesHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid bounce ID")
		return
	}

	if err := h.service.DeleteBounce(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrBounceNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "bounce not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete bounce")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
