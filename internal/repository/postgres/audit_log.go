package postgres

import (
	"context"
	"fmt"
	"time"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	"github.com/jackc/pgx/v5"
)

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

func (s *Store) insertAuditLogTx(ctx context.Context, tx pgx.Tx, userID string, eventID *string, triggerSource entdom.Source, prev, next entdom.Entitlement) error {
	if prev.Active == next.Active &&
		prev.Source == next.Source &&
		prev.Reason == next.Reason {
		return nil
	}

	now := s.now().UTC()
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO audit_log
			(user_id, event_id, source, previous_active, previous_source, previous_reason,
			 next_active, next_source, next_reason, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userID,
		eventID,
		string(triggerSource),
		prev.Active,
		string(prev.Source),
		prev.Reason,
		next.Active,
		string(next.Source),
		next.Reason,
		now,
	); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}
