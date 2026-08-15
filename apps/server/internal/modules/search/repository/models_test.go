package searchrepository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToCoreSearchStoryKeepsComplexitySeparateFromDuration(t *testing.T) {
	complexity := int16(5)
	duration := 90
	reason := "Waiting for an owner"
	updatedAt := time.Date(2026, time.August, 15, 9, 30, 0, 0, time.UTC)
	story := toCoreSearchStory(dbStory{
		EstimateValue:            &complexity,
		EstimateScheme:           "tshirt",
		EstimatedDurationMinutes: &duration,
		AutoSchedulingEnabled:    true,
		AutoSchedulingLocked:     true,
		AutoSchedulingStatus:     "needs_owner",
		AutoSchedulingReason:     &reason,
		AutoSchedulingUpdatedAt:  &updatedAt,
	})

	require.Equal(t, &complexity, story.EstimateValue)
	require.Equal(t, "tshirt", story.EstimateScheme)
	require.NotNil(t, story.EstimateLabel)
	require.Equal(t, "L", *story.EstimateLabel)
	require.Equal(t, &duration, story.EstimatedDurationMinutes)
	require.True(t, story.AutoSchedulingEnabled)
	require.True(t, story.AutoSchedulingLocked)
	require.Equal(t, "needs_owner", story.AutoSchedulingStatus)
	require.Equal(t, &reason, story.AutoSchedulingReason)
	require.Equal(t, &updatedAt, story.AutoSchedulingUpdatedAt)
}

func TestSearchEstimateLabelSupportsPoints(t *testing.T) {
	complexity := int16(3)

	require.Equal(t, "3", *searchEstimateLabel("points", &complexity))
	require.Nil(t, searchEstimateLabel("tshirt", nil))
}
