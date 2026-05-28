package worker

import (
	"context"
	"log"
	"time"

	notifdom "github.com/dinocodesx/subscription-reconciler/internal/domain/notification"
)

// RunNotificationSender runs the notification worker loop on a fixed interval.
func RunNotificationSender(ctx context.Context, logger *log.Logger, interval time.Duration, sender *notifdom.Sender) {
	runLoop(ctx, logger, interval, "notification sender", sender.RunOnce)
}
