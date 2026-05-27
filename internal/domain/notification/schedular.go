package notification

import "time"

// DesiredSchedule computes the one-time notification time for expiring access.
func DesiredSchedule(now time.Time, active bool, expiresAt *time.Time) (*time.Time, bool) {
	if !active || expiresAt == nil {
		return nil, false
	}

	scheduledFor := expiresAt.UTC().Add(-24 * time.Hour)
	if scheduledFor.Before(now.UTC()) {
		scheduledFor = now.UTC()
	}

	return &scheduledFor, true
}
