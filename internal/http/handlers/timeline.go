package handlers

import (
	"log"
	"net/http"

	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
)

type TimelineHandler struct {
	store *postgres.Store
}

type timelineEntry struct {
	EventID        *string `json:"eventId"`
	Source         string  `json:"source"`
	PreviousActive bool    `json:"previousActive"`
	PreviousSource string  `json:"previousSource"`
	PreviousReason string  `json:"previousReason"`
	NextActive     bool    `json:"nextActive"`
	NextSource     string  `json:"nextSource"`
	NextReason     string  `json:"nextReason"`
	Timestamp      string  `json:"timestamp"`
}

// NewTimelineHandler builds the stretch-goal timeline endpoint.
func NewTimelineHandler(store *postgres.Store) *TimelineHandler {
	return &TimelineHandler{store: store}
}

// Handle returns the reconstructed entitlement history from the audit log.
func (h *TimelineHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user id is required")
		return
	}

	entries, err := h.store.GetTimeline(r.Context(), userID)
	if err != nil {
		log.Printf("get timeline for user %s: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := make([]timelineEntry, 0, len(entries))
	for _, entry := range entries {
		response = append(response, timelineEntry{
			EventID:        entry.EventID,
			Source:         entry.Source,
			PreviousActive: entry.PreviousActive,
			PreviousSource: entry.PreviousSource,
			PreviousReason: entry.PreviousReason,
			NextActive:     entry.NextActive,
			NextSource:     entry.NextSource,
			NextReason:     entry.NextReason,
			Timestamp:      entry.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"userId":   userID,
		"timeline": response,
	})
}
