package handlers

import (
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
