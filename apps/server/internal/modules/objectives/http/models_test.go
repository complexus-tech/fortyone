package objectiveshttp

import (
	"testing"
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

func dateValue(value string) *date.Date {
	parsed := date.Date(time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC))
	if value == "end" {
		parsed = date.Date(time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC))
	}
	return &parsed
}

func TestObjectiveShortSummaryMappings(t *testing.T) {
	t.Parallel()

	shortSummary := "Launch a reliable first version for early customers."
	createdBy := uuid.New()

	coreNewObjective := toCoreNewObjective(AppNewObjective{
		Name:         "Launch MVP",
		ShortSummary: &shortSummary,
	}, createdBy)
	if coreNewObjective.ShortSummary == nil || *coreNewObjective.ShortSummary != shortSummary {
		t.Fatalf("expected create mapping to preserve short summary, got %#v", coreNewObjective.ShortSummary)
	}

	appObjective := toAppObjective(objectives.CoreObjective{
		ID:           uuid.New(),
		Name:         "Launch MVP",
		ShortSummary: &shortSummary,
		CreatedBy:    createdBy,
	})
	if appObjective.ShortSummary == nil || *appObjective.ShortSummary != shortSummary {
		t.Fatalf("expected response mapping to preserve short summary, got %#v", appObjective.ShortSummary)
	}
}

func TestObjectiveSequenceMapping(t *testing.T) {
	t.Parallel()

	const sequenceID = 42
	appObjective := toAppObjective(objectives.CoreObjective{
		ID:         uuid.New(),
		SequenceID: sequenceID,
		Name:       "Launch MVP",
	})

	if appObjective.SequenceID != sequenceID {
		t.Fatalf("expected response mapping to preserve sequence ID %d, got %d", sequenceID, appObjective.SequenceID)
	}
}

func TestObjectiveForecastMapping(t *testing.T) {
	t.Parallel()

	forecast := time.Date(2026, time.June, 29, 0, 0, 0, 0, time.UTC)
	storyID := uuid.New()
	appObjective := toAppObjective(objectives.CoreObjective{
		ID:                uuid.New(),
		Name:              "Launch MVP",
		Color:             "#686DE0",
		ForecastEndDate:   &forecast,
		ScheduleStatus:    objectives.ScheduleStatusAtRisk,
		ForecastDaysDelta: 4,
		ForecastCauseStory: &objectives.CoreForecastCauseStory{
			ID:         storyID,
			SequenceID: 17,
			Title:      "Complete billing integration",
			Source:     "calendar",
		},
	})

	if appObjective.Color != "#686DE0" || appObjective.ForecastEndDate == nil {
		t.Fatalf("expected color and forecast to be mapped, got %#v", appObjective)
	}
	if appObjective.ScheduleStatus != "at_risk" || appObjective.ForecastDaysDelta != 4 {
		t.Fatalf("expected at-risk forecast metadata, got %#v", appObjective)
	}
	if appObjective.ForecastCauseStory == nil || appObjective.ForecastCauseStory.ID != storyID {
		t.Fatalf("expected forecast cause story to be mapped, got %#v", appObjective.ForecastCauseStory)
	}
}

func TestCreateKeyResultsRequestValidation(t *testing.T) {
	t.Parallel()

	valid := AppNewKeyResult{
		Name:            "Reduce page load time from 10s to 1s",
		MeasurementType: "number",
		StartValue:      10,
		CurrentValue:    10,
		TargetValue:     1,
		StartDate:       dateValue("start"),
		EndDate:         dateValue("end"),
	}
	if err := (AppCreateKeyResultsRequest{KeyResults: []AppNewKeyResult{valid}}).Validate(); err != nil {
		t.Fatalf("expected valid batch, got %v", err)
	}

	invalid := valid
	invalid.EndDate = nil
	if err := (AppCreateKeyResultsRequest{KeyResults: []AppNewKeyResult{invalid}}).Validate(); err == nil {
		t.Fatal("expected missing end date to fail validation")
	}
}
