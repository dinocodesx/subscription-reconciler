package handlers

import "github.com/subscription-reconciler/internal/repository/postgres"

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
