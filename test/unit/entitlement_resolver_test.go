package unit

import (
	"testing"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

func TestResolveUsesSourcePrecedence(t *testing.T) {
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceMarketplace,
			Active:        true,
			LastChangedAt: util.MustParseTime(t, "2026-01-01T00:00:00Z"),
			Reason:        "seeded_market",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceCarrier,
			Active:        true,
			LastChangedAt: util.MustParseTime(t, "2026-01-02T00:00:00Z"),
			Reason:        "active",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceStore,
			Active:        true,
			LastChangedAt: util.MustParseTime(t, "2026-01-03T00:00:00Z"),
			Reason:        "RENEWAL",
		},
	}

	result := entdom.Resolve("u_42", states)

	if result.Source != entdom.SourceStore {
		t.Fatalf("expected STORE to win precedence, got %s", result.Source)
	}
}

func TestResolveReturnsNoneWhenNoSourcesAreActive(t *testing.T) {
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceMarketplace,
			Active:        false,
			LastChangedAt: util.MustParseTime(t, "2026-01-04T00:00:00Z"),
			Reason:        "MARKETPLACE_REVOKE",
		},
	}

	result := entdom.Resolve("u_42", states)

	if result.Source != entdom.SourceNone {
		t.Fatalf("expected NONE, got %s", result.Source)
	}
	if result.Reason != "NO_ACTIVE_ENTITLEMENT" {
		t.Fatalf("expected default reason, got %s", result.Reason)
	}
}

func TestResolveCarrierWinsWhenStoreInactive(t *testing.T) {
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceStore,
			Active:        false,
			LastChangedAt: util.MustParseTime(t, "2026-01-03T00:00:00Z"),
			Reason:        "EXPIRATION",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceCarrier,
			Active:        true,
			LastChangedAt: util.MustParseTime(t, "2026-01-02T00:00:00Z"),
			Reason:        "active",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceMarketplace,
			Active:        true,
			LastChangedAt: util.MustParseTime(t, "2026-01-01T00:00:00Z"),
			Reason:        "seeded_market",
		},
	}

	result := entdom.Resolve("u_42", states)

	if result.Source != entdom.SourceCarrier {
		t.Fatalf("expected CARRIER to win when STORE is inactive, got %s", result.Source)
	}
	if !result.Active {
		t.Fatalf("expected active entitlement")
	}
}

func TestResolveLastChangedAtFromMostRecentInactiveSource(t *testing.T) {
	later := util.MustParseTime(t, "2026-01-10T00:00:00Z")
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceStore,
			Active:        false,
			LastChangedAt: util.MustParseTime(t, "2026-01-05T00:00:00Z"),
			Reason:        "EXPIRATION",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceCarrier,
			Active:        false,
			LastChangedAt: later,
			Reason:        "inactive",
		},
	}

	result := entdom.Resolve("u_42", states)

	if result.Source != entdom.SourceNone {
		t.Fatalf("expected NONE, got %s", result.Source)
	}
	if !result.LastChangedAt.Equal(later) {
		t.Fatalf("expected LastChangedAt %s, got %s", later, result.LastChangedAt)
	}
}

func TestResolveEmptyStatesReturnsDefault(t *testing.T) {
	result := entdom.Resolve("u_42", nil)

	if result.Source != entdom.SourceNone {
		t.Fatalf("expected NONE, got %s", result.Source)
	}
	if result.Active {
		t.Fatalf("expected inactive")
	}
	if result.Reason != "NO_ACTIVE_ENTITLEMENT" {
		t.Fatalf("expected default reason, got %s", result.Reason)
	}
}
