package domain_test

import (
	"errors"
	"testing"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
)

func TestPatchPreservesClearAndNormalizesValues(t *testing.T) {
	t.Parallel()

	patch, err := (sprintdomain.Patch{
		Name: platformpatch.Set("  Planning  "),
		Goal: platformpatch.Clear[string](),
	}).Normalize()
	if err != nil {
		t.Fatalf("normalize patch: %v", err)
	}
	name, _ := patch.Name.Value()
	goal, goalSet := patch.Goal.Value()
	if name == nil || *name != "Planning" || !goalSet || goal != nil {
		t.Fatalf("normalized patch = name %v, goal (%v, %t)", name, goal, goalSet)
	}
}

func TestNewSprintRejectsCrossedDates(t *testing.T) {
	t.Parallel()

	_, err := (sprintdomain.NewSprint{
		Name:   "Sprint",
		TeamID: uuid.New(), WorkspaceID: uuid.New(),
		StartDate: time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}).Normalize()
	if !errors.Is(err, sprintdomain.ErrInvalid) {
		t.Fatalf("error = %v, want invalid", err)
	}
}

func TestListFilterIsBounded(t *testing.T) {
	t.Parallel()

	filter, err := (sprintdomain.ListFilter{}).Normalize()
	if err != nil {
		t.Fatalf("normalize list filter: %v", err)
	}
	if filter.Limit != sprintdomain.DefaultListLimit {
		t.Fatalf("limit = %d, want %d", filter.Limit, sprintdomain.DefaultListLimit)
	}
	if _, err := (sprintdomain.ListFilter{Limit: sprintdomain.MaximumListLimit + 1}).Normalize(); !errors.Is(err, sprintdomain.ErrInvalid) {
		t.Fatalf("unbounded error = %v, want invalid", err)
	}
}
