package postgres

import "time"

type AuditEntry struct {
	ID             int64
	UserID         string
	EventID        *string
	Source         string
	PreviousActive bool
	PreviousSource string
	PreviousReason string
	NextActive     bool
	NextSource     string
	NextReason     string
	CreatedAt      time.Time
}
