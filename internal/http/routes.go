package http

import (
	"net/http"

	"github.com/subscription-reconciler/internal/repository/postgres"
)

func NewRouter(store *postgres.Store) http.Handler {
	mux := http.NewServeMux()

	return mux
}
