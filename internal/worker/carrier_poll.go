package worker

import (
	"context"
	"log"
	"time"

	"github.com/dinocodesx/subscription-reconciler/internal/domain/carrier"
)

// RunCarrierPoller runs the carrier reconciliation loop on a fixed interval.
func RunCarrierPoller(ctx context.Context, logger *log.Logger, interval time.Duration, poller *carrier.Poller) {
	runLoop(ctx, logger, interval, "carrier poller", poller.RunOnce)
}
