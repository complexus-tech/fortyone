package okractivities

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// Repository provides access to the OKR activities storage.
type Repository interface {
	Create(ctx context.Context, na CoreNewActivity) error
	GetObjectiveActivities(ctx context.Context, objectiveID uuid.UUID, page, pageSize int) ([]CoreActivity, bool, error)
	GetKeyResultActivities(ctx context.Context, keyResultID uuid.UUID, page, pageSize int) ([]CoreActivity, bool, error)
}

// Service manages the OKR activities operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs an OKR activities service for API access.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Create adds a new OKR activity to the system.
func (s *Service) Create(ctx context.Context, na CoreNewActivity) error {
	s.log.Info(ctx, "creating new OKR activity", "type", na.Type, "updateType", na.UpdateType)
	ctx, span := web.AddSpan(ctx, "business.core.okractivities.Create")
	defer span.End()

	if err := s.repo.Create(ctx, na); err != nil {
		span.RecordError(err)
		return fmt.Errorf("creating OKR activity: %w", err)
	}

	return nil
}

// CreateBatch adds multiple OKR activities to the system.
func (s *Service) CreateBatch(ctx context.Context, activities []CoreNewActivity) error {
	s.log.Info(ctx, "creating batch of OKR activities", "count", len(activities))
	ctx, span := web.AddSpan(ctx, "business.core.okractivities.CreateBatch")
	defer span.End()

	for _, activity := range activities {
		if err := s.repo.Create(ctx, activity); err != nil {
			span.RecordError(err)
			return fmt.Errorf("creating OKR activity batch: %w", err)
		}
	}

	return nil
}

// GetObjectiveActivities retrieves activities for a specific objective.
func (s *Service) GetObjectiveActivities(ctx context.Context, objectiveID uuid.UUID, page, pageSize int) ([]CoreActivity, bool, error) {
	s.log.Info(ctx, "getting objective activities", "objectiveID", objectiveID, "page", page, "pageSize", pageSize)
	ctx, span := web.AddSpan(ctx, "business.core.okractivities.GetObjectiveActivities")
	defer span.End()

	activities, hasMore, err := s.repo.GetObjectiveActivities(ctx, objectiveID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		return nil, false, fmt.Errorf("getting objective activities: %w", err)
	}

	return activities, hasMore, nil
}

// GetKeyResultActivities retrieves activities for a specific key result.
func (s *Service) GetKeyResultActivities(ctx context.Context, keyResultID uuid.UUID, page, pageSize int) ([]CoreActivity, bool, error) {
	s.log.Info(ctx, "getting key result activities", "keyResultID", keyResultID, "page", page, "pageSize", pageSize)
	ctx, span := web.AddSpan(ctx, "business.core.okractivities.GetKeyResultActivities")
	defer span.End()

	activities, hasMore, err := s.repo.GetKeyResultActivities(ctx, keyResultID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		return nil, false, fmt.Errorf("getting key result activities: %w", err)
	}

	return activities, hasMore, nil
}
