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

// NewSender constructs the notification worker dependency.
func NewSender(repo SendRepository, batchSize int, logger *log.Logger) *Sender {
	return &Sender{
		repo:      repo,
		batchSize: batchSize,
		logger:    logger,
	}
}

// RunOnce claims and marks one batch of due notifications.
func (s *Sender) RunOnce(ctx context.Context) error {
	rows, err := s.repo.ClaimDueNotifications(ctx, s.batchSize)
	if err != nil {
		return err
	}

	if len(rows) > 0 {
		s.logger.Printf("marked %d notification(s) as sent", len(rows))
	}

	return nil
}
