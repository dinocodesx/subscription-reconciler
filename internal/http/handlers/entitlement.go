package handlers

import (
	"log"
	"net/http"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

type EntitlementHandler struct {
	store *postgres.Store
}

type entitlementResponse struct {
	Active        bool    `json:"active"`
	Source        string  `json:"source"`
	ExpiresAt     *string `json:"expiresAt"`
	LastChangedAt string  `json:"lastChangedAt"`
	Reason        string  `json:"reason"`
}

func NewEntitlementHandler(store *postgres.Store) *EntitlementHandler {
	return &EntitlementHandler{store: store}
}

// Handle reads the already-computed entitlement projection for a single user.
func (h *EntitlementHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	entitlement, err := h.store.GetEntitlement(r.Context(), userID)
	if err != nil {
		log.Printf("get entitlement for user %s: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, entitlementResponseFromDomain(entitlement))
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
