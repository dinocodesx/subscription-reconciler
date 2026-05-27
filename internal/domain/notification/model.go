package notification

import "time"

type Kind string

const KindPremiumExpiresSoon Kind = "PREMIUM_EXPIRES_SOON"

// Notification is the persisted work item used by the assignment's "sending" worker.
type Notification struct {
	ID           int64
	UserID       string
	Type         Kind
	ScheduledFor time.Time
	SentAt       *time.Time
	CreatedAt    time.Time
}
