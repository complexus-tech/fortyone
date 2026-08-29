package adminrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type adminQueryFake struct {
	adminsql.Querier
	lockAdmin        func(context.Context, adminsql.LockActiveInternalAdminParams) (uuid.UUID, error)
	dashboard        func(context.Context, adminsql.GetAdminDashboardSummaryParams) (adminsql.GetAdminDashboardSummaryRow, error)
	lockParticipants func(context.Context, adminsql.LockAdminUserMutationParticipantsParams) ([]adminsql.LockAdminUserMutationParticipantsRow, error)
	lockWorkspace    func(context.Context, adminsql.LockAdminWorkspaceMutationTargetParams) (adminsql.LockAdminWorkspaceMutationTargetRow, error)
	updateTrial      func(context.Context, adminsql.UpdateAdminWorkspaceTrialParams) (adminsql.UpdateAdminWorkspaceTrialRow, error)
	insertAudit      func(context.Context, adminsql.InsertAdminAuditLogParams) (adminsql.InsertAdminAuditLogRow, error)
	lockUser         func(context.Context, adminsql.LockUserTargetParams) (uuid.UUID, error)
	getUser          func(context.Context, adminsql.GetAdminUserParams) (adminsql.GetAdminUserRow, error)
	listMemberships  func(context.Context, adminsql.ListAdminUserMembershipsParams) ([]adminsql.ListAdminUserMembershipsRow, error)
	revokeSessions   func(context.Context, adminsql.RevokeAdminUserBrowserSessionsParams) (int64, error)
}

func (fake *adminQueryFake) LockActiveInternalAdmin(ctx context.Context, params adminsql.LockActiveInternalAdminParams) (uuid.UUID, error) {
	return fake.lockAdmin(ctx, params)
}

func (fake *adminQueryFake) GetAdminDashboardSummary(ctx context.Context, params adminsql.GetAdminDashboardSummaryParams) (adminsql.GetAdminDashboardSummaryRow, error) {
	return fake.dashboard(ctx, params)
}

func (fake *adminQueryFake) LockAdminUserMutationParticipants(ctx context.Context, params adminsql.LockAdminUserMutationParticipantsParams) ([]adminsql.LockAdminUserMutationParticipantsRow, error) {
	return fake.lockParticipants(ctx, params)
}

func (fake *adminQueryFake) LockAdminWorkspaceMutationTarget(ctx context.Context, params adminsql.LockAdminWorkspaceMutationTargetParams) (adminsql.LockAdminWorkspaceMutationTargetRow, error) {
	return fake.lockWorkspace(ctx, params)
}

func (fake *adminQueryFake) UpdateAdminWorkspaceTrial(ctx context.Context, params adminsql.UpdateAdminWorkspaceTrialParams) (adminsql.UpdateAdminWorkspaceTrialRow, error) {
	return fake.updateTrial(ctx, params)
}

func (fake *adminQueryFake) InsertAdminAuditLog(ctx context.Context, params adminsql.InsertAdminAuditLogParams) (adminsql.InsertAdminAuditLogRow, error) {
	return fake.insertAudit(ctx, params)
}

func (fake *adminQueryFake) LockUserTarget(ctx context.Context, params adminsql.LockUserTargetParams) (uuid.UUID, error) {
	return fake.lockUser(ctx, params)
}

func (fake *adminQueryFake) GetAdminUser(ctx context.Context, params adminsql.GetAdminUserParams) (adminsql.GetAdminUserRow, error) {
	return fake.getUser(ctx, params)
}

func (fake *adminQueryFake) ListAdminUserMemberships(ctx context.Context, params adminsql.ListAdminUserMembershipsParams) ([]adminsql.ListAdminUserMembershipsRow, error) {
	return fake.listMemberships(ctx, params)
}

func (fake *adminQueryFake) RevokeAdminUserBrowserSessions(ctx context.Context, params adminsql.RevokeAdminUserBrowserSessionsParams) (int64, error) {
	return fake.revokeSessions(ctx, params)
}

func TestSessionRevocationAdvancesEpochBeforeDurableAudit(t *testing.T) {
	t.Parallel()

	actorID, userID := uuid.New(), uuid.New()
	now := time.Now().UTC()
	sequence := make([]string, 0, 6)
	fake := &adminQueryFake{
		lockAdmin: func(context.Context, adminsql.LockActiveInternalAdminParams) (uuid.UUID, error) {
			sequence = append(sequence, "authorize")
			return actorID, nil
		},
		lockUser: func(_ context.Context, params adminsql.LockUserTargetParams) (uuid.UUID, error) {
			sequence = append(sequence, "lock")
			require.Equal(t, userID, params.UserID)
			return userID, nil
		},
		getUser: func(_ context.Context, params adminsql.GetAdminUserParams) (adminsql.GetAdminUserRow, error) {
			sequence = append(sequence, "read")
			return adminsql.GetAdminUserRow{
				UserID: userID, Username: "ada", Email: "ada@example.com",
				IsActive: true, CreatedAt: now, UpdatedAt: now,
			}, nil
		},
		listMemberships: func(context.Context, adminsql.ListAdminUserMembershipsParams) ([]adminsql.ListAdminUserMembershipsRow, error) {
			sequence = append(sequence, "memberships")
			return nil, nil
		},
		revokeSessions: func(_ context.Context, params adminsql.RevokeAdminUserBrowserSessionsParams) (int64, error) {
			sequence = append(sequence, "revoke")
			require.Equal(t, userID, params.UserID)
			return 2, nil
		},
		insertAudit: func(_ context.Context, params adminsql.InsertAdminAuditLogParams) (adminsql.InsertAdminAuditLogRow, error) {
			sequence = append(sequence, "audit")
			require.Equal(t, string(admindomain.AuditUserSessionRevocationRequested), params.Action)
			return adminsql.InsertAdminAuditLogRow{}, nil
		},
	}
	repository := newWithQueries(fake)

	_, err := repository.RequestSessionRevocation(t.Context(), admindomain.RequestSessionRevocationCommand{
		ActorID: actorID, UserID: userID, Reason: "credential exposure",
	})

	require.NoError(t, err)
	require.Equal(t, []string{"authorize", "lock", "read", "memberships", "revoke", "audit"}, sequence)
}

func TestDashboardDoesNotRunAfterLiveAuthorizationFails(t *testing.T) {
	called := false
	fake := &adminQueryFake{
		lockAdmin: func(context.Context, adminsql.LockActiveInternalAdminParams) (uuid.UUID, error) {
			return uuid.Nil, pgx.ErrNoRows
		},
		dashboard: func(context.Context, adminsql.GetAdminDashboardSummaryParams) (adminsql.GetAdminDashboardSummaryRow, error) {
			called = true
			return adminsql.GetAdminDashboardSummaryRow{}, nil
		},
	}
	repository := newWithQueries(fake)

	_, err := repository.GetDashboardSummary(t.Context(), admindomain.DashboardSummaryQuery{
		ActorID: uuid.New(), Now: time.Now(),
	})

	require.ErrorIs(t, err, admindomain.ErrForbidden)
	require.False(t, called)
}

func TestUserStateRejectsAuthorizedSelfMutationBeforeUpdate(t *testing.T) {
	actorID := uuid.New()
	fake := &adminQueryFake{
		lockParticipants: func(context.Context, adminsql.LockAdminUserMutationParticipantsParams) ([]adminsql.LockAdminUserMutationParticipantsRow, error) {
			return []adminsql.LockAdminUserMutationParticipantsRow{{
				UserID: actorID, IsActive: true, IsInternal: true,
			}}, nil
		},
	}
	repository := newWithQueries(fake)

	_, err := repository.UpdateUserState(t.Context(), admindomain.UpdateUserStateCommand{
		ActorID: actorID, UserID: actorID, Reason: "security review", Now: time.Now(),
		Patch: admindomain.UserStatePatch{IsActive: platformpatch.Set(false)},
	})

	require.ErrorIs(t, err, admindomain.ErrSelfMutation)
}

func TestWorkspaceMutationStopsWhenDurableAuditFails(t *testing.T) {
	actorID, workspaceID := uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	auditFailure := errors.New("audit unavailable")
	sequence := make([]string, 0, 4)
	fake := &adminQueryFake{
		lockAdmin: func(context.Context, adminsql.LockActiveInternalAdminParams) (uuid.UUID, error) {
			sequence = append(sequence, "authorize")
			return actorID, nil
		},
		lockWorkspace: func(context.Context, adminsql.LockAdminWorkspaceMutationTargetParams) (adminsql.LockAdminWorkspaceMutationTargetRow, error) {
			sequence = append(sequence, "lock")
			return adminsql.LockAdminWorkspaceMutationTargetRow{
				WorkspaceID: workspaceID, Name: "Acme", Slug: "acme", UpdatedAt: now,
			}, nil
		},
		updateTrial: func(context.Context, adminsql.UpdateAdminWorkspaceTrialParams) (adminsql.UpdateAdminWorkspaceTrialRow, error) {
			sequence = append(sequence, "update")
			trial := now.Add(48 * time.Hour)
			return adminsql.UpdateAdminWorkspaceTrialRow{TrialEndsOn: &trial, UpdatedAt: now.Add(time.Second)}, nil
		},
		insertAudit: func(context.Context, adminsql.InsertAdminAuditLogParams) (adminsql.InsertAdminAuditLogRow, error) {
			sequence = append(sequence, "audit")
			return adminsql.InsertAdminAuditLogRow{}, auditFailure
		},
	}
	repository := newWithQueries(fake)

	_, err := repository.UpdateWorkspaceTrial(t.Context(), admindomain.UpdateWorkspaceTrialCommand{
		ActorID: actorID, WorkspaceID: workspaceID, TrialEndsOn: now.Add(48 * time.Hour),
		Reason: "approved extension", Now: now.Add(time.Second),
	})

	require.ErrorIs(t, err, auditFailure)
	require.Equal(t, []string{"authorize", "lock", "update", "audit"}, sequence)
}

func TestWorkspaceTrialCannotShortenAnActiveTrial(t *testing.T) {
	actorID, workspaceID := uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	currentEnd := now.Add(72 * time.Hour)
	updateCalled := false
	fake := &adminQueryFake{
		lockAdmin: func(context.Context, adminsql.LockActiveInternalAdminParams) (uuid.UUID, error) {
			return actorID, nil
		},
		lockWorkspace: func(context.Context, adminsql.LockAdminWorkspaceMutationTargetParams) (adminsql.LockAdminWorkspaceMutationTargetRow, error) {
			return adminsql.LockAdminWorkspaceMutationTargetRow{
				WorkspaceID: workspaceID, TrialEndsOn: &currentEnd, UpdatedAt: now,
			}, nil
		},
		updateTrial: func(context.Context, adminsql.UpdateAdminWorkspaceTrialParams) (adminsql.UpdateAdminWorkspaceTrialRow, error) {
			updateCalled = true
			return adminsql.UpdateAdminWorkspaceTrialRow{}, nil
		},
	}
	repository := newWithQueries(fake)

	_, err := repository.UpdateWorkspaceTrial(t.Context(), admindomain.UpdateWorkspaceTrialCommand{
		ActorID: actorID, WorkspaceID: workspaceID, TrialEndsOn: now.Add(48 * time.Hour),
		Reason: "shorten", Now: now,
	})

	require.ErrorIs(t, err, admindomain.ErrInvalidTrialEndsOn)
	require.False(t, updateCalled)
}

func TestAdminPaginationRejectsOversizedAndOverflowingOffsets(t *testing.T) {
	_, err := newSQLPage(pagination.OffsetParams{Page: 1, PageSize: pagination.MaximumPageSize + 1})
	require.ErrorIs(t, err, admindomain.ErrInvalidPagination)

	_, err = newSQLPage(pagination.OffsetParams{Page: int(^uint(0) >> 1), PageSize: pagination.MaximumPageSize})
	require.ErrorIs(t, err, admindomain.ErrInvalidPagination)
}
