//go:build integration

package agentinvocationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestInvocationApprovalQuestionIsAtomicRecoverableAndSeparateFromResult(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := acceptInvocation(t, repo, input)
	run, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.SaveProposal(ctx, run, domain.Proposal{Reply: "Should I close this story?", Review: "Status: Open → Closed", Action: json.RawMessage(`{}`), RequiresApproval: true})
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusRunning {
		t.Fatal("paused before question publication")
	}
	if _, err := postgres.Pool.Exec(ctx, `CREATE TABLE invocation_effects (id uuid PRIMARY KEY, kind text NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	write := func(txCtx context.Context, id uuid.UUID, kind string) error {
		tx, ok := database.UnitTransaction(txCtx)
		if !ok {
			return errors.New("missing unit")
		}
		_, err := tx.Exec(txCtx, `INSERT INTO invocation_effects (id,kind) VALUES ($1,$2)`, id, kind)
		return err
	}
	questionID := uuid.New()
	rollback := errors.New("publication failed after write")
	if _, err := repo.PublishProposal(ctx, run, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) {
		if err := write(txCtx, questionID, "question"); err != nil {
			return uuid.Nil, err
		}
		return uuid.Nil, rollback
	}); !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	assertEffects(t, postgres.Pool, 0)
	run, err = repo.PublishProposal(ctx, run, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) {
		return questionID, write(txCtx, questionID, "question")
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.ProposalReplyID == nil || *run.ProposalReplyID != questionID || run.Outcome != nil || run.ReplyCommentID != nil {
		t.Fatal("proposal crossed final result boundary")
	}
	assertEffects(t, postgres.Pool, 1)
	if err := repo.Fail(ctx, run, "Notification temporarily unavailable", false); err != nil {
		t.Fatal(err)
	}
	run, err = repo.Get(ctx, input.WorkspaceID, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Control(ctx, controlFor(run, run.Version), "retry")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusQueued {
		t.Fatal("undelivered question skipped recovery")
	}
	run, err = repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.PublishProposal(ctx, run, func(context.Context, domain.Run) (uuid.UUID, error) {
		t.Fatal("question duplicated on recovery")
		return uuid.Nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.AwaitApproval(ctx, run)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusAwaitingApproval || run.ProposalDeliveredAt == nil {
		t.Fatal("question delivery not acknowledged")
	}
	run, err = repo.Control(ctx, controlFor(run, run.Version), "approve")
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Execute(ctx, run, func(txCtx context.Context, current domain.Run) (domain.Outcome, error) {
		return domain.Outcome{Reply: "Closed the story."}, write(txCtx, current.ID, "action")
	})
	if err != nil {
		t.Fatal(err)
	}
	resultID := uuid.New()
	run, err = repo.Publish(ctx, run, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) {
		return resultID, write(txCtx, resultID, "result")
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Complete(ctx, run)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusSucceeded || run.ReplyCommentID == nil || *run.ReplyCommentID != resultID || run.ProposalReplyID == nil || *run.ProposalReplyID != questionID {
		t.Fatalf("question/result identities changed: %+v", run)
	}
	assertEffects(t, postgres.Pool, 3)
}

func TestInvocationTeammateContinuationConsumesOldApprovalWithoutInheritingIt(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	parent := acceptInvocation(t, repo, input)
	parent, err := repo.Claim(ctx, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	parent, err = repo.SaveProposal(ctx, parent, domain.Proposal{Reply: "Should I change priority?", Review: "Priority: Low → High", Action: json.RawMessage(`{}`), RequiresApproval: true})
	if err != nil {
		t.Fatal(err)
	}
	questionID := uuid.New()
	parent, err = repo.PublishProposal(ctx, parent, func(context.Context, domain.Run) (uuid.UUID, error) { return questionID, nil })
	if err != nil {
		t.Fatal(err)
	}
	parent, err = repo.AwaitApproval(ctx, parent)
	if err != nil {
		t.Fatal(err)
	}
	responder := uuid.New()
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO public.users (user_id,username,email,full_name) VALUES ($1,$2,$3,'Responder')`, []any{responder, responder.String(), responder.String() + "@example.com"}},
		{`INSERT INTO public.workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')`, []any{input.WorkspaceID, responder}},
		{`INSERT INTO public.team_members (team_id,user_id) VALUES ($1,$2)`, []any{input.TeamID, responder}},
	} {
		if _, err := postgres.Pool.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	childInput := input
	childInput.SourceID = uuid.New()
	childInput.RequesterID = responder
	childInput.IdempotencyKey = uuid.NewString()
	parentID, err := repo.ResolveContinuation(ctx, domain.ContinuationQuery{WorkspaceID: input.WorkspaceID, RequesterID: responder, Surface: input.Surface, EntityID: input.EntityID, ConversationID: input.ConversationID, SourceID: childInput.SourceID, ReplyToID: questionID})
	if err != nil {
		t.Fatal(err)
	}
	if parentID == nil || *parentID != parent.ID {
		t.Fatal("proposal reply did not resolve continuation")
	}
	childInput.ContinuationID = parentID
	child := acceptInvocation(t, repo, childInput)
	if child.RequesterID != responder || child.ApprovedAt != nil || child.Proposal != nil || child.Outcome != nil {
		t.Fatal("continuation inherited previous requester's authority")
	}
	parent, err = repo.Get(ctx, input.WorkspaceID, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if parent.Status != domain.StatusSucceeded {
		t.Fatal("old approval remained executable")
	}
	if _, err := repo.Control(ctx, controlFor(parent, parent.Version), "approve"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("superseded proposal approved: %v", err)
	}
}
