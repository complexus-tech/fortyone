package objectiveshttp

import (
	"testing"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
)

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
