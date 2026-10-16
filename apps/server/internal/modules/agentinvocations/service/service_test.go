package agentinvocations

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type transactionMarker struct{}

type lifecycleStore struct {
	Store
	run                                       Run
	saveError                                 error
	executeCalls, publishCalls, completeCalls int
	proposalPublishCalls                      int
	failureRun                                Run
	failureSummary                            string
	retry                                     bool
}

func (s *lifecycleStore) Claim(context.Context, uuid.UUID) (Run, error) {
	s.run.Status = StatusRunning
	s.run.AttemptCount++
	s.run.RetryAttemptCount++
	return s.run, nil
}
func (s *lifecycleStore) SaveProposal(_ context.Context, run Run, proposal Proposal) (Run, error) {
	if s.saveError != nil {
		return Run{}, s.saveError
	}
	run.Proposal = &proposal
	s.run = run
	return run, nil
}
func (s *lifecycleStore) Execute(ctx context.Context, run Run, callback func(context.Context, Run) (Outcome, error)) (Run, error) {
	s.executeCalls++
	outcome, err := callback(context.WithValue(ctx, transactionMarker{}, true), run)
	if err != nil {
		return Run{}, err
	}
	run.Outcome = &outcome
	s.run = run
	return run, nil
}
func (s *lifecycleStore) Publish(ctx context.Context, run Run, callback func(context.Context, Run) (uuid.UUID, error)) (Run, error) {
	s.publishCalls++
	id, err := callback(context.WithValue(ctx, transactionMarker{}, true), run)
	if err != nil {
		return Run{}, err
	}
	run.ReplyCommentID = &id
	s.run = run
	return run, nil
}
func (s *lifecycleStore) Complete(_ context.Context, run Run) (Run, error) {
	s.completeCalls++
	run.Status = StatusSucceeded
	if run.Proposal.NeedsInput {
		run.Status = StatusNeedsInput
	}
	s.run = run
	return run, nil
}

func (s *lifecycleStore) PublishProposal(ctx context.Context, run Run, callback func(context.Context, Run) (uuid.UUID, error)) (Run, error) {
	s.proposalPublishCalls++
	id, err := callback(context.WithValue(ctx, transactionMarker{}, true), run)
	if err != nil {
		return Run{}, err
	}
	run.ProposalReplyID = &id
	s.run = run
	return run, nil
}

func (s *lifecycleStore) AwaitApproval(_ context.Context, run Run) (Run, error) {
	now := time.Now()
	run.ProposalDeliveredAt = &now
	run.Status = StatusAwaitingApproval
	s.run = run
	return run, nil
}
func (s *lifecycleStore) Fail(_ context.Context, run Run, summary string, retry bool) error {
	s.run.FailurePending = false
	s.failureRun = run
	s.failureSummary = summary
	s.retry = retry
	if retry && (run.RetryAttemptCount < 3 || run.HasDurableResponse()) {
		s.run.Status = StatusQueued
	} else {
		s.run.Status = StatusFailed
	}
	return nil
}

type lifecycleRuntime struct {
	proposal                                                                                         Proposal
	authorizations, transactionAuthorizations, prepareCalls, applyCalls, publishCalls, deliveryCalls int
	denyTransaction                                                                                  bool
	prepareError, deliveryError                                                                      error
	authorizationError, publishError                                                                 error
	proposalPublishCalls                                                                             int
}

func (r *lifecycleRuntime) Authorize(ctx context.Context, _ Run) error {
	r.authorizations++
	if r.authorizationError != nil {
		return r.authorizationError
	}
	if ctx.Value(transactionMarker{}) == true {
		r.transactionAuthorizations++
		if r.denyTransaction {
			return ErrPermissionDenied
		}
	}
	return nil
}
func (r *lifecycleRuntime) Prepare(context.Context, Run) (Proposal, error) {
	r.prepareCalls++
	return r.proposal, r.prepareError
}
func (r *lifecycleRuntime) Apply(ctx context.Context, _ Run, _ Proposal) (Outcome, error) {
	r.applyCalls++
	if ctx.Value(transactionMarker{}) != true {
		return Outcome{}, errors.New("apply escaped transaction")
	}
	return Outcome{Reply: "Updated.", Receipt: json.RawMessage(`{"applied":true}`)}, nil
}
func (r *lifecycleRuntime) Publish(ctx context.Context, _ Run, _ Outcome) (uuid.UUID, error) {
	r.publishCalls++
	if ctx.Value(transactionMarker{}) != true {
		return uuid.Nil, errors.New("publish escaped transaction")
	}
	if r.publishError != nil {
		return uuid.Nil, r.publishError
	}
	return uuid.New(), nil
}

func (r *lifecycleRuntime) PublishProposal(ctx context.Context, _ Run, _ Proposal) (uuid.UUID, error) {
	r.proposalPublishCalls++
	if ctx.Value(transactionMarker{}) != true {
		return uuid.Nil, errors.New("proposal publication escaped transaction")
	}
	return uuid.New(), nil
}
func (r *lifecycleRuntime) Deliver(ctx context.Context, run Run, _ Outcome, replyID uuid.UUID) error {
	r.deliveryCalls++
	committed := (run.ReplyCommentID != nil && *run.ReplyCommentID == replyID) || (run.ProposalReplyID != nil && *run.ProposalReplyID == replyID) || (run.FailureReplyID != nil && *run.FailureReplyID == replyID)
	if ctx.Value(transactionMarker{}) != nil || !committed {
		return errors.New("delivery ran before reply committed")
	}
	return r.deliveryError
}

func TestQuestionAndClarificationAuthorizeInsidePublication(t *testing.T) {
	for _, needsInput := range []bool{false, true} {
		store := &lifecycleStore{run: Run{ID: uuid.New()}}
		runtime := &lifecycleRuntime{proposal: Proposal{Reply: "Story context", NeedsInput: needsInput}}
		if err := New(store, runtime).Process(t.Context(), store.run.ID); err != nil {
			t.Fatal(err)
		}
		want := StatusSucceeded
		if needsInput {
			want = StatusNeedsInput
		}
		if store.run.Status != want || runtime.applyCalls != 0 || runtime.publishCalls != 1 || runtime.transactionAuthorizations != 2 || runtime.deliveryCalls != 1 {
			t.Fatalf("unexpected lifecycle: status=%s runtime=%+v", store.run.Status, runtime)
		}
	}
}

func TestApprovalPublishesQuestionBeforePausingAction(t *testing.T) {
	store := &lifecycleStore{run: Run{ID: uuid.New()}}
	runtime := &lifecycleRuntime{proposal: Proposal{Summary: "Close story", Reply: "Should I close this story?", Review: "Status: Open → Closed", Action: json.RawMessage(`{"token":"opaque"}`), RequiresApproval: true}}
	if err := New(store, runtime).Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if store.run.Status != StatusAwaitingApproval || store.executeCalls != 0 || store.publishCalls != 0 || runtime.deliveryCalls != 1 || runtime.proposalPublishCalls != 1 || store.run.ProposalDeliveredAt == nil {
		t.Fatalf("approval did not pause: %+v", store)
	}
}

func TestApprovalDeliveryRecoveryKeepsQuestionSeparateFromResult(t *testing.T) {
	store := &lifecycleStore{run: Run{ID: uuid.New()}}
	runtime := &lifecycleRuntime{proposal: Proposal{Reply: "Should I close it?", Review: "Status: Open → Closed", Action: json.RawMessage(`{}`), RequiresApproval: true}, deliveryError: errors.New("temporary delivery failure")}
	service := New(store, runtime)
	if err := service.Process(t.Context(), store.run.ID); err == nil {
		t.Fatal("expected delivery failure")
	}
	if store.run.ProposalReplyID == nil || store.run.ProposalDeliveredAt != nil || store.run.Outcome != nil || store.run.ReplyCommentID != nil || runtime.applyCalls != 0 {
		t.Fatalf("approval crossed action boundary: %+v", store.run)
	}
	questionID := *store.run.ProposalReplyID
	runtime.deliveryError = nil
	if err := service.Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if runtime.prepareCalls != 1 || runtime.proposalPublishCalls != 1 || runtime.deliveryCalls != 2 || store.run.Status != StatusAwaitingApproval {
		t.Fatalf("question recovery replayed work: %+v", runtime)
	}
	approved := time.Now()
	store.run.ApprovedAt = &approved
	if err := service.Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if runtime.applyCalls != 1 || runtime.publishCalls != 1 || runtime.proposalPublishCalls != 1 || store.run.ReplyCommentID == nil || *store.run.ReplyCommentID == questionID || store.run.Status != StatusSucceeded {
		t.Fatalf("result reused approval question: %+v", store.run)
	}
}

func TestDeliveryRecoveryReusesActionAndReply(t *testing.T) {
	store := &lifecycleStore{run: Run{ID: uuid.New()}}
	runtime := &lifecycleRuntime{proposal: Proposal{Action: json.RawMessage(`{"token":"opaque"}`)}, deliveryError: errors.New("temporary transport failure")}
	service := New(store, runtime)
	if err := service.Process(t.Context(), store.run.ID); err == nil {
		t.Fatal("expected delivery failure")
	}
	if store.run.Outcome == nil || store.run.ReplyCommentID == nil || store.completeCalls != 0 || !store.retry {
		t.Fatalf("lost durable stages: %+v", store)
	}
	replyID := *store.run.ReplyCommentID
	runtime.deliveryError = nil
	if err := service.Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if store.run.Status != StatusSucceeded || runtime.prepareCalls != 1 || runtime.applyCalls != 1 || runtime.publishCalls != 1 || runtime.deliveryCalls != 2 || *store.run.ReplyCommentID != replyID {
		t.Fatalf("recovery duplicated committed work: %+v", runtime)
	}
}

func TestRevocationInsideTransactionPreventsAction(t *testing.T) {
	store := &lifecycleStore{run: Run{ID: uuid.New()}}
	runtime := &lifecycleRuntime{proposal: Proposal{Action: json.RawMessage(`{}`)}, denyTransaction: true}
	if err := New(store, runtime).Process(t.Context(), store.run.ID); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("error=%v", err)
	}
	if runtime.applyCalls != 0 || runtime.publishCalls != 0 || store.retry || store.run.Status != StatusFailed {
		t.Fatalf("revoked work executed: %+v", runtime)
	}
}

func TestProposalSaveFailurePreservesFenceAndHidesRawError(t *testing.T) {
	token := uuid.New()
	store := &lifecycleStore{run: Run{ID: uuid.New(), LeaseToken: &token, LeaseGeneration: 9}, saveError: errors.New("secret provider key")}
	runtime := &lifecycleRuntime{proposal: Proposal{Reply: "Hello"}}
	if err := New(store, runtime).Process(t.Context(), store.run.ID); err == nil {
		t.Fatal("expected failure")
	}
	if store.failureRun.LeaseToken == nil || *store.failureRun.LeaseToken != token || store.failureRun.LeaseGeneration != 9 {
		t.Fatal("failure lost claimed fence")
	}
	if strings.Contains(store.failureSummary, "secret") || store.failureSummary == "" {
		t.Fatalf("unsafe summary %q", store.failureSummary)
	}
}

func TestLostLeaseDoesNotFailOrPublishCancelledRun(t *testing.T) {
	store := &lifecycleStore{run: Run{ID: uuid.New()}, saveError: ErrLeaseLost}
	runtime := &lifecycleRuntime{proposal: Proposal{Reply: "Hello"}}
	if err := New(store, runtime).Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if store.failureSummary != "" || runtime.publishCalls != 0 {
		t.Fatal("lost lease wrote a result")
	}
}
