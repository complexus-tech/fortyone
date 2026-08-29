package storiesrepository

import (
	"context"

	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repo struct {
	pool       *pgxpool.Pool
	reads      storyreadsql.Querier
	transactor platformdatabase.Transactor
	log        *logger.Logger

	attachmentObjectStorage *attachmentObjectStorageRoute

	retention               storyRetentionQueries
	runRetentionTransaction func(context.Context, func(storyRetentionQueries) error) error

	runStoryAutomationTransaction func(context.Context, func(storyAutomationQueries) error) error
}

// New constructs the stories PostgreSQL adapter. All story persistence uses
// generated SQLC queries over native pgx; there is no legacy SQLx connection.
func New(log *logger.Logger, pool *pgxpool.Pool, options ...Option) *repo {
	var reads storyreadsql.Querier
	if pool != nil {
		reads = storyreadsql.New(pool)
	}
	repository := &repo{
		pool:       pool,
		reads:      reads,
		transactor: platformdatabase.NewTransactor(pool),
		log:        log,
	}
	for _, option := range options {
		if option != nil {
			option(repository)
		}
	}
	if reads != nil {
		repository.retention = reads
		repository.runRetentionTransaction = func(
			ctx context.Context,
			operation func(storyRetentionQueries) error,
		) error {
			return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(storyreadsql.New(tx))
			})
		}
		repository.runStoryAutomationTransaction = func(
			ctx context.Context,
			operation func(storyAutomationQueries) error,
		) error {
			return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(storyreadsql.New(tx))
			})
		}
	}
	return repository
}

// NewMutationRepository constructs the pgx-only story mutation/outbox adapter
// for workers that must not depend on the legacy SQLx connection.
func NewMutationRepository(log *logger.Logger, pool *pgxpool.Pool, options ...Option) *repo {
	return New(log, pool, options...)
}
