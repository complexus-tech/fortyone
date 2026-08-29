package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
)

const (
	MaximumNameLength = 255
	MaximumGoalLength = 10_000
)

var (
	ErrInvalid          = errors.New("invalid sprint request")
	ErrForbidden        = errors.New("sprint access is forbidden")
	ErrNotFound         = errors.New("sprint not found")
	ErrVersionConflict  = errors.New("sprint changed since it was reviewed")
	ErrInvalidReference = errors.New("sprint reference is invalid")
)

// NewSprint contains the values needed to create a sprint.
type NewSprint struct {
	Name        string
	Goal        *string
	ObjectiveID *uuid.UUID
	TeamID      uuid.UUID
	WorkspaceID uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
}

// CreateCommand binds a sprint mutation to its authenticated actor.
type CreateCommand struct {
	Sprint  NewSprint
	ActorID uuid.UUID
}

func (command CreateCommand) Normalize() (CreateCommand, error) {
	if command.ActorID == uuid.Nil {
		return CreateCommand{}, fmt.Errorf("%w: actor is required", ErrInvalid)
	}
	normalized, err := command.Sprint.Normalize()
	if err != nil {
		return CreateCommand{}, err
	}
	command.Sprint = normalized
	return command, nil
}

func (sprint NewSprint) Normalize() (NewSprint, error) {
	sprint.Name = strings.TrimSpace(sprint.Name)
	if err := validateName(sprint.Name); err != nil {
		return NewSprint{}, err
	}
	if sprint.TeamID == uuid.Nil || sprint.WorkspaceID == uuid.Nil {
		return NewSprint{}, fmt.Errorf("%w: workspace and team are required", ErrInvalid)
	}
	if sprint.ObjectiveID != nil && *sprint.ObjectiveID == uuid.Nil {
		return NewSprint{}, fmt.Errorf("%w: objective cannot be a zero id", ErrInvalid)
	}
	goal, err := normalizeOptionalGoal(sprint.Goal)
	if err != nil {
		return NewSprint{}, err
	}
	sprint.Goal = goal
	sprint.StartDate = normalizeDate(sprint.StartDate)
	sprint.EndDate = normalizeDate(sprint.EndDate)
	if sprint.StartDate.IsZero() || sprint.EndDate.IsZero() || sprint.EndDate.Before(sprint.StartDate) {
		return NewSprint{}, fmt.Errorf("%w: end date cannot be before start date", ErrInvalid)
	}
	return sprint, nil
}

// Patch is the finite sprint update surface. Nullable fields can be cleared;
// required fields reject explicit null.
type Patch struct {
	Name        platformpatch.Field[string]
	Goal        platformpatch.Field[string]
	ObjectiveID platformpatch.Field[uuid.UUID]
	StartDate   platformpatch.Field[time.Time]
	EndDate     platformpatch.Field[time.Time]
}

func (patch Patch) Empty() bool {
	return !patch.Name.Specified() && !patch.Goal.Specified() &&
		!patch.ObjectiveID.Specified() && !patch.StartDate.Specified() &&
		!patch.EndDate.Specified()
}

func (patch Patch) Normalize() (Patch, error) {
	if patch.Empty() {
		return Patch{}, fmt.Errorf("%w: at least one update is required", ErrInvalid)
	}
	if value, specified := patch.Name.Value(); specified {
		if value == nil {
			return Patch{}, fmt.Errorf("%w: name cannot be null", ErrInvalid)
		}
		trimmed := strings.TrimSpace(*value)
		if err := validateName(trimmed); err != nil {
			return Patch{}, err
		}
		patch.Name = platformpatch.Set(trimmed)
	}
	if value, specified := patch.Goal.Value(); specified && value != nil {
		normalized, err := normalizeOptionalGoal(value)
		if err != nil {
			return Patch{}, err
		}
		if normalized == nil {
			patch.Goal = platformpatch.Clear[string]()
		} else {
			patch.Goal = platformpatch.Set(*normalized)
		}
	}
	if value, specified := patch.ObjectiveID.Value(); specified && value != nil && *value == uuid.Nil {
		return Patch{}, fmt.Errorf("%w: objective cannot be a zero id", ErrInvalid)
	}
	for _, field := range []*platformpatch.Field[time.Time]{&patch.StartDate, &patch.EndDate} {
		value, specified := field.Value()
		if !specified {
			continue
		}
		if value == nil || value.IsZero() {
			return Patch{}, fmt.Errorf("%w: sprint dates cannot be null", ErrInvalid)
		}
		*field = platformpatch.Set(normalizeDate(*value))
	}
	if start, startSet := patch.StartDate.Value(); startSet && start != nil {
		if end, endSet := patch.EndDate.Value(); endSet && end != nil && end.Before(*start) {
			return Patch{}, fmt.Errorf("%w: end date cannot be before start date", ErrInvalid)
		}
	}
	return patch, nil
}

// UpdateCommand scopes a partial update and optionally carries optimistic
// concurrency state from a client that previously read the sprint.
type UpdateCommand struct {
	SprintID          uuid.UUID
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	ExpectedUpdatedAt *time.Time
	Patch             Patch
}

func (command UpdateCommand) Normalize() (UpdateCommand, error) {
	if command.SprintID == uuid.Nil || command.WorkspaceID == uuid.Nil || command.ActorID == uuid.Nil {
		return UpdateCommand{}, fmt.Errorf("%w: sprint, workspace, and actor are required", ErrInvalid)
	}
	normalized, err := command.Patch.Normalize()
	if err != nil {
		return UpdateCommand{}, err
	}
	command.Patch = normalized
	if command.ExpectedUpdatedAt != nil {
		value := command.ExpectedUpdatedAt.UTC()
		command.ExpectedUpdatedAt = &value
	}
	return command, nil
}

// DeleteCommand scopes deletion to an authenticated workspace actor.
type DeleteCommand struct {
	SprintID    uuid.UUID
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
}

func (command DeleteCommand) Validate() error {
	if command.SprintID == uuid.Nil || command.WorkspaceID == uuid.Nil || command.ActorID == uuid.Nil {
		return fmt.Errorf("%w: sprint, workspace, and actor are required", ErrInvalid)
	}
	return nil
}

func validateName(value string) error {
	if value == "" || utf8.RuneCountInString(value) > MaximumNameLength {
		return fmt.Errorf("%w: name must contain at most %d characters", ErrInvalid, MaximumNameLength)
	}
	return nil
}

func normalizeOptionalGoal(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > MaximumGoalLength {
		return nil, fmt.Errorf("%w: goal cannot exceed %d characters", ErrInvalid, MaximumGoalLength)
	}
	return &trimmed, nil
}

func normalizeDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
