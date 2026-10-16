//go:build integration

package agentinvocationsrepository

import (
	"context"
	"errors"
	"sync"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestInvocationContinuationIsAtomicBoundAndReplaySafe(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	parent := clarificationInvocation(t, repo, input)
	childInput := input
	childInput.SourceID = uuid.New()
	childInput.SourceRevision = "2"
	childInput.SourceHash = "new-answer"
	childInput.IdempotencyKey = uuid.NewString()
	childInput.ContinuationID = &parent.ID
	query := domain.ContinuationQuery{WorkspaceID: input.WorkspaceID, RequesterID: input.RequesterID, Surface: input.Surface, EntityID: input.EntityID, ConversationID: input.ConversationID, ReplyToID: parent.SourceID, SourceID: childInput.SourceID}
	for _, replyTo := range []uuid.UUID{parent.SourceID, *parent.ReplyCommentID} {
		query.ReplyToID = replyTo
		resolved, err := repo.ResolveContinuation(ctx, query)
		if err != nil || resolved == nil || *resolved != parent.ID {
			t.Fatalf("resolve source/reply: %v %v", resolved, err)
		}
	}
	for _, test := range []struct {
		name   string
		change func(*domain.Input)
	}{
		{"workspace", func(input *domain.Input) { input.WorkspaceID = uuid.New() }},
		{"team", func(input *domain.Input) { input.TeamID = uuid.New() }},
		{"surface", func(input *domain.Input) { input.Surface = "native_message" }},
		{"entity", func(input *domain.Input) { input.EntityID = uuid.New() }},
		{"conversation", func(input *domain.Input) { input.ConversationID = uuid.New() }},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := childInput
			test.change(&invalid)
			if _, _, err := repo.Accept(ctx, invalid); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("mismatched continuation=%v", err)
			}
		})
	}
	unknownRequester := childInput
	unknownRequester.RequesterID = uuid.New()
	if _, _, err := repo.Accept(ctx, unknownRequester); !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("unknown requester continued clarification: %v", err)
	}
	rollback := errors.New("new source rolled back")
	var rolledBackID uuid.UUID
	err := database.WithinUnitOfWork(ctx, postgres.Pool, func(txCtx context.Context, _ pgx.Tx) error {
		child, created, err := repo.Accept(txCtx, childInput)
		if err != nil {
			return err
		}
		if !created {
			t.Fatal("first child replayed")
		}
		rolledBackID = child.ID
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	stillWaiting, err := repo.Get(ctx, input.WorkspaceID, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stillWaiting.Status != domain.StatusNeedsInput || stillWaiting.Version != parent.Version {
		t.Fatal("rollback completed clarification")
	}
	if _, err := repo.Get(ctx, input.WorkspaceID, rolledBackID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("rollback kept child: %v", err)
	}
	// Other teammates may answer a clarification, using their own authority.
	// Independent requests remain independent when no continuation is supplied.
	otherID := uuid.New()
	if _, err := postgres.Pool.Exec(ctx, `INSERT INTO public.users (user_id,username,email,full_name) VALUES ($1,$2,$3,'Other requester')`, otherID, otherID.String(), otherID.String()+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.Pool.Exec(ctx, `INSERT INTO public.workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')`, input.WorkspaceID, otherID); err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.Pool.Exec(ctx, `INSERT INTO public.team_members (team_id,user_id) VALUES ($1,$2)`, input.TeamID, otherID); err != nil {
		t.Fatal(err)
	}
	independent := childInput
	independent.RequesterID = otherID
	independent.ContinuationID = nil
	independent.IdempotencyKey = uuid.NewString()
	independent.SourceID = uuid.New()
	otherRun := acceptInvocation(t, repo, independent)
	if otherRun.ParentRunID != nil {
		t.Fatal("cross-author request was linked")
	}
	stillWaiting, err = repo.Get(ctx, input.WorkspaceID, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stillWaiting.Status != domain.StatusNeedsInput {
		t.Fatal("cross-author request completed clarification")
	}
	childInput.RequesterID = otherID
	query.RequesterID = otherID
	if resolved, err := repo.ResolveContinuation(ctx, query); err != nil || resolved == nil || *resolved != parent.ID {
		t.Fatalf("teammate could not resolve clarification: %v %v", resolved, err)
	}
	child := acceptInvocation(t, repo, childInput)
	if child.ParentRunID == nil || *child.ParentRunID != parent.ID || child.RequesterID != otherID || child.Status != domain.StatusQueued || child.ApprovedAt != nil || child.Proposal != nil || child.Outcome != nil {
		t.Fatalf("new child inherited execution state: %+v", child)
	}
	continued, err := repo.Get(ctx, input.WorkspaceID, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if continued.RequesterID != input.RequesterID || continued.Status != domain.StatusSucceeded || continued.Summary != "Continued in a new request." || continued.Version != parent.Version+1 {
		t.Fatalf("clarification not completed: %+v", continued)
	}
	replayed, created, err := repo.Accept(ctx, childInput)
	if err != nil || created || replayed.ID != child.ID {
		t.Fatalf("child replay created=%v error=%v", created, err)
	}
	replayedParent, err := repo.Get(ctx, input.WorkspaceID, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if replayedParent.Version != continued.Version {
		t.Fatal("replay transitioned parent again")
	}
	resolved, err := repo.ResolveContinuation(ctx, query)
	if err != nil || resolved == nil || *resolved != parent.ID {
		t.Fatalf("replay lost original continuation: %v %v", resolved, err)
	}
	query.SourceID = uuid.New()
	if resolved, err := repo.ResolveContinuation(ctx, query); err != nil || resolved != nil {
		t.Fatalf("new reply reused completed clarification: %v %v", resolved, err)
	}
	query.SourceID = independent.SourceID
	query.RequesterID = otherID
	if resolved, err := repo.ResolveContinuation(ctx, query); err != nil || resolved != nil {
		t.Fatalf("independent replay gained parent linkage: %v %v", resolved, err)
	}
	withoutBinding := childInput
	withoutBinding.ContinuationID = nil
	if _, _, err := repo.Accept(ctx, withoutBinding); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("idempotency ignored continuation binding: %v", err)
	}
	second := childInput
	second.SourceID = uuid.New()
	second.IdempotencyKey = uuid.NewString()
	if _, _, err := repo.Accept(ctx, second); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second continuation accepted: %v", err)
	}
}

func TestInvocationClarificationAcceptsOnlyOneConcurrentContinuation(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	parent := clarificationInvocation(t, repo, input)
	var group sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		group.Go(func() {
			child := input
			child.SourceID = uuid.New()
			child.IdempotencyKey = uuid.NewString()
			child.ContinuationID = &parent.ID
			_, created, err := repo.Accept(t.Context(), child)
			if err == nil && !created {
				err = errors.New("unexpected replay")
			}
			results <- err
		})
	}
	group.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, domain.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	runs, err := repo.List(t.Context(), input.WorkspaceID, input.EntityID)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatalf("unexpected persisted runs=%d", len(runs))
	}
}

func clarificationInvocation(t *testing.T, repo *Repository, input domain.Input) domain.Run {
	t.Helper()
	run := acceptInvocation(t, repo, input)
	run, err := repo.Claim(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.SaveProposal(t.Context(), run, domain.Proposal{Reply: "Which status did you mean?", NeedsInput: true})
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Execute(t.Context(), run, func(_ context.Context, current domain.Run) (domain.Outcome, error) {
		return domain.Outcome{Reply: current.Proposal.Reply}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Publish(t.Context(), run, func(context.Context, domain.Run) (uuid.UUID, error) { return uuid.New(), nil })
	if err != nil {
		t.Fatal(err)
	}
	run, err = repo.Complete(t.Context(), run)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusNeedsInput {
		t.Fatal(run.Status)
	}
	return run
}
