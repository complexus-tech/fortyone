package messaging

import (
	"context"
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNativeAcceptanceRebindsExactProposalToRespondingUser(t *testing.T) {
	t.Parallel()
	executor, original, backend, _, _ := nativeFixture(t)
	proposal := nativeProposal(t, executor, original, map[string]any{"description": "The exact proposed prose.", "command_intent": "proposal"})
	before, err := executor.InspectStoryMutation(context.Background(), original, proposal.Token)
	require.NoError(t, err)
	responding := original
	responding.UserID = uuid.New()
	teamService := executor.teams.(*teamsServiceStub)
	teamService.calls = nil
	// Previous identity is provenance only. Current permission flags and team
	// access must come from the responding participant, not the previous user.
	previous := original
	previous.AllowMutations = false
	previous.AllowedTeamIDs = []uuid.UUID{}
	accepted, err := executor.RebindStoryMutation(context.Background(), previous, responding, proposal.Token)
	require.NoError(t, err)
	require.NotEqual(t, proposal.Token, accepted.Token)
	require.Equal(t, proposal.ExpiresAt, accepted.ExpiresAt)
	require.Empty(t, backend.updateCalls)
	for _, call := range teamService.calls {
		require.Equal(t, responding.UserID, call.userID)
	}
	inspection, err := executor.InspectStoryMutation(context.Background(), responding, accepted.Token)
	require.NoError(t, err)
	require.True(t, inspection.DirectEligible)
	require.Equal(t, before.Review, inspection.Review)
	require.Equal(t, before.ChangedFields, inspection.ChangedFields)
	_, err = executor.ConfirmStoryMutation(context.Background(), original, accepted.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	_, err = executor.ConfirmStoryMutation(context.Background(), responding, accepted.Token)
	require.NoError(t, err)
	require.Len(t, backend.updateCalls, 1)
	require.Equal(t, responding.UserID, backend.updateCalls[0].actorID)
	require.Equal(t, nativeStoryMutationReason, backend.updateCalls[0].reason)
	require.Equal(t, "The exact proposed prose.", backend.updateCalls[0].updates["description"])
}

func TestNativeAcceptanceFreezesOriginalMeAssignment(t *testing.T) {
	t.Parallel()
	executor, original, backend, members, _ := nativeFixture(t)
	members.current = users.CoreUser{ID: original.UserID, FullName: "Original requester", IsActive: true}
	members.items = append(members.items, members.current)
	proposal := nativeProposal(t, executor, original, map[string]any{"assignee": "me", "command_intent": "proposal"})
	responding := original
	responding.UserID = uuid.New()
	accepted, err := executor.RebindStoryMutation(context.Background(), original, responding, proposal.Token)
	require.NoError(t, err)
	inspection, err := executor.InspectStoryMutation(context.Background(), responding, accepted.Token)
	require.NoError(t, err)
	require.Contains(t, inspection.Review, "After:\nOriginal requester")
	_, err = executor.ConfirmStoryMutation(context.Background(), responding, accepted.Token)
	require.NoError(t, err)
	require.Equal(t, original.UserID, backend.updateCalls[0].updates["assignee_id"])
	require.Equal(t, responding.UserID, backend.updateCalls[0].actorID)
}

func TestNativeAcceptanceRejectsUnavailableOrUnrelatedProposals(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"not_allowed", "team_access", "workspace", "story", "old_identity", "cancelled", "expired", "applied", "stale"} {
		t.Run(scenario, func(t *testing.T) {
			executor, original, backend, _, _ := nativeFixture(t)
			proposal := nativeProposal(t, executor, original, map[string]any{"title": "Exact title", "command_intent": "proposal"})
			responding := original
			responding.UserID = uuid.New()
			previous := original
			expected := ErrInvalidConfirmation
			switch scenario {
			case "not_allowed":
				responding.AllowMutations = false
				expected = ErrMutationNotAllowed
			case "team_access":
				responding.AllowedTeamIDs = []uuid.UUID{}
				expected = ErrTeamNotAccessible
			case "workspace":
				responding.WorkspaceID = uuid.New()
			case "story":
				responding.SourceScope = &SourceScope{Surface: SourceSurfaceStoryComment, StoryID: uuid.New(), TeamID: original.SourceScope.TeamID}
			case "old_identity":
				previous.UserID = uuid.New()
			case "cancelled":
				_, err := executor.CancelStoryMutation(context.Background(), original, proposal.Token)
				require.NoError(t, err)
				expected = ErrCancelledConfirmation
			case "expired":
				executor.mutations.now = func() time.Time { return proposal.ExpiresAt.Add(time.Second) }
				expected = ErrExpiredConfirmation
			case "applied":
				_, err := executor.ConfirmStoryMutation(context.Background(), original, proposal.Token)
				require.NoError(t, err)
				expected = ErrAppliedConfirmation
			case "stale":
				story := backend.persisted[original.SourceScope.StoryID]
				story.UpdatedAt = story.UpdatedAt.Add(time.Second)
				backend.persisted[story.ID] = story
				expected = ErrStaleMutation
			}
			_, err := executor.RebindStoryMutation(context.Background(), previous, responding, proposal.Token)
			require.ErrorIs(t, err, expected)
			require.Len(t, backend.confirmations.entries, 1, "a rejected acceptance must not create another durable proposal")
		})
	}
}
