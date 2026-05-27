package http

import (
	"net/http"

	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

// NewRouter registers the assignment endpoints on Go's standard library mux.
func NewRouter(store *postgres.Store) http.Handler {
	mux := http.NewServeMux()

	return mux
}
