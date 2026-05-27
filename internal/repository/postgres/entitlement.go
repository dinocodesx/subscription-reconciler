package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	notifdom "github.com/dinocodesx/subscription-reconciler/internal/domain/notification"

	"github.com/jackc/pgx/v5"
)

func scanEntitlement(row pgx.Row) (entdom.Entitlement, error) {
	var entitlement entdom.Entitlement
	var source string
	var expiresAt *time.Time

	if err := row.Scan(
		&entitlement.UserID,
		&entitlement.Active,
		&source,
		&expiresAt,
		&entitlement.LastChangedAt,
		&entitlement.Reason,
	); err != nil {
		return entdom.Entitlement{}, err
	}

	entitlement.Source = entdom.Source(source)
	entitlement.ExpiresAt = cloneTime(expiresAt)
	entitlement.LastChangedAt = entitlement.LastChangedAt.UTC()

	return entitlement, nil
}

func (s *Store) getCanonicalEntitlementTx(ctx context.Context, tx pgx.Tx, userID string) (entdom.Entitlement, error) {
	row := tx.QueryRow(
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
		return entdom.Entitlement{}, fmt.Errorf("get canonical entitlement for audit: %w", err)
	}

	return entitlement, nil
}

func (s *Store) upsertCanonicalEntitlementTx(ctx context.Context, tx pgx.Tx, entitlement entdom.Entitlement) error {
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO entitlements (user_id, active, source, expires_at, last_changed_at, reason, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id) DO UPDATE
		 SET active = EXCLUDED.active,
		     source = EXCLUDED.source,
		     expires_at = EXCLUDED.expires_at,
		     last_changed_at = EXCLUDED.last_changed_at,
		     reason = EXCLUDED.reason,
		     updated_at = EXCLUDED.updated_at`,
		entitlement.UserID,
		entitlement.Active,
		entitlement.Source,
		nullableTime(entitlement.ExpiresAt),
		entitlement.LastChangedAt.UTC(),
		entitlement.Reason,
		s.now().UTC(),
	); err != nil {
		return fmt.Errorf("upsert canonical entitlement: %w", err)
	}

	return nil
}

// syncExpiringSoonNotificationTx keeps one pending expiry notification aligned with the latest entitlement state.
func (s *Store) syncExpiringSoonNotificationTx(ctx context.Context, tx pgx.Tx, entitlement entdom.Entitlement) error {
	now := s.now().UTC()
	desired, shouldSchedule := notifdom.DesiredSchedule(now, entitlement.Active, entitlement.ExpiresAt)

	// If expiry moved or access disappeared, drop the stale unsent row before inserting the fresh one.
	if _, err := tx.Exec(
		ctx,
		`DELETE FROM notifications
		 WHERE user_id = $1
		   AND type = $2
		   AND sent_at IS NULL
		   AND ($3::timestamptz IS NULL OR scheduled_for <> $3)`,
		entitlement.UserID,
		notifdom.KindPremiumExpiresSoon,
		nullableTime(desired),
	); err != nil {
		return fmt.Errorf("delete stale notifications: %w", err)
	}

	if !shouldSchedule {
		return nil
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO notifications (user_id, type, scheduled_for, sent_at, created_at)
		 VALUES ($1, $2, $3, NULL, $4)
		 ON CONFLICT DO NOTHING`,
		entitlement.UserID,
		notifdom.KindPremiumExpiresSoon,
		desired.UTC(),
		now,
	); err != nil {
		return fmt.Errorf("insert expiring soon notification: %w", err)
	}

	return nil
}
