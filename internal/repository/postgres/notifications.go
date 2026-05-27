package postgres

import "time"

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC()
}

func unixMillisToTime(value int64) time.Time {
	return time.UnixMilli(value).UTC()
}
