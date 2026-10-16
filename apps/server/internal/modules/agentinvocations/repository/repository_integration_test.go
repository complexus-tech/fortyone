//go:build integration

package agentinvocationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInvocationAcceptanceIsAtomicIdempotentAndTenantScoped(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	rollback := errors.New("source creation rolled back")
	var rolledBackID uuid.UUID
	err := database.WithinUnitOfWork(ctx, postgres.Pool, func(txCtx context.Context, _ pgx.Tx) error {
		run, created, err := repo.Accept(txCtx, input)
		if err != nil {
			return err
		}
		if !created {
			t.Fatal("first accept replayed")
		}
		rolledBackID = run.ID
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, input.WorkspaceID, rolledBackID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("rollback left invocation: %v", err)
	}
	first, created, err := repo.Accept(ctx, input)
	if err != nil || !created {
		t.Fatalf("accept: created=%v error=%v", created, err)
	}
	replay, created, err := repo.Accept(ctx, input)
	if err != nil || created || replay.ID != first.ID {
		t.Fatalf("replay: created=%v error=%v", created, err)
	}
	metadataChanged := input
	metadataChanged.Payload = json.RawMessage(`{"request":"display metadata changed"}`)
	replay, created, err = repo.Accept(ctx, metadataChanged)
	if err != nil || created || replay.ID != first.ID || string(replay.Payload) != string(first.Payload) {
		t.Fatalf("metadata replay changed original snapshot: created=%v error=%v payload=%s", created, err, replay.Payload)
	}
	changed := input
	changed.SourceHash = "different"
	if _, _, err := repo.Accept(ctx, changed); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("changed replay=%v", err)
	}
	if _, err := repo.Get(ctx, uuid.New(), first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("cross-tenant get=%v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `DELETE FROM public.team_members WHERE team_id=$1 AND user_id=$2`, input.TeamID, input.RequesterID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.Accept(ctx, input); !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("removed member replay=%v", err)
	}
	changed = input
	changed.IdempotencyKey = uuid.NewString()
	if _, _, err := repo.Accept(ctx, changed); !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("removed member insert=%v", err)
	}
}

func TestInvocationLeaseFencesCancelledAndExpiredWorkers(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := acceptInvocation(t, repo, input)
	var group sync.WaitGroup
	claims := make(chan domain.Run, 2)
	failures := make(chan error, 2)
	for range 2 {
		group.Go(func() {
			claimed, err := repo.Claim(ctx, run.ID)
			if err != nil {
				failures <- err
			} else {
				claims <- claimed
			}
		})
	}
	group.Wait()
	close(claims)
	close(failures)
	if len(claims) != 1 || len(failures) != 1 {
		t.Fatalf("claims=%d failures=%d", len(claims), len(failures))
	}
	if err := <-failures; !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatal(err)
	}
	claimed := <-claims
	if _, err := repo.Control(ctx, domain.Control{RunID: run.ID, WorkspaceID: input.WorkspaceID, RequesterID: uuid.New(), ExpectedVersion: claimed.Version}, "cancel"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("foreign control=%v", err)
	}
	if _, err := repo.Control(ctx, controlFor(claimed, claimed.Version-1), "cancel"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale control=%v", err)
	}
	cancelled, err := repo.Control(ctx, controlFor(claimed, claimed.Version), "cancel")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Fatal(cancelled.Status)
	}
	if _, err := repo.SaveProposal(ctx, claimed, domain.Proposal{Reply: "Too late"}); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("cancelled worker saved proposal: %v", err)
	}
	called := false
	if _, err := repo.Execute(ctx, claimed, func(context.Context, domain.Run) (domain.Outcome, error) {
		called = true
		return domain.Outcome{Reply: "Too late"}, nil
	}); !errors.Is(err, domain.ErrLeaseLost) || called {
		t.Fatalf("cancelled execute: called=%v error=%v", called, err)
	}
	if _, err := repo.Control(ctx, controlFor(cancelled, cancelled.Version), "retry"); err != nil {
		t.Fatal(err)
	}
	old, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.Pool.Exec(ctx, `UPDATE public.agent_invocations SET lease_expires_at=clock_timestamp()-INTERVAL '1 second' WHERE id=$1`, run.ID); err != nil {
		t.Fatal(err)
	}
	ids, err := repo.Pending(ctx, 10)
	if err != nil || len(ids) != 1 || ids[0] != run.ID {
		t.Fatalf("dead lease not recovered: %v %v", ids, err)
	}
	recovered, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.LeaseGeneration <= old.LeaseGeneration || *recovered.LeaseToken == *old.LeaseToken {
		t.Fatal("recovery reused fencing identity")
	}
	if _, err := repo.SaveProposal(ctx, old, domain.Proposal{Reply: "Old worker"}); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("expired worker saved proposal: %v", err)
	}
}

func TestInvocationActionReceiptAndReplyCommitExactlyOnce(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := acceptInvocation(t, repo, input)
	claimed, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := repo.SaveProposal(ctx, claimed, domain.Proposal{Summary: "Update story", Reply: "Should I close this story?", Review: "Status: Open → Closed", Action: json.RawMessage(`{"action":"update"}`), RequiresApproval: true})
	if err != nil {
		t.Fatal(err)
	}
	question, err := repo.PublishProposal(ctx, prepared, func(context.Context, domain.Run) (uuid.UUID, error) { return uuid.New(), nil })
	if err != nil {
		t.Fatal(err)
	}
	paused, err := repo.AwaitApproval(ctx, question)
	if err != nil {
		t.Fatal(err)
	}
	if paused.Status != domain.StatusAwaitingApproval {
		t.Fatal(paused.Status)
	}
	if _, err := repo.Claim(ctx, run.ID); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("unapproved run claimed: %v", err)
	}
	cancelled, err := repo.Control(ctx, controlFor(paused, paused.Version), "cancel")
	if err != nil {
		t.Fatal(err)
	}
	paused, err = repo.Control(ctx, controlFor(cancelled, cancelled.Version), "retry")
	if err != nil {
		t.Fatal(err)
	}
	if paused.Status != domain.StatusAwaitingApproval {
		t.Fatal("retry bypassed approval")
	}
	if _, err := repo.Control(ctx, controlFor(paused, paused.Version), "approve"); err != nil {
		t.Fatal(err)
	}
	claimed, err = repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.ApprovedAt == nil {
		t.Fatal("approval not persisted")
	}
	if _, err := postgres.Pool.Exec(ctx, `CREATE TABLE invocation_effects (id uuid PRIMARY KEY, kind text NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	writeEffect := func(txCtx context.Context, id uuid.UUID, kind string) error {
		tx, ok := database.UnitTransaction(txCtx)
		if !ok {
			return errors.New("missing unit of work")
		}
		_, err := tx.Exec(txCtx, `INSERT INTO invocation_effects (id,kind) VALUES ($1,$2)`, id, kind)
		return err
	}
	rollback := errors.New("action failed after write")
	if _, err := repo.Execute(ctx, claimed, func(txCtx context.Context, current domain.Run) (domain.Outcome, error) {
		if err := writeEffect(txCtx, current.ID, "action"); err != nil {
			return domain.Outcome{}, err
		}
		return domain.Outcome{}, rollback
	}); !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	assertEffects(t, postgres.Pool, 0)
	if _, err := repo.Execute(ctx, claimed, func(txCtx context.Context, current domain.Run) (domain.Outcome, error) {
		if err := writeEffect(txCtx, current.ID, "expired action"); err != nil {
			return domain.Outcome{}, err
		}
		tx, _ := database.UnitTransaction(txCtx)
		if _, err := tx.Exec(txCtx, `UPDATE public.agent_invocations SET lease_expires_at=clock_timestamp()-INTERVAL '1 second' WHERE id=$1`, current.ID); err != nil {
			return domain.Outcome{}, err
		}
		return domain.Outcome{Reply: "Expired while applying"}, nil
	}); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("expired action committed: %v", err)
	}
	assertEffects(t, postgres.Pool, 0)
	applied, err := repo.Execute(ctx, claimed, func(txCtx context.Context, current domain.Run) (domain.Outcome, error) {
		if err := writeEffect(txCtx, current.ID, "action"); err != nil {
			return domain.Outcome{}, err
		}
		return domain.Outcome{Reply: "Done", Receipt: json.RawMessage(`{"changed":true}`)}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if applied.Outcome == nil {
		t.Fatal("receipt missing")
	}
	assertEffects(t, postgres.Pool, 1)
	if _, err := repo.Control(ctx, controlFor(applied, applied.Version), "cancel"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("cancel hid committed action: %v", err)
	}
	if err := repo.Fail(ctx, applied, "Retry publication", false); err != nil {
		t.Fatal(err)
	}
	failed, err := repo.Get(ctx, input.WorkspaceID, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Control(ctx, controlFor(failed, failed.Version), "retry"); err != nil {
		t.Fatal(err)
	}
	recovered, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err = repo.Execute(ctx, recovered, func(context.Context, domain.Run) (domain.Outcome, error) {
		t.Fatal("action replayed after commit")
		return domain.Outcome{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	replyID := uuid.New()
	if _, err := repo.Publish(ctx, recovered, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) {
		if err := writeEffect(txCtx, replyID, "reply"); err != nil {
			return uuid.Nil, err
		}
		return uuid.Nil, rollback
	}); !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	assertEffects(t, postgres.Pool, 1)
	published, err := repo.Publish(ctx, recovered, func(txCtx context.Context, _ domain.Run) (uuid.UUID, error) {
		return replyID, writeEffect(txCtx, replyID, "reply")
	})
	if err != nil {
		t.Fatal(err)
	}
	if published.ReplyCommentID == nil || *published.ReplyCommentID != replyID {
		t.Fatal("reply ID missing")
	}
	assertEffects(t, postgres.Pool, 2)
	published, err = repo.Publish(ctx, published, func(context.Context, domain.Run) (uuid.UUID, error) {
		t.Fatal("reply replayed after commit")
		return uuid.Nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, err := repo.Complete(ctx, published)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != domain.StatusSucceeded || completed.LeaseToken != nil {
		t.Fatalf("completion=%+v", completed)
	}
	if completed.Summary != "Update story" {
		t.Fatalf("completion retained stale failure summary: %q", completed.Summary)
	}
	if _, err := repo.Claim(ctx, run.ID); !errors.Is(err, domain.ErrLeaseLost) {
		t.Fatalf("terminal run reclaimed: %v", err)
	}
}

func TestInvocationAutomaticRetriesAreBoundedAndManualRetryPreservesAttemptIdentity(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := acceptInvocation(t, repo, input)
	for attempt := int32(1); attempt <= maximumAttempts; attempt++ {
		claimed, err := repo.Claim(ctx, run.ID)
		if err != nil {
			t.Fatal(err)
		}
		if claimed.AttemptCount != attempt || claimed.RetryAttemptCount != attempt {
			t.Fatalf("attempt counters: %+v", claimed)
		}
		if err := repo.Fail(ctx, claimed, "Temporary failure", true); err != nil {
			t.Fatal(err)
		}
		run, err = repo.Get(ctx, input.WorkspaceID, run.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := domain.StatusQueued
		if attempt == maximumAttempts {
			want = domain.StatusFailed
		}
		if run.Status != want {
			t.Fatalf("attempt %d status=%s, want %s", attempt, run.Status, want)
		}
		if _, err := postgres.Pool.Exec(ctx, `UPDATE public.agent_invocations SET available_at=clock_timestamp() WHERE id=$1`, run.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := repo.Control(ctx, controlFor(run, run.Version), "retry"); err != nil {
		t.Fatal(err)
	}
	claimed, err := repo.Claim(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.AttemptCount != maximumAttempts+1 || claimed.RetryAttemptCount != 1 {
		t.Fatalf("manual retry reused provider attempt identity: total=%d cycle=%d", claimed.AttemptCount, claimed.RetryAttemptCount)
	}
	if err := repo.Fail(ctx, claimed, "Temporary failure", true); err != nil {
		t.Fatal(err)
	}
	queued, err := repo.Get(ctx, input.WorkspaceID, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if queued.Status != domain.StatusQueued {
		t.Fatal("manual retry did not restart bounded retry cycle")
	}
}

func invocationFixture(t *testing.T, pool *pgxpool.Pool) domain.Input {
	t.Helper()
	input := domain.Input{WorkspaceID: uuid.New(), RequesterID: uuid.New(), TeamID: uuid.New(), Surface: "story_comment", SourceID: uuid.New(), ConversationID: uuid.New(), EntityID: uuid.New(), SourceRevision: "1", SourceHash: "source-hash", IdempotencyKey: uuid.NewString(), Payload: json.RawMessage(`{"request":"question"}`)}
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO public.workspaces (workspace_id,name,slug) VALUES ($1,'Agent tests',$2)`, []any{input.WorkspaceID, "agent-" + input.WorkspaceID.String()}},
		{`INSERT INTO public.users (user_id,username,email,full_name) VALUES ($1,$2,$3,'Agent requester')`, []any{input.RequesterID, input.RequesterID.String(), input.RequesterID.String() + "@example.com"}},
		{`INSERT INTO public.teams (team_id,workspace_id,name,code,color) VALUES ($1,$2,'Agent team','AGT','#000000')`, []any{input.TeamID, input.WorkspaceID}},
		{`INSERT INTO public.workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')`, []any{input.WorkspaceID, input.RequesterID}},
		{`INSERT INTO public.team_members (team_id,user_id) VALUES ($1,$2)`, []any{input.TeamID, input.RequesterID}},
	} {
		if _, err := pool.Exec(t.Context(), statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	return input
}

func acceptInvocation(t *testing.T, repo *Repository, input domain.Input) domain.Run {
	t.Helper()
	run, created, err := repo.Accept(t.Context(), input)
	if err != nil || !created {
		t.Fatalf("accept: created=%v error=%v", created, err)
	}
	return run
}

func controlFor(run domain.Run, version int64) domain.Control {
	return domain.Control{WorkspaceID: run.WorkspaceID, RequesterID: run.RequesterID, RunID: run.ID, ExpectedVersion: version}
}

func assertEffects(t *testing.T, pool *pgxpool.Pool, want int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM invocation_effects`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("effects=%d, want %d", count, want)
	}
}
