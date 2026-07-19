package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
)

// ContactsHandler handles HTTP requests for contact/address book CRUD.
type ContactsHandler struct {
	service *domain.ContactService
	router  *chi.Mux
}

// NewContactsHandler creates a new ContactsHandler with routes configured.
func NewContactsHandler(service *domain.ContactService) *ContactsHandler {
	h := &ContactsHandler{
		service: service,
		router:  chi.NewRouter(),
	}

	h.router.Route("/api/contacts", func(r chi.Router) {
		r.Get("/{userEmail}", h.handleList)
		r.Post("/{userEmail}", h.handleCreate)
		r.Put("/{userEmail}/{contactID}", h.handleUpdate)
		r.Delete("/{userEmail}/{contactID}", h.handleDelete)
	})

	return h
}

// Router returns the chi router for testing.
func (h *ContactsHandler) Router() *chi.Mux {
	return h.router
}

// RegisterRoutes mounts contact routes onto an existing router.
func (h *ContactsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/contacts", func(r chi.Router) {
		r.Get("/{userEmail}", h.handleList)
		r.Post("/{userEmail}", h.handleCreate)
		r.Put("/{userEmail}/{contactID}", h.handleUpdate)
		r.Delete("/{userEmail}/{contactID}", h.handleDelete)
	})
}

func (h *ContactsHandler) handleList(w http.ResponseWriter, r *http.Request) {
	userEmail, err := url.PathUnescape(chi.URLParam(r, "userEmail"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if userEmail == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "user email is required")
		return
	}

	contacts, err := h.service.List(r.Context(), userEmail)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list contacts")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contacts)
}

func (h *ContactsHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userEmail, err := url.PathUnescape(chi.URLParam(r, "userEmail"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if userEmail == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "user email is required")
		return
	}

	var req domain.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Name == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name and email are required")
		return
	}

	contact, err := h.service.Create(r.Context(), userEmail, req)
	if err != nil {
		if errors.Is(err, domain.ErrContactExists) {
			writeError(w, http.StatusConflict, "contact_exists", "contact with this email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create contact")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

func (h *ContactsHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	userEmail, err := url.PathUnescape(chi.URLParam(r, "userEmail"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email", "invalid email encoding")
		return
	}
	if userEmail == "" {
		writeError(w, http.StatusBadRequest, "invalid_email", "user email is required")
		return
	}

	contactID, err := strconv.ParseInt(chi.URLParam(r, "contactID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid contact ID")
		return
	}

	var req domain.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	contact, err := h.service.Update(r.Context(), contactID, req)
	if err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "contact not found")
			return
		}
		if errors.Is(err, domain.ErrContactExists) {
			writeError(w, http.StatusConflict, "contact_exists", "contact with this email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update contact")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contact)
}

func (h *ContactsHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	contactID, err := strconv.ParseInt(chi.URLParam(r, "contactID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid contact ID")
		return
	}

	if err := h.service.Delete(r.Context(), contactID); err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "contact not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete contact")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
