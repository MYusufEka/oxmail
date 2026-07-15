package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/mail"
)

// SieveHandler handles HTTP requests for sieve script management.
type SieveHandler struct {
	manager *mail.SieveManager
}

// NewSieveHandler creates a new SieveHandler.
func NewSieveHandler(manager *mail.SieveManager) *SieveHandler {
	return &SieveHandler{manager: manager}
}

// RegisterRoutes mounts sieve endpoints on the given router.
func (h *SieveHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/mail/filters", func(r chi.Router) {
		r.Get("/{email}", h.handleGet)
		r.Post("/{email}", h.handleSet)
		r.Delete("/{email}", h.handleDelete)
	})
	r.Route("/api/mail/vacation", func(r chi.Router) {
		r.Get("/{email}", h.handleVacationGet)
		r.Post("/{email}", h.handleVacationSet)
		r.Delete("/{email}", h.handleVacationDelete)
	})
}

// sieveSetRequest is the JSON body for setting a sieve script.
type sieveSetRequest struct {
	Script string `json:"script"`
}

// sieveResponse is the JSON response for sieve operations.
type sieveResponse struct {
	Email  string `json:"email"`
	Script string `json:"script,omitempty"`
	Active bool   `json:"active,omitempty"`
	Status string `json:"status,omitempty"`
}

func (h *SieveHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return
	}

	script, err := h.manager.GetScript(email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	resp := sieveResponse{
		Email:  email,
		Script: script,
		Active: script != "",
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *SieveHandler) handleSet(w http.ResponseWriter, r *http.Request) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return
	}

	var req sieveSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Script == "" {
		writeError(w, http.StatusBadRequest, "invalid_script", "script content is required")
		return
	}

	if err := h.manager.SetScript(email, req.Script); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	resp := sieveResponse{
		Email:  email,
		Status: "saved",
		Active: true,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *SieveHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return
	}

	if err := h.manager.DeleteScript(email); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	resp := sieveResponse{
		Email:  email,
		Status: "deleted",
	}
	writeJSON(w, http.StatusOK, resp)
}

type vacationSetRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Enabled bool   `json:"enabled"`
}

// vacationResponse is the JSON response for vacation operations.
type vacationResponse struct {
	Email   string `json:"email"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status,omitempty"`
}

func (h *SieveHandler) handleVacationGet(w http.ResponseWriter, r *http.Request) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return
	}

	subject, body, enabled, err := h.manager.GetVacation(email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vacationResponse{
		Email:   email,
		Subject: subject,
		Body:    body,
		Enabled: enabled,
	})
}

func (h *SieveHandler) handleVacationSet(w http.ResponseWriter, r *http.Request) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return
	}

	var req vacationSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Enabled && req.Subject == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "subject is required when enabling vacation")
		return
	}

	if err := h.manager.SetVacation(email, req.Subject, req.Body, req.Enabled); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}
	writeJSON(w, http.StatusOK, vacationResponse{
		Email:   email,
		Enabled: req.Enabled,
		Status:  status,
	})
}

func (h *SieveHandler) handleVacationDelete(w http.ResponseWriter, r *http.Request) {
	email, err := url.PathUnescape(chi.URLParam(r, "email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if email == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "email parameter is required")
		return
	}

	if err := h.manager.DeleteVacation(email); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vacationResponse{
		Email:   email,
		Enabled: false,
		Status:  "deleted",
	})
}
