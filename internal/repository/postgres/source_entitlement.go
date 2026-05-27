package postgres

import (
	"context"
	"fmt"
	"time"

	entdom "github.com/subscription-reconciler/internal/domain/entitlement"

	"github.com/jackc/pgx/v5"
)

func (s *Store) recomputeCanonicalEntitlementTx(ctx context.Context, tx pgx.Tx, userID string, eventID *string, triggerSource entdom.Source) (entdom.Entitlement, error) {
	_, err := s.getCanonicalEntitlementTx(ctx, tx, userID)
	if err != nil {
		return entdom.Entitlement{}, err
	}

	states, err := s.listSourceEntitlementsTx(ctx, tx, userID)
	if err != nil {
		return entdom.Entitlement{}, err
	}

	entitlement := entdom.Resolve(userID, states)
	if err := s.upsertCanonicalEntitlementTx(ctx, tx, entitlement); err != nil {
		return entdom.Entitlement{}, err
	}

	// if err := s.insertAuditLogTx(ctx, tx, userID, eventID, triggerSource, prev, entitlement); err != nil {
	// 	return entdom.Entitlement{}, err
	// }

	if err := s.syncExpiringSoonNotificationTx(ctx, tx, entitlement); err != nil {
		return entdom.Entitlement{}, err
	}

	return entitlement, nil
}

func (s *Store) listSourceEntitlementsTx(ctx context.Context, tx pgx.Tx, userID string) ([]entdom.SourceEntitlement, error) {
	rows, err := tx.Query(
		ctx,
		`SELECT user_id, source, active, expires_at, last_changed_at, reason, updated_at, next_poll_at
		 FROM source_entitlements
		 WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query source entitlements: %w", err)
	}
	defer rows.Close()

	states := make([]entdom.SourceEntitlement, 0)
	for rows.Next() {
		var state entdom.SourceEntitlement
		var expiresAt, nextPollAt *time.Time
		var source string
		if err := rows.Scan(
			&state.UserID,
			&source,
			&state.Active,
			&expiresAt,
			&state.LastChangedAt,
			&state.Reason,
			&state.UpdatedAt,
			&nextPollAt,
		); err != nil {
			return nil, fmt.Errorf("scan source entitlement: %w", err)
		}

		state.Source = entdom.Source(source)
		state.ExpiresAt = cloneTime(expiresAt)
		state.NextPollAt = cloneTime(nextPollAt)
		state.LastChangedAt = state.LastChangedAt.UTC()
		state.UpdatedAt = state.UpdatedAt.UTC()
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate source entitlements: %w", err)
	}

	return states, nil
}
