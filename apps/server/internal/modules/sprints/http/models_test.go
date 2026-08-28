package sprintshttp

import (
	"encoding/json"
	"errors"
	"testing"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
)

func TestAppUpdateSprintPreservesOmittedValueAndNull(t *testing.T) {
	t.Parallel()

	var request AppUpdateSprint
	if err := json.Unmarshal([]byte(`{
		"name":"  Planning cycle  ",
		"goal":null,
		"objectiveId":null,
		"startDate":"2026-08-24"
	}`), &request); err != nil {
		t.Fatalf("decode sprint patch: %v", err)
	}
	patch, err := request.SprintPatch().Normalize()
	if err != nil {
		t.Fatalf("normalize sprint patch: %v", err)
	}
	if value, specified := patch.Name.Value(); !specified || value == nil || *value != "Planning cycle" {
		t.Fatalf("name patch = (%v, %t)", value, specified)
	}
	if value, specified := patch.Goal.Value(); !specified || value != nil {
		t.Fatalf("goal patch = (%v, %t), want explicit clear", value, specified)
	}
	if value, specified := patch.ObjectiveID.Value(); !specified || value != nil {
		t.Fatalf("objective patch = (%v, %t), want explicit clear", value, specified)
	}
	if _, specified := patch.EndDate.Value(); specified {
		t.Fatal("omitted end date was marked specified")
	}
}

func TestAppUpdateSprintRejectsNullRequiredFieldsAndEmptyPatches(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"null name": `{ "name": null }`,
		"null date": `{ "startDate": null }`,
		"empty":     `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			var request AppUpdateSprint
			if err := json.Unmarshal([]byte(body), &request); err != nil {
				t.Fatalf("decode sprint patch: %v", err)
			}
			if err := request.Validate(); !errors.Is(err, sprintdomain.ErrInvalid) {
				t.Fatalf("validation error = %v, want ErrInvalid", err)
			}
		})
	}
}
