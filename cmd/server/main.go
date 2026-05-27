package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dinocodesx/subscription-reconciler/internal/app"
	"github.com/dinocodesx/subscription-reconciler/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
