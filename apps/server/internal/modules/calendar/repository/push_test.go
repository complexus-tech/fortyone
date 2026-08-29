package calendarrepository

import (
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	"github.com/google/uuid"
)

func TestManagedScheduleEventCanonicalState(t *testing.T) {
	t.Parallel()
	blockID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	block := managedScheduleBlockState{
		ID: blockID, WorkspaceID: workspaceID, StoryID: &storyID,
		Title: "Private focus", StartAt: startAt, EndAt: startAt.Add(time.Hour),
	}
	canonical := calendar.ManagedScheduleEventChange{
		EventID: calendar.StableGoogleScheduleEventID(blockID), Title: block.Title,
		StartAt: block.StartAt, EndAt: block.EndAt, Visibility: "private", Transparency: "opaque", Status: "confirmed",
		Source: "maya_schedule", BlockID: blockID.String(), StoryID: storyID.String(), WorkspaceID: workspaceID.String(),
	}
	if !managedScheduleEventMatchesBlock(canonical, block) {
		t.Fatal("an equal canonical FortyOne self-write must not invalidate its sync hash")
	}

	for name, mutate := range map[string]func(*calendar.ManagedScheduleEventChange){
		"public":             func(change *calendar.ManagedScheduleEventChange) { change.Visibility = "public" },
		"transparent":        func(change *calendar.ManagedScheduleEventChange) { change.Transparency = "transparent" },
		"missing provenance": func(change *calendar.ManagedScheduleEventChange) { change.BlockID = "" },
		"attendees":          func(change *calendar.ManagedScheduleEventChange) { change.HasAttendees = true },
		"recurrence":         func(change *calendar.ManagedScheduleEventChange) { change.Recurring = true },
		"moved":              func(change *calendar.ManagedScheduleEventChange) { change.StartAt = change.StartAt.Add(time.Hour) },
	} {
		t.Run(name, func(t *testing.T) {
			change := canonical
			mutate(&change)
			if managedScheduleEventMatchesBlock(change, block) {
				t.Fatalf("%s managed event drift must invalidate the provider sync hash", name)
			}
		})
	}
}
