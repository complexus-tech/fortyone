//go:build integration

package agentinvocationsrepository

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestInvocationFailureReplySurvivesCrashAndDoesNotBecomeActionReceipt(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := acceptInvocation(t, repo, input)
	run, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.BeginFailure(ctx, run, "Maya could not finish this request. Please try again.")
	if err != nil {
		t.Fatal(err)
	}
	if !run.FailurePending || run.Outcome != nil {
		t.Fatal("failure became an action receipt")
	}
	if _, err := postgres.Pool.Exec(ctx, `CREATE TABLE invocation_effects (id uuid PRIMARY KEY, kind text NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	rollback := errors.New("crash during failure publication")
	write := func(txCtx context.Context) error {
		tx, ok := database.UnitTransaction(txCtx)
		if !ok {
			return errors.New("missing transaction")
		}
		_, err := tx.Exec(txCtx, `INSERT INTO invocation_effects (id,kind) VALUES ($1,'failure')`, id)
		return err
	}
	if _, err := repo.PublishFailure(ctx, run, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) {
		if err := write(txCtx); err != nil {
			return uuid.Nil, err
		}
		return uuid.Nil, rollback
	}); !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	assertEffects(t, postgres.Pool, 0)
	run, err = repo.PublishFailure(ctx, run, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) { return id, write(txCtx) })
	if err != nil {
		t.Fatal(err)
	}
	if run.FailureReplyID == nil || *run.FailureReplyID != id || run.ReplyCommentID != nil || run.Outcome != nil {
		t.Fatalf("failure overwrote result: %+v", run)
	}
	if err := repo.DeferFailure(ctx, run); err != nil {
		t.Fatal(err)
	}
	run, err = repo.Get(ctx, input.WorkspaceID, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusQueued || !run.FailurePending {
		t.Fatal("delivery retry lost failure stage")
	}
	if _, err := postgres.Pool.Exec(ctx, `UPDATE public.agent_invocations SET available_at=clock_timestamp() WHERE id=$1`, run.ID); err != nil {
		t.Fatal(err)
	}
	run, err = repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.PublishFailure(ctx, run, func(context.Context, domain.Run) (uuid.UUID, error) {
		t.Fatal("failure reply duplicated")
		return uuid.Nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.CompleteFailure(ctx, run)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusFailed || run.FailurePending {
		t.Fatal("failure delivery did not terminate")
	}
	assertEffects(t, postgres.Pool, 1)
	run, err = repo.Control(ctx, controlFor(run, run.Version), "retry")
	if err != nil {
		t.Fatal(err)
	}
	if run.FailurePending || run.Outcome != nil {
		t.Fatal("explicit retry inherited failure as action result")
	}
	run, err = repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.SaveProposal(ctx, run, domain.Proposal{Reply: "Retry succeeded."})
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Execute(ctx, run, func(context.Context, domain.Run) (domain.Outcome, error) {
		return domain.Outcome{Reply: "Retry succeeded."}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.BeginFailure(ctx, run, "Must not overwrite action"); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("failure overwrote committed result: %v", err)
	}
}
