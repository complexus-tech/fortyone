package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestObjectivePatchPreservesStableFieldOrder(t *testing.T) {
	t.Parallel()

	patch := ObjectivePatch{
		Color: SetField("#123456"), Name: SetField("Customer trust"),
		Description: ClearField[string](), Health: SetField(HealthOnTrack),
	}
	want := []string{"name", "description", "health", "color"}
	got := patch.Fields()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Fields() = %v, want %v", got, want)
	}
}

func TestObjectivePatchValidation(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, -1)
	tests := []struct {
		name  string
		patch ObjectivePatch
	}{
		{name: "empty", patch: ObjectivePatch{}},
		{name: "null name", patch: ObjectivePatch{Name: ClearField[string]()}},
		{name: "blank name", patch: ObjectivePatch{Name: SetField("  ")}},
		{name: "long name", patch: ObjectivePatch{Name: SetField(strings.Repeat("a", MaximumObjectiveNameLength+1))}},
		{name: "long summary", patch: ObjectivePatch{ShortSummary: SetField(strings.Repeat("a", 501))}},
		{name: "long priority", patch: ObjectivePatch{Priority: SetField(strings.Repeat("a", MaximumObjectivePriorityLength+1))}},
		{name: "null privacy", patch: ObjectivePatch{IsPrivate: ClearField[bool]()}},
		{name: "zero status", patch: ObjectivePatch{Status: SetField(uuid.Nil)}},
		{name: "invalid health", patch: ObjectivePatch{Health: SetField(ObjectiveHealth("unknown"))}},
		{name: "invalid color", patch: ObjectivePatch{Color: SetField("purple")}},
		{name: "reversed dates", patch: ObjectivePatch{StartDate: SetField(start), EndDate: SetField(end)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.patch.Validate(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Validate() error = %v, want ErrInvalid", err)
			}
		})
	}

	valid := ObjectivePatch{
		Name: SetField("Customer trust"), Description: ClearField[string](),
		Color: SetField("#123ABC"), Health: SetField(HealthOnTrack),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid patch rejected: %v", err)
	}
}

func TestCreateCommandValidatesAggregateInvariants(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	command := CreateCommand{
		WorkspaceID: uuid.New(),
		Objective: NewObjective{
			Name: "Launch", Team: uuid.New(), Status: uuid.New(), CreatedBy: uuid.New(),
			Color: "#686DE0", StartDate: &start, EndDate: &end,
		},
		KeyResults: []NewKeyResult{{
			Name: "Adoption", MeasurementType: "percentage", StartDate: &start, EndDate: &end,
		}},
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("valid create command rejected: %v", err)
	}

	command.KeyResults[0].EndDate = &start
	command.KeyResults[0].StartDate = &end
	if err := command.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid key-result dates error = %v, want ErrInvalid", err)
	}

	command.KeyResults[0].StartDate = &start
	command.KeyResults[0].EndDate = &end
	command.KeyResults[0].Contributors = []uuid.UUID{uuid.Nil}
	if err := command.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("zero contributor error = %v, want ErrInvalid", err)
	}

	command.KeyResults[0].Contributors = make([]uuid.UUID, MaximumKeyResultContributors+1)
	for index := range command.KeyResults[0].Contributors {
		command.KeyResults[0].Contributors[index] = uuid.New()
	}
	if err := command.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("too many contributors error = %v, want ErrInvalid", err)
	}
}

func TestListQueryNormalizeBoundsUntrustedInput(t *testing.T) {
	t.Parallel()

	query, err := (ListQuery{
		WorkspaceID: uuid.New(), ActorID: uuid.New(), Search: "  roadmap  ", Limit: 10_000,
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if query.Search != "roadmap" || query.Limit != MaximumPageSize+1 {
		t.Fatalf("Normalize() = %#v, want trimmed search and bounded lookahead", query)
	}

	_, err = (ListQuery{WorkspaceID: uuid.New(), ActorID: uuid.New(), Offset: -1}).Normalize()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("negative offset error = %v, want ErrInvalid", err)
	}
}

func TestStrategicPillarPatchRequiresIntent(t *testing.T) {
	t.Parallel()

	if err := (UpdateStrategicPillar{}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty patch error = %v, want ErrInvalid", err)
	}
	if err := (UpdateStrategicPillar{Name: ClearField[string](), OrderIndex: ClearField[int]()}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("null non-nullable fields error = %v, want ErrInvalid", err)
	}
	if err := (UpdateStrategicPillar{Description: ClearField[string]()}).Validate(); err != nil {
		t.Fatalf("clear description rejected: %v", err)
	}
	if err := (UpdateStrategicPillar{Name: SetField(strings.Repeat("a", MaximumStrategyNameLength+1))}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("long pillar name error = %v, want ErrInvalid", err)
	}
}

func TestObjectiveQueriesAndDeleteRequireTenantActorIdentity(t *testing.T) {
	t.Parallel()

	if err := (GetQuery{ObjectiveID: uuid.New(), WorkspaceID: uuid.New()}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("actorless get error = %v, want ErrInvalid", err)
	}
	if err := (GetQuery{ObjectiveID: uuid.New(), WorkspaceID: uuid.New(), Internal: true}).Validate(); err != nil {
		t.Fatalf("trusted internal get rejected: %v", err)
	}
	if err := (AnalyticsQuery{ObjectiveID: uuid.New(), WorkspaceID: uuid.New()}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("actorless analytics error = %v, want ErrInvalid", err)
	}
	if err := (DeleteCommand{ObjectiveID: uuid.New(), WorkspaceID: uuid.New()}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("actorless delete error = %v, want ErrInvalid", err)
	}
	if err := (StrategyQuery{WorkspaceID: uuid.New()}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("actorless strategy error = %v, want ErrInvalid", err)
	}
}

func TestStrategyInputsRespectDatabaseLimits(t *testing.T) {
	t.Parallel()

	if err := (StrategyUpdate{UltimateGoal: strings.Repeat("a", MaximumStrategyNameLength+1)}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("long ultimate goal error = %v, want ErrInvalid", err)
	}
	if err := (NewStrategicPillar{Name: strings.Repeat("a", MaximumStrategyNameLength+1)}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("long pillar name error = %v, want ErrInvalid", err)
	}
}
