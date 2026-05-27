package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	storeevent "github.com/dinocodesx/subscription-reconciler/internal/domain/store"
	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

type StoreWebhookHandler struct {
	store *postgres.Store
}

type storeWebhookRequest struct {
	EventID     string `json:"eventId"`
	UserID      string `json:"userId"`
	Type        string `json:"type"`
	EventTimeMS int64  `json:"eventTimeMs"`
	ProductID   string `json:"productId"`
}

func NewStoreWebhookHandler(store *postgres.Store) *StoreWebhookHandler {
	return &StoreWebhookHandler{store: store}
}

func (h *StoreWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var payload storeWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	eventType := storeevent.EventType(strings.ToUpper(payload.Type))
	if payload.EventID == "" || payload.UserID == "" || payload.ProductID == "" || payload.EventTimeMS == 0 || !eventType.Valid() {
		writeError(w, http.StatusBadRequest, "missing or invalid webhook fields")
		return
	}

	inserted, entitlement, err := h.store.ApplyStoreEvent(r.Context(), storeevent.Event{
		EventID:    payload.EventID,
		UserID:     payload.UserID,
		Type:       eventType,
		EventTime:  time.UnixMilli(payload.EventTimeMS).UTC(),
		ProductID:  payload.ProductID,
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"duplicate":   !inserted,
		"entitlement": entitlementResponseFromDomain(entitlement),
	})
}
