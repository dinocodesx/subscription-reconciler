package handlers

import (
	"math/rand/v2"
	"net/http"
	"sync"
)

type MockCarrierHandler struct {
	random *rand.Rand
	mu     sync.Mutex
}

func NewMockCarrierHandler(random *rand.Rand) *MockCarrierHandler {
	return &MockCarrierHandler{random: random}
}

func (h *MockCarrierHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("userId") == "" {
		writeError(w, http.StatusBadRequest, "userId is required")
		return
	}

	h.mu.Lock()
	value := h.random.IntN(100)
	h.mu.Unlock()

	status := "active"
	switch {
	case value < 85:
		status = "active"
	case value < 95:
		status = "inactive"
	default:
		status = "api_error"
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": status,
	})
}
