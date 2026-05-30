package postgres

import (
	"time"

	"github.com/dinocodesx/subscription-reconciler/internal/util"
)

// cloneTime prevents callers from sharing mutable time pointers.
func cloneTime(value *time.Time) *time.Time {
	return util.CloneTime(value)
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC()
}

func unixMillisToTime(value int64) time.Time {
	return time.UnixMilli(value).UTC()
}
