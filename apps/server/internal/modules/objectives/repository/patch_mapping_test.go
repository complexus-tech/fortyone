package objectivesrepository

import (
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestObjectivePatchParamsPreserveOmittedNullAndValue(t *testing.T) {
	t.Parallel()

	workspaceID, objectiveID, actorID := uuid.New(), uuid.New(), uuid.New()
	statusID := uuid.New()
	dueDate := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	params := objectivePatchParams(objectivesdomain.UpdateCommand{
		WorkspaceID: workspaceID, ObjectiveID: objectiveID, ActorID: actorID,
		Patch: objectivesdomain.ObjectivePatch{
			Description: objectivesdomain.ClearField[string](),
			EndDate:     objectivesdomain.SetField(dueDate),
			Status:      objectivesdomain.SetField(statusID),
		},
	})

	require.True(t, params.SetDescription)
	require.Nil(t, params.Description)
	require.False(t, params.SetName)
	require.True(t, params.SetEndDate)
	require.NotNil(t, params.EndDate)
	require.Equal(t, dueDate, *params.EndDate)
	require.True(t, params.SetStatusID)
	require.Equal(t, statusID, *params.StatusID)
	require.Equal(t, objectiveID, params.ObjectiveID)
	require.Equal(t, workspaceID, *params.WorkspaceID)
}

func TestObjectivePatchChangesHaveStableOrderAndSkipRichTextAutosaves(t *testing.T) {
	t.Parallel()

	patch := objectivesdomain.ObjectivePatch{
		Health:      objectivesdomain.SetField(objectivesdomain.HealthAtRisk),
		Name:        objectivesdomain.SetField("Launch reliably"),
		Description: objectivesdomain.SetField("Long-form context"),
		Priority:    objectivesdomain.ClearField[string](),
	}

	require.Equal(t, []objectiveChange{
		{field: "name", value: "Launch reliably"},
		{field: "priority", value: "nil"},
		{field: "health", value: "At Risk"},
	}, objectivePatchChanges(patch))
}

func TestDeduplicateUUIDsIsStableAndDropsZeroValues(t *testing.T) {
	t.Parallel()

	first, second := uuid.New(), uuid.New()
	require.Equal(t, []uuid.UUID{first, second}, deduplicateUUIDs([]uuid.UUID{
		first, uuid.Nil, second, first,
	}))
}
