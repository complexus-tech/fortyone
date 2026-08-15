package searchhttp

import (
	"testing"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/stretchr/testify/require"
)

func TestToAppSearchStoryIncludesComplexityAndDuration(t *testing.T) {
	complexity := int16(3)
	label := "M"
	duration := 120
	story := toAppSearchStory(search.CoreSearchStory{
		EstimateValue:            &complexity,
		EstimateLabel:            &label,
		EstimateScheme:           "tshirt",
		EstimatedDurationMinutes: &duration,
	})

	require.Equal(t, &complexity, story.EstimateValue)
	require.Equal(t, &label, story.EstimateLabel)
	require.Equal(t, "tshirt", story.EstimateScheme)
	require.Equal(t, &duration, story.EstimatedDurationMinutes)
}
