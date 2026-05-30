package store

import (
	"sort"
	"time"

	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

type State struct {
	Active        bool
	ExpiresAt     *time.Time
	LastChangedAt time.Time
	Reason        string
}

// SortEvents gives the reducer a stable chronological order even when deliveries are delayed.
func SortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].EventTime.Equal(events[j].EventTime) {
			if events[i].ReceivedAt.Equal(events[j].ReceivedAt) {
				return events[i].EventID < events[j].EventID
			}
			return events[i].ReceivedAt.Before(events[j].ReceivedAt)
		}
		return events[i].EventTime.Before(events[j].EventTime)
	})
}

// Reduce rebuilds store-derived entitlement state from the full event history for one user.
func Reduce(events []Event, now time.Time) State {
	now = now.UTC()

	var expiry *time.Time
	lastChanged := time.Unix(0, 0).UTC()
	reason := "NO_ACTIVE_ENTITLEMENT"

	for _, event := range events {
		eventTime := event.EventTime.UTC()
		switch event.Type {
		case EventInitialPurchase, EventRenewal:
			// Renewals extend from the later of "event time" or the current paid-through date.
			base := eventTime
			if expiry != nil && expiry.After(base) {
				base = expiry.UTC()
			}
			next := base.Add(SubscriptionDuration).UTC()
			expiry = &next
			lastChanged = eventTime
			reason = string(event.Type)
		case EventExpiration:
			expiredAt := eventTime
			expiry = &expiredAt
			lastChanged = eventTime
			reason = string(event.Type)
		case EventCancellation, EventBillingIssue, EventUnCancellation:
			lastChanged = eventTime
			reason = string(event.Type)
		}
	}

	active := expiry != nil && expiry.After(now)
	return State{
		Active:        active,
		ExpiresAt:     cloneTime(expiry),
		LastChangedAt: lastChanged,
		Reason:        reason,
	}
}

func cloneTime(value *time.Time) *time.Time {
	return util.CloneTime(value)
}
