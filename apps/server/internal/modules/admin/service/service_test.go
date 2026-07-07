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
	users                    map[uuid.UUID]UserSummary
	listWorkspacesIn         ListWorkspacesInput
	workspaces               ListResult[WorkspaceSummary]
	workspace                WorkspaceOverview
	updateWorkspaceDeletedIn *UpdateWorkspaceDeletedInput
	listUsers                ListResult[UserSummary]
	user                     UserOverview
	updateUserStateIn        *UpdateUserStateInput
	listAuditLogsIn          ListAuditLogsInput
	notes                    ListResult[AdminNote]
	createNoteIn             *CreateAdminNoteInput
	auditEntries             []AuditEntryInput
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

func (r *adminTestRepo) UpdateWorkspaceDeleted(ctx context.Context, input UpdateWorkspaceDeletedInput) error {
	r.updateWorkspaceDeletedIn = &input
	if input.Deleted {
		now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
		r.workspace.Workspace.DeletedAt = &now
		return nil
	}
	r.workspace.Workspace.DeletedAt = nil
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

func (r *adminTestRepo) UpdateUserState(ctx context.Context, input UpdateUserStateInput) error {
	r.updateUserStateIn = &input
	if input.IsActive != nil {
		r.user.User.IsActive = *input.IsActive
	}
	if input.IsInternal != nil {
		r.user.User.IsInternal = *input.IsInternal
	}
	return nil
}

func (r *adminTestRepo) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (ListResult[AuditLog], error) {
	r.listAuditLogsIn = input
	return ListResult[AuditLog]{}, nil
}

func (r *adminTestRepo) ListAdminNotes(ctx context.Context, input ListAdminNotesInput) (ListResult[AdminNote], error) {
	return r.notes, nil
}

func (r *adminTestRepo) CreateAdminNote(ctx context.Context, input CreateAdminNoteInput) (AdminNote, error) {
	r.createNoteIn = &input
	return AdminNote{
		ID:              uuid.New(),
		TargetType:      input.TargetType,
		TargetID:        input.TargetID,
		WorkspaceID:     input.WorkspaceID,
		Body:            input.Body,
		CreatedByUserID: uuid.New(),
		CreatedAt:       time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC),
	}, nil
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

func TestUpdateWorkspaceDeletedWritesAudit(t *testing.T) {
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
		workspace: WorkspaceOverview{
			Workspace: WorkspaceSummary{
				ID:   workspaceID,
				Name: "Acme",
				Slug: "acme",
			},
		},
	}
	service := New(repo, WithNow(func() time.Time {
		return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	}))

	overview, err := service.UpdateWorkspaceDeleted(context.Background(), actorID, workspaceID, UpdateWorkspaceDeletedInput{
		Deleted: true,
		Reason:  " customer requested workspace closure ",
	})

	require.NoError(t, err)
	require.True(t, overview.Workspace.DeletedAt != nil)
	require.NotNil(t, repo.updateWorkspaceDeletedIn)
	require.Equal(t, workspaceID, repo.updateWorkspaceDeletedIn.WorkspaceID)
	require.True(t, repo.updateWorkspaceDeletedIn.Deleted)
	require.Equal(t, "customer requested workspace closure", repo.updateWorkspaceDeletedIn.Reason)
	require.Len(t, repo.auditEntries, 1)
	entry := repo.auditEntries[0]
	require.Equal(t, "workspace.deleted", entry.Action)
	require.Equal(t, "deleted_at", entry.FieldName)
	require.Equal(t, workspaceID, *entry.TargetID)
	require.Equal(t, "customer requested workspace closure", entry.Reason)
}

func TestUpdateWorkspaceDeletedRequiresReason(t *testing.T) {
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
		workspace: WorkspaceOverview{Workspace: WorkspaceSummary{ID: workspaceID}},
	}
	service := New(repo)

	_, err := service.UpdateWorkspaceDeleted(context.Background(), actorID, workspaceID, UpdateWorkspaceDeletedInput{
		Deleted: true,
	})

	require.ErrorIs(t, err, ErrReasonRequired)
	require.Nil(t, repo.updateWorkspaceDeletedIn)
	require.Empty(t, repo.auditEntries)
}

func TestRequestWorkspaceSubscriptionSyncCallsSyncerAndWritesAudit(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	stripeSubscriptionID := "sub_123"
	syncer := &adminTestSubscriptionSyncer{}
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
				ID:                   workspaceID,
				Name:                 "Acme",
				StripeSubscriptionID: &stripeSubscriptionID,
				SubscriptionStatus:   stringPtr("past_due"),
			},
		},
	}
	service := New(repo, WithSubscriptionSyncer(syncer))

	overview, err := service.RequestWorkspaceSubscriptionSync(context.Background(), actorID, workspaceID, RequestWorkspaceSubscriptionSyncInput{
		Reason: "billing support follow-up",
	})

	require.NoError(t, err)
	require.Equal(t, workspaceID, overview.Workspace.ID)
	require.Equal(t, workspaceID, syncer.workspaceID)
	require.Len(t, repo.auditEntries, 1)
	entry := repo.auditEntries[0]
	require.Equal(t, "subscription.synced", entry.Action)
	require.Equal(t, "subscription_status", entry.FieldName)
	require.Equal(t, "past_due", entry.OldValue)
	require.Equal(t, "billing support follow-up", entry.Reason)
}

func TestUpdateUserStateWritesAudit(t *testing.T) {
	actorID := uuid.New()
	userID := uuid.New()
	isActive := false
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
				ID:         userID,
				Email:      "customer@example.com",
				IsActive:   true,
				IsInternal: false,
			},
		},
	}
	service := New(repo)

	overview, err := service.UpdateUserState(context.Background(), actorID, userID, UpdateUserStateInput{
		IsActive: &isActive,
		Reason:   "security review",
	})

	require.NoError(t, err)
	require.False(t, overview.User.IsActive)
	require.NotNil(t, repo.updateUserStateIn)
	require.Equal(t, userID, repo.updateUserStateIn.UserID)
	require.Len(t, repo.auditEntries, 1)
	entry := repo.auditEntries[0]
	require.Equal(t, "user.deactivated", entry.Action)
	require.Equal(t, "is_active", entry.FieldName)
	require.Equal(t, true, entry.OldValue)
	require.Equal(t, false, entry.NewValue)
	require.Equal(t, "security review", entry.Reason)
}

func TestUpdateUserStateRejectsSelfAccessMutation(t *testing.T) {
	actorID := uuid.New()
	isActive := false
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
				ID:         actorID,
				IsActive:   true,
				IsInternal: true,
			},
		},
	}
	service := New(repo)

	_, err := service.UpdateUserState(context.Background(), actorID, actorID, UpdateUserStateInput{
		IsActive: &isActive,
		Reason:   "testing",
	})

	require.ErrorIs(t, err, ErrSelfMutation)
	require.Nil(t, repo.updateUserStateIn)
	require.Empty(t, repo.auditEntries)
}

func TestRequestUserSessionRevocationWritesAudit(t *testing.T) {
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
				ID:       userID,
				Email:    "customer@example.com",
				FullName: "Customer",
			},
		},
	}
	service := New(repo)

	overview, err := service.RequestUserSessionRevocation(context.Background(), actorID, userID, RequestUserSessionRevocationInput{
		Reason: "suspected token sharing",
	})

	require.NoError(t, err)
	require.Equal(t, userID, overview.User.ID)
	require.Len(t, repo.auditEntries, 1)
	entry := repo.auditEntries[0]
	require.Equal(t, "user.session_revocation_requested", entry.Action)
	require.Equal(t, userID, *entry.TargetID)
	require.Equal(t, "suspected token sharing", entry.Reason)
}

func TestCreateAdminNoteValidatesTargetAndWritesAudit(t *testing.T) {
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
		workspace: WorkspaceOverview{
			Workspace: WorkspaceSummary{
				ID:   workspaceID,
				Name: "Acme",
				Slug: "acme",
			},
		},
	}
	service := New(repo)

	note, err := service.CreateAdminNote(context.Background(), actorID, CreateAdminNoteInput{
		TargetType: " workspace ",
		TargetID:   workspaceID,
		Body:       " Customer asked for annual procurement follow-up. ",
	})

	require.NoError(t, err)
	require.Equal(t, "workspace", note.TargetType)
	require.Equal(t, "Customer asked for annual procurement follow-up.", note.Body)
	require.NotNil(t, repo.createNoteIn)
	require.Equal(t, workspaceID, repo.createNoteIn.TargetID)
	require.NotNil(t, repo.createNoteIn.WorkspaceID)
	require.Equal(t, workspaceID, *repo.createNoteIn.WorkspaceID)
	require.Len(t, repo.auditEntries, 1)
	require.Equal(t, "admin_note.created", repo.auditEntries[0].Action)
}

func TestCreateAdminNoteRejectsEmptyBody(t *testing.T) {
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
	}
	service := New(repo)

	_, err := service.CreateAdminNote(context.Background(), actorID, CreateAdminNoteInput{
		TargetType: "workspace",
		TargetID:   workspaceID,
		Body:       "   ",
	})

	require.ErrorIs(t, err, ErrInvalidAdminNote)
	require.Nil(t, repo.createNoteIn)
}

func TestListAuditLogsNormalizesFilters(t *testing.T) {
	actorID := uuid.New()
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	to := time.Date(2026, 7, 7, 0, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	repo := &adminTestRepo{
		users: map[uuid.UUID]UserSummary{
			actorID: {
				ID:         actorID,
				Email:      "ops@fortyone.app",
				IsActive:   true,
				IsInternal: true,
			},
		},
	}
	service := New(repo)

	_, err := service.ListAuditLogs(context.Background(), actorID, ListAuditLogsInput{
		Pagination: PaginationInput{Page: 0, Limit: 500},
		TargetType: " Workspace ",
		Query:      " Acme ",
		Action:     " Workspace.Trial_Updated ",
		ActorQuery: " Joseph ",
		From:       &from,
		To:         &to,
	})

	require.NoError(t, err)
	require.Equal(t, 1, repo.listAuditLogsIn.Pagination.Page)
	require.Equal(t, maxPageLimit, repo.listAuditLogsIn.Pagination.Limit)
	require.Equal(t, "workspace", repo.listAuditLogsIn.TargetType)
	require.Equal(t, "acme", repo.listAuditLogsIn.Query)
	require.Equal(t, "workspace.trial_updated", repo.listAuditLogsIn.Action)
	require.Equal(t, "joseph", repo.listAuditLogsIn.ActorQuery)
	require.Equal(t, from.UTC(), *repo.listAuditLogsIn.From)
	require.Equal(t, to.UTC(), *repo.listAuditLogsIn.To)
}

func stringPtr(value string) *string {
	return &value
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

type adminTestSubscriptionSyncer struct {
	workspaceID uuid.UUID
}

func (s *adminTestSubscriptionSyncer) SyncSubscription(ctx context.Context, workspaceID uuid.UUID) error {
	s.workspaceID = workspaceID
	return nil
}

func (r *adminTestAssetResolver) ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error) {
	r.profileExpiry = expiry
	return "profile:" + avatar, nil
}

func (r *adminTestAssetResolver) ResolveWorkspaceLogoURL(ctx context.Context, logo string, expiry time.Duration) (string, error) {
	r.workspaceExpiry = expiry
	return "workspace:" + logo, nil
}
