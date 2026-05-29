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

func (s *Store) GetTimeline(ctx context.Context, userID string) ([]AuditEntry, error) {
	rows, err := s.pool.Query(
		ctx,
		`SELECT id, user_id, event_id, source,
		        previous_active, previous_source, previous_reason,
		        next_active, next_source, next_reason, created_at
		 FROM audit_log
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit log: %w", err)
	}
	defer rows.Close()

	entries := make([]AuditEntry, 0)
	for rows.Next() {
		var entry AuditEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.EventID,
			&entry.Source,
			&entry.PreviousActive,
			&entry.PreviousSource,
			&entry.PreviousReason,
			&entry.NextActive,
			&entry.NextSource,
			&entry.NextReason,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log entry: %w", err)
		}

		entry.CreatedAt = entry.CreatedAt.UTC()
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit log: %w", err)
	}

	return entries, nil
}
