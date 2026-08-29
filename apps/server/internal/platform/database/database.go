package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connections owns the single native pgx pool shared by SQLC repositories in
// one process.
type Connections struct {
	Pool *pgxpool.Pool
}

// OpenPool opens and verifies the native pgx pool used by generated sqlc
// queries. The caller is responsible for closing the returned pool.
func OpenPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolConfig, err := newPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	startupCtx, cancel := context.WithTimeout(ctx, poolConfig.ConnConfig.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(startupCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL pool: %w", err)
	}

	if err := pool.Ping(startupCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return pool, nil
}

// Open opens the process's single native pgx pool.
func Open(ctx context.Context, cfg Config) (*Connections, error) {
	pool, err := OpenPool(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Connections{Pool: pool}, nil
}

// Close drains the native pool. It returns an error-shaped result so process
// lifecycle registries can compose it with other resource closers.
func (c *Connections) Close() error {
	if c == nil {
		return nil
	}
	if c.Pool != nil {
		c.Pool.Close()
	}
	return nil
}

func newPoolConfig(cfg Config) (*pgxpool.Config, error) {
	if cfg.MaxOpenConns < 0 {
		return nil, fmt.Errorf("max open PostgreSQL connections cannot be negative: %d", cfg.MaxOpenConns)
	}
	if cfg.MaxOpenConns > math.MaxInt32 {
		return nil, fmt.Errorf("max open PostgreSQL connections exceeds pgx limit: %d", cfg.MaxOpenConns)
	}
	if cfg.MinConns < 0 || cfg.MinConns > math.MaxInt32 {
		return nil, fmt.Errorf("minimum PostgreSQL connections is outside the pgx range: %d", cfg.MinConns)
	}
	if cfg.ConnectTimeout < 0 || cfg.MaxConnIdleTime < 0 || cfg.MaxConnLifetime < 0 || cfg.HealthCheckPeriod < 0 {
		return nil, errors.New("PostgreSQL pool durations cannot be negative")
	}

	connectionString, err := ConnectionString(cfg)
	if err != nil {
		return nil, err
	}
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL pool config: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = int32(cfg.MinConns)
	}
	if poolConfig.MinConns > poolConfig.MaxConns {
		return nil, fmt.Errorf("minimum PostgreSQL connections %d exceeds maximum %d", poolConfig.MinConns, poolConfig.MaxConns)
	}
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	if poolConfig.ConnConfig.ConnectTimeout == 0 {
		poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	return poolConfig, nil
}
