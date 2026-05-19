package api

import (
	"encoding/json"
	"net/http"
)

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":  "ok",
		"version": "0.1.0",
	}

	json.NewEncoder(w).Encode(response)
}
