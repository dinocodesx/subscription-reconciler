package store

import "time"

type State struct {
	Active        bool
	ExpiresAt     *time.Time
	LastChangedAt time.Time
	Reason        string
}
