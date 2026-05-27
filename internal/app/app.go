package app

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/subscription-reconciler/internal/config"
	"github.com/subscription-reconciler/internal/repository/postgres"
)

func Run(ctx context.Context, cfg config.Config) error {
	// logger := log.Default()

	pool, err := connectPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	store := postgres.NewStore(pool, time.Now, cfg.CarrierPollInterval)
	if err := store.Migrate(ctx, filepath.Join(".", "migrations")); err != nil {
		return err
	}
	// if err := store.SeedDemoData(ctx); err != nil {
	// 	return err
	// }

	// router := httpapi.NewRouter(store)
	// server := &http.Server{
	// 	Addr:    cfg.HTTPAddr,
	// 	Handler: router,
	// }

	return nil
}

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
