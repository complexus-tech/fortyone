package admin

import (
	"context"
	"testing"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListWorkspacesBuildsTypedBoundedQuery(t *testing.T) {
	actorID := uuid.New()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	repository := &adminTestRepository{}
	service := New(repository, WithNow(func() time.Time { return now }))

	_, err := service.ListWorkspaces(context.Background(), actorID, ListWorkspacesInput{
		Pagination: PaginationInput{Page: -4, Limit: 500},
		Query:      "  Acme  ",
		Status:     " TRIALING ",
	})

	require.NoError(t, err)
	require.Equal(t, actorID, repository.listWorkspacesQuery.ActorID)
	require.Equal(t, 1, repository.listWorkspacesQuery.Page.Page)
	require.Equal(t, maxPageLimit, repository.listWorkspacesQuery.Page.PageSize)
	require.Equal(t, "acme", repository.listWorkspacesQuery.Search)
	require.Equal(t, admindomain.WorkspaceStatusTrialing, repository.listWorkspacesQuery.Status)
	require.Equal(t, now.UTC(), repository.listWorkspacesQuery.Now)
}

func TestListWorkspacesRejectsUnknownStatus(t *testing.T) {
	service := New(&adminTestRepository{})

	_, err := service.ListWorkspaces(context.Background(), uuid.New(), ListWorkspacesInput{Status: "invented"})

	require.ErrorIs(t, err, ErrInvalidFilter)
}

func TestListWorkspaceIntegrationsBuildsTypedBoundedQuery(t *testing.T) {
	actorID := uuid.New()
	now := time.Date(2026, 8, 31, 10, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	repository := &adminTestRepository{}
	service := New(repository, WithNow(func() time.Time { return now }))

	_, err := service.ListWorkspaceIntegrations(context.Background(), actorID, ListWorkspaceIntegrationsInput{
		Pagination: PaginationInput{Page: 0, Limit: 500},
		Query:      "  Acme  ",
		Provider:   " GITHUB ",
		Status:     " ATTENTION ",
	})

	require.NoError(t, err)
	require.Equal(t, actorID, repository.listIntegrationsQuery.ActorID)
	require.Equal(t, 1, repository.listIntegrationsQuery.Page.Page)
	require.Equal(t, maxPageLimit, repository.listIntegrationsQuery.Page.PageSize)
	require.Equal(t, "acme", repository.listIntegrationsQuery.Search)
	require.Equal(t, admindomain.IntegrationProviderGitHub, repository.listIntegrationsQuery.Provider)
	require.Equal(t, admindomain.IntegrationStatusAttention, repository.listIntegrationsQuery.Status)
	require.Equal(t, now.UTC(), repository.listIntegrationsQuery.Now)
}

func TestListWorkspaceIntegrationsRejectsUnknownFilters(t *testing.T) {
	service := New(&adminTestRepository{})

	_, err := service.ListWorkspaceIntegrations(context.Background(), uuid.New(), ListWorkspaceIntegrationsInput{
		Provider: "drop table",
	})
	require.ErrorIs(t, err, ErrInvalidFilter)

	_, err = service.ListWorkspaceIntegrations(context.Background(), uuid.New(), ListWorkspaceIntegrationsInput{
		Status: "broken",
	})
	require.ErrorIs(t, err, ErrInvalidFilter)
}

func TestListAuditLogsBuildsTypedUTCFilter(t *testing.T) {
	actorID := uuid.New()
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	repository := &adminTestRepository{}
	service := New(repository)

	_, err := service.ListAuditLogs(context.Background(), actorID, ListAuditLogsInput{
		TargetType: " WORKSPACE ", Action: " Workspace.Trial_Updated ",
		ActorQuery: " Ops ", From: &from,
	})

	require.NoError(t, err)
	require.Equal(t, admindomain.TargetWorkspace, repository.auditQuery.TargetType)
	require.Equal(t, admindomain.AuditWorkspaceTrialUpdated, repository.auditQuery.Action)
	require.Equal(t, "ops", repository.auditQuery.ActorSearch)
	require.Equal(t, from.UTC(), *repository.auditQuery.From)
}

func TestListUsersResolvesProfileImages(t *testing.T) {
	resolver := &adminTestAssetResolver{}
	repository := &adminTestRepository{users: admindomain.ListResult[admindomain.UserSummary]{
		Items: []admindomain.UserSummary{{ID: uuid.New(), AvatarURL: "profiles/member.png"}},
	}}
	service := New(repository, WithAssetResolver(resolver))

	result, err := service.ListUsers(context.Background(), uuid.New(), ListUsersInput{})

	require.NoError(t, err)
	require.Equal(t, "profile:profiles/member.png", result.Items[0].AvatarURL)
	require.Equal(t, adminAssetURLExpiry, resolver.profileExpiry)
}
