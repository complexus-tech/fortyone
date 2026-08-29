package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Field distinguishes an omitted update from an explicit zero or null value.
// It is intentionally local to key results because validation and activity
// semantics remain domain-specific even though the representation is generic.
type Field[T any] struct {
	Value T
	Set   bool
}

func SetField[T any](value T) Field[T] {
	return Field[T]{Value: value, Set: true}
}

func ClearField[T any]() Field[*T] {
	return Field[*T]{Set: true}
}

type Patch struct {
	Name            Field[string]
	MeasurementType Field[string]
	StartValue      Field[float64]
	CurrentValue    Field[float64]
	TargetValue     Field[float64]
	Lead            Field[*uuid.UUID]
	Contributors    Field[[]uuid.UUID]
	StartDate       Field[*time.Time]
	EndDate         Field[*time.Time]
}

func (patch Patch) Empty() bool {
	return !patch.Name.Set && !patch.MeasurementType.Set && !patch.StartValue.Set &&
		!patch.CurrentValue.Set && !patch.TargetValue.Set && !patch.Lead.Set &&
		!patch.Contributors.Set && !patch.StartDate.Set && !patch.EndDate.Set
}

func (patch Patch) Normalize() (Patch, error) {
	if patch.Empty() {
		return Patch{}, fmt.Errorf("%w: at least one update is required", ErrInvalid)
	}
	if patch.Name.Set {
		patch.Name.Value = strings.TrimSpace(patch.Name.Value)
		if patch.Name.Value == "" || len([]rune(patch.Name.Value)) > 255 {
			return Patch{}, fmt.Errorf("%w: name must contain at most 255 characters", ErrInvalid)
		}
	}
	if patch.MeasurementType.Set {
		measurement := MeasurementType(strings.TrimSpace(patch.MeasurementType.Value))
		if !measurement.Valid() {
			return Patch{}, fmt.Errorf("%w: unsupported measurement type", ErrInvalid)
		}
		patch.MeasurementType.Value = string(measurement)
	}
	for _, value := range []Field[float64]{patch.StartValue, patch.CurrentValue, patch.TargetValue} {
		if value.Set && !finite(value.Value) {
			return Patch{}, fmt.Errorf("%w: measurement values must be finite", ErrInvalid)
		}
	}
	if patch.Lead.Set && patch.Lead.Value != nil && *patch.Lead.Value == uuid.Nil {
		return Patch{}, fmt.Errorf("%w: lead cannot be a zero id", ErrInvalid)
	}
	if patch.Contributors.Set {
		contributors, err := normalizeUUIDs(patch.Contributors.Value)
		if err != nil {
			return Patch{}, err
		}
		patch.Contributors.Value = contributors
	}
	for _, field := range []*Field[*time.Time]{&patch.StartDate, &patch.EndDate} {
		if !field.Set {
			continue
		}
		if field.Value == nil {
			return Patch{}, fmt.Errorf("%w: key result dates cannot be cleared", ErrInvalid)
		}
		value := normalizeDate(*field.Value)
		field.Value = &value
	}
	return patch, nil
}

type AccessScope struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	AllTeams    bool
	TeamIDs     []uuid.UUID
}

func (scope AccessScope) Validate() error {
	if scope.WorkspaceID == uuid.Nil || scope.ActorID == uuid.Nil {
		return fmt.Errorf("%w: workspace and actor are required", ErrInvalid)
	}
	if scope.AllTeams && len(scope.TeamIDs) > 0 {
		return fmt.Errorf("%w: team access cannot be both unrestricted and restricted", ErrInvalid)
	}
	if !scope.AllTeams && len(scope.TeamIDs) == 0 {
		return fmt.Errorf("%w: restricted access requires at least one team", ErrForbidden)
	}
	for _, teamID := range scope.TeamIDs {
		if teamID == uuid.Nil {
			return fmt.Errorf("%w: team access contains a zero id", ErrInvalid)
		}
	}
	return nil
}

type CreateCommand struct {
	Access     AccessScope
	KeyResults []NewKeyResult
}

func (command CreateCommand) Normalize() (CreateCommand, error) {
	if err := command.Access.Validate(); err != nil {
		return CreateCommand{}, err
	}
	if len(command.KeyResults) == 0 || len(command.KeyResults) > MaximumBatchSize {
		return CreateCommand{}, fmt.Errorf("%w: key result batch size must be between 1 and %d", ErrInvalid, MaximumBatchSize)
	}
	objectiveID := command.KeyResults[0].ObjectiveID
	for index := range command.KeyResults {
		normalized, err := command.KeyResults[index].Normalize()
		if err != nil {
			return CreateCommand{}, fmt.Errorf("key result %d: %w", index, err)
		}
		if normalized.ObjectiveID != objectiveID || normalized.CreatedBy != command.Access.ActorID {
			return CreateCommand{}, fmt.Errorf("%w: a batch must use one objective and the authenticated actor", ErrInvalid)
		}
		command.KeyResults[index] = normalized
	}
	return command, nil
}

type UpdateCommand struct {
	Access            AccessScope
	KeyResultID       uuid.UUID
	Patch             Patch
	Comment           string
	ExpectedUpdatedAt *time.Time
}

func (command UpdateCommand) Normalize() (UpdateCommand, error) {
	if err := command.Access.Validate(); err != nil {
		return UpdateCommand{}, err
	}
	if command.KeyResultID == uuid.Nil {
		return UpdateCommand{}, fmt.Errorf("%w: key result id is required", ErrInvalid)
	}
	patch, err := command.Patch.Normalize()
	if err != nil {
		return UpdateCommand{}, err
	}
	command.Patch = patch
	command.Comment = strings.TrimSpace(command.Comment)
	if len([]rune(command.Comment)) > 10_000 {
		return UpdateCommand{}, fmt.Errorf("%w: comment cannot exceed 10000 characters", ErrInvalid)
	}
	if command.ExpectedUpdatedAt != nil {
		if command.ExpectedUpdatedAt.IsZero() {
			return UpdateCommand{}, fmt.Errorf("%w: expected update time is required", ErrInvalid)
		}
		expected := command.ExpectedUpdatedAt.UTC()
		command.ExpectedUpdatedAt = &expected
	}
	return command, nil
}

type MutationResult struct {
	Before        KeyResult
	After         KeyResult
	ChangedFields []string
}

type DeleteCommand struct {
	Access      AccessScope
	KeyResultID uuid.UUID
}

func (command DeleteCommand) Validate() error {
	if err := command.Access.Validate(); err != nil {
		return err
	}
	if command.KeyResultID == uuid.Nil {
		return fmt.Errorf("%w: key result id is required", ErrInvalid)
	}
	return nil
}
