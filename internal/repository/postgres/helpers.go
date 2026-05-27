package postgres

import "time"

// cloneTime prevents callers from sharing mutable time pointers.
func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := value.UTC()
	return &cloned
}
