package searchrepository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToCoreSearchStoryKeepsComplexitySeparateFromDuration(t *testing.T) {
	complexity := int16(5)
	duration := 90
	story := toCoreSearchStory(dbStory{
		EstimateValue:            &complexity,
		EstimateScheme:           "tshirt",
		EstimatedDurationMinutes: &duration,
	})

	require.Equal(t, &complexity, story.EstimateValue)
	require.Equal(t, "tshirt", story.EstimateScheme)
	require.NotNil(t, story.EstimateLabel)
	require.Equal(t, "L", *story.EstimateLabel)
	require.Equal(t, &duration, story.EstimatedDurationMinutes)
}

func TestSearchEstimateLabelSupportsPoints(t *testing.T) {
	complexity := int16(3)

	require.Equal(t, "3", *searchEstimateLabel("points", &complexity))
	require.Nil(t, searchEstimateLabel("tshirt", nil))
}
