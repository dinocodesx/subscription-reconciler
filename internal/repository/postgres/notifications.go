package postgres

import (
	"context"
	"time"

	notifdom "github.com/dinocodesx/subscription-reconciler/internal/domain/notification"
)

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC()
}

func unixMillisToTime(value int64) time.Time {
	return time.UnixMilli(value).UTC()
}

// ClaimDueNotifications atomically marks due rows as sent and returns the claimed batch.
func (s *Store) ClaimDueNotifications(ctx context.Context, limit int) ([]notifdom.Notification, error) {

}
