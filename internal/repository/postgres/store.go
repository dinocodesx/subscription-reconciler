package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool                *pgxpool.Pool
	now                 func() time.Time
	carrierPollInterval time.Duration
}

// NewStore wraps the Postgres pool with the repository methods used by the service.
func NewStore(pool *pgxpool.Pool, now func() time.Time, carrierPollInterval time.Duration) *Store {
	return &Store{
		pool:                pool,
		now:                 now,
		carrierPollInterval: carrierPollInterval,
	}
}

// Close releases the underlying database pool.
func (s *Store) Close() {
	s.pool.Close()
}
