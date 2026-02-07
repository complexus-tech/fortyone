package documents

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the documents storage.
type Repository interface {
	List(ctx context.Context) ([]CoreDocument, error)
}

// Service provides story-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new stories service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// List returns a list of documents.
func (s *Service) List(ctx context.Context) ([]CoreDocument, error) {
	s.log.Info(ctx, "business.core.documents.list")
	ctx, span := web.AddSpan(ctx, "business.core.documents.List")
	defer span.End()

	documents, err := s.repo.List(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("documents retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(documents)),
	))
	return documents, nil
}
