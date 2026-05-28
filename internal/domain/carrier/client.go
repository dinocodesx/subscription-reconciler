package carrier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
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

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPClient) PlanStatus(ctx context.Context, userID string) (string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	base.Path = path.Join(base.Path, "/mock/carrier/plan")
	query := base.Query()
	query.Set("userId", userID)
	base.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call carrier plan endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("carrier plan endpoint returned %d", resp.StatusCode)
	}

	var payload planResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode carrier plan endpoint response: %w", err)
	}

	return payload.Status, nil
}
