package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
	entdom "github.com/subscription-reconciler/internal/domain/entitlement"
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
