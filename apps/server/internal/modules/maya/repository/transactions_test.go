package mayarepository

import (
	"context"
	"errors"
	"testing"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestCreateActionsUsesOneTransactionAndPreservesOrder(t *testing.T) {
	t.Parallel()

	queries := &transactionQueryStub{}
	repository := newWithQueries(queries)
	transactionCalls := 0
	repository.withinTransaction = func(
		_ context.Context,
		_ pgx.TxOptions,
		operation func(mayasql.Querier) error,
	) error {
		transactionCalls++
		return operation(queries)
	}
	runID := uuid.New()
	workspaceID := uuid.New()
	storyID := uuid.New()
	actions := []mayadomain.CoreAction{
		{
			RunID: runID, WorkspaceID: workspaceID, StoryID: storyID,
			Type: mayadomain.ActionTypeAssignStory, Status: mayadomain.ActionStatusProposed,
			Reason: "first", Payload: mayadomain.ActionPayload{AssignStory: &mayadomain.AssignStoryPayload{AssigneeID: uuid.New()}},
		},
		{
			RunID: runID, WorkspaceID: workspaceID, StoryID: storyID,
			Type: mayadomain.ActionTypeFlagScheduleRisk, Status: mayadomain.ActionStatusProposed,
			Reason: "second", Payload: mayadomain.ActionPayload{Risk: &mayadomain.RiskPayload{Code: "capacity"}},
		},
	}

	created, err := repository.CreateActions(context.Background(), actions)
	if err != nil {
		t.Fatalf("CreateActions() error = %v", err)
	}
	if transactionCalls != 1 {
		t.Fatalf("CreateActions() transaction calls = %d, want 1", transactionCalls)
	}
	if len(queries.createdActions) != 2 || len(created) != 2 {
		t.Fatalf("CreateActions() created %d/%d actions, want 2/2", len(queries.createdActions), len(created))
	}
	if queries.createdActions[0].Reason != "first" || queries.createdActions[1].Reason != "second" {
		t.Fatalf("CreateActions() order = %q, %q", queries.createdActions[0].Reason, queries.createdActions[1].Reason)
	}
}

func TestCompleteInterruptedScheduleRunKeepsActionAndRunUpdatesAtomic(t *testing.T) {
	t.Parallel()

	queries := &transactionQueryStub{}
	repository := newWithQueries(queries)
	transactionCalls := 0
	repository.withinTransaction = func(
		_ context.Context,
		_ pgx.TxOptions,
		operation func(mayasql.Querier) error,
	) error {
		transactionCalls++
		return operation(queries)
	}
	runID := uuid.New()

	if err := repository.CompleteInterruptedScheduleRun(context.Background(), runID, "recovered"); err != nil {
		t.Fatalf("CompleteInterruptedScheduleRun() error = %v", err)
	}
	if transactionCalls != 1 || len(queries.operations) != 2 {
		t.Fatalf("transaction calls/operations = %d/%v, want 1/[fail-actions complete-run]", transactionCalls, queries.operations)
	}
	if queries.operations[0] != "fail-actions" || queries.operations[1] != "complete-run" {
		t.Fatalf("operation order = %v, want [fail-actions complete-run]", queries.operations)
	}
}

func TestCompleteInterruptedScheduleRunStopsAfterActionFailure(t *testing.T) {
	t.Parallel()

	queries := &transactionQueryStub{failActionsErr: errQueryFailed}
	repository := newWithQueries(queries)
	err := repository.CompleteInterruptedScheduleRun(context.Background(), uuid.New(), "recovered")
	if !errors.Is(err, errQueryFailed) {
		t.Fatalf("CompleteInterruptedScheduleRun() error = %v, want %v", err, errQueryFailed)
	}
	if len(queries.operations) != 1 || queries.operations[0] != "fail-actions" {
		t.Fatalf("operations after failure = %v, want [fail-actions]", queries.operations)
	}
}

func TestWithScheduleStoryLockAcquiresLockBeforeCallback(t *testing.T) {
	t.Parallel()

	queries := &transactionQueryStub{}
	repository := newWithQueries(queries)
	callbackCalls := 0
	err := repository.WithScheduleStoryLock(context.Background(), uuid.New(), uuid.New(), func() error {
		callbackCalls++
		if len(queries.operations) != 1 || queries.operations[0] != "lock-story" {
			t.Fatalf("callback observed operations = %v, want lock first", queries.operations)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithScheduleStoryLock() error = %v", err)
	}
	if callbackCalls != 1 {
		t.Fatalf("WithScheduleStoryLock() callback calls = %d, want 1", callbackCalls)
	}
}

type transactionQueryStub struct {
	mayasql.Querier
	createdActions []mayasql.CreateMayaActionParams
	operations     []string
	failActionsErr error
}

func (q *transactionQueryStub) CreateMayaAction(
	_ context.Context,
	params mayasql.CreateMayaActionParams,
) (mayasql.MayaAgentAction, error) {
	q.createdActions = append(q.createdActions, params)
	return mayasql.MayaAgentAction{
		ActionID: uuid.New(), RunID: params.RunID, WorkspaceID: params.WorkspaceID,
		StoryID: params.StoryID, ActionType: params.ActionType, Status: params.Status,
		Reason: params.Reason, Payload: params.Payload, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, nil
}

func (q *transactionQueryStub) FailInterruptedScheduleActions(
	_ context.Context,
	_ mayasql.FailInterruptedScheduleActionsParams,
) (int64, error) {
	q.operations = append(q.operations, "fail-actions")
	return 1, q.failActionsErr
}

func (q *transactionQueryStub) CompleteInterruptedScheduleRun(
	_ context.Context,
	_ mayasql.CompleteInterruptedScheduleRunParams,
) (int64, error) {
	q.operations = append(q.operations, "complete-run")
	return 1, nil
}

func (q *transactionQueryStub) LockMayaStorySchedule(
	_ context.Context,
	_ mayasql.LockMayaStoryScheduleParams,
) error {
	q.operations = append(q.operations, "lock-story")
	return nil
}
