package workspacesrepository

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repo struct {
	queries         workspacesql.Querier
	bindTransaction func(pgx.Tx) workspacesql.Querier
	runTransaction  func(context.Context, func(workspacesql.Querier) error) error
}

// New constructs the workspace persistence adapter over the shared native pgx
// pool. All handwritten SQL for this module lives in the sqlc query files.
func New(pool *pgxpool.Pool) *repo {
	queries := workspacesql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)

	return &repo{
		queries: queries,
		bindTransaction: func(tx pgx.Tx) workspacesql.Querier {
			return queries.WithTx(tx)
		},
		runTransaction: func(ctx context.Context, operation func(workspacesql.Querier) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
	}
}

func newWithQueries(queries workspacesql.Querier) *repo {
	return &repo{queries: queries}
}

func (r *repo) withinTransaction(ctx context.Context, operation func(workspacesql.Querier) error) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if r == nil || r.runTransaction == nil {
		return errors.New("workspace repository transactions are unavailable")
	}
	return r.runTransaction(ctx, operation)
}

var colors = []string{
	"#FFE066", "#FF6B6B", "#C0392B", "#FFA07A", "#FFB6C1",
	"#E056FD", "#686DE0", "#E67E22", "#A8E6CF", "#9B59B6", "#8E44AD",
	"#6BCB77", "#4ECDC4", "#4A90E2", "#95A5A6", "#27AE60", "#2ECC71",
	"#30336B", "#B4A6AB", "#636E72", "#34495E", "#2C3E50",
}

// generateRandomColor returns a random color from the Colors slice
func generateRandomColor() (string, error) {
	if len(colors) == 0 {
		return "", fmt.Errorf("workspace color palette is empty")
	}
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		return "", fmt.Errorf("select workspace color: %w", err)
	}
	return colors[index.Int64()], nil
}
