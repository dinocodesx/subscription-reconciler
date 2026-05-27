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
