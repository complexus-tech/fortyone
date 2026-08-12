package emailreply

import (
	"context"
	"errors"
	"testing"
	"time"

	emailagent "github.com/complexus-tech/projects-api/internal/modules/emailagent/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type objectiveMutationStub struct {
	id, workspaceID, userID uuid.UUID
	expected                time.Time
	comment                 string
	updates                 map[string]any
	err                     error
}

func (stub *objectiveMutationStub) UpdateExternalUserActionIfUnchanged(
	_ context.Context,
	id, workspaceID, userID uuid.UUID,
	expected time.Time,
	comment string,
	updates map[string]any,
) error {
	stub.id, stub.workspaceID, stub.userID = id, workspaceID, userID
	stub.expected, stub.comment, stub.updates = expected, comment, updates
	return stub.err
}

type keyResultMutationStub struct{}

func (*keyResultMutationStub) UpdateExternalUserActionIfUnchanged(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time, map[string]any, string) error {
	return nil
}

type storyMutationStub struct {
	updates map[string]any
}

func (stub *storyMutationStub) UpdateExternalUserActionIfUnchanged(_ context.Context, _, _, _ uuid.UUID, _ time.Time, updates map[string]any) error {
	stub.updates = updates
	return nil
}

type feedbackMutationStub struct{}

func (*feedbackMutationStub) UpdateItemStatusIfUnchanged(context.Context, uuid.UUID, uuid.UUID, time.Time, feedback.CoreUpdateItemStatusInput) (feedback.CoreItem, error) {
	return feedback.CoreItem{}, nil
}

type versionReaderStub struct {
	version time.Time
	err     error
}

func (stub versionReaderStub) CurrentVersion(context.Context, emailagent.ActionProposal) (time.Time, error) {
	return stub.version, stub.err
}

func newMutationApplierForTest(t *testing.T, objective *objectiveMutationStub, story *storyMutationStub, version time.Time) *DomainMutationApplier {
	t.Helper()
	applier, err := NewDomainMutationApplier(objective, &keyResultMutationStub{}, story, &feedbackMutationStub{}, versionReaderStub{version: version})
	require.NoError(t, err)
	return applier
}

func TestDomainMutationApplierMapsObjectiveCASAndCheckIn(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, objectiveID, teamID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	expected := time.Now().UTC()
	health := emailagent.ObjectiveHealthOnTrack
	checkIn := "The launch blocker is resolved."
	objective := &objectiveMutationStub{}
	applier := newMutationApplierForTest(t, objective, &storyMutationStub{}, expected)

	err := applier.Apply(context.Background(), emailagent.ActionProposal{
		WorkspaceID: workspaceID, ActorID: actorID, Kind: emailagent.ActionObjectiveUpdate,
		Objective: &emailagent.ObjectiveAction{
			Target: emailagent.TargetSnapshot{ID: objectiveID, TeamID: teamID, ExpectedUpdatedAt: expected},
			Health: &health, CheckIn: &checkIn,
		},
	})

	require.NoError(t, err)
	require.Equal(t, objectiveID, objective.id)
	require.Equal(t, workspaceID, objective.workspaceID)
	require.Equal(t, actorID, objective.userID)
	require.Equal(t, expected, objective.expected)
	require.Equal(t, checkIn, objective.comment)
	require.Equal(t, map[string]any{"health": "On Track"}, objective.updates)
}

func TestDomainMutationApplierMapsStoryClearAndUnassign(t *testing.T) {
	t.Parallel()

	story := &storyMutationStub{}
	expected := time.Now()
	applier := newMutationApplierForTest(t, &objectiveMutationStub{}, story, expected)
	err := applier.Apply(context.Background(), emailagent.ActionProposal{
		WorkspaceID: uuid.New(), ActorID: uuid.New(), Kind: emailagent.ActionStoryUpdate,
		Story: &emailagent.StoryAction{
			Target:   emailagent.TargetSnapshot{ID: uuid.New(), TeamID: uuid.New(), ExpectedUpdatedAt: expected},
			DueDate:  &emailagent.DateChange{Operation: emailagent.DateClear},
			Assignee: &emailagent.AssigneeChange{Operation: emailagent.AssigneeUnassign},
		},
	})

	require.NoError(t, err)
	require.Contains(t, story.updates, "end_date")
	require.Nil(t, story.updates["end_date"])
	require.Contains(t, story.updates, "assignee_id")
	require.Nil(t, story.updates["assignee_id"])
}

func TestDomainMutationApplierNormalizesCASConflict(t *testing.T) {
	t.Parallel()

	health := emailagent.ObjectiveHealthAtRisk
	objective := &objectiveMutationStub{err: objectives.ErrVersionConflict}
	expected := time.Now()
	applier := newMutationApplierForTest(t, objective, &storyMutationStub{}, expected)
	err := applier.Apply(context.Background(), emailagent.ActionProposal{
		WorkspaceID: uuid.New(), ActorID: uuid.New(), Kind: emailagent.ActionObjectiveUpdate,
		Objective: &emailagent.ObjectiveAction{
			Target: emailagent.TargetSnapshot{ID: uuid.New(), TeamID: uuid.New(), ExpectedUpdatedAt: expected},
			Health: &health,
		},
	})

	require.ErrorIs(t, err, ErrActionConflict)
	require.True(t, errors.Is(err, ErrActionConflict))
}
