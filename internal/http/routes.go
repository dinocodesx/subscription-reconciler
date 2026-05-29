package http

import (
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/dinocodesx/subscription-reconciler/internal/http/handlers"
	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

// NewRouter registers the assignment endpoints on Go's standard library mux.
func NewRouter(store *postgres.Store) http.Handler {
	mux := http.NewServeMux()

	storeHandler := handlers.NewStoreWebhookHandler(store)
	marketplaceHandler := handlers.NewMarketplaceWebhookHandler(store)
	entitlementHandler := handlers.NewEntitlementHandler(store)
	timelineHandler := handlers.NewTimelineHandler(store)

	mockCarrierHandler := handlers.NewMockCarrierHandler(rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)))

	mux.HandleFunc("POST /webhooks/store", storeHandler.Handle)
	mux.HandleFunc("POST /webhooks/marketplace/revoke", marketplaceHandler.Handle)
	mux.HandleFunc("GET /users/{id}/entitlement", entitlementHandler.Handle)
	mux.HandleFunc("GET /users/{id}/timeline", timelineHandler.Handle)

	mux.HandleFunc("GET /mock/carrier/plan", mockCarrierHandler.Handle)

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	mux.HandleFunc("GET /api-docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger_ui.html")
	})

	return mux
}
