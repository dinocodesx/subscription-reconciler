package handlers

import (
	"log"
	"net/http"

	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

type EntitlementHandler struct {
	store *postgres.Store
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
