package usersrepository

import (
	"context"
	"errors"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound                = usersdomain.ErrNotFound
	errTransactionsUnavailable = errors.New("users repository transactions are unavailable")
)

type repo struct {
	queries         usersql.Querier
	bindTransaction func(pgx.Tx) usersql.Querier
	runTransaction  func(context.Context, func(usersql.Querier) error) error
}

// New constructs the users persistence adapter over the shared native pgx
// pool. Generated sqlc contracts stay private to this package.
func New(pool *pgxpool.Pool) *repo {
	queries := usersql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)

	return &repo{
		queries: queries,
		bindTransaction: func(tx pgx.Tx) usersql.Querier {
			return queries.WithTx(tx)
		},
		runTransaction: func(ctx context.Context, operation func(usersql.Querier) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
	}
}

func newWithQueries(queries usersql.Querier) *repo {
	return &repo{queries: queries}
}

func (r *repo) withinTransaction(
	ctx context.Context,
	operation func(usersql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if r == nil || r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return r.runTransaction(ctx, operation)
}

func (r *repo) transactionQueries(tx pgx.Tx) (usersql.Querier, error) {
	if tx == nil {
		return nil, errors.New("users transaction is required")
	}
	if r == nil || r.bindTransaction == nil {
		return nil, errTransactionsUnavailable
	}
	return r.bindTransaction(tx), nil
}

// WorkspaceTransaction is the account capability made available only inside
// the workspace creation unit of work.
type WorkspaceTransaction interface {
	UpdateLastUsedWorkspace(context.Context, uuid.UUID, uuid.UUID) error
}

type workspaceTransaction struct {
	queries usersql.Querier
}

func (r *repo) BindWorkspaceTransaction(tx pgx.Tx) (WorkspaceTransaction, error) {
	queries, err := r.transactionQueries(tx)
	if err != nil {
		return nil, err
	}
	return &workspaceTransaction{queries: queries}, nil
}

func (transaction *workspaceTransaction) UpdateLastUsedWorkspace(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) error {
	return updateLastUsedWorkspace(ctx, transaction.queries, userID, workspaceID)
}
