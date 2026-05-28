package carrier

import (
	"context"
	"net/http"
)

type Client interface {
	PlanStatus(ctx context.Context, userID string) (string, error)
}

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

type planResponse struct {
	Status string `json:"status"`
}
