package objectives

import (
	"context"
	"errors"
	"fmt"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service errors
var (
	ErrNotFound        = errors.New("objective not found")
	ErrNameExists      = errors.New("an objective with this name already exists")
	ErrVersionConflict = errors.New("objective changed since it was reviewed")
)

// Repository provides access to the objectives storage.
type Repository interface {
	List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]CoreObjective, error)
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreObjective, error)
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	UpdateIfUnchanged(ctx context.Context, id, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any) (bool, error)
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Create(ctx context.Context, objective CoreNewObjective, workspaceID uuid.UUID, keyResults []keyresults.CoreNewKeyResult) (CoreObjective, []keyresults.CoreKeyResult, error)
	GetAnalytics(ctx context.Context, objectiveID uuid.UUID, workspaceID uuid.UUID) (CoreObjectiveAnalytics, error)
	GetStrategyMap(ctx context.Context, workspaceID uuid.UUID) (CoreStrategyMap, error)
	UpdateStrategy(ctx context.Context, workspaceID uuid.UUID, strategy CoreStrategyUpdate) error
	CreateStrategicPillar(ctx context.Context, workspaceID uuid.UUID, pillar CoreNewStrategicPillar) (CoreStrategicPillar, error)
	UpdateStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID, pillar CoreUpdateStrategicPillar) (CoreStrategicPillar, error)
	DeleteStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID) error
	AlignObjective(ctx context.Context, workspaceID, objectiveID uuid.UUID, pillarID *uuid.UUID) error
}

// Service provides objective-related operations.
type Service struct {
	repo          Repository
	okrActivities *okractivities.Service
	log           *logger.Logger
	publisher     *publisher.Publisher
}

// Option configures optional objective service dependencies.
type Option func(*Service)

// WithPublisher publishes objective changes for notification and integration consumers.
func WithPublisher(eventPublisher *publisher.Publisher) Option {
	return func(service *Service) {
		service.publisher = eventPublisher
	}
}

func (s *Service) GetStrategyMap(ctx context.Context, workspaceID uuid.UUID) (CoreStrategyMap, error) {
	return s.repo.GetStrategyMap(ctx, workspaceID)
}

func (s *Service) UpdateStrategy(ctx context.Context, workspaceID uuid.UUID, strategy CoreStrategyUpdate) error {
	return s.repo.UpdateStrategy(ctx, workspaceID, strategy)
}

func (s *Service) CreateStrategicPillar(ctx context.Context, workspaceID uuid.UUID, pillar CoreNewStrategicPillar) (CoreStrategicPillar, error) {
	return s.repo.CreateStrategicPillar(ctx, workspaceID, pillar)
}

func (s *Service) UpdateStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID, pillar CoreUpdateStrategicPillar) (CoreStrategicPillar, error) {
	return s.repo.UpdateStrategicPillar(ctx, workspaceID, pillarID, pillar)
}

func (s *Service) DeleteStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID) error {
	return s.repo.DeleteStrategicPillar(ctx, workspaceID, pillarID)
}

func (s *Service) AlignObjective(ctx context.Context, workspaceID, objectiveID uuid.UUID, pillarID *uuid.UUID) error {
	return s.repo.AlignObjective(ctx, workspaceID, objectiveID, pillarID)
}

// New constructs a new objectives service instance with the provided repository.
func New(log *logger.Logger, repo Repository, okrActivities *okractivities.Service, options ...Option) *Service {
	service := &Service{
		repo:          repo,
		okrActivities: okrActivities,
		log:           log,
	}
	for _, option := range options {
		option(service)
	}
	return service
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

	s.recordUpdateSideEffects(ctx, id, workspaceId, userId, comment, updates)

	span.AddEvent("objective updated", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

// UpdateExternalUserActionIfUnchanged applies a user-requested external update
// only while the objective still has the version shown in the confirmation
// preview. Callers must separately reauthorize the actor immediately before
// invoking this method.
func (s *Service) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	expectedUpdatedAt time.Time,
	comment string,
	updates map[string]any,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected objective update time is required")
	}
	updated, err := s.repo.UpdateIfUnchanged(ctx, id, workspaceID, expectedUpdatedAt.UTC(), updates)
	if err != nil {
		return err
	}
	if !updated {
		current, getErr := s.repo.Get(ctx, id, workspaceID)
		if getErr == nil && objectiveExternalUpdatesAlreadyApplied(current, updates) {
			return nil
		}
		return ErrVersionConflict
	}
	s.recordUpdateSideEffects(ctx, id, workspaceID, userID, comment, updates)
	return nil
}

// objectiveExternalUpdatesAlreadyApplied makes an external retry idempotent
// after the domain write succeeded but its delivery/proposal receipt did not.
// Unknown fields deliberately return false so this never weakens CAS for a
// future action added without an explicit comparison here.
func objectiveExternalUpdatesAlreadyApplied(objective CoreObjective, updates map[string]any) bool {
	if len(updates) == 0 {
		return true
	}
	for field, value := range updates {
		switch field {
		case "health":
			if objective.Health == nil || string(*objective.Health) != fmt.Sprint(value) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (s *Service) recordUpdateSideEffects(ctx context.Context, id, workspaceID, userID uuid.UUID, comment string, updates map[string]any) {
	activities := []okractivities.CoreNewActivity{}
	for field, value := range updates {
		if field == "description" || field == "short_summary" {
			continue
		}
		activity := okractivities.CoreNewActivity{
			ObjectiveID:  id,
			KeyResultID:  nil,
			UserID:       userID,
			Type:         okractivities.ActivityTypeUpdate,
			UpdateType:   okractivities.UpdateTypeObjective,
			Field:        field,
			CurrentValue: s.formatValue(value),
			Comment:      comment,
			WorkspaceID:  workspaceID,
		}
		activities = append(activities, activity)
	}

	if err := s.okrActivities.CreateBatch(ctx, activities); err != nil {
		s.log.Error(ctx, "failed to record objective update activities", "error", err, "objectiveID", id)
		// Don't fail the update operation if activity recording fails
	}

	if s.publisher != nil && hasNotifiableObjectiveUpdate(updates) {
		objective, getErr := s.repo.Get(ctx, id, workspaceID)
		if getErr != nil {
			s.log.Error(ctx, "failed to load objective for update event", "error", getErr, "objectiveID", id)
		} else {
			event := events.Event{
				Type: events.ObjectiveUpdated,
				Payload: events.ObjectiveUpdatedPayload{
					ObjectiveID: id,
					WorkspaceID: workspaceID,
					Updates:     updates,
					LeadID:      objective.LeadUser,
				},
				Timestamp: time.Now().UTC(),
				ActorID:   userID,
			}
			if publishErr := s.publisher.Publish(ctx, event); publishErr != nil {
				s.log.Error(ctx, "failed to publish objective update event", "error", publishErr, "objectiveID", id)
			}
		}
	}

}

func hasNotifiableObjectiveUpdate(updates map[string]any) bool {
	for _, field := range []string{"lead_user_id", "status_id", "health", "end_date"} {
		if _, exists := updates[field]; exists {
			return true
		}
	}
	return false
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

	if newObjective.Color == "" {
		newObjective.Color = DefaultObjectiveColor
	}

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
