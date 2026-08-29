package calendarrepository

import (
	"context"
	"errors"
	"fmt"

	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("calendar repository is not configured")

type Repo struct {
	queries           calendarsql.Querier
	withinTransaction func(context.Context, func(calendarsql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repo {
	if pool == nil {
		return &Repo{}
	}
	queries := calendarsql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.withinTransaction = func(ctx context.Context, operation func(calendarsql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries calendarsql.Querier) *Repo {
	repository := &Repo{queries: queries}
	if queries != nil {
		repository.withinTransaction = func(_ context.Context, operation func(calendarsql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (r *Repo) configured() error {
	if r == nil || r.queries == nil || r.withinTransaction == nil {
		return errRepositoryNotConfigured
	}
	return nil
}

// lockCalendarUser serializes calendar mutations for one user across all of
// their workspaces. This closes races between provider snapshots, reconnects,
// and workspace-specific schedule blocks.
func lockCalendarUser(ctx context.Context, queries calendarsql.Querier, userID uuid.UUID) error {
	if err := queries.LockCalendarUser(ctx, calendarsql.LockCalendarUserParams{UserID: userID}); err != nil {
		return fmt.Errorf("lock calendar user: %w", err)
	}
	return nil
}
