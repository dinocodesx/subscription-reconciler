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
