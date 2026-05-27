package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

type MarketplaceWebhookHandler struct {
	store *postgres.Store
}

type marketplaceWebhookRequest struct {
	UserIDs []string `json:"userIds"`
}

func NewMarketplaceWebhookHandler(store *postgres.Store) *MarketplaceWebhookHandler {
	return &MarketplaceWebhookHandler{store: store}
}

func (h *MarketplaceWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var payload marketplaceWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if len(payload.UserIDs) == 0 {
		writeError(w, http.StatusBadRequest, "userIds is required")
		return
	}

	if err := h.store.RevokeMarketplaceUsers(r.Context(), payload.UserIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"revokedCount": len(payload.UserIDs),
	})
}
