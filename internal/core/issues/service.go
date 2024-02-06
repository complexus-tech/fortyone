package issues

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Repository interface {
	MyIssues(ctx context.Context) ([]Issue, error)
	Get(ctx context.Context, id int) (Issue, error)
}

type Service struct {
	repo Repository
	log  *logger.Logger
}

func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// MyIssues returns a list of issues.
func (s *Service) MyIssues(ctx context.Context) ([]Issue, error) {
	s.log.Info(ctx, "business.core.issues.list")
	ctx, span := web.AddSpan(ctx, "business.core.issues.List")
	defer span.End()

	issues, err := s.repo.MyIssues(ctx)
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
func (s *Service) Get(ctx context.Context, id int) (Issue, error) {
	s.log.Info(ctx, "business.core.issues.Get")
	ctx, span := web.AddSpan(ctx, "business.core.issues.Get")
	defer span.End()

	issue, err := s.repo.Get(ctx, id)
	if err != nil {
		span.RecordError(err)
		return Issue{}, err
	}
	return issue, nil
}
