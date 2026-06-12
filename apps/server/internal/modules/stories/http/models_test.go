package storieshttp

import (
	"net/http/httptest"
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

func TestParseStoryQueryMapsEstimateValues(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"/stories?estimateValues=1,5&estimateValues=8",
		nil,
	)

	query, err := parseStoryQuery(request, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("expected query to parse, got error: %v", err)
	}

	expected := []int16{1, 5, 8}
	if len(query.Filters.EstimateValues) != len(expected) {
		t.Fatalf("expected %d estimates, got %d", len(expected), len(query.Filters.EstimateValues))
	}

	for i, estimateValue := range expected {
		if query.Filters.EstimateValues[i] != estimateValue {
			t.Fatalf("expected estimate %d to be %d, got %d", i, estimateValue, query.Filters.EstimateValues[i])
		}
	}
}
