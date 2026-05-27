package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON keeps the handlers focused on request validation and domain flow.
func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError returns a small JSON error envelope for consistent API responses.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
