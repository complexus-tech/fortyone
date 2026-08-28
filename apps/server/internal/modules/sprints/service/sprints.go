package sprints

import (
	"context"
	"fmt"
	"math"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/pkg/logger"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository is the persistence capability consumed by the sprint service.
type Repository interface {
	List(ctx context.Context, query sprintdomain.ListQuery) ([]sprintdomain.Sprint, error)
	Running(ctx context.Context, workspaceID, actorID uuid.UUID, today time.Time) ([]sprintdomain.Sprint, error)
	GetByID(ctx context.Context, sprintID, workspaceID, actorID uuid.UUID) (sprintdomain.Sprint, error)
	Create(ctx context.Context, command sprintdomain.CreateCommand) (sprintdomain.Sprint, error)
	Update(ctx context.Context, command sprintdomain.UpdateCommand) (sprintdomain.Sprint, error)
	Delete(ctx context.Context, command sprintdomain.DeleteCommand) error
	GetAnalytics(ctx context.Context, sprintID, workspaceID, actorID uuid.UUID, now time.Time) (sprintdomain.Analytics, error)
}

// Service coordinates sprint policy and persistence.
type Service struct {
	repo  Repository
	log   *logger.Logger
	clock platformclock.Clock
}

// New constructs a sprint service using wall-clock decision time.
func New(log *logger.Logger, repo Repository) *Service {
	return NewWithClock(log, repo, platformclock.System{})
}

// NewWithClock constructs a sprint service with deterministic decision time.
func NewWithClock(log *logger.Logger, repo Repository, clock platformclock.Clock) *Service {
	if clock == nil {
		clock = platformclock.System{}
	}
	return &Service{repo: repo, log: log, clock: clock}
}

// List retains the established in-process map boundary while immediately
// converting it to the module's finite typed filter. Unknown filters fail
// closed instead of becoming dynamic SQL identifiers.
func (s *Service) List(ctx context.Context, workspaceID, actorID uuid.UUID, filters map[string]any) ([]CoreSprint, error) {
	filter, err := listFilterFromCompatibilityMap(filters)
	if err != nil {
		return nil, err
	}
	return s.ListQuery(ctx, sprintdomain.ListQuery{WorkspaceID: workspaceID, ActorID: actorID, Filter: filter})
}

// ListQuery is the strongly typed sprint-list entry point.
func (s *Service) ListQuery(ctx context.Context, query sprintdomain.ListQuery) ([]CoreSprint, error) {
	query, err := query.Normalize()
	if err != nil {
		return nil, err
	}

	s.log.Info(ctx, "business.core.sprints.list")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.List")
	defer span.End()

	items, err := s.repo.List(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("sprints retrieved", trace.WithAttributes(attribute.Int("sprint.count", len(items))))
	return items, nil
}

func (s *Service) Running(ctx context.Context, workspaceID, actorID uuid.UUID) ([]CoreSprint, error) {
	if workspaceID == uuid.Nil || actorID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace and actor are required", sprintdomain.ErrInvalid)
	}

	s.log.Info(ctx, "business.core.sprints.running")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.Running")
	defer span.End()

	items, err := s.repo.Running(ctx, workspaceID, actorID, s.clock.Now())
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("running sprints retrieved", trace.WithAttributes(attribute.Int("sprint.count", len(items))))
	return items, nil
}

func (s *Service) GetByID(ctx context.Context, sprintID, workspaceID, actorID uuid.UUID) (CoreSprint, error) {
	if sprintID == uuid.Nil || workspaceID == uuid.Nil || actorID == uuid.Nil {
		return CoreSprint{}, fmt.Errorf("%w: sprint, workspace, and actor are required", sprintdomain.ErrInvalid)
	}

	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.GetByID")
	defer span.End()
	item, err := s.repo.GetByID(ctx, sprintID, workspaceID, actorID)
	if err != nil {
		span.RecordError(err)
		return CoreSprint{}, err
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, sprint CoreNewSprint, actorID *uuid.UUID) (CoreSprint, error) {
	if actorID == nil {
		return CoreSprint{}, fmt.Errorf("%w: actor is required", sprintdomain.ErrInvalid)
	}
	command, err := (sprintdomain.CreateCommand{Sprint: sprint, ActorID: *actorID}).Normalize()
	if err != nil {
		return CoreSprint{}, err
	}

	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.Create")
	defer span.End()
	created, err := s.repo.Create(ctx, command)
	if err != nil {
		span.RecordError(err)
		return CoreSprint{}, err
	}
	return created, nil
}

// UpdatePatch applies a finite typed patch and optional compare-and-swap token.
func (s *Service) UpdatePatch(
	ctx context.Context,
	sprintID, workspaceID, actorID uuid.UUID,
	patch SprintPatch,
	expectedUpdatedAt *time.Time,
) (CoreSprint, error) {
	command, err := (sprintdomain.UpdateCommand{
		SprintID: sprintID, WorkspaceID: workspaceID, ActorID: actorID,
		Patch: patch, ExpectedUpdatedAt: expectedUpdatedAt,
	}).Normalize()
	if err != nil {
		return CoreSprint{}, err
	}

	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.Update")
	defer span.End()
	updated, err := s.repo.Update(ctx, command)
	if err != nil {
		span.RecordError(err)
		return CoreSprint{}, err
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, sprintID, workspaceID uuid.UUID, actorID *uuid.UUID) error {
	if actorID == nil {
		return fmt.Errorf("%w: actor is required", sprintdomain.ErrInvalid)
	}
	command := sprintdomain.DeleteCommand{SprintID: sprintID, WorkspaceID: workspaceID, ActorID: *actorID}
	if err := command.Validate(); err != nil {
		return err
	}

	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.Delete")
	defer span.End()
	if err := s.repo.Delete(ctx, command); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *Service) GetAnalytics(
	ctx context.Context,
	sprintID, workspaceID, actorID uuid.UUID,
) (CoreSprintAnalytics, error) {
	if sprintID == uuid.Nil || workspaceID == uuid.Nil || actorID == uuid.Nil {
		return CoreSprintAnalytics{}, fmt.Errorf("%w: sprint, workspace, and actor are required", sprintdomain.ErrInvalid)
	}

	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.sprints.GetAnalytics")
	defer span.End()
	analytics, err := s.repo.GetAnalytics(ctx, sprintID, workspaceID, actorID, s.clock.Now())
	if err != nil {
		span.RecordError(err)
		return CoreSprintAnalytics{}, err
	}
	return analytics, nil
}

func listFilterFromCompatibilityMap(filters map[string]any) (sprintdomain.ListFilter, error) {
	var result sprintdomain.ListFilter
	for name, value := range filters {
		var err error
		switch name {
		case "sprint_id":
			result.SprintID, err = optionalUUIDFilter(name, value)
		case "objective_id":
			result.ObjectiveID, err = optionalUUIDFilter(name, value)
		case "team_id":
			result.TeamID, err = optionalUUIDFilter(name, value)
		case "search":
			result.Search, err = stringFilter(name, value)
		case "limit":
			result.Limit, err = integerFilter(name, value)
		case "offset":
			result.Offset, err = integerFilter(name, value)
		case "page", "pageSize":
			// HTTP pagination metadata is converted to limit/offset by the handler.
		default:
			return sprintdomain.ListFilter{}, fmt.Errorf("%w: unsupported sprint filter %q", sprintdomain.ErrInvalid, name)
		}
		if err != nil {
			return sprintdomain.ListFilter{}, err
		}
	}
	return result.Normalize()
}

func optionalUUIDFilter(name string, value any) (*uuid.UUID, error) {
	switch typed := value.(type) {
	case uuid.UUID:
		return &typed, nil
	case *uuid.UUID:
		return typed, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("%w: filter %s must be a UUID", sprintdomain.ErrInvalid, name)
	}
}

func stringFilter(name string, value any) (string, error) {
	if value == nil {
		return "", nil
	}
	result, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: filter %s must be a string", sprintdomain.ErrInvalid, name)
	}
	return result, nil
}

func integerFilter(name string, value any) (int, error) {
	switch typed := value.(type) {
	case int:
		return typed, nil
	case int32:
		return int(typed), nil
	case int64:
		if typed > int64(math.MaxInt) || typed < int64(math.MinInt) {
			return 0, fmt.Errorf("%w: filter %s is outside integer range", sprintdomain.ErrInvalid, name)
		}
		return int(typed), nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: filter %s must be an integer", sprintdomain.ErrInvalid, name)
	}
}
