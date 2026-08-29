package mayarepository

import (
	"context"
	"errors"
	"testing"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/google/uuid"
)

func TestMarkActionAppliedRequiresOneProposedAction(t *testing.T) {
	t.Parallel()

	actionID := uuid.New()
	tests := []struct {
		name     string
		rows     int64
		queryErr error
		wantErr  error
	}{
		{name: "updated", rows: 1},
		{name: "already terminal", rows: 0, wantErr: mayadomain.ErrActionNotProposed},
		{name: "query failure", queryErr: errQueryFailed, wantErr: errQueryFailed},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			queries := &mayaQueryStub{appliedRows: test.rows, appliedErr: test.queryErr}
			err := newWithQueries(queries).MarkActionApplied(context.Background(), actionID)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("MarkActionApplied() error = %v, want %v", err, test.wantErr)
			}
			if queries.appliedID != actionID {
				t.Fatalf("MarkActionApplied() action = %s, want %s", queries.appliedID, actionID)
			}
		})
	}
}

func TestMarkActionFailedRequiresOneProposedAction(t *testing.T) {
	t.Parallel()

	actionID := uuid.New()
	queries := &mayaQueryStub{failedRows: 0}
	err := newWithQueries(queries).MarkActionFailed(context.Background(), actionID, "application failed")
	if !errors.Is(err, mayadomain.ErrActionNotProposed) {
		t.Fatalf("MarkActionFailed() error = %v, want %v", err, mayadomain.ErrActionNotProposed)
	}
	if queries.failedID != actionID || queries.failedMessage != "application failed" {
		t.Fatalf("MarkActionFailed() persisted (%s, %q), want (%s, %q)", queries.failedID, queries.failedMessage, actionID, "application failed")
	}
}

func TestUnconfiguredRepositoryFailsClosed(t *testing.T) {
	t.Parallel()

	_, err := (*Repo)(nil).CreateRun(context.Background(), mayadomain.CreateRunInput{})
	if !errors.Is(err, mayadomain.ErrPersistenceNotConfigured) {
		t.Fatalf("CreateRun() error = %v, want %v", err, mayadomain.ErrPersistenceNotConfigured)
	}
	if err := (&Repo{}).MarkActionApplied(context.Background(), uuid.New()); !errors.Is(err, mayadomain.ErrPersistenceNotConfigured) {
		t.Fatalf("MarkActionApplied() error = %v, want %v", err, mayadomain.ErrPersistenceNotConfigured)
	}
}

func TestWorkspaceCanUseMayaDelegatesToTypedEntitlementQuery(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	queries := &mayaQueryStub{workspaceAccess: true}

	hasAccess, err := newWithQueries(queries).WorkspaceCanUseMaya(context.Background(), workspaceID)

	if err != nil {
		t.Fatalf("WorkspaceCanUseMaya() error = %v", err)
	}
	if !hasAccess {
		t.Fatal("WorkspaceCanUseMaya() = false, want true")
	}
	if queries.workspaceAccessParams.WorkspaceID != workspaceID {
		t.Fatalf(
			"WorkspaceCanUseMaya() workspace = %s, want %s",
			queries.workspaceAccessParams.WorkspaceID,
			workspaceID,
		)
	}
}

var errQueryFailed = errors.New("query failed")

type mayaQueryStub struct {
	mayasql.Querier
	appliedRows           int64
	appliedErr            error
	appliedID             uuid.UUID
	failedRows            int64
	failedErr             error
	failedID              uuid.UUID
	failedMessage         string
	workspaceAccess       bool
	workspaceAccessErr    error
	workspaceAccessParams mayasql.WorkspaceCanUseMayaParams
}

func (q *mayaQueryStub) MarkMayaActionApplied(
	_ context.Context,
	params mayasql.MarkMayaActionAppliedParams,
) (int64, error) {
	q.appliedID = params.ActionID
	return q.appliedRows, q.appliedErr
}

func (q *mayaQueryStub) MarkMayaActionFailed(
	_ context.Context,
	params mayasql.MarkMayaActionFailedParams,
) (int64, error) {
	q.failedID = params.ActionID
	if params.ErrorMessage != nil {
		q.failedMessage = *params.ErrorMessage
	}
	return q.failedRows, q.failedErr
}

func (q *mayaQueryStub) WorkspaceCanUseMaya(
	_ context.Context,
	params mayasql.WorkspaceCanUseMayaParams,
) (bool, error) {
	q.workspaceAccessParams = params
	return q.workspaceAccess, q.workspaceAccessErr
}
