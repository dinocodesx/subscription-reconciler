package postgres

import (
	"context"
	"fmt"

	storedom "github.com/subscription-reconciler/internal/domain/store"

	"github.com/jackc/pgx/v5"
)

func (s *Store) insertStoreEventTx(ctx context.Context, tx pgx.Tx, event storedom.Event) (bool, error) {
	commandTag, err := tx.Exec(
		ctx,
		`INSERT INTO store_events (event_id, user_id, type, event_time_ms, product_id, received_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (event_id) DO NOTHING`,
		event.EventID,
		event.UserID,
		string(event.Type),
		event.EventTime.UTC().UnixMilli(),
		event.ProductID,
		event.ReceivedAt.UTC(),
	)
	if err != nil {
		return false, fmt.Errorf("insert store event: %w", err)
	}

	return commandTag.RowsAffected() == 1, nil
}

func (s *Store) loadStoreEventsTx(ctx context.Context, tx pgx.Tx, userID string) ([]storedom.Event, error) {
	rows, err := tx.Query(
		ctx,
		`SELECT event_id, user_id, type, event_time_ms, product_id, received_at
		 FROM store_events
		 WHERE user_id = $1
		 ORDER BY event_time_ms ASC, received_at ASC, event_id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query store events: %w", err)
	}
	defer rows.Close()

	events := make([]storedom.Event, 0)
	for rows.Next() {
		var event storedom.Event
		var eventTimeMs int64
		var eventType string
		if err := rows.Scan(&event.EventID, &event.UserID, &eventType, &eventTimeMs, &event.ProductID, &event.ReceivedAt); err != nil {
			return nil, fmt.Errorf("scan store event: %w", err)
		}

		event.Type = storedom.EventType(eventType)
		event.EventTime = unixMillisToTime(eventTimeMs)
		event.ReceivedAt = event.ReceivedAt.UTC()
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate store events: %w", err)
	}

	return events, nil
}
