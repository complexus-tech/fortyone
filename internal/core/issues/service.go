package issues

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/repository/issues"
	"github.com/complexus-tech/projects-api/internal/web"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service interface {
	List(ctx context.Context) ([]issues.Issue, error)
	Get(ctx context.Context, id int) (issues.Issue, error)
}

type service struct {
	repo issues.Repository
	log  *logger.Logger
}

func NewService(log *logger.Logger, db *sqlx.DB) Service {
	return &service{
		repo: issues.NewRepository(log, db),
		log:  log,
	}
}

// List returns a list of issues.
func (s *service) List(ctx context.Context) ([]issues.Issue, error) {
	s.log.Info(ctx, "business.core.issues.list")
	ctx, span := web.AddSpan(ctx, "business.core.issues.List")
	defer span.End()

	issues, err := s.repo.List(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("issues retrieved.", trace.WithAttributes(
		attribute.Int("issue.count", len(issues)),
	))
	return issues, nil
}

// Get returns the issue with the specified ID.
func (s *service) Get(ctx context.Context, id int) (issues.Issue, error) {
	s.log.Info(ctx, "business.core.issues.Get")
	ctx, span := web.AddSpan(ctx, "business.core.issues.Get")
	defer span.End()

	issue, err := s.repo.Get(ctx, id)
	if err != nil {
		span.RecordError(err)
		return issues.Issue{}, err
	}
	return issue, nil
}
