package teamshttp

import (
	"encoding/json"
	"testing"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
)

func TestWorkNamingPatchDistinguishesOmissionAndReset(t *testing.T) {
	for _, test := range []struct {
		name    string
		body    string
		present bool
		value   *string
	}{
		{name: "omitted", body: `{"name":"Sales"}`},
		{name: "reset", body: `{"storyTerm":null}`, present: true},
		{name: "deal", body: `{"storyTerm":"deal"}`, present: true, value: stringPointer("deal")},
	} {
		t.Run(test.name, func(t *testing.T) {
			var input AppUpdateTeam
			if err := json.Unmarshal([]byte(test.body), &input); err != nil {
				t.Fatal(err)
			}
			if input.StoryTerm.Present != test.present {
				t.Fatalf("present = %v, want %v", input.StoryTerm.Present, test.present)
			}
			if test.value == nil {
				if input.StoryTerm.Value != nil {
					t.Fatal("expected workspace fallback")
				}
			} else if input.StoryTerm.Value == nil || *input.StoryTerm.Value != *test.value {
				t.Fatalf("unexpected story term: %v", input.StoryTerm.Value)
			}
		})
	}
}

func TestWorkNamingRejectsInvalidJSONTypesAndTerms(t *testing.T) {
	for _, body := range []string{`{"storyTerm":true}`, `{"storyTerm":4}`, `{"storyTerm":[]}`} {
		var input AppUpdateTeam
		if err := json.Unmarshal([]byte(body), &input); err == nil {
			t.Fatalf("accepted invalid naming body %s", body)
		}
	}
	for _, term := range []string{"deal", "task", "story", "issue", "ticket", "work item"} {
		if !teamsdomain.ValidStoryTerm(&term) {
			t.Fatalf("rejected supported term %q", term)
		}
	}
	for _, term := range []string{"", "deals", "Deal", " arbitrary ", "<script>"} {
		if teamsdomain.ValidStoryTerm(&term) {
			t.Fatalf("accepted unsupported term %q", term)
		}
	}
	if !teamsdomain.ValidStoryTerm(nil) {
		t.Fatal("workspace fallback must be valid")
	}
}

func stringPointer(value string) *string { return &value }
