package activities

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// Repository provides access to the activities storage.
type Repository interface {
	Create(ctx context.Context, na CoreNewActivity) error
	GetActivities(ctx context.Context, userID uuid.UUID, limit int, workspaceId uuid.UUID, filters ActivityFilters) ([]CoreActivity, error)
}

// Service manages the activities operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a activities service for api access.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Create adds a new activity to the system.
func (s *Service) Create(ctx context.Context, na CoreNewActivity) error {
	s.log.Info(ctx, "creating new activity", "type", na.Type)
	ctx, span := web.AddSpan(ctx, "business.core.activities.Create")
	defer span.End()

	if err := s.repo.Create(ctx, na); err != nil {
		span.RecordError(err)
		return fmt.Errorf("creating activity: %w", err)
	}

	return nil
}

// GetActivities retrieves activities for a user.
func (s *Service) GetActivities(ctx context.Context, userID uuid.UUID, limit int, workspaceId uuid.UUID, filters ActivityFilters) ([]CoreActivity, error) {
	s.log.Info(ctx, "getting activities")
	ctx, span := web.AddSpan(ctx, "business.core.activities.GetActivities")
	defer span.End()

	activities, err := s.repo.GetActivities(ctx, userID, limit, workspaceId, filters)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("getting activities: %w", err)
	}

	return activities, nil
}
