package store

import "time"

type EventType string

const (
	EventInitialPurchase EventType = "INITIAL_PURCHASE"
	EventRenewal         EventType = "RENEWAL"
	EventCancellation    EventType = "CANCELLATION"
	EventBillingIssue    EventType = "BILLING_ISSUE"
	EventExpiration      EventType = "EXPIRATION"
	EventUnCancellation  EventType = "UN_CANCELLATION"
)

const SubscriptionDuration = 30 * 24 * time.Hour

type Event struct {
	EventID    string
	UserID     string
	Type       EventType
	EventTime  time.Time
	ProductID  string
	ReceivedAt time.Time
}

func (t EventType) Valid() bool {
	switch t {
	case EventInitialPurchase, EventRenewal, EventCancellation, EventBillingIssue, EventExpiration, EventUnCancellation:
		return true
	default:
		return false
	}
}
