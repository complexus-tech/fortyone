package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestFilterStoriesAssignedByResolvesExactAndFirstNameMatches(t *testing.T) {
	t.Parallel()

	dariaID := uuid.New()
	otherID := uuid.New()
	stories := []StoryList{
		{ID: uuid.New(), Title: "Daria task", AssignedBy: &AssignmentActor{ID: dariaID, Username: "daria", FullName: "Daria Jones"}},
		{ID: uuid.New(), Title: "Other task", AssignedBy: &AssignmentActor{ID: otherID, Username: "sam", FullName: "Sam Lee"}},
		{ID: uuid.New(), Title: "Unknown attribution"},
	}

	for _, query := range []string{"DARIA", "Daria Jones", " daria "} {
		result := FilterStoriesAssignedBy(stories, query)
		if result.Resolved == nil || result.Resolved.ID != dariaID {
			t.Fatalf("query %q resolved actor = %#v", query, result.Resolved)
		}
		if len(result.Stories) != 1 || result.Stories[0].Title != "Daria task" {
			t.Fatalf("query %q stories = %#v", query, result.Stories)
		}
	}
}

func TestFilterStoriesAssignedByReturnsAmbiguousCandidates(t *testing.T) {
	t.Parallel()

	stories := []StoryList{
		{ID: uuid.New(), AssignedBy: &AssignmentActor{ID: uuid.New(), Username: "daria.j", FullName: "Daria Jones"}},
		{ID: uuid.New(), AssignedBy: &AssignmentActor{ID: uuid.New(), Username: "daria.m", FullName: "Daria Morgan"}},
	}

	result := FilterStoriesAssignedBy(stories, "Daria")
	if result.Resolved != nil || len(result.Stories) != 0 || len(result.Candidates) != 2 {
		t.Fatalf("ambiguous result = %#v", result)
	}
}

func TestFilterStoriesAssignedByDoesNotInventUnknownAttribution(t *testing.T) {
	t.Parallel()

	stories := []StoryList{{ID: uuid.New(), Title: "Legacy task"}}
	result := FilterStoriesAssignedBy(stories, "Daria")
	if result.Resolved != nil || len(result.Stories) != 0 || len(result.Candidates) != 0 {
		t.Fatalf("unknown attribution result = %#v", result)
	}
}
