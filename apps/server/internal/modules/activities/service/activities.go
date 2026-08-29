package activities

import (
	"context"
	"fmt"

	activitiesdomain "github.com/complexus-tech/projects-api/internal/modules/activities/domain"
	"github.com/google/uuid"
)

type CoreActivity = activitiesdomain.Activity
type UserDetails = activitiesdomain.UserDetails
type CoreNewActivity = activitiesdomain.NewActivity
type ActivityFilters = activitiesdomain.Filters

// Repository provides access to the activities storage.
type Repository interface {
	Create(ctx context.Context, na CoreNewActivity) error
	GetActivities(ctx context.Context, userID uuid.UUID, limit int, workspaceId uuid.UUID, filters ActivityFilters) ([]CoreActivity, error)
}

// Service manages the activities operations.
type Service struct {
	repo Repository
}

// New constructs a activities service for api access.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create adds a new activity to the system.
func (s *Service) Create(ctx context.Context, na CoreNewActivity) error {
	if err := s.repo.Create(ctx, na); err != nil {
		return fmt.Errorf("creating activity: %w", err)
	}

	return nil
}

// GetActivities retrieves activities for a user.
func (s *Service) GetActivities(ctx context.Context, userID uuid.UUID, limit int, workspaceId uuid.UUID, filters ActivityFilters) ([]CoreActivity, error) {
	activities, err := s.repo.GetActivities(ctx, userID, limit, workspaceId, filters)
	if err != nil {
		return nil, fmt.Errorf("getting activities: %w", err)
	}

	return activities, nil
}
