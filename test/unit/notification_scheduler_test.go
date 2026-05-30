package unit

import (
	"testing"

	notifdom "github.com/dinocodesx/subscription-reconciler/internal/domain/notification"
	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

func TestDesiredScheduleUsesExpiryMinus24Hours(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-01T10:00:00Z")
	expiresAt := util.MustParseTime(t, "2026-01-03T10:00:00Z")

	scheduledFor, ok := notifdom.DesiredSchedule(now, true, &expiresAt)
	if !ok {
		t.Fatalf("expected notification schedule")
	}

	want := util.MustParseTime(t, "2026-01-02T10:00:00Z")
	if scheduledFor == nil || !scheduledFor.Equal(want) {
		t.Fatalf("expected schedule %s, got %#v", want, scheduledFor)
	}
}

func TestDesiredScheduleRunsImmediatelyInside24Hours(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-01T10:00:00Z")
	expiresAt := util.MustParseTime(t, "2026-01-01T20:00:00Z")

	scheduledFor, ok := notifdom.DesiredSchedule(now, true, &expiresAt)
	if !ok {
		t.Fatalf("expected notification schedule")
	}

	if scheduledFor == nil || !scheduledFor.Equal(now) {
		t.Fatalf("expected immediate schedule at %s, got %#v", now, scheduledFor)
	}
}

func TestDesiredScheduleNoNotificationForInactiveEntitlement(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-01T10:00:00Z")
	expiresAt := util.MustParseTime(t, "2026-01-03T10:00:00Z")

	_, ok := notifdom.DesiredSchedule(now, false, &expiresAt)
	if ok {
		t.Fatalf("expected no notification for inactive entitlement")
	}
}

func TestDesiredScheduleNoNotificationForNilExpiry(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-01T10:00:00Z")

	_, ok := notifdom.DesiredSchedule(now, true, nil)
	if ok {
		t.Fatalf("expected no notification when expiresAt is nil")
	}
}

func TestDesiredScheduleNoNotificationForInactiveAndNilExpiry(t *testing.T) {
	now := util.MustParseTime(t, "2026-01-01T10:00:00Z")

	_, ok := notifdom.DesiredSchedule(now, false, nil)
	if ok {
		t.Fatalf("expected no notification for inactive entitlement with nil expiry")
	}
}
