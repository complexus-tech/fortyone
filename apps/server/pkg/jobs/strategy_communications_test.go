package jobs

import (
	"fmt"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNextQuarterStart(t *testing.T) {
	tests := []struct {
		name     string
		current  time.Time
		expected time.Time
	}{
		{
			name:     "advances within the same year",
			current:  time.Date(2026, time.August, 2, 9, 0, 0, 0, time.UTC),
			expected: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "rolls Q4 into the next year",
			current:  time.Date(2026, time.December, 20, 9, 0, 0, 0, time.UTC),
			expected: time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, nextQuarterStart(test.current))
		})
	}
}

func TestCalendarDaysBetweenIgnoresDaylightSavingTransitions(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	beforeTransition := time.Date(2026, time.March, 7, 9, 0, 0, 0, location)
	afterTransition := time.Date(2026, time.March, 9, 9, 0, 0, 0, location)

	require.Equal(t, 2, calendarDaysBetween(beforeTransition, afterTransition))
}

func TestStrategyWeeklyCheckInDedupeKeyUsesISOWeekYearAcrossCalendarBoundary(t *testing.T) {
	workspaceID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	december31 := time.Date(2025, time.December, 31, 9, 0, 0, 0, time.UTC)
	january1 := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)

	decemberISOYear, decemberISOWeek := december31.ISOWeek()
	januaryISOYear, januaryISOWeek := january1.ISOWeek()

	require.Equal(t, 2026, decemberISOYear)
	require.Equal(t, 1, decemberISOWeek)
	require.Equal(t, decemberISOYear, januaryISOYear)
	require.Equal(t, decemberISOWeek, januaryISOWeek)
	require.Equal(
		t,
		strategyWeeklyCheckInDedupeKey(workspaceID, userID, decemberISOYear, decemberISOWeek),
		strategyWeeklyCheckInDedupeKey(workspaceID, userID, januaryISOYear, januaryISOWeek),
	)
	require.Contains(t, strategyWeeklyCheckInDedupeKey(workspaceID, userID, decemberISOYear, decemberISOWeek), ":2026-01")
}

func TestMissingStrategyElements(t *testing.T) {
	require.Equal(
		t,
		"your ultimate goal, strategic pillars, the next period's objectives",
		missingStrategyElements(strategyFoundation{}),
	)
	require.Equal(
		t,
		"the next period's objectives",
		missingStrategyElements(strategyFoundation{HasUltimateGoal: true, PillarCount: 3}),
	)
	require.Equal(
		t,
		[]string{
			notifications.StrategyMissingElementUltimateGoal,
			notifications.StrategyMissingElementPillars,
			notifications.StrategyMissingElementObjectives,
		},
		missingStrategyElementKeys(strategyFoundation{}),
	)
	require.Equal(
		t,
		[]string{notifications.StrategyMissingElementObjectives},
		missingStrategyElementKeys(strategyFoundation{HasUltimateGoal: true, PillarCount: 3}),
	)
}

func TestStrategyCheckInSummaryOmitsZeroSignals(t *testing.T) {
	summary := strategyCheckInSummary(strategyCheckIn{
		AtRiskObjectives: 2,
		StaleKeyResults:  3,
	})

	require.Equal(t, "2 at-risk objectives, 3 stalled key results", summary)
}

func TestStrategyPlanningReminderIsSentOnceNearThePlanningPeriod(t *testing.T) {
	require.False(t, isStrategyPlanningReminderDue(21))
	require.False(t, isStrategyPlanningReminderDue(8))
	require.True(t, isStrategyPlanningReminderDue(7))
	require.False(t, isStrategyPlanningReminderDue(6))
}

func TestBuildStrategyCheckInsDeduplicatesAndPreservesSignalCounts(t *testing.T) {
	workspaceID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	teamID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	statusID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	offTrackObjectiveID := uuid.MustParse("50000000-0000-0000-0000-000000000001")
	overlappingObjectiveID := uuid.MustParse("50000000-0000-0000-0000-000000000002")
	keyResultOnlyObjectiveID := uuid.MustParse("50000000-0000-0000-0000-000000000003")
	firstKeyResultID := uuid.MustParse("60000000-0000-0000-0000-000000000001")
	secondKeyResultID := uuid.MustParse("60000000-0000-0000-0000-000000000002")
	thirdKeyResultID := uuid.MustParse("60000000-0000-0000-0000-000000000003")

	statusName := "In progress"
	statusCategory := "started"
	atRisk := "At Risk"
	offTrack := "Off Track"
	onTrack := "On Track"
	measurementType := "number"
	firstKeyResultName := "Reach 100 enterprise customers"
	secondKeyResultName := "Reach $1M ARR"
	thirdKeyResultName := "Launch the partner program"
	startValue := 0.0
	currentValue := 35.0
	targetValue := 100.0
	oldestUpdate := time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC)
	middleUpdate := oldestUpdate.Add(24 * time.Hour)
	latestUpdate := middleUpdate.Add(24 * time.Hour)

	recipient := strategyRecipient{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Timezone:    "Africa/Harare",
	}
	overlappingRecord := strategyCheckInRecord{
		strategyRecipient:        recipient,
		ObjectiveID:              overlappingObjectiveID,
		TeamID:                   teamID,
		ObjectiveName:            "Grow enterprise revenue",
		ObjectiveHealth:          &atRisk,
		ObjectiveStatusID:        &statusID,
		ObjectiveStatusName:      &statusName,
		ObjectiveStatusCategory:  &statusCategory,
		ObjectiveUpdatedAt:       middleUpdate,
		IsStaleObjective:         true,
		IsAtRiskObjective:        true,
		KeyResultID:              &secondKeyResultID,
		KeyResultName:            &secondKeyResultName,
		KeyResultMeasurementType: &measurementType,
		KeyResultStartValue:      &startValue,
		KeyResultCurrentValue:    &currentValue,
		KeyResultTargetValue:     &targetValue,
		KeyResultUpdatedAt:       &latestUpdate,
	}

	records := []strategyCheckInRecord{
		overlappingRecord,
		{
			strategyRecipient:       recipient,
			ObjectiveID:             offTrackObjectiveID,
			TeamID:                  teamID,
			ObjectiveName:           "Fix retention",
			ObjectiveHealth:         &offTrack,
			ObjectiveStatusID:       &statusID,
			ObjectiveStatusName:     &statusName,
			ObjectiveStatusCategory: &statusCategory,
			ObjectiveUpdatedAt:      latestUpdate,
			IsAtRiskObjective:       true,
		},
		{
			strategyRecipient:        recipient,
			ObjectiveID:              keyResultOnlyObjectiveID,
			TeamID:                   teamID,
			ObjectiveName:            "Build the partner channel",
			ObjectiveHealth:          &onTrack,
			ObjectiveStatusID:        &statusID,
			ObjectiveStatusName:      &statusName,
			ObjectiveStatusCategory:  &statusCategory,
			ObjectiveUpdatedAt:       latestUpdate,
			KeyResultID:              &thirdKeyResultID,
			KeyResultName:            &thirdKeyResultName,
			KeyResultMeasurementType: &measurementType,
			KeyResultStartValue:      &startValue,
			KeyResultCurrentValue:    &currentValue,
			KeyResultTargetValue:     &targetValue,
			KeyResultUpdatedAt:       &oldestUpdate,
		},
	}
	firstKeyResultRecord := overlappingRecord
	firstKeyResultRecord.KeyResultID = &firstKeyResultID
	firstKeyResultRecord.KeyResultName = &firstKeyResultName
	firstKeyResultRecord.KeyResultUpdatedAt = &middleUpdate
	records = append(records, firstKeyResultRecord)

	checkIns := buildStrategyCheckIns(records)

	require.Len(t, checkIns, 1)
	checkIn := checkIns[0]
	require.Equal(t, 2, checkIn.AtRiskObjectives)
	require.Equal(t, 1, checkIn.StaleObjectives)
	require.Equal(t, 3, checkIn.StaleKeyResults)
	require.Equal(t, 3, uniqueStrategyObjectiveCount(checkIn))
	require.Len(t, checkIn.Objectives, 2)
	require.Equal(t, offTrackObjectiveID, checkIn.Objectives[0].ID)
	require.Equal(t, overlappingObjectiveID, checkIn.Objectives[1].ID)
	require.Equal(
		t,
		[]string{
			notifications.StrategySignalReasonAtRisk,
			notifications.StrategySignalReasonStale,
		},
		checkIn.Objectives[1].Reasons,
	)
	require.Equal(t, statusID, checkIn.Objectives[1].Status.ID)
	require.Equal(t, "started", checkIn.Objectives[1].Status.Category)

	require.Len(t, checkIn.KeyResults, 3)
	require.Equal(t, thirdKeyResultID, checkIn.KeyResults[0].ID)
	require.Equal(t, firstKeyResultID, checkIn.KeyResults[1].ID)
	require.Equal(t, secondKeyResultID, checkIn.KeyResults[2].ID)
	require.Equal(t, keyResultOnlyObjectiveID, checkIn.KeyResults[0].ObjectiveID)
	require.Equal(t, "Build the partner channel", checkIn.KeyResults[0].ObjectiveName)
	require.Equal(
		t,
		[]string{
			notifications.StrategySignalReasonStale,
			notifications.StrategySignalReasonIncomplete,
		},
		checkIn.KeyResults[0].Reasons,
	)
}

func TestStrategyWeeklyLocalTimeFallsBackToUTCForInvalidTimezone(t *testing.T) {
	now := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	localNow, due := strategyWeeklyLocalTime(now, "not/a-timezone")

	require.True(t, due)
	require.Equal(t, time.UTC, localNow.Location())
	require.Equal(t, now, localNow)
}

func TestStrategyWeeklyLocalTimeTracksDaylightSavingBoundary(t *testing.T) {
	beforeDST := time.Date(2026, time.March, 4, 14, 0, 0, 0, time.UTC)
	afterDST := time.Date(2026, time.March, 11, 13, 0, 0, 0, time.UTC)

	beforeLocal, beforeDue := strategyWeeklyLocalTime(beforeDST, "America/New_York")
	afterLocal, afterDue := strategyWeeklyLocalTime(afterDST, "America/New_York")

	require.True(t, beforeDue)
	require.True(t, afterDue)
	require.Equal(t, 9, beforeLocal.Hour())
	require.Equal(t, 9, afterLocal.Hour())
	_, beforeOffset := beforeLocal.Zone()
	_, afterOffset := afterLocal.Zone()
	require.Equal(t, -5*60*60, beforeOffset)
	require.Equal(t, -4*60*60, afterOffset)
}

func TestStrategyMonthlySummaryTextDoesNotInventZeroProgressWithoutKeyResults(t *testing.T) {
	progress := 46.0
	withKeyResults := strategyMonthlySummary{
		AtRiskObjectives:    2,
		UnalignedObjectives: 3,
		KeyResultCount:      6,
		KeyResultProgress:   &progress,
		CompletedStories:    12,
	}

	withProgress := strategyMonthlySummaryText(withKeyResults)
	require.Contains(t, withProgress, "46% average progress across 6 key results")
	require.Contains(t, withProgress, "12 linked stories completed last month")

	withoutProgress := strategyMonthlySummaryText(strategyMonthlySummary{
		AtRiskObjectives:    2,
		UnalignedObjectives: 3,
		CompletedStories:    12,
	})
	require.Contains(t, withoutProgress, "no key results in the current snapshot")
	require.NotContains(t, withoutProgress, "0%")
}

func TestStrategyMonthlySummaryRequiresAnActionableSignal(t *testing.T) {
	require.False(t, (strategyMonthlySummary{PillarCount: 2, ObjectiveCount: 4, KeyResultCount: 8}).hasActionableSignal())
	require.True(t, (strategyMonthlySummary{PillarsNeedingReview: 1}).hasActionableSignal())
	require.True(t, (strategyMonthlySummary{AtRiskObjectives: 1}).hasActionableSignal())
	require.True(t, (strategyMonthlySummary{UnalignedObjectives: 1}).hasActionableSignal())
}

func TestStrategyWeeklyTeamCountsAreCompleteDeterministicAndDeduplicateObjectiveOverlap(t *testing.T) {
	firstTeamID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	secondTeamID := uuid.MustParse("30000000-0000-0000-0000-000000000002")
	firstObjectiveID := uuid.MustParse("50000000-0000-0000-0000-000000000001")
	secondObjectiveID := uuid.MustParse("50000000-0000-0000-0000-000000000002")
	thirdObjectiveID := uuid.MustParse("50000000-0000-0000-0000-000000000003")
	checkIn := strategyCheckIn{
		AtRiskObjectives: 1,
		StaleObjectives:  2,
		StaleKeyResults:  3,
		Objectives: []notifications.StrategyObjectiveSnapshot{
			{
				ID:     thirdObjectiveID,
				TeamID: secondTeamID,
				Reasons: []string{
					notifications.StrategySignalReasonStale,
				},
			},
			{
				ID:     firstObjectiveID,
				TeamID: firstTeamID,
				Reasons: []string{
					notifications.StrategySignalReasonAtRisk,
					notifications.StrategySignalReasonStale,
				},
			},
		},
		KeyResults: []notifications.StrategyKeyResultSnapshot{
			{ID: uuid.MustParse("60000000-0000-0000-0000-000000000001"), ObjectiveID: firstObjectiveID, TeamID: firstTeamID},
			{ID: uuid.MustParse("60000000-0000-0000-0000-000000000002"), ObjectiveID: secondObjectiveID, TeamID: firstTeamID},
			{ID: uuid.MustParse("60000000-0000-0000-0000-000000000003"), ObjectiveID: thirdObjectiveID, TeamID: secondTeamID},
		},
	}

	teamCounts := strategyWeeklyTeamCounts(checkIn, checkIn.Objectives, checkIn.KeyResults)

	require.Equal(t, []notifications.StrategyWeeklyCheckInTeamCountsSnapshot{
		{
			TeamID: firstTeamID,
			Counts: notifications.StrategyWeeklyCheckInCounts{
				AtRiskObjectives: 1,
				StaleObjectives:  1,
				StaleKeyResults:  2,
				UniqueObjectives: 2,
			},
		},
		{
			TeamID: secondTeamID,
			Counts: notifications.StrategyWeeklyCheckInCounts{
				StaleObjectives:  1,
				StaleKeyResults:  1,
				UniqueObjectives: 1,
			},
		},
	}, teamCounts)

	var summed notifications.StrategyWeeklyCheckInCounts
	for _, teamCount := range teamCounts {
		summed.AtRiskObjectives += teamCount.Counts.AtRiskObjectives
		summed.StaleObjectives += teamCount.Counts.StaleObjectives
		summed.StaleKeyResults += teamCount.Counts.StaleKeyResults
		summed.UniqueObjectives += teamCount.Counts.UniqueObjectives
	}
	require.Equal(t, checkIn.AtRiskObjectives, summed.AtRiskObjectives)
	require.Equal(t, checkIn.StaleObjectives, summed.StaleObjectives)
	require.Equal(t, checkIn.StaleKeyResults, summed.StaleKeyResults)
	require.Equal(t, uniqueStrategyObjectiveCount(checkIn), summed.UniqueObjectives)
}

func TestBoundedStrategyCheckInDetailsBalancesTypesAndReportsOmissions(t *testing.T) {
	firstTeamID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	secondTeamID := uuid.MustParse("30000000-0000-0000-0000-000000000002")
	objectives := make([]notifications.StrategyObjectiveSnapshot, 12)
	for index := range objectives {
		objectives[index].ID = uuid.MustParse(fmt.Sprintf("50000000-0000-0000-0000-%012d", index+1))
		objectives[index].TeamID = firstTeamID
		if index >= len(objectives)/2 {
			objectives[index].TeamID = secondTeamID
		}
		objectives[index].Reasons = []string{notifications.StrategySignalReasonStale}
	}
	keyResults := make([]notifications.StrategyKeyResultSnapshot, 8)
	for index := range keyResults {
		keyResults[index].ID = uuid.MustParse(fmt.Sprintf("60000000-0000-0000-0000-%012d", index+1))
		keyResults[index].ObjectiveID = uuid.MustParse(fmt.Sprintf("70000000-0000-0000-0000-%012d", index+1))
		keyResults[index].TeamID = firstTeamID
		if index >= len(keyResults)/2 {
			keyResults[index].TeamID = secondTeamID
		}
	}
	checkIn := strategyCheckIn{Objectives: objectives, KeyResults: keyResults}

	selectedObjectives, selectedKeyResults, omitted := boundedStrategyCheckInDetails(checkIn, strategyWeeklyDetailLimit)
	teamCounts := strategyWeeklyTeamCounts(checkIn, selectedObjectives, selectedKeyResults)

	require.Len(t, selectedObjectives, 5)
	require.Len(t, selectedKeyResults, 5)
	require.Equal(t, objectives[:5], selectedObjectives)
	require.Equal(t, keyResults[:5], selectedKeyResults)
	require.Equal(t, &notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot{
		Objectives: 7,
		KeyResults: 3,
	}, omitted)
	require.Len(t, checkIn.Objectives, 12)
	require.Len(t, checkIn.KeyResults, 8)
	require.Equal(t, 20, uniqueStrategyObjectiveCount(checkIn))
	require.Equal(t, []notifications.StrategyWeeklyCheckInTeamCountsSnapshot{
		{
			TeamID: firstTeamID,
			Counts: notifications.StrategyWeeklyCheckInCounts{
				StaleObjectives:  6,
				StaleKeyResults:  4,
				UniqueObjectives: 10,
			},
			OmittedDetails: &notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot{
				Objectives: 1,
				KeyResults: 0,
			},
		},
		{
			TeamID: secondTeamID,
			Counts: notifications.StrategyWeeklyCheckInCounts{
				StaleObjectives:  6,
				StaleKeyResults:  4,
				UniqueObjectives: 10,
			},
			OmittedDetails: &notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot{
				Objectives: 6,
				KeyResults: 3,
			},
		},
	}, teamCounts)
}

func TestBoundedStrategyCheckInDetailsUsesSpareCapacityAndOmitsMetadataWhenComplete(t *testing.T) {
	objectives := make([]notifications.StrategyObjectiveSnapshot, strategyWeeklyDetailLimit)
	for index := range objectives {
		objectives[index].ID = uuid.MustParse(fmt.Sprintf("50000000-0000-0000-0000-%012d", index+1))
	}

	selectedObjectives, selectedKeyResults, omitted := boundedStrategyCheckInDetails(
		strategyCheckIn{Objectives: objectives},
		strategyWeeklyDetailLimit,
	)

	require.Len(t, selectedObjectives, strategyWeeklyDetailLimit)
	require.Empty(t, selectedKeyResults)
	require.Nil(t, omitted)
}
