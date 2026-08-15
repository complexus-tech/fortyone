package searchhttp

import (
	"testing"
	"time"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/stretchr/testify/require"
)

func TestToAppSearchStoryIncludesComplexityAndDuration(t *testing.T) {
	complexity := int16(3)
	label := "M"
	duration := 120
	reason := "No eligible capacity"
	updatedAt := time.Date(2026, time.August, 15, 9, 30, 0, 0, time.UTC)
	story := toAppSearchStory(search.CoreSearchStory{
		EstimateValue:            &complexity,
		EstimateLabel:            &label,
		EstimateScheme:           "tshirt",
		EstimatedDurationMinutes: &duration,
		AutoSchedulingEnabled:    true,
		AutoSchedulingLocked:     true,
		AutoSchedulingStatus:     "cannot_fit",
		AutoSchedulingReason:     &reason,
		AutoSchedulingUpdatedAt:  &updatedAt,
	})

	require.Equal(t, &complexity, story.EstimateValue)
	require.Equal(t, &label, story.EstimateLabel)
	require.Equal(t, "tshirt", story.EstimateScheme)
	require.Equal(t, &duration, story.EstimatedDurationMinutes)
	require.True(t, story.AutoSchedulingEnabled)
	require.True(t, story.AutoSchedulingLocked)
	require.Equal(t, "cannot_fit", story.AutoSchedulingStatus)
	require.Equal(t, &reason, story.AutoSchedulingReason)
	require.Equal(t, &updatedAt, story.AutoSchedulingUpdatedAt)
}
