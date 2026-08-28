package teamsrepository

import (
	"context"
	"errors"

	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("teams repository transactions are unavailable")

type repo struct {
	queries         teamsql.Querier
	bindTransaction func(pgx.Tx) teamsql.Querier
	runTransaction  func(context.Context, func(teamsql.Querier) error) error
}

// New constructs the native pgx/sqlc teams repository. The pool owns both
// ordinary queries and every repository-created transaction.
func New(pool *pgxpool.Pool) *repo {
	queries := teamsql.New(pool)
	transactor := database.NewTransactor(pool)

	return &repo{
		queries: queries,
		bindTransaction: func(tx pgx.Tx) teamsql.Querier {
			return queries.WithTx(tx)
		},
		runTransaction: func(ctx context.Context, operation func(teamsql.Querier) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
	}
}

func newWithQueries(queries teamsql.Querier) *repo {
	return &repo{queries: queries}
}

func (r *repo) withinTransaction(
	ctx context.Context,
	operation func(teamsql.Querier) error,
) error {
	if operation == nil {
		return database.ErrNilTransactionOperation
	}
	if r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return r.runTransaction(ctx, operation)
}

func (r *repo) transactionQueries(tx pgx.Tx) (teamsql.Querier, error) {
	if tx == nil {
		return nil, errors.New("teams transaction is required")
	}
	if r.bindTransaction == nil {
		return nil, errTransactionsUnavailable
	}
	return r.bindTransaction(tx), nil
}

// WorkspaceTransaction is the team capability exposed to the workspace unit
// of work after binding all generated queries to one caller-owned transaction.
type WorkspaceTransaction interface {
	CreateTeam(context.Context, WorkspaceTeamInput) (WorkspaceTeam, error)
	AddMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type WorkspaceTeamInput struct {
	Name      string
	Code      string
	Color     string
	Workspace uuid.UUID
}

type WorkspaceTeam struct {
	ID uuid.UUID
}

type workspaceTransaction struct {
	queries teamsql.Querier
}

func (r *repo) BindWorkspaceTransaction(tx pgx.Tx) (WorkspaceTransaction, error) {
	queries, err := r.transactionQueries(tx)
	if err != nil {
		return nil, err
	}
	return &workspaceTransaction{queries: queries}, nil
}
