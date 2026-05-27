package handlers

import "github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"

type MarketplaceWebhookHandler struct {
	store *postgres.Store
}

type marketplaceWebhookRequest struct {
	UserIDs []string `json:"userIds"`
}
