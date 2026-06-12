package storieshttp

import (
	"testing"

	"github.com/google/uuid"
)

func TestToCoreNewStoryMapsLabelIDs(t *testing.T) {
	userID := uuid.New()
	labelIDs := []uuid.UUID{uuid.New(), uuid.New()}

	coreStory := toCoreNewStory(AppNewStory{
		Title:    "Add reporting filters",
		LabelIDs: labelIDs,
		Team:     uuid.New(),
		Priority: "High",
	}, userID)

	if len(coreStory.LabelIDs) != len(labelIDs) {
		t.Fatalf("expected %d labels, got %d", len(labelIDs), len(coreStory.LabelIDs))
	}

	for i, labelID := range labelIDs {
		if coreStory.LabelIDs[i] != labelID {
			t.Fatalf("expected label %d to be %s, got %s", i, labelID, coreStory.LabelIDs[i])
		}
	}
}
