package unit

import (
	"testing"

	storedom "github.com/dinocodesx/subscription-reconciler/internal/domain/store"
)

func TestStoreEventTypeValidation(t *testing.T) {
	valid := []storedom.EventType{
		storedom.EventInitialPurchase,
		storedom.EventRenewal,
		storedom.EventCancellation,
		storedom.EventBillingIssue,
		storedom.EventExpiration,
		storedom.EventUnCancellation,
	}

	for _, et := range valid {
		if !et.Valid() {
			t.Errorf("expected %s to be valid", et)
		}
	}
}

func TestStoreEventTypeRejectsInvalid(t *testing.T) {
	invalid := []storedom.EventType{
		"REFUND",
		"UPGRADE",
		"",
		"initial_purchase",
	}

	for _, et := range invalid {
		if et.Valid() {
			t.Errorf("expected %q to be invalid", et)
		}
	}
}

func TestStoreSortEventsChronological(t *testing.T) {
	events := []storedom.Event{
		{EventID: "evt_3", EventTime: mustParseTimeHelper("2026-01-03T00:00:00Z")},
		{EventID: "evt_1", EventTime: mustParseTimeHelper("2026-01-01T00:00:00Z")},
		{EventID: "evt_2", EventTime: mustParseTimeHelper("2026-01-02T00:00:00Z")},
	}

	storedom.SortEvents(events)

	if events[0].EventID != "evt_1" || events[1].EventID != "evt_2" || events[2].EventID != "evt_3" {
		t.Fatalf("expected chronological order, got %s %s %s", events[0].EventID, events[1].EventID, events[2].EventID)
	}
}

func TestStoreSortEventsTieBreakByReceivedAt(t *testing.T) {
	sameTime := mustParseTimeHelper("2026-01-01T00:00:00Z")
	events := []storedom.Event{
		{EventID: "evt_b", EventTime: sameTime, ReceivedAt: mustParseTimeHelper("2026-01-02T00:00:00Z")},
		{EventID: "evt_a", EventTime: sameTime, ReceivedAt: mustParseTimeHelper("2026-01-01T00:00:00Z")},
	}

	storedom.SortEvents(events)

	if events[0].EventID != "evt_a" {
		t.Fatalf("expected evt_a first (earlier ReceivedAt), got %s", events[0].EventID)
	}
}

func TestStoreSortEventsTieBreakByEventID(t *testing.T) {
	sameTime := mustParseTimeHelper("2026-01-01T00:00:00Z")
	events := []storedom.Event{
		{EventID: "evt_z", EventTime: sameTime, ReceivedAt: sameTime},
		{EventID: "evt_a", EventTime: sameTime, ReceivedAt: sameTime},
	}

	storedom.SortEvents(events)

	if events[0].EventID != "evt_a" {
		t.Fatalf("expected evt_a first (lexicographic tie-break), got %s", events[0].EventID)
	}
}

func TestStoreSubscriptionDuration(t *testing.T) {
	expected := 30 * 24 * 60 * 60 // 30 days in seconds
	actual := int(storedom.SubscriptionDuration.Seconds())

	if actual != expected {
		t.Fatalf("expected subscription duration %d seconds, got %d", expected, actual)
	}
}
