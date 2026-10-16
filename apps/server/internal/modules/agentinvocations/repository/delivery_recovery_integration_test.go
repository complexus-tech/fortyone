//go:build integration

package agentinvocationsrepository

import (
	"context"
	"encoding/json"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInvocationDurableCompletionRetriesRemainQueuedWithBackoff(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	for _, stage := range []string{"action receipt", "result reply", "approval reply"} {
		t.Run(stage, func(t *testing.T) {
			input := invocationFixture(t, postgres.Pool)
			run := acceptInvocation(t, repo, input)
			run, err := repo.Claim(ctx, run.ID)
			require.NoError(t, err)
			proposal := domain.Proposal{Reply: "Should I close it?", Review: "Status: Open → Closed", Action: json.RawMessage(`{}`), RequiresApproval: stage == "approval reply"}
			run, err = repo.SaveProposal(ctx, run, proposal)
			require.NoError(t, err)
			if stage == "approval reply" {
				run, err = repo.PublishProposal(ctx, run, func(context.Context, domain.Run) (uuid.UUID, error) {
					return uuid.New(), nil
				})
			} else {
				run, err = repo.Execute(ctx, run, func(context.Context, domain.Run) (domain.Outcome, error) {
					return domain.Outcome{Reply: "Updated", Receipt: json.RawMessage(`{"applied":true}`)}, nil
				})
				if err == nil && stage == "result reply" {
					run, err = repo.Publish(ctx, run, func(context.Context, domain.Run) (uuid.UUID, error) {
						return uuid.New(), nil
					})
				}
			}
			require.NoError(t, err)
			for range maximumAttempts + 2 {
				// Persistence must retain completion retries even if a commit's
				// response was lost and the worker has an older in-memory copy.
				stale := run
				stale.Outcome, stale.ProposalReplyID = nil, nil
				require.NoError(t, repo.Fail(ctx, stale, "Temporary delivery outage", true))
				queued, err := repo.Get(ctx, input.WorkspaceID, run.ID)
				require.NoError(t, err)
				require.Equal(t, domain.StatusQueued, queued.Status)
				require.True(t, queued.HasDurableResponse())
				var backedOff bool
				require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT available_at > clock_timestamp() FROM agent_invocations WHERE id=$1`, run.ID).Scan(&backedOff))
				require.True(t, backedOff, "recovery must not hot-loop during an outage")
				pending, err := repo.Pending(ctx, 100)
				require.NoError(t, err)
				require.NotContains(t, pending, run.ID)
				_, err = repo.Claim(ctx, run.ID)
				require.ErrorIs(t, err, domain.ErrLeaseLost)
				_, err = postgres.Pool.Exec(ctx, `UPDATE agent_invocations SET available_at=clock_timestamp() WHERE id=$1`, run.ID)
				require.NoError(t, err)
				run, err = repo.Claim(ctx, run.ID)
				require.NoError(t, err)
			}
			require.Greater(t, run.RetryAttemptCount, int32(maximumAttempts))
			// A permanent authorization/source failure still halts recovery.
			require.NoError(t, repo.Fail(ctx, run, "Access revoked", false))
			terminal, err := repo.Get(ctx, input.WorkspaceID, run.ID)
			require.NoError(t, err)
			require.Equal(t, domain.StatusFailed, terminal.Status)
			pending, err := repo.Pending(ctx, 100)
			require.NoError(t, err)
			require.NotContains(t, pending, run.ID)
		})
	}
}
