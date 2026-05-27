package http

import (
	"net/http"

	"github.com/dinocodesx/subscription-reconciler/internal/http/handlers"
	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

// NewRouter registers the assignment endpoints on Go's standard library mux.
func NewRouter(store *postgres.Store) http.Handler {
	mux := http.NewServeMux()

	storeHandler := handlers.NewStoreWebhookHandler(store)

	mux.HandleFunc("POST /webhooks/store", storeHandler.Handle)

	return mux
}
