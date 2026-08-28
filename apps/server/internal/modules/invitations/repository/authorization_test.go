package invitationsrepository

import (
	"context"
	"errors"
	"testing"

	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type adminLockQueries struct {
	invitationsql.Querier
	lock func(context.Context, invitationsql.LockActiveWorkspaceAdminParams) (bool, error)
}

func (q adminLockQueries) LockActiveWorkspaceAdmin(
	ctx context.Context,
	params invitationsql.LockActiveWorkspaceAdminParams,
) (bool, error) {
	if q.lock == nil {
		panic("unexpected LockActiveWorkspaceAdmin call")
	}
	return q.lock(ctx, params)
}

func TestLockActiveWorkspaceAdminMapsDenialAndPreservesDatabaseErrors(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	databaseErr := errors.New("database unavailable")
	tests := []struct {
		name    string
		lockErr error
		wantErr error
	}{
		{name: "authorized"},
		{name: "not found is authorization denial", lockErr: pgx.ErrNoRows, wantErr: authorization.ErrWorkspaceAdminRequired},
		{name: "database error remains discoverable", lockErr: databaseErr, wantErr: databaseErr},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			queries := adminLockQueries{lock: func(_ context.Context, params invitationsql.LockActiveWorkspaceAdminParams) (bool, error) {
				if params.WorkspaceID != workspaceID || params.ActorID != actorID {
					t.Fatalf("authorization params = %s/%s, want %s/%s", params.WorkspaceID, params.ActorID, workspaceID, actorID)
				}
				return test.lockErr == nil, test.lockErr
			}}
			err := lockActiveWorkspaceAdmin(context.Background(), queries, workspaceID, actorID)
			if test.wantErr == nil && err != nil {
				t.Fatalf("lock active admin error = %v, want nil", err)
			}
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("lock active admin error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestLockInvitationWorkspaceAdminsRejectsActorMismatchBeforeSQL(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	called := false
	queries := adminLockQueries{lock: func(context.Context, invitationsql.LockActiveWorkspaceAdminParams) (bool, error) {
		called = true
		return true, nil
	}}
	err := lockInvitationWorkspaceAdmins(context.Background(), queries, actorID, []invitationsdomain.NewWorkspaceInvitation{{
		Invitation: invitationsdomain.WorkspaceInvitation{WorkspaceID: uuid.New(), InviterID: uuid.New()},
	}})
	if !errors.Is(err, authorization.ErrWorkspaceAdminRequired) {
		t.Fatalf("actor mismatch error = %v, want ErrWorkspaceAdminRequired", err)
	}
	if called {
		t.Fatal("actor mismatch reached the database authorization query")
	}
}

func TestLockInvitationWorkspaceAdminsDeduplicatesAndOrdersWorkspaces(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	workspaceA := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	workspaceB := uuid.MustParse("20000000-0000-0000-0000-000000000002")
	var locked []uuid.UUID
	queries := adminLockQueries{lock: func(_ context.Context, params invitationsql.LockActiveWorkspaceAdminParams) (bool, error) {
		if params.ActorID != actorID {
			t.Fatalf("authorization actor = %s, want %s", params.ActorID, actorID)
		}
		locked = append(locked, params.WorkspaceID)
		return true, nil
	}}
	commands := []invitationsdomain.NewWorkspaceInvitation{
		{Invitation: invitationsdomain.WorkspaceInvitation{WorkspaceID: workspaceB, InviterID: actorID}},
		{Invitation: invitationsdomain.WorkspaceInvitation{WorkspaceID: workspaceA, InviterID: actorID}},
		{Invitation: invitationsdomain.WorkspaceInvitation{WorkspaceID: workspaceB, InviterID: actorID}},
	}
	if err := lockInvitationWorkspaceAdmins(context.Background(), queries, actorID, commands); err != nil {
		t.Fatalf("lock invitation workspaces: %v", err)
	}
	if len(locked) != 2 || locked[0] != workspaceA || locked[1] != workspaceB {
		t.Fatalf("locked workspaces = %v, want [%s %s]", locked, workspaceA, workspaceB)
	}
}
