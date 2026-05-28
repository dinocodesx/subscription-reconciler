package carrier

import (
	"context"
	"log"
)

type PollRepository interface {
	ClaimCarrierUsers(ctx context.Context, limit int) ([]string, error)
	ApplyCarrierStatus(ctx context.Context, userID, status string) error
}

type Poller struct {
	repo      PollRepository
	client    Client
	batchSize int
	logger    *log.Logger
}

func NewPoller(repo PollRepository, client Client, batchSize int, logger *log.Logger) *Poller {
	return &Poller{
		repo:      repo,
		client:    client,
		batchSize: batchSize,
		logger:    logger,
	}
}

func (p *Poller) RunOnce(ctx context.Context) error {
	userIDs, err := p.repo.ClaimCarrierUsers(ctx, p.batchSize)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		status, err := p.client.PlanStatus(ctx, userID)
		if err != nil {
			p.logger.Printf("carrier poll failed for %s: %v", userID, err)
			continue
		}

		if err := p.repo.ApplyCarrierStatus(ctx, userID, status); err != nil {
			p.logger.Printf("carrier status apply failed for %s: %v", userID, err)
		}
	}

	return nil
}
