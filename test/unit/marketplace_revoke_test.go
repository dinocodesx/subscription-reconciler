package unit

import (
	"testing"
	"time"

	entdom "github.com/dinocodesx/subscription-reconciler/internal/domain/entitlement"
	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

func TestMarketplaceRevokeOnlyAffectsMarketplaceSource(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-10T00:00:00Z")
	expiresAt := util.MustParseTime(t, "2026-02-10T00:00:00Z")

	// Simulate post-revoke state: store and carrier are active, marketplace is revoked.
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceStore,
			Active:        true,
			ExpiresAt:     &expiresAt,
			LastChangedAt: now,
			Reason:        "RENEWAL",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceCarrier,
			Active:        true,
			LastChangedAt: now,
			Reason:        "active",
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceMarketplace,
			Active:        false,
			LastChangedAt: now,
			Reason:        entdom.MarketplaceRevokeReason,
		},
	}

	result := entdom.Resolve("u_42", states)

	// STORE should win because it has highest precedence and is still active.
	if result.Source != entdom.SourceStore {
		t.Fatalf("expected STORE to win, got %s", result.Source)
	}
	if !result.Active {
		t.Fatalf("expected entitlement to stay active despite marketplace revoke")
	}
}

func TestMarketplaceRevokeDeactivatesMarketplaceOnlyUser(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-10T00:00:00Z")

	// User only has marketplace source, and it was revoked.
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_91",
			Source:        entdom.SourceMarketplace,
			Active:        false,
			LastChangedAt: now,
			Reason:        entdom.MarketplaceRevokeReason,
		},
	}

	result := entdom.Resolve("u_91", states)

	if result.Active {
		t.Fatalf("expected inactive after marketplace revoke for marketplace-only user")
	}
	if result.Source != entdom.SourceNone {
		t.Fatalf("expected source NONE, got %s", result.Source)
	}
}

func TestMarketplaceRevokeReasonConstant(t *testing.T) {
	if entdom.MarketplaceRevokeReason != "MARKETPLACE_REVOKE" {
		t.Fatalf("expected revoke reason 'MARKETPLACE_REVOKE', got %q", entdom.MarketplaceRevokeReason)
	}
}

func TestMarketplaceRevokePreservesLastChangedAt(t *testing.T) {
	revokeTime := util.MustParseTime(t, "2026-01-15T12:00:00Z")

	// After revoke, the source entitlement's LastChangedAt should be the revoke time.
	se := entdom.SourceEntitlement{
		UserID:        "u_42",
		Source:        entdom.SourceMarketplace,
		Active:        false,
		LastChangedAt: revokeTime,
		Reason:        entdom.MarketplaceRevokeReason,
		UpdatedAt:     revokeTime,
	}

	result := entdom.Resolve("u_42", []entdom.SourceEntitlement{se})
	if !result.LastChangedAt.Equal(revokeTime) {
		t.Fatalf("expected canonical LastChangedAt %s, got %s", revokeTime, result.LastChangedAt)
	}
}

func TestMarketplaceRevokeCarrierFallback(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-10T00:00:00Z")

	// Marketplace revoked, but carrier is active. Carrier should win.
	states := []entdom.SourceEntitlement{
		{
			UserID:        "u_42",
			Source:        entdom.SourceMarketplace,
			Active:        false,
			LastChangedAt: now.Add(1 * time.Hour),
			Reason:        entdom.MarketplaceRevokeReason,
		},
		{
			UserID:        "u_42",
			Source:        entdom.SourceCarrier,
			Active:        true,
			LastChangedAt: now,
			Reason:        "active",
		},
	}

	result := entdom.Resolve("u_42", states)

	if result.Source != entdom.SourceCarrier {
		t.Fatalf("expected CARRIER fallback, got %s", result.Source)
	}
	if !result.Active {
		t.Fatalf("expected active entitlement from carrier fallback")
	}
}
