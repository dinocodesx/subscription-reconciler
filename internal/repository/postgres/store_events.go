package postgres

import (
	"context"
	"errors"
	"fmt"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	storedom "github.com/dinocodesx/subscription-reconciler/internal/domain/store"

	"github.com/jackc/pgx/v5"
)

func (s *Store) ApplyStoreEvent(ctx context.Context, event storedom.Event) (bool, entdom.Entitlement, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, entdom.Entitlement{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	inserted, err := s.insertStoreEventTx(ctx, tx, event)
	if err != nil {
		return false, entdom.Entitlement{}, err
	}

	if inserted {
		events, err := s.loadStoreEventsTx(ctx, tx, event.UserID)
		if err != nil {
			return false, entdom.Entitlement{}, err
		}

		// Replay from source-of-truth events on every accepted insert so late arrivals can repair state.
		storedom.SortEvents(events)
		state := storedom.Reduce(events, s.now())
		if err := s.upsertSourceEntitlementTx(ctx, tx, entdom.SourceEntitlement{
			UserID:        event.UserID,
			Source:        entdom.SourceStore,
			Active:        state.Active,
			ExpiresAt:     state.ExpiresAt,
			LastChangedAt: state.LastChangedAt,
			Reason:        state.Reason,
			UpdatedAt:     s.now().UTC(),
		}); err != nil {
			return false, entdom.Entitlement{}, err
		}
	}

	eventIDStr := event.EventID
	entitlement, err := s.recomputeCanonicalEntitlementTx(ctx, tx, event.UserID, &eventIDStr, entdom.SourceStore)
	if err != nil {
		return false, entdom.Entitlement{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, entdom.Entitlement{}, fmt.Errorf("commit transaction: %w", err)
	}

	return inserted, entitlement, nil
}

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

func (s *Store) GetEntitlement(ctx context.Context, userID string) (entdom.Entitlement, error) {
	row := s.pool.QueryRow(
		ctx,
		`SELECT user_id, active, source, expires_at, last_changed_at, reason
		 FROM entitlements
		 WHERE user_id = $1`,
		userID,
	)

	entitlement, err := scanEntitlement(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return entdom.DefaultEntitlement(userID), nil
	}
	if err != nil {
		return entdom.Entitlement{}, fmt.Errorf("get entitlement: %w", err)
	}

	return entitlement, nil
}
