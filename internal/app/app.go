package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/dinocodesx/subscription-reconciler/internal/config"
	"github.com/dinocodesx/subscription-reconciler/internal/domain/carrier"
	"github.com/dinocodesx/subscription-reconciler/internal/domain/notification"
	httpapi "github.com/dinocodesx/subscription-reconciler/internal/http"
	"github.com/dinocodesx/subscription-reconciler/internal/repository/postgres"
	"github.com/dinocodesx/subscription-reconciler/internal/worker"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run boots the HTTP server, background workers, schema migrations, and demo seed data.
func Run(ctx context.Context, cfg config.Config) error {
	logger := log.Default()

	pool, err := connectPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	store := postgres.NewStore(pool, time.Now, cfg.CarrierPollInterval)
	if err := store.Migrate(ctx, filepath.Join(".", "migrations")); err != nil {
		return err
	}
	if err := store.SeedDemoData(ctx); err != nil {
		return err
	}

	router := httpapi.NewRouter(store)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	// Both workers run in-process for the assignment, but their database claim logic is
	// written to tolerate multiple live service instances.
	go worker.RunCarrierPoller(
		ctx,
		logger,
		cfg.CarrierPollInterval,
		carrier.NewPoller(store, carrier.NewHTTPClient(cfg.SelfBaseURL), cfg.CarrierPollBatchSize, logger),
	)
	go worker.RunNotificationSender(
		ctx,
		logger,
		cfg.NotificationInterval,
		notificationSender(store, cfg, logger),
	)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	logger.Printf("listening on %s", cfg.HTTPAddr)
	err = server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

// connectPool retries database startup so docker compose can bring up Postgres and the app together.
func connectPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	var lastErr error

	for attempt := 1; attempt <= cfg.DBConnectMaxAttempts; attempt++ {
		pool, err := openPool(ctx, cfg)
		if err == nil {
			return pool, nil
		}

		lastErr = err
		log.Printf("database connection attempt %d/%d failed: %v", attempt, cfg.DBConnectMaxAttempts, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.DBConnectTimeout):
		}
	}

	return nil, fmt.Errorf("connect database: %w", lastErr)
}

// openPool creates a pgx connection pool and validates it with a ping before returning.
func openPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	connCtx, cancel := context.WithTimeout(ctx, cfg.DBConnectTimeout)
	defer cancel()

	pool, err := pgxpool.New(connCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(connCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// notificationSender keeps construction of the notification worker dependency local to app wiring.
func notificationSender(store *postgres.Store, cfg config.Config, logger *log.Logger) *notification.Sender {
	return notification.NewSender(store, cfg.NotificationBatchSize, logger)
}
