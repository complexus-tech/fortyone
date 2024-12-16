package activities

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the story storage.
type Repository interface {
	RecordActivity(ctx context.Context, activity *CoreActivity) (CoreActivity, error)
	GetActivities(ctx context.Context, storyID uuid.UUID) ([]CoreActivity, error)
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

func (s *Service) RecordActivity(ctx context.Context, activity CoreNewActivity) (CoreActivity, error) {
	s.log.Info(ctx, "business.core.activities.RecordActivity")
	ctx, span := web.AddSpan(ctx, "business.core.activities.RecordActivity")
	defer span.End()

	ca := toCoreActivity(activity)

	ca, err := s.repo.RecordActivity(ctx, &ca)
	if err != nil {
		span.RecordError(err)
		return CoreActivity{}, err
	}

	span.AddEvent("activity recorded.", trace.WithAttributes(
		attribute.String("activity.story_id", ca.StoryID.String()),
	))

	return ca, nil
}

func (s *Service) GetActivities(ctx context.Context, storyID uuid.UUID) ([]CoreActivity, error) {
	s.log.Info(ctx, "business.core.activities.GetActivities")
	ctx, span := web.AddSpan(ctx, "business.core.activities.GetActivities")
	defer span.End()

	activities, err := s.repo.GetActivities(ctx, storyID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("activities retrieved.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
	))

	return activities, nil
}
