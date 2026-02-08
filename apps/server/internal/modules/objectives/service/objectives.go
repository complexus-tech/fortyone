package objectives

import (
	"context"
	"errors"
	"fmt"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service errors
var (
	ErrNotFound   = errors.New("objective not found")
	ErrNameExists = errors.New("an objective with this name already exists")
)

// Repository provides access to the objectives storage.
type Repository interface {
	List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]CoreObjective, error)
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreObjective, error)
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Create(ctx context.Context, objective CoreNewObjective, workspaceID uuid.UUID, keyResults []keyresults.CoreNewKeyResult) (CoreObjective, []keyresults.CoreKeyResult, error)
	GetAnalytics(ctx context.Context, objectiveID uuid.UUID, workspaceID uuid.UUID) (CoreObjectiveAnalytics, error)
}

// Service provides objective-related operations.
type Service struct {
	repo          Repository
	okrActivities *okractivities.Service
	log           *logger.Logger
}

// New constructs a new objectives service instance with the provided repository.
func New(log *logger.Logger, repo Repository, okrActivities *okractivities.Service) *Service {
	return &Service{
		repo:          repo,
		okrActivities: okrActivities,
		log:           log,
	}
}

// Get returns an objective by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreObjective, error) {
	s.log.Info(ctx, "business.core.objectives.Get")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Get")
	defer span.End()

	objective, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		span.RecordError(err)
		return CoreObjective{}, err
	}

	span.AddEvent("objective retrieved.", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return objective, nil
}

// List returns a list of objectives.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]CoreObjective, error) {
	s.log.Info(ctx, "business.core.objectives.list")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.List")
	defer span.End()

	objectives, err := s.repo.List(ctx, workspaceId, userID, filters)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("objectives retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(objectives)),
	))
	return objectives, nil
}

// Update updates an objective in the system
func (s *Service) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, userId uuid.UUID, comment string, updates map[string]any) error {
	s.log.Info(ctx, "business.core.objectives.Update")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Update")
	defer span.End()

	if err := s.repo.Update(ctx, id, workspaceId, updates); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		span.RecordError(err)
		return err
	}

	activities := []okractivities.CoreNewActivity{}
	for field, value := range updates {
		if field == "description" {
			continue
		}
		activity := okractivities.CoreNewActivity{
			ObjectiveID:  id,
			KeyResultID:  nil,
			UserID:       userId,
			Type:         okractivities.ActivityTypeUpdate,
			UpdateType:   okractivities.UpdateTypeObjective,
			Field:        field,
			CurrentValue: s.formatValue(value),
			Comment:      comment,
			WorkspaceID:  workspaceId,
		}
		activities = append(activities, activity)
	}

	if err := s.okrActivities.CreateBatch(ctx, activities); err != nil {
		s.log.Error(ctx, "failed to record objective update activities", "error", err, "objectiveID", id)
		// Don't fail the update operation if activity recording fails
	}

	span.AddEvent("objective updated", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

// Delete removes an objective from the system
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.objectives.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		span.RecordError(err)
		return err
	}

	span.AddEvent("objective deleted", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

// Create creates a new objective with optional key results
func (s *Service) Create(ctx context.Context, newObjective CoreNewObjective, workspaceID uuid.UUID, keyResults []keyresults.CoreNewKeyResult) (CoreObjective, []keyresults.CoreKeyResult, error) {
	s.log.Info(ctx, "business.core.objectives.Create")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Create")
	defer span.End()

	createdObj, createdKRs, err := s.repo.Create(ctx, newObjective, workspaceID, keyResults)
	if err != nil {
		span.RecordError(err)
		return CoreObjective{}, nil, err
	}

	// Record the create activity
	ca := []okractivities.CoreNewActivity{}
	activity := okractivities.CoreNewActivity{
		ObjectiveID:  createdObj.ID,
		KeyResultID:  nil,
		UserID:       newObjective.CreatedBy,
		Type:         okractivities.ActivityTypeCreate,
		UpdateType:   okractivities.UpdateTypeObjective,
		Field:        "all",
		CurrentValue: createdObj.Name,
		Comment:      "",
		WorkspaceID:  workspaceID,
	}
	ca = append(ca, activity)

	for _, kr := range createdKRs {
		activity := okractivities.CoreNewActivity{
			ObjectiveID:  createdObj.ID,
			KeyResultID:  &kr.ID,
			UserID:       newObjective.CreatedBy,
			Type:         okractivities.ActivityTypeCreate,
			UpdateType:   okractivities.UpdateTypeKeyResult,
			CurrentValue: kr.Name,
			Comment:      "",
			WorkspaceID:  workspaceID,
		}
		ca = append(ca, activity)
	}

	if err := s.okrActivities.CreateBatch(ctx, ca); err != nil {
		s.log.Error(ctx, "failed to record objective create activity", "error", err, "objectiveID", createdObj.ID)
		// Don't fail the create operation if activity recording fails
	}

	span.AddEvent("objective created.", trace.WithAttributes(
		attribute.String("objective.id", createdObj.ID.String()),
		attribute.Int("key_results.count", len(createdKRs)),
	))

	return createdObj, createdKRs, nil
}

// GetAnalytics returns analytics data for an objective.
func (s *Service) GetAnalytics(ctx context.Context, objectiveID uuid.UUID, workspaceID uuid.UUID) (CoreObjectiveAnalytics, error) {
	s.log.Info(ctx, "business.core.objectives.GetAnalytics")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.GetAnalytics")
	defer span.End()

	analytics, err := s.repo.GetAnalytics(ctx, objectiveID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreObjectiveAnalytics{}, err
	}

	span.AddEvent("objective analytics retrieved.", trace.WithAttributes(
		attribute.String("objective.id", objectiveID.String()),
		attribute.Int("priority_breakdown.count", len(analytics.PriorityBreakdown)),
		attribute.Int("team_allocation.count", len(analytics.TeamAllocation)),
	))

	return analytics, nil
}

func (s *Service) formatValue(value any) string {
	if value == nil {
		return "nil"
	}
	switch v := value.(type) {
	case *float64:
		if v != nil {
			return fmt.Sprintf("%.2f", *v)
		}
		return "nil"
	case *uuid.UUID:
		if v != nil {
			return v.String()
		}
		return "nil"
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}
