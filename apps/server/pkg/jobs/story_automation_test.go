package jobs

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildAutoCloseActivityRecordsIncludesReason(t *testing.T) {
	systemUserID := uuid.New()
	story := ClosedStory{
		ID:          uuid.New(),
		WorkspaceID: uuid.New(),
		StatusID:    uuid.New(),
	}

	records := buildAutoCloseActivityRecords([]ClosedStory{story}, systemUserID)

	if len(records) != 1 {
		t.Fatalf("expected one record, got %d", len(records))
	}
	if records[0].Reason == "" {
		t.Fatal("expected auto-close activity reason")
	}
	if records[0].Reason != storyAutoCloseActivityReason {
		t.Fatalf("expected reason %q, got %q", storyAutoCloseActivityReason, records[0].Reason)
	}
}

func TestBuildSprintMigrationActivityRecordsIncludesReason(t *testing.T) {
	systemUserID := uuid.New()
	story := MigratedStory{
		ID:               uuid.New(),
		WorkspaceID:      uuid.New(),
		PreviousSprintID: uuid.New(),
		NewSprintID:      uuid.New(),
	}

	records := buildSprintMigrationActivityRecords([]MigratedStory{story}, systemUserID)

	if len(records) != 1 {
		t.Fatalf("expected one record, got %d", len(records))
	}
	if records[0].Reason == "" {
		t.Fatal("expected sprint migration activity reason")
	}
	if records[0].Reason != sprintMigrationActivityReason {
		t.Fatalf("expected reason %q, got %q", sprintMigrationActivityReason, records[0].Reason)
	}
}
