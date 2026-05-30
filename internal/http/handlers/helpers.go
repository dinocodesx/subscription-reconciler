package handlers

import (
	"encoding/json"
	"net/http"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
)

type entitlementResponse struct {
	Active        bool    `json:"active"`
	Source        string  `json:"source"`
	ExpiresAt     *string `json:"expiresAt"`
	LastChangedAt string  `json:"lastChangedAt"`
	Reason        string  `json:"reason"`
}

func entitlementResponseFromDomain(entitlement entdom.Entitlement) entitlementResponse {
	var expiresAt *string
	if entitlement.ExpiresAt != nil {
		value := entitlement.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		expiresAt = &value
	}

	return entitlementResponse{
		Active:        entitlement.Active,
		Source:        string(entitlement.Source),
		ExpiresAt:     expiresAt,
		LastChangedAt: entitlement.LastChangedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Reason:        entitlement.Reason,
	}
}

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
