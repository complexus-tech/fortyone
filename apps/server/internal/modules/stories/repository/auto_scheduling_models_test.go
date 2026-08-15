package storiesrepository

import (
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

func TestStoryAutoSchedulingModelRoundTrip(t *testing.T) {
	t.Parallel()

	reason := "Waiting for an owner"
	updatedAt := time.Date(2026, time.August, 15, 9, 30, 0, 0, time.UTC)
	databaseStory := toDBStory(stories.CoreSingleStory{
		AutoSchedulingEnabled:   true,
		AutoSchedulingLocked:    true,
		AutoSchedulingStatus:    stories.AutoSchedulingStatusNeedsOwner,
		AutoSchedulingReason:    &reason,
		AutoSchedulingUpdatedAt: &updatedAt,
	})

	coreStory := toCoreStory(databaseStory)
	if !coreStory.AutoSchedulingEnabled || !coreStory.AutoSchedulingLocked {
		t.Fatalf("auto-scheduling preferences were not preserved: %#v", coreStory)
	}
	if coreStory.AutoSchedulingStatus != stories.AutoSchedulingStatusNeedsOwner {
		t.Fatalf("auto-scheduling status = %q, want %q", coreStory.AutoSchedulingStatus, stories.AutoSchedulingStatusNeedsOwner)
	}
	if coreStory.AutoSchedulingReason != &reason || coreStory.AutoSchedulingUpdatedAt != &updatedAt {
		t.Fatalf("auto-scheduling metadata pointers were not preserved: %#v", coreStory)
	}
}
