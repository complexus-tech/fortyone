//go:build integration

package agentinvocationsrepository

import (
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInvocationListExcludesIgnoredDecisionsBeforeLimit(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	visible := acceptInvocation(t, repo, input)
	var ignored domain.Run
	for range 105 {
		input.SourceID = uuid.New()
		input.IdempotencyKey = uuid.NewString()
		ignored = acceptInvocation(t, repo, input)
		claimed, err := repo.Claim(t.Context(), ignored.ID)
		require.NoError(t, err)
		claimed, err = repo.SaveProposal(t.Context(), claimed, domain.Proposal{Ignore: true})
		require.NoError(t, err)
		_, err = repo.Complete(t.Context(), claimed)
		require.NoError(t, err)
	}
	runs, err := repo.List(t.Context(), input.WorkspaceID, input.EntityID)
	require.NoError(t, err)
	require.Len(t, runs, 1, "ignored traffic must not displace real requests from the latest 100")
	require.Equal(t, visible.ID, runs[0].ID)
	persisted, err := repo.Get(t.Context(), input.WorkspaceID, ignored.ID)
	require.NoError(t, err)
	require.True(t, persisted.Proposal.Ignore, "internal recovery retains access to its durable decision")
}
