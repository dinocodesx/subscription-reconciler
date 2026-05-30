package unit

import (
	"testing"
	"time"

	storedom "github.com/dinocodesx/subscription-reconciler/internal/domain/store"
	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

func TestReduceOutOfOrderEvents(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-20T00:00:00Z")
	events := []storedom.Event{
		{
			EventID:    "evt_renewal",
			UserID:     "u_42",
			Type:       storedom.EventRenewal,
			EventTime:  util.MustParseTime(t, "2026-01-15T00:00:00Z"),
			ReceivedAt: util.MustParseTime(t, "2026-01-15T00:00:01Z"),
		},
		{
			EventID:    "evt_purchase",
			UserID:     "u_42",
			Type:       storedom.EventInitialPurchase,
			EventTime:  util.MustParseTime(t, "2025-12-16T00:00:00Z"),
			ReceivedAt: util.MustParseTime(t, "2026-01-16T00:00:00Z"),
		},
	}

	storedom.SortEvents(events)
	state := storedom.Reduce(events, now)

	if !state.Active {
		t.Fatalf("expected store entitlement to stay active")
	}

	wantExpiry := util.MustParseTime(t, "2026-02-14T00:00:00Z")
	if state.ExpiresAt == nil || !state.ExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expected expiry %s, got %#v", wantExpiry.Format(time.RFC3339), state.ExpiresAt)
	}
}

func TestReduceLateExpirationThenRenewal(t *testing.T) {
	now := util.MustParseTime(t, "2026-03-15T00:00:00Z")
	events := []storedom.Event{
		{
			EventID:    "evt_renewal",
			UserID:     "u_99",
			Type:       storedom.EventRenewal,
			EventTime:  util.MustParseTime(t, "2026-03-01T00:00:00Z"),
			ReceivedAt: util.MustParseTime(t, "2026-03-16T00:00:00Z"),
		},
		{
			EventID:    "evt_expired",
			UserID:     "u_99",
			Type:       storedom.EventExpiration,
			EventTime:  util.MustParseTime(t, "2026-02-15T00:00:00Z"),
			ReceivedAt: util.MustParseTime(t, "2026-02-15T00:00:01Z"),
		},
	}

	storedom.SortEvents(events)
	state := storedom.Reduce(events, now)

	if !state.Active {
		t.Fatalf("expected renewal after expiration to restore active entitlement")
	}

	wantExpiry := util.MustParseTime(t, "2026-03-31T00:00:00Z")
	if state.ExpiresAt == nil || !state.ExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expected expiry %s, got %#v", wantExpiry.Format(time.RFC3339), state.ExpiresAt)
	}
}
