package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/MYusufEka/oxmail/internal/domain"
)

// RegisterDKIMRoutes mounts DKIM endpoints on the given router.
func RegisterDKIMRoutes(r chi.Router, dkimService *domain.DKIMService) {
	r.Route("/api/domains/{domain}/dkim", func(r chi.Router) {
		r.Post("/", handleDKIMGenerate(dkimService))
		r.Get("/", handleDKIMGet(dkimService))
		r.Delete("/", handleDKIMDelete(dkimService))
		r.Post("/rotate", handleDKIMRotate(dkimService))
	})
}

func handleDKIMGenerate(dkimService *domain.DKIMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domainName := chi.URLParam(r, "domain")

		result, err := dkimService.Generate(domainName, "default")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, result)
	}
}

func handleDKIMGet(dkimService *domain.DKIMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domainName := chi.URLParam(r, "domain")

		result, err := dkimService.Get(domainName, "default")
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, "not_found", err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func handleDKIMDelete(dkimService *domain.DKIMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domainName := chi.URLParam(r, "domain")

		err := dkimService.Delete(domainName, "default")
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, "not_found", err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func handleDKIMRotate(dkimService *domain.DKIMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domainName := chi.URLParam(r, "domain")

		result, err := dkimService.Rotate(domainName, "default")
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, "not_found", err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		response := map[string]interface{}{
			"domain":    result.Domain,
			"selector":  result.Selector,
			"publicKey": result.PublicKey,
			"dnsRecord": result.DNSRecord,
			"createdAt": result.CreatedAt,
			"message":   "DKIM key rotated. Update your DNS TXT record. Allow 24-48h for propagation.",
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
