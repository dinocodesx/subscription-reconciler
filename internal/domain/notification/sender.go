package notification

import (
	"context"
	"log"
)

type SendRepository interface {
	ClaimDueNotifications(ctx context.Context, limit int) ([]Notification, error)
}

// Sender marks due notifications as sent after claiming them safely from Postgres.
type Sender struct {
	repo      SendRepository
	batchSize int
	logger    *log.Logger
}
