package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
)

type AuditHandler struct {
	service *domain.AuditService
	router  *chi.Mux
}

func NewAuditHandler(service *domain.AuditService) *AuditHandler {
	h := &AuditHandler{
		service: service,
		router:  chi.NewRouter(),
	}
	h.router.Get("/api/audit", h.handleList)
	return h
}

func (h *AuditHandler) Router() *chi.Mux {
	return h.router
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/audit", h.handleList)
}

func (h *AuditHandler) handleList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	actor := r.URL.Query().Get("actor")
	action := r.URL.Query().Get("action")

	ctx := r.Context()

	var (
		entries []domain.AuditEntry
		total   int
		err     error
	)

	if actor != "" || action != "" {
		entries, err = h.service.ListFiltered(ctx, actor, action, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to list audit log")
			return
		}
		total, err = h.service.CountFiltered(ctx, actor, action)
	} else {
		entries, err = h.service.List(ctx, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to list audit log")
			return
		}
		total, err = h.service.Count(ctx)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to count audit log")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"entries": entries,
		"total":   total,
	})
}
