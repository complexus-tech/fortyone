package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalid          = errors.New("invalid objective request")
	ErrForbidden        = errors.New("objective access is forbidden")
	ErrNotFound         = errors.New("objective not found")
	ErrNameExists       = errors.New("an objective with this name already exists")
	ErrVersionConflict  = errors.New("objective changed since it was reviewed")
	ErrInvalidReference = errors.New("objective reference is invalid")
)

type Field[T any] struct {
	specified bool
	value     *T
}

func SetField[T any](value T) Field[T] { return Field[T]{specified: true, value: &value} }
func ClearField[T any]() Field[T]      { return Field[T]{specified: true} }
func (field Field[T]) Specified() bool { return field.specified }
func (field Field[T]) Value() (*T, bool) {
	return field.value, field.specified
}

type ObjectivePatch struct {
	Name         Field[string]
	Description  Field[string]
	ShortSummary Field[string]
	LeadUser     Field[uuid.UUID]
	StartDate    Field[time.Time]
	EndDate      Field[time.Time]
	IsPrivate    Field[bool]
	Status       Field[uuid.UUID]
	Priority     Field[string]
	Health       Field[ObjectiveHealth]
	Color        Field[string]
}

var objectivePatchOrder = []string{
	"name", "description", "short_summary", "lead_user_id", "start_date", "end_date",
	"is_private", "status_id", "priority", "health", "color",
}

func (patch ObjectivePatch) Fields() []string {
	fields := make([]string, 0, len(objectivePatchOrder))
	for _, field := range objectivePatchOrder {
		if patch.specified(field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func (patch ObjectivePatch) Empty() bool { return len(patch.Fields()) == 0 }

func (patch ObjectivePatch) specified(field string) bool {
	switch field {
	case "name":
		return patch.Name.Specified()
	case "description":
		return patch.Description.Specified()
	case "short_summary":
		return patch.ShortSummary.Specified()
	case "lead_user_id":
		return patch.LeadUser.Specified()
	case "start_date":
		return patch.StartDate.Specified()
	case "end_date":
		return patch.EndDate.Specified()
	case "is_private":
		return patch.IsPrivate.Specified()
	case "status_id":
		return patch.Status.Specified()
	case "priority":
		return patch.Priority.Specified()
	case "health":
		return patch.Health.Specified()
	case "color":
		return patch.Color.Specified()
	default:
		return false
	}
}

var objectiveColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

const (
	MaximumObjectiveNameLength     = 255
	MaximumObjectivePriorityLength = 100
	MaximumShortSummaryLength      = 500
	MaximumKeyResultNameLength     = 255
	MaximumKeyResultContributors   = 100
)

func (patch ObjectivePatch) Validate() error {
	if patch.Empty() {
		return fmt.Errorf("%w: at least one field is required", ErrInvalid)
	}
	if value, ok := patch.Name.Value(); ok {
		if value == nil || strings.TrimSpace(*value) == "" {
			return fmt.Errorf("%w: name cannot be blank or null", ErrInvalid)
		}
		if len([]rune(*value)) > MaximumObjectiveNameLength {
			return fmt.Errorf("%w: name cannot exceed %d characters", ErrInvalid, MaximumObjectiveNameLength)
		}
	}
	if value, ok := patch.ShortSummary.Value(); ok && value != nil && len([]rune(*value)) > MaximumShortSummaryLength {
		return fmt.Errorf("%w: short summary cannot exceed %d characters", ErrInvalid, MaximumShortSummaryLength)
	}
	if value, ok := patch.IsPrivate.Value(); ok && value == nil {
		return fmt.Errorf("%w: is private cannot be null", ErrInvalid)
	}
	if value, ok := patch.Status.Value(); ok && (value == nil || *value == uuid.Nil) {
		return fmt.Errorf("%w: status cannot be null", ErrInvalid)
	}
	if value, ok := patch.Color.Value(); ok && (value == nil || !objectiveColorPattern.MatchString(*value)) {
		return fmt.Errorf("%w: color must be a six-digit hexadecimal color", ErrInvalid)
	}
	if value, ok := patch.Health.Value(); ok && value != nil && !slices.Contains(
		[]ObjectiveHealth{HealthAtRisk, HealthOnTrack, HealthOffTrack}, *value,
	) {
		return fmt.Errorf("%w: unsupported health", ErrInvalid)
	}
	if value, ok := patch.Priority.Value(); ok && value != nil && len([]rune(*value)) > MaximumObjectivePriorityLength {
		return fmt.Errorf("%w: priority cannot exceed %d characters", ErrInvalid, MaximumObjectivePriorityLength)
	}
	if start, startSet := patch.StartDate.Value(); startSet && start != nil {
		if end, endSet := patch.EndDate.Value(); endSet && end != nil && end.Before(*start) {
			return fmt.Errorf("%w: end date cannot be before start date", ErrInvalid)
		}
	}
	return nil
}

type NewObjective struct {
	Name         string
	Description  *string
	ShortSummary *string
	LeadUser     *uuid.UUID
	Team         uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	IsPrivate    bool
	Status       uuid.UUID
	Priority     *string
	Color        string
	CreatedBy    uuid.UUID
}

type NewKeyResult struct {
	Name            string
	MeasurementType string
	StartValue      float64
	CurrentValue    float64
	TargetValue     float64
	Lead            *uuid.UUID
	Contributors    []uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
}

type KeyResult struct {
	ID              uuid.UUID
	SequenceID      int
	ObjectiveID     uuid.UUID
	Name            string
	MeasurementType string
	StartValue      float64
	CurrentValue    float64
	TargetValue     float64
	Lead            *uuid.UUID
	Contributors    []uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       uuid.UUID
}

type CreateCommand struct {
	WorkspaceID uuid.UUID
	Objective   NewObjective
	KeyResults  []NewKeyResult
}

func (command CreateCommand) Validate() error {
	objective := command.Objective
	if command.WorkspaceID == uuid.Nil || objective.Team == uuid.Nil || objective.Status == uuid.Nil || objective.CreatedBy == uuid.Nil {
		return fmt.Errorf("%w: workspace, team, status, and creator are required", ErrInvalid)
	}
	if strings.TrimSpace(objective.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalid)
	}
	if len([]rune(objective.Name)) > MaximumObjectiveNameLength {
		return fmt.Errorf("%w: name cannot exceed %d characters", ErrInvalid, MaximumObjectiveNameLength)
	}
	if objective.ShortSummary != nil && len([]rune(*objective.ShortSummary)) > MaximumShortSummaryLength {
		return fmt.Errorf("%w: short summary cannot exceed %d characters", ErrInvalid, MaximumShortSummaryLength)
	}
	if objective.Priority != nil && len([]rune(*objective.Priority)) > MaximumObjectivePriorityLength {
		return fmt.Errorf("%w: priority cannot exceed %d characters", ErrInvalid, MaximumObjectivePriorityLength)
	}
	if objective.Color == "" || !objectiveColorPattern.MatchString(objective.Color) {
		return fmt.Errorf("%w: color must be a six-digit hexadecimal color", ErrInvalid)
	}
	if objective.StartDate != nil && objective.EndDate != nil && objective.EndDate.Before(*objective.StartDate) {
		return fmt.Errorf("%w: end date cannot be before start date", ErrInvalid)
	}
	if len(command.KeyResults) > 20 {
		return fmt.Errorf("%w: no more than 20 key results can be created at once", ErrInvalid)
	}
	for index, keyResult := range command.KeyResults {
		if strings.TrimSpace(keyResult.Name) == "" || !slices.Contains([]string{"percentage", "number", "boolean"}, keyResult.MeasurementType) {
			return fmt.Errorf("%w: key result %d has an invalid name or measurement type", ErrInvalid, index)
		}
		if len([]rune(keyResult.Name)) > MaximumKeyResultNameLength {
			return fmt.Errorf("%w: key result %d name cannot exceed %d characters", ErrInvalid, index, MaximumKeyResultNameLength)
		}
		if len(keyResult.Contributors) > MaximumKeyResultContributors {
			return fmt.Errorf("%w: key result %d cannot have more than %d contributors", ErrInvalid, index, MaximumKeyResultContributors)
		}
		if keyResult.StartDate == nil || keyResult.EndDate == nil || keyResult.EndDate.Before(*keyResult.StartDate) {
			return fmt.Errorf("%w: key result %d requires a valid date range", ErrInvalid, index)
		}
		if keyResult.Lead != nil && *keyResult.Lead == uuid.Nil {
			return fmt.Errorf("%w: key result %d has an invalid lead", ErrInvalid, index)
		}
		for contributorIndex, contributorID := range keyResult.Contributors {
			if contributorID == uuid.Nil {
				return fmt.Errorf("%w: key result %d contributor %d is invalid", ErrInvalid, index, contributorIndex)
			}
		}
	}
	return nil
}

type CreateResult struct {
	Objective  Objective
	KeyResults []KeyResult
}

type UpdateCommand struct {
	ObjectiveID       uuid.UUID
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	Patch             ObjectivePatch
	Comment           string
	ExpectedUpdatedAt *time.Time
}

func (command UpdateCommand) Validate() error {
	if command.ObjectiveID == uuid.Nil || command.WorkspaceID == uuid.Nil || command.ActorID == uuid.Nil {
		return fmt.Errorf("%w: objective, workspace, and actor are required", ErrInvalid)
	}
	return command.Patch.Validate()
}

type DeleteCommand struct {
	ObjectiveID uuid.UUID
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
}

func (command DeleteCommand) Validate() error {
	if command.ObjectiveID == uuid.Nil || command.WorkspaceID == uuid.Nil || command.ActorID == uuid.Nil {
		return fmt.Errorf("%w: objective, workspace, and actor are required", ErrInvalid)
	}
	return nil
}
