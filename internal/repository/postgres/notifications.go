package postgres

import (
	"context"
	"fmt"
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
	now := s.now().UTC()

	rows, err := s.pool.Query(
		ctx,
		`WITH claimed AS (
			SELECT id
			FROM notifications
			WHERE sent_at IS NULL
			  AND scheduled_for <= $1
			ORDER BY scheduled_for ASC, id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		)
		UPDATE notifications n
		SET sent_at = $1
		FROM claimed
		WHERE n.id = claimed.id
		RETURNING n.id, n.user_id, n.type, n.scheduled_for, n.sent_at, n.created_at`,
		now,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("claim due notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]notifdom.Notification, 0)
	for rows.Next() {
		var entry notifdom.Notification
		var notificationType string
		var sentAt *time.Time
		if err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&notificationType,
			&entry.ScheduledFor,
			&sentAt,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		entry.Type = notifdom.Kind(notificationType)
		entry.ScheduledFor = entry.ScheduledFor.UTC()
		entry.CreatedAt = entry.CreatedAt.UTC()
		entry.SentAt = cloneTime(sentAt)
		notifications = append(notifications, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}
