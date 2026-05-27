package postgres

import (
	"context"

	entdom "github.com/subscription-reconciler/internal/domain/entitlement"

	"github.com/jackc/pgx/v5"
)

func (s *Store) recomputeCanonicalEntitlementTx(ctx context.Context, tx pgx.Tx, userID string, eventID *string, triggerSource entdom.Source) (entdom.Entitlement, error) {
	prev, err := s.getCanonicalEntitlementTx(ctx, tx, userID)
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

	if err := s.insertAuditLogTx(ctx, tx, userID, eventID, triggerSource, prev, entitlement); err != nil {
		return entdom.Entitlement{}, err
	}

	if err := s.syncExpiringSoonNotificationTx(ctx, tx, entitlement); err != nil {
		return entdom.Entitlement{}, err
	}

	return entitlement, nil
}
