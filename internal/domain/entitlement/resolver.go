package entitlement

import (
	"time"

	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

var sourcePrecedence = []Source{
	SourceStore,
	SourceCarrier,
	SourceMarketplace,
}

// Resolve picks the winning active source using the assignment's deterministic precedence.
func Resolve(userID string, states []SourceEntitlement) Entitlement {
	for _, source := range sourcePrecedence {
		for _, state := range states {
			if state.Source == source && state.Active {
				return Entitlement{
					UserID:        userID,
					Active:        true,
					Source:        state.Source,
					ExpiresAt:     cloneTime(state.ExpiresAt),
					LastChangedAt: state.LastChangedAt.UTC(),
					Reason:        state.Reason,
				}
			}
		}
	}

	result := DefaultEntitlement(userID)
	for _, state := range states {
		if state.LastChangedAt.After(result.LastChangedAt) {
			result.LastChangedAt = state.LastChangedAt.UTC()
		}
	}

	return result
}

func cloneTime(value *time.Time) *time.Time {
	return util.CloneTime(value)
}
