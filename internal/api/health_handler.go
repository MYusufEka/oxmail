package api

import (
	"encoding/json"
	"net/http"

	"github.com/MYusufEka/oxmail/internal/health"
)

// newHealthHandler creates an http.HandlerFunc that uses the health service.
func newHealthHandler(svc *health.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := svc.Check(r.Context())

		w.Header().Set("Content-Type", "application/json")

		switch result.Status {
		case "unhealthy":
			w.WriteHeader(http.StatusServiceUnavailable)
		case "degraded":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(result)
	}
}
