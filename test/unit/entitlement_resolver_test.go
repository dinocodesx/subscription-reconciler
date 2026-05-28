package unit

import (
	"testing"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	"github.com/dinocodesx/subscription-reconciler/internal/testutil"
)

func TestResolveUsesSourcePrecedence(t *testing.T) {
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceMarketplace,
			Active:        true,
			LastChangedAt: testutil.MustParseTime(t, "2026-01-01T00:00:00Z"),
			Reason:        "seeded_market",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceCarrier,
			Active:        true,
			LastChangedAt: testutil.MustParseTime(t, "2026-01-02T00:00:00Z"),
			Reason:        "active",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceStore,
			Active:        true,
			LastChangedAt: testutil.MustParseTime(t, "2026-01-03T00:00:00Z"),
			Reason:        "RENEWAL",
		},
	}

	result := entdom.Resolve("u_42", states)

	if result.Source != entdom.SourceStore {
		t.Fatalf("expected STORE to win precedence, got %s", result.Source)
	}
}
