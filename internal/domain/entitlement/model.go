package entitlement

import "time"

type Source string

const (
	SourceStore       Source = "STORE"
	SourceCarrier     Source = "CARRIER"
	SourceMarketplace Source = "MARKETPLACE"
	SourceNone        Source = "NONE"
)

// Entitlement is the canonical answer returned by the read endpoint.
type Entitlement struct {
	UserID        string
	Active        bool
	Source        Source
	ExpiresAt     *time.Time
	LastChangedAt time.Time
	Reason        string
}

// SourceEntitlement stores the per-channel view that feeds canonical resolution.
type SourceEntitlement struct {
	UserID        string
	Source        Source
	Active        bool
	ExpiresAt     *time.Time
	LastChangedAt time.Time
	Reason        string
	UpdatedAt     time.Time
	NextPollAt    *time.Time
}

// DefaultEntitlement returns the empty canonical state for users without active access.
func DefaultEntitlement(userID string) Entitlement {
	return Entitlement{
		UserID:        userID,
		Active:        false,
		Source:        SourceNone,
		LastChangedAt: time.Unix(0, 0).UTC(),
		Reason:        "NO_ACTIVE_ENTITLEMENT",
	}
}
