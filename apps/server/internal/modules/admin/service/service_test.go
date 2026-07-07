package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type adminTestRepo struct {
	users            map[uuid.UUID]UserSummary
	listWorkspacesIn ListWorkspacesInput
	workspaces       ListResult[WorkspaceSummary]
	workspace        WorkspaceOverview
	listUsers        ListResult[UserSummary]
	user             UserOverview
	auditEntries     []AuditEntryInput
}

func (r *adminTestRepo) GetAdminUser(ctx context.Context, userID uuid.UUID) (UserSummary, error) {
	user, ok := r.users[userID]
	if !ok {
		return UserSummary{}, ErrNotFound
	}
	return user, nil
}

func (r *adminTestRepo) GetDashboardSummary(ctx context.Context) (DashboardSummary, error) {
	return DashboardSummary{}, nil
}

func (r *adminTestRepo) ListWorkspaces(ctx context.Context, input ListWorkspacesInput) (ListResult[WorkspaceSummary], error) {
	r.listWorkspacesIn = input
	return r.workspaces, nil
}

func (r *adminTestRepo) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID) (WorkspaceOverview, error) {
	if r.workspace.Workspace.ID == uuid.Nil {
		return WorkspaceOverview{}, ErrNotFound
	}
	return r.workspace, nil
}

func (r *adminTestRepo) UpdateWorkspaceTrial(ctx context.Context, input UpdateWorkspaceTrialInput) error {
	trialEndsOn := input.TrialEndsOn
	r.workspace.Workspace.TrialEndsOn = &trialEndsOn
	return nil
}

func (r *adminTestRepo) ListUsers(ctx context.Context, input ListUsersInput) (ListResult[UserSummary], error) {
	return r.listUsers, nil
}

func (r *adminTestRepo) GetUserOverview(ctx context.Context, userID uuid.UUID) (UserOverview, error) {
	if r.user.User.ID == uuid.Nil {
		return UserOverview{}, ErrNotFound
	}
	return r.user, nil
}

func (r *adminTestRepo) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (ListResult[AuditLog], error) {
	return ListResult[AuditLog]{}, nil
}

func (r *adminTestRepo) InsertAuditEntry(ctx context.Context, input AuditEntryInput) error {
	r.auditEntries = append(r.auditEntries, input)
	return nil
}

func TestListWorkspacesRejectsNonInternalUsers(t *testing.T) {
	actorID := uuid.New()
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "member@example.com",
				IsInternal: false,
			},
		},
	}
	service := New(repo)

	_, err := service.ListWorkspaces(context.Background(), actorID, ListWorkspacesInput{})

	require.ErrorIs(t, err, ErrForbidden)
}

func TestListWorkspacesNormalizesSearchAndPagination(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
		workspaces: ListResult[WorkspaceSummary]{
			Items: []WorkspaceSummary{{ID: workspaceID, Name: "Acme"}},
			Pagination: Pagination{
				Total:  1,
				Page:   2,
				Limit:  100,
				Offset: 100,
			},
		},
	}
	service := New(repo)

	result, err := service.ListWorkspaces(context.Background(), actorID, ListWorkspacesInput{
		Pagination: PaginationInput{Page: -5, Limit: 500},
		Query:      "  acme  ",
		Status:     "  trialing ",
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, "acme", repo.listWorkspacesIn.Query)
	require.Equal(t, "trialing", repo.listWorkspacesIn.Status)
	require.Equal(t, 1, repo.listWorkspacesIn.Pagination.Page)
	require.Equal(t, maxPageLimit, repo.listWorkspacesIn.Pagination.Limit)
}

func TestListWorkspacesResolvesLogoURLs(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	avatarURL := "logos/acme.png"
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
		workspaces: ListResult[WorkspaceSummary]{
			Items: []WorkspaceSummary{{
				ID:        workspaceID,
				Name:      "Acme",
				AvatarURL: &avatarURL,
			}},
		},
	}
	resolver := &adminTestAssetResolver{}
	service := New(repo, WithAssetResolver(resolver))

	result, err := service.ListWorkspaces(context.Background(), actorID, ListWorkspacesInput{})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.NotNil(t, result.Items[0].AvatarURL)
	require.Equal(t, "workspace:"+avatarURL, *result.Items[0].AvatarURL)
	require.Equal(t, adminAssetURLExpiry, resolver.workspaceExpiry)
}

func TestListUsersResolvesProfileImageURLs(t *testing.T) {
	actorID := uuid.New()
	userID := uuid.New()
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
		listUsers: ListResult[UserSummary]{
			Items: []UserSummary{{
				ID:        userID,
				Email:     "member@example.com",
				AvatarURL: "profiles/member.png",
			}},
		},
	}
	resolver := &adminTestAssetResolver{}
	service := New(repo, WithAssetResolver(resolver))

	result, err := service.ListUsers(context.Background(), actorID, ListUsersInput{})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, "profile:profiles/member.png", result.Items[0].AvatarURL)
	require.Equal(t, adminAssetURLExpiry, resolver.profileExpiry)
}

func TestGetUserOverviewResolvesProfileImageURL(t *testing.T) {
	actorID := uuid.New()
	userID := uuid.New()
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
		user: UserOverview{
			User: UserSummary{
				ID:        userID,
				Email:     "member@example.com",
				AvatarURL: "profiles/member.png",
			},
		},
	}
	resolver := &adminTestAssetResolver{}
	service := New(repo, WithAssetResolver(resolver))

	overview, err := service.GetUserOverview(context.Background(), actorID, userID)

	require.NoError(t, err)
	require.Equal(t, "profile:profiles/member.png", overview.User.AvatarURL)
	require.Equal(t, adminAssetURLExpiry, resolver.profileExpiry)
}

func TestUpdateWorkspaceTrialExtendsTrialAndWritesAudit(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	currentTrialEndsOn := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	newTrialEndsOn := currentTrialEndsOn.Add(7 * 24 * time.Hour)
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
		workspace: WorkspaceOverview{
			Workspace: WorkspaceSummary{
				ID:          workspaceID,
				Name:        "Acme",
				TrialEndsOn: &currentTrialEndsOn,
			},
		},
	}
	service := New(repo, WithNow(func() time.Time {
		return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	}))

	overview, err := service.UpdateWorkspaceTrial(context.Background(), actorID, workspaceID, UpdateWorkspaceTrialInput{
		TrialEndsOn: newTrialEndsOn,
		Reason:      "  sales-led onboarding extension  ",
	})

	require.NoError(t, err)
	require.Equal(t, newTrialEndsOn, *overview.Workspace.TrialEndsOn)
	require.Len(t, repo.auditEntries, 1)
	entry := repo.auditEntries[0]
	require.Equal(t, actorID, entry.ActorUserID)
	require.Equal(t, "workspace", entry.TargetType)
	require.Equal(t, workspaceID, *entry.TargetID)
	require.Equal(t, workspaceID, *entry.WorkspaceID)
	require.Equal(t, "workspace.trial_updated", entry.Action)
	require.Equal(t, "trial_ends_on", entry.FieldName)
	require.Equal(t, "sales-led onboarding extension", entry.Reason)
	require.Equal(t, currentTrialEndsOn, entry.OldValue)
	require.Equal(t, newTrialEndsOn, entry.NewValue)
}

func TestUpdateWorkspaceTrialRejectsShorteningActiveTrial(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	currentTrialEndsOn := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
		workspace: WorkspaceOverview{
			Workspace: WorkspaceSummary{
				ID:          workspaceID,
				Name:        "Acme",
				TrialEndsOn: &currentTrialEndsOn,
			},
		},
	}
	service := New(repo, WithNow(func() time.Time {
		return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	}))

	_, err := service.UpdateWorkspaceTrial(context.Background(), actorID, workspaceID, UpdateWorkspaceTrialInput{
		TrialEndsOn: currentTrialEndsOn.Add(-24 * time.Hour),
		Reason:      "testing",
	})

	require.ErrorIs(t, err, ErrInvalidTrialEndsOn)
	require.Empty(t, repo.auditEntries)
}

func TestUpdateWorkspaceTrialReturnsAuditErrors(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	auditErr := errors.New("audit unavailable")
	repo := &adminTestRepoWithAuditError{
		adminTestRepo: adminTestRepo{
			users: map[uuid.UUID]UserSummary{
				actorID: {
					ID:         actorID,
					Email:      "ops@fortyone.app",
					IsActive:   true,
					IsInternal: true,
				},
			},
			workspace: WorkspaceOverview{
				Workspace: WorkspaceSummary{
					ID:   workspaceID,
					Name: "Acme",
				},
			},
		},
		auditErr: auditErr,
	}
	service := New(repo, WithNow(func() time.Time {
		return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	}))

	_, err := service.UpdateWorkspaceTrial(context.Background(), actorID, workspaceID, UpdateWorkspaceTrialInput{
		TrialEndsOn: time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
		Reason:      "sales request",
	})

	require.ErrorIs(t, err, auditErr)
}

type adminTestRepoWithAuditError struct {
	adminTestRepo
	auditErr error
}

func (r *adminTestRepoWithAuditError) InsertAuditEntry(ctx context.Context, input AuditEntryInput) error {
	return r.auditErr
}

type adminTestAssetResolver struct {
	profileExpiry   time.Duration
	workspaceExpiry time.Duration
}

func (r *adminTestAssetResolver) ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error) {
	r.profileExpiry = expiry
	return "profile:" + avatar, nil
}

func (r *adminTestAssetResolver) ResolveWorkspaceLogoURL(ctx context.Context, logo string, expiry time.Duration) (string, error) {
	r.workspaceExpiry = expiry
	return "workspace:" + logo, nil
}
