package unit

import (
	"testing"
	"time"

	storedom "github.com/dinocodesx/subscription-reconciler/internal/domain/store"
	"github.com/dinocodesx/subscription-reconciler/internal/testutil"
)

func TestReduceOutOfOrderEvents(t *testing.T) {
	now := testutil.MustParseTime(t, "2026-01-20T00:00:00Z")
	events := []storedom.Event{
		{
			EventID:    "evt_renewal",
			UserID:     "u_42",
			Type:       storedom.EventRenewal,
			EventTime:  testutil.MustParseTime(t, "2026-01-15T00:00:00Z"),
			ReceivedAt: testutil.MustParseTime(t, "2026-01-15T00:00:01Z"),
		},
		{
			EventID:    "evt_purchase",
			UserID:     "u_42",
			Type:       storedom.EventInitialPurchase,
			EventTime:  testutil.MustParseTime(t, "2025-12-16T00:00:00Z"),
			ReceivedAt: testutil.MustParseTime(t, "2026-01-16T00:00:00Z"),
		},
	}

	storedom.SortEvents(events)
	state := storedom.Reduce(events, now)

	if !state.Active {
		t.Fatalf("expected store entitlement to stay active")
	}

	wantExpiry := testutil.MustParseTime(t, "2026-02-14T00:00:00Z")
	if state.ExpiresAt == nil || !state.ExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expected expiry %s, got %#v", wantExpiry.Format(time.RFC3339), state.ExpiresAt)
	}
}
