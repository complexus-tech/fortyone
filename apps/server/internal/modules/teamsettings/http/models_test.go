package teamsettingshttp

import "testing"

func TestSprintPatchMappingPreservesExplicitZeroAndFalse(t *testing.T) {
	t.Parallel()

	disabled := false
	zero := 0
	workingDays := []int{1, 2, 3, 4, 5}
	updates := toCoreUpdateTeamSprintSettings(AppUpdateTeamSprintSettings{
		AutoCreateSprints:            &disabled,
		UpcomingSprintsCount:         &zero,
		WorkingDays:                  &workingDays,
		MoveIncompleteStoriesEnabled: &disabled,
	})

	if !updates.AutoCreateSprints.Present || updates.AutoCreateSprints.Value {
		t.Fatalf("auto-create patch = %#v", updates.AutoCreateSprints)
	}
	if !updates.UpcomingSprintsCount.Present || updates.UpcomingSprintsCount.Value != 0 {
		t.Fatalf("upcoming-count patch = %#v", updates.UpcomingSprintsCount)
	}
	if !updates.MoveIncompleteStoriesEnabled.Present || updates.MoveIncompleteStoriesEnabled.Value {
		t.Fatalf("move-incomplete patch = %#v", updates.MoveIncompleteStoriesEnabled)
	}
	if !updates.WorkingDays.Present || len(updates.WorkingDays.Value) != len(workingDays) {
		t.Fatalf("working-days patch = %#v", updates.WorkingDays)
	}
	if updates.SprintDurationWeeks.Present || updates.SprintStartDay.Present || updates.NextAutoSprintNumber.Present {
		t.Fatal("omitted sprint fields were marked present")
	}
}

func TestEstimationPatchMappingPreservesOmission(t *testing.T) {
	t.Parallel()

	if update := toCoreUpdateTeamEstimationSettings(AppUpdateTeamEstimationSettings{}); update.Scheme.Present {
		t.Fatalf("omitted estimation scheme patch = %#v", update.Scheme)
	}
	scheme := "points"
	if update := toCoreUpdateTeamEstimationSettings(AppUpdateTeamEstimationSettings{Scheme: &scheme}); !update.Scheme.Present || update.Scheme.Value != scheme {
		t.Fatalf("present estimation scheme patch = %#v", update.Scheme)
	}
}
