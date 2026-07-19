package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
)

// SignatureHandler handles HTTP requests for webmail signatures.
type SignatureHandler struct {
	service *domain.SignatureService
	router  *chi.Mux
}

// NewSignatureHandler creates a new SignatureHandler with routes configured.
func NewSignatureHandler(service *domain.SignatureService) *SignatureHandler {
	h := &SignatureHandler{
		service: service,
		router:  chi.NewRouter(),
	}
	h.RegisterRoutes(h.router)
	return h
}

// Router returns the chi router for testing.
func (h *SignatureHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts signature routes onto an existing router.
func (h *SignatureHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/mail/signature", func(r chi.Router) {
		r.Get("/{email}", h.handleGet)
		r.Post("/{email}", h.handleUpsert)
		r.Delete("/{email}", h.handleDelete)
	})
}

type signatureUpsertRequest struct {
	Content string `json:"content"`
	Enabled bool   `json:"enabled"`
}

func (h *SignatureHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	email, ok := signatureEmailParam(w, r)
	if !ok {
		return
	}

	signature, err := h.service.Get(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get signature")
		return
	}

	writeJSON(w, http.StatusOK, signature)
}

func (h *SignatureHandler) handleUpsert(w http.ResponseWriter, r *http.Request) {
	email, ok := signatureEmailParam(w, r)
	if !ok {
		return
	}

	var req signatureUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	signature, err := h.service.Upsert(r.Context(), email, req.Content, req.Enabled)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to save signature")
		return
	}

	writeJSON(w, http.StatusOK, signature)
}

func (h *SignatureHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	email, ok := signatureEmailParam(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), email); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete signature")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func signatureEmailParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return "", false
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return "", false
	}
	return email, true
}
