package worker

import (
	"context"
	"log"
	"time"
)

// runLoop provides shared ticker-and-logging behavior for the in-process workers.
func runLoop(ctx context.Context, logger *log.Logger, interval time.Duration, name string, run func(context.Context) error) {
	if err := run(ctx); err != nil {
		logger.Printf("%s initial run failed: %v", name, err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := run(ctx); err != nil {
				logger.Printf("%s run failed: %v", name, err)
			}
		}
	}
}
