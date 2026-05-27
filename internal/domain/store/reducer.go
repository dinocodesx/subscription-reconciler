package store

import (
	"sort"
	"time"
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
