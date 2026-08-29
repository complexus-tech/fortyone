package domain

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewKeyResultNormalize(t *testing.T) {
	t.Parallel()

	objectiveID, actorID, leadID, contributorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	start := time.Date(2026, time.August, 28, 15, 30, 0, 0, time.FixedZone("CAT", 2*60*60))
	end := start.AddDate(0, 0, 2)
	draft := NewKeyResult{
		ObjectiveID: objectiveID, CreatedBy: actorID, Name: "  Ship API  ",
		MeasurementType: "number", StartValue: 1, CurrentValue: 2, TargetValue: 3,
		Lead: &leadID, Contributors: []uuid.UUID{contributorID, contributorID},
		StartDate: &start, EndDate: &end,
	}

	normalized, err := draft.Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if normalized.Name != "Ship API" || normalized.MeasurementType != "number" {
		t.Fatalf("Normalize() text = %q/%q", normalized.Name, normalized.MeasurementType)
	}
	if len(normalized.Contributors) != 1 || normalized.Contributors[0] != contributorID {
		t.Fatalf("Normalize() contributors = %#v", normalized.Contributors)
	}
	wantStart := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	if normalized.StartDate == nil || !normalized.StartDate.Equal(wantStart) {
		t.Fatalf("Normalize() start = %v, want %v", normalized.StartDate, wantStart)
	}
}

func TestNewKeyResultNormalizeRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	valid := validNewKeyResult()
	tests := []struct {
		name   string
		mutate func(*NewKeyResult)
	}{
		{name: "missing objective", mutate: func(value *NewKeyResult) { value.ObjectiveID = uuid.Nil }},
		{name: "missing actor", mutate: func(value *NewKeyResult) { value.CreatedBy = uuid.Nil }},
		{name: "blank name", mutate: func(value *NewKeyResult) { value.Name = "  " }},
		{name: "unsupported measurement", mutate: func(value *NewKeyResult) { value.MeasurementType = "currency" }},
		{name: "nan measurement", mutate: func(value *NewKeyResult) { value.CurrentValue = math.NaN() }},
		{name: "infinite measurement", mutate: func(value *NewKeyResult) { value.TargetValue = math.Inf(1) }},
		{name: "missing start", mutate: func(value *NewKeyResult) { value.StartDate = nil }},
		{name: "missing end", mutate: func(value *NewKeyResult) { value.EndDate = nil }},
		{name: "reversed dates", mutate: func(value *NewKeyResult) { value.StartDate, value.EndDate = value.EndDate, value.StartDate }},
		{name: "zero lead", mutate: func(value *NewKeyResult) { id := uuid.Nil; value.Lead = &id }},
		{name: "zero contributor", mutate: func(value *NewKeyResult) { value.Contributors = []uuid.UUID{uuid.Nil} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			value := valid
			test.mutate(&value)
			if _, err := value.Normalize(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Normalize() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestPatchNormalizePreservesTypedUpdateIntent(t *testing.T) {
	t.Parallel()

	contributorID := uuid.New()
	patch, err := (Patch{
		CurrentValue: SetField(0.0),
		Lead:         ClearField[uuid.UUID](),
		Contributors: SetField([]uuid.UUID{contributorID, contributorID}),
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if !patch.CurrentValue.Set || patch.CurrentValue.Value != 0 {
		t.Fatalf("current-value intent = %#v", patch.CurrentValue)
	}
	if !patch.Lead.Set || patch.Lead.Value != nil {
		t.Fatalf("lead clear intent = %#v", patch.Lead)
	}
	if len(patch.Contributors.Value) != 1 || patch.Contributors.Value[0] != contributorID {
		t.Fatalf("contributors = %#v", patch.Contributors.Value)
	}
}

func TestPatchNormalizeRejectsInvalidIntent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		patch Patch
	}{
		{name: "empty", patch: Patch{}},
		{name: "blank name", patch: Patch{Name: SetField(" ")}},
		{name: "unsupported measurement", patch: Patch{MeasurementType: SetField("currency")}},
		{name: "nan", patch: Patch{CurrentValue: SetField(math.NaN())}},
		{name: "zero lead", patch: Patch{Lead: SetField(uuidPointerForTest(uuid.Nil))}},
		{name: "zero contributor", patch: Patch{Contributors: SetField([]uuid.UUID{uuid.Nil})}},
		{name: "clear start", patch: Patch{StartDate: ClearField[time.Time]()}},
		{name: "clear end", patch: Patch{EndDate: ClearField[time.Time]()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := test.patch.Normalize(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Normalize() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestAccessScopeValidateFailsClosed(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, teamID := uuid.New(), uuid.New(), uuid.New()
	if err := (AccessScope{WorkspaceID: workspaceID, ActorID: actorID, AllTeams: true}).Validate(); err != nil {
		t.Fatalf("unrestricted scope error = %v", err)
	}
	if err := (AccessScope{WorkspaceID: workspaceID, ActorID: actorID, TeamIDs: []uuid.UUID{teamID}}).Validate(); err != nil {
		t.Fatalf("restricted scope error = %v", err)
	}
	if err := (AccessScope{WorkspaceID: workspaceID, ActorID: actorID}).Validate(); !errors.Is(err, ErrForbidden) {
		t.Fatalf("empty restricted scope error = %v, want ErrForbidden", err)
	}
	if err := (AccessScope{WorkspaceID: workspaceID, ActorID: actorID, AllTeams: true, TeamIDs: []uuid.UUID{teamID}}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ambiguous scope error = %v, want ErrInvalid", err)
	}
}

func validNewKeyResult() NewKeyResult {
	start := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	return NewKeyResult{
		ObjectiveID: uuid.New(), CreatedBy: uuid.New(), Name: "Ship API",
		MeasurementType: "percentage", TargetValue: 100,
		StartDate: &start, EndDate: &end,
	}
}

func uuidPointerForTest(value uuid.UUID) *uuid.UUID {
	return &value
}
