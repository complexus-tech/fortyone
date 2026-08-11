package objectivesrepository

import (
	"context"
	"testing"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeObjectiveKeyResultWriter struct {
	finalSequenceID    int
	allocatedWorkspace uuid.UUID
	allocatedTeam      uuid.UUID
	allocatedCount     int
	inserted           []dbKeyResult
	contributors       map[uuid.UUID][]uuid.UUID
}

func (w *fakeObjectiveKeyResultWriter) AllocateSequences(
	_ context.Context,
	workspaceID, teamID uuid.UUID,
	count int,
) (int, error) {
	w.allocatedWorkspace = workspaceID
	w.allocatedTeam = teamID
	w.allocatedCount = count
	return w.finalSequenceID, nil
}

func (w *fakeObjectiveKeyResultWriter) InsertKeyResult(
	_ context.Context,
	keyResult dbKeyResult,
) (dbKeyResult, error) {
	keyResult.ID = uuid.New()
	w.inserted = append(w.inserted, keyResult)
	return keyResult, nil
}

func (w *fakeObjectiveKeyResultWriter) InsertContributor(
	_ context.Context,
	keyResultID, contributorID uuid.UUID,
) error {
	if w.contributors == nil {
		w.contributors = make(map[uuid.UUID][]uuid.UUID)
	}
	w.contributors[keyResultID] = append(w.contributors[keyResultID], contributorID)
	return nil
}

func TestBuildObjectiveUpdateStatementScopesAndOrdersUpdates(t *testing.T) {
	objectiveID := uuid.New()
	workspaceID := uuid.New()

	query, params := buildObjectiveUpdateStatement(objectiveID, workspaceID, map[string]any{
		"priority": "high",
		"name":     "Grow enterprise adoption",
	})

	require.Equal(t,
		"UPDATE objectives SET name = :name, priority = :priority, updated_at = NOW() WHERE objective_id = :id AND workspace_id = :workspace_id",
		query,
	)
	require.Equal(t, objectiveID, params["id"])
	require.Equal(t, workspaceID, params["workspace_id"])
	require.Equal(t, "Grow enterprise adoption", params["name"])
	require.Equal(t, "high", params["priority"])
}

func TestCreateObjectiveKeyResultsAllocatesTeamSequencesAndContributors(t *testing.T) {
	objectiveID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	creatorID := uuid.New()
	firstContributorID := uuid.New()
	secondContributorID := uuid.New()
	writer := &fakeObjectiveKeyResultWriter{finalSequenceID: 12}

	created, err := createObjectiveKeyResults(
		context.Background(),
		writer,
		objectiveID,
		workspaceID,
		teamID,
		[]keyresults.CoreNewKeyResult{
			{
				Name:            "Increase activation",
				MeasurementType: "percentage",
				StartValue:      20,
				CurrentValue:    20,
				TargetValue:     60,
				CreatedBy:       creatorID,
				Contributors:    []uuid.UUID{firstContributorID, secondContributorID, firstContributorID},
			},
			{
				Name:            "Reduce time to value",
				MeasurementType: "number",
				StartValue:      10,
				CurrentValue:    10,
				TargetValue:     3,
				CreatedBy:       creatorID,
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, workspaceID, writer.allocatedWorkspace)
	require.Equal(t, teamID, writer.allocatedTeam)
	require.Equal(t, 2, writer.allocatedCount)
	require.Len(t, writer.inserted, 2)
	require.Equal(t, 11, writer.inserted[0].SequenceID)
	require.Equal(t, 12, writer.inserted[1].SequenceID)
	require.Equal(t, teamID, writer.inserted[0].TeamID)
	require.Equal(t, objectiveID, writer.inserted[0].ObjectiveID)
	require.Len(t, created, 2)
	require.Equal(t, 11, created[0].SequenceID)
	require.Equal(t,
		[]uuid.UUID{firstContributorID, secondContributorID},
		created[0].Contributors,
	)
	require.Equal(t,
		[]uuid.UUID{firstContributorID, secondContributorID},
		writer.contributors[created[0].ID],
	)
}
