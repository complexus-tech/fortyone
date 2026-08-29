package documentsrepository

import (
	"context"
	"errors"
	"fmt"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maximumTransactionAttempts = 4

// Repository is the native pgx adapter for documents. It deliberately exposes
// domain values rather than sqlc-generated rows.
type Repository struct {
	pool       *pgxpool.Pool
	queries    *documentssql.Queries
	transactor platformdatabase.Transactor
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{
		pool:       pool,
		queries:    documentssql.New(pool),
		transactor: platformdatabase.NewTransactor(pool),
	}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.pool == nil || repository.queries == nil {
		return documentdomain.ErrNotConfigured
	}
	return nil
}

func (repository *Repository) withinSerializable(
	ctx context.Context,
	operation func(*documentssql.Queries) error,
) error {
	var transactionErr error
	for attempt := 0; attempt < maximumTransactionAttempts; attempt++ {
		transactionErr = repository.transactor.WithinTransaction(
			ctx,
			pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: pgx.ReadWrite},
			func(tx pgx.Tx) error {
				return operation(documentssql.New(tx))
			},
		)
		if transactionErr == nil || !platformdatabase.IsRetryableTransactionError(transactionErr) {
			return transactionErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return fmt.Errorf("document transaction remained contended after %d attempts: %w", maximumTransactionAttempts, transactionErr)
}

func mapNotFound(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return documentdomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func mapCreateError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return documentdomain.ErrForbidden
	}
	return fmt.Errorf("create document: %w", err)
}
