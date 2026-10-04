package notificationsrepository

import (
	"context"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the native pgx persistence adapter for notifications. SQLC's
// generated package remains private to this adapter so database types do not
// leak into services or transports.
type Repository struct {
	queries           notificationssql.Querier
	pool              *pgxpool.Pool
	feedbackCompleter RoutineFeedbackCompleter
}

// RoutineFeedbackCompleter joins feedback cursor advancement to the caller's
// transaction without coupling this adapter to another concrete repository.
type RoutineFeedbackCompleter interface {
	CompleteDigestDeliveryTx(context.Context, pgx.Tx, feedback.CoreDigestDeliveryCompletion) error
}

type Option func(*Repository)

func WithFeedbackDigestCompletion(completer RoutineFeedbackCompleter) Option {
	return func(repository *Repository) { repository.feedbackCompleter = completer }
}

func New(pool *pgxpool.Pool, options ...Option) *Repository {
	repository := &Repository{queries: notificationssql.New(pool), pool: pool}
	for _, option := range options {
		option(repository)
	}
	return repository
}
