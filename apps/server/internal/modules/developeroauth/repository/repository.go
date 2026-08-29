package developeroauthrepository

import (
	"context"
	"errors"
	"fmt"

	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool    *pgxpool.Pool
	queries *developeroauthsql.Queries
}

func New(pool *pgxpool.Pool) (*Store, error) {
	if pool == nil {
		return nil, errors.New("developer OAuth database pool is required")
	}
	return &Store{pool: pool, queries: developeroauthsql.New(pool)}, nil
}

func (store *Store) begin(ctx context.Context) (pgx.Tx, *developeroauthsql.Queries, error) {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite})
	if err != nil {
		return nil, nil, fmt.Errorf("begin developer OAuth transaction: %w", err)
	}
	return tx, store.queries.WithTx(tx), nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}
