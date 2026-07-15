package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
)

// StatsHandler handles HTTP requests for mail statistics.
type StatsHandler struct {
	service *domain.StatsService
	router  *chi.Mux
}

// NewStatsHandler creates a new StatsHandler with routes configured.
func NewStatsHandler(service *domain.StatsService) *StatsHandler {
	h := &StatsHandler{
		service: service,
		router:  chi.NewRouter(),
	}

	h.router.Route("/api/stats", func(r chi.Router) {
		r.Get("/", h.handleGetStats)
		r.Get("/summary", h.handleGetSummary)
	})

	return h
}

// Router returns the chi router for testing.
func (h *StatsHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts stats routes onto an existing router.
func (h *StatsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/stats", func(r chi.Router) {
		r.Get("/", h.handleGetStats)
		r.Get("/summary", h.handleGetSummary)
	})
}

func (h *StatsHandler) handleGetStats(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		parsed, err := strconv.Atoi(daysStr)
		if err != nil || parsed < 1 {
			writeError(w, http.StatusBadRequest, "invalid_days", "days must be a positive integer")
			return
		}
		days = parsed
	}

	stats, err := h.service.GetStats(r.Context(), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to fetch stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *StatsHandler) handleGetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to fetch summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
