package searchrepository

import (
	"testing"
	"time"

	searchsql "github.com/complexus-tech/projects-api/internal/modules/search/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestToCoreSearchStoryKeepsComplexitySeparateFromDuration(t *testing.T) {
	complexity := int16(5)
	duration := int32(90)
	sequenceID := int32(41)
	priority := "high"
	reason := "Waiting for an owner"
	statusName := "In Progress"
	statusColor := "#eab308"
	statusCategory := "started"
	assigneeFullName := "Joseph Mukorivo"
	assigneeUsername := "joseph"
	updatedAt := time.Date(2026, time.August, 15, 9, 30, 0, 0, time.UTC)
	story, err := toCoreSearchStory(searchsql.SearchStoriesRow{
		ID:                       uuid.New(),
		SequenceID:               &sequenceID,
		Priority:                 &priority,
		TeamID:                   uuid.New(),
		WorkspaceID:              uuid.New(),
		EstimateUnit:             &complexity,
		EstimateScheme:           "tshirt",
		EstimatedDurationMinutes: &duration,
		AutoSchedulingEnabled:    true,
		AutoSchedulingLocked:     true,
		AutoSchedulingStatus:     "needs_owner",
		AutoSchedulingReason:     &reason,
		AutoSchedulingUpdatedAt:  &updatedAt,
		StatusName:               &statusName,
		StatusColor:              &statusColor,
		StatusCategory:           &statusCategory,
		TeamName:                 "Product",
		TeamCode:                 "PROD",
		AssigneeFullName:         &assigneeFullName,
		AssigneeUsername:         &assigneeUsername,
	})
	require.NoError(t, err)

	require.Equal(t, &complexity, story.EstimateValue)
	require.Equal(t, "tshirt", story.EstimateScheme)
	require.NotNil(t, story.EstimateLabel)
	require.Equal(t, "L", *story.EstimateLabel)
	require.Equal(t, 90, *story.EstimatedDurationMinutes)
	require.True(t, story.AutoSchedulingEnabled)
	require.True(t, story.AutoSchedulingLocked)
	require.Equal(t, "needs_owner", story.AutoSchedulingStatus)
	require.Equal(t, &reason, story.AutoSchedulingReason)
	require.Equal(t, &updatedAt, story.AutoSchedulingUpdatedAt)
	require.Equal(t, &statusName, story.StatusName)
	require.Equal(t, &statusColor, story.StatusColor)
	require.Equal(t, &statusCategory, story.StatusCategory)
	require.Equal(t, "Product", story.TeamName)
	require.Equal(t, "PROD", story.TeamCode)
	require.Equal(t, &assigneeFullName, story.AssigneeFullName)
	require.Equal(t, &assigneeUsername, story.AssigneeUsername)
}

func TestSearchProjectionFailsClosedWhenRequiredColumnsAreNull(t *testing.T) {
	_, err := toCoreSearchStory(searchsql.SearchStoriesRow{})
	require.ErrorIs(t, err, errInvalidProjection)

	_, err = toCoreSearchObjective(searchsql.SearchObjectivesRow{})
	require.ErrorIs(t, err, errInvalidProjection)

	_, err = toCoreSimilarStory(searchsql.FindSimilarStoriesRow{})
	require.ErrorIs(t, err, errInvalidProjection)
}

func TestSearchEstimateLabelSupportsPoints(t *testing.T) {
	complexity := int16(3)

	require.Equal(t, "3", *searchEstimateLabel("points", &complexity))
	require.Nil(t, searchEstimateLabel("tshirt", nil))
}
