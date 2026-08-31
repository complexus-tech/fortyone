package adminrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) ListWorkspaceIntegrations(
	ctx context.Context,
	query admindomain.ListWorkspaceIntegrationsQuery,
) (admindomain.ListResult[admindomain.WorkspaceIntegrationSummary], error) {
	if _, err := admindomain.ParseIntegrationProvider(string(query.Provider)); err != nil {
		return admindomain.ListResult[admindomain.WorkspaceIntegrationSummary]{}, err
	}
	if _, err := admindomain.ParseIntegrationStatus(string(query.Status)); err != nil {
		return admindomain.ListResult[admindomain.WorkspaceIntegrationSummary]{}, err
	}
	page, err := newSQLPage(query.Page)
	if err != nil {
		return admindomain.ListResult[admindomain.WorkspaceIntegrationSummary]{}, err
	}

	var result admindomain.ListResult[admindomain.WorkspaceIntegrationSummary]
	err = repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		rows, err := queries.ListAdminWorkspaceIntegrations(ctx, adminsql.ListAdminWorkspaceIntegrationsParams{
			SearchText: query.Search, ProviderFilter: string(query.Provider),
			StatusFilter: string(query.Status), NowAt: query.Now,
			RowLimit: page.limit, RowOffset: page.offset,
		})
		if err != nil {
			return fmt.Errorf("list admin workspace integrations: %w", err)
		}
		items := make([]admindomain.WorkspaceIntegrationSummary, 0, len(rows))
		for _, row := range rows {
			item, err := workspaceIntegrationSummaryFromRow(row, query.Now)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		total := int64(0)
		if len(rows) > 0 {
			total = rows[0].TotalCount
		}
		pagination, err := paginationResult(query.Page, total)
		if err != nil {
			return err
		}
		result = admindomain.ListResult[admindomain.WorkspaceIntegrationSummary]{
			Items: items, Pagination: pagination,
		}
		return nil
	})
	return result, err
}

func (repository *Repository) GetWorkspaceIntegrations(
	ctx context.Context,
	query admindomain.GetWorkspaceIntegrationsQuery,
) (admindomain.WorkspaceIntegrations, error) {
	var result admindomain.WorkspaceIntegrations
	err := repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		if _, err := queries.GetAdminWorkspace(ctx, adminsql.GetAdminWorkspaceParams{WorkspaceID: query.WorkspaceID}); errors.Is(err, pgx.ErrNoRows) {
			return admindomain.ErrNotFound
		} else if err != nil {
			return fmt.Errorf("get integration workspace: %w", err)
		}

		mappings, err := listIntegrationMemberMappings(ctx, queries, query.WorkspaceID)
		if err != nil {
			return err
		}
		slack, err := getSlackIntegration(ctx, queries, query.WorkspaceID)
		if err != nil {
			return err
		}
		github, err := listGitHubInstallations(ctx, queries, query.WorkspaceID)
		if err != nil {
			return err
		}
		calendar, err := listCalendarConnections(ctx, queries, query.WorkspaceID)
		if err != nil {
			return err
		}
		figma, err := getFigmaIntegration(ctx, queries, query.WorkspaceID, query.Now)
		if err != nil {
			return err
		}

		result = admindomain.WorkspaceIntegrations{
			WorkspaceID:    query.WorkspaceID,
			Slack:          slack,
			GitHub:         github,
			Calendar:       calendar,
			Figma:          figma,
			MemberMappings: mappings,
		}
		result.Summaries = integrationSummaries(result, len(mappings))
		return nil
	})
	return result, err
}

func workspaceIntegrationSummaryFromRow(
	row adminsql.ListAdminWorkspaceIntegrationsRow,
	now time.Time,
) (admindomain.WorkspaceIntegrationSummary, error) {
	members, err := integrationCount(row.MemberCount, "workspace members")
	if err != nil {
		return admindomain.WorkspaceIntegrationSummary{}, err
	}
	slackMappings, err := integrationCount(row.SlackMappingCount, "Slack mappings")
	if err != nil {
		return admindomain.WorkspaceIntegrationSummary{}, err
	}
	githubConnections, err := integrationCount(row.GithubConnectionCount, "GitHub connections")
	if err != nil {
		return admindomain.WorkspaceIntegrationSummary{}, err
	}
	githubMappings, err := integrationCount(row.GithubMappingCount, "GitHub mappings")
	if err != nil {
		return admindomain.WorkspaceIntegrationSummary{}, err
	}
	calendarConnections, err := integrationCount(row.CalendarConnectionCount, "Calendar connections")
	if err != nil {
		return admindomain.WorkspaceIntegrationSummary{}, err
	}

	slackState := "not_connected"
	var slackLastSyncedAt *time.Time
	if row.SlackID != nil {
		slackState = "connected"
		slackLastSyncedAt = &row.SlackLastSyncedAt
	}
	githubState := "not_connected"
	var githubLastSyncedAt *time.Time
	if githubConnections > 0 {
		githubState = "connected"
		githubLastSyncedAt = &row.GithubLastSyncedAt
		if row.GithubNeedsAttention {
			githubState = "suspended"
		}
	}
	calendarState := "not_connected"
	var calendarLastSyncedAt *time.Time
	if calendarConnections > 0 {
		calendarState = "connected"
		calendarLastSyncedAt = &row.CalendarLastSyncedAt
		if row.CalendarNeedsAttention {
			calendarState = "degraded"
		}
	}
	figmaState := "not_connected"
	var figmaLastSyncedAt *time.Time
	if row.FigmaID != nil {
		figmaState = "connected"
		figmaLastSyncedAt = row.FigmaCreatedAt
		if row.FigmaExpiresAt != nil && !row.FigmaExpiresAt.After(now) {
			figmaState = "reauthorization_required"
		}
	}

	return admindomain.WorkspaceIntegrationSummary{
		WorkspaceID: row.WorkspaceID, WorkspaceName: row.Name, WorkspaceSlug: row.Slug,
		WorkspaceAvatar: row.AvatarURL, SubscriptionTier: enumString(row.SubscriptionTier),
		MemberCount: members,
		Integrations: []admindomain.IntegrationSummary{
			{Provider: "slack", State: slackState, AccountLabel: row.SlackTeamName,
				ConnectionCount: boolCount(row.SlackID != nil), MappingCount: slackMappings,
				UnmappedMemberCount: unmappedCount(members, slackMappings), LastSyncedAt: slackLastSyncedAt},
			{Provider: "github", State: githubState, AccountLabel: optionalNonBlankString(row.GithubAccountLabel),
				ConnectionCount: githubConnections, MappingCount: githubMappings,
				UnmappedMemberCount: unmappedCount(members, githubMappings), LastSyncedAt: githubLastSyncedAt},
			{Provider: "calendar", State: calendarState, ConnectionCount: calendarConnections,
				MappingCount: calendarConnections, UnmappedMemberCount: unmappedCount(members, calendarConnections),
				LastSyncedAt: calendarLastSyncedAt},
			{Provider: "figma", State: figmaState, AccountLabel: row.FigmaAccountLabel,
				ConnectionCount: boolCount(row.FigmaID != nil), LastSyncedAt: figmaLastSyncedAt},
		},
	}, nil
}

func listIntegrationMemberMappings(ctx context.Context, queries adminsql.Querier, workspaceID uuid.UUID) ([]admindomain.IntegrationMemberMapping, error) {
	rows, err := queries.ListAdminIntegrationMemberMappings(ctx, adminsql.ListAdminIntegrationMemberMappingsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list admin integration member mappings: %w", err)
	}
	items := make([]admindomain.IntegrationMemberMapping, 0, len(rows))
	for _, row := range rows {
		var slackLinkedAt *time.Time
		if row.SlackUserID != "" {
			slackLinkedAt = &row.SlackLinkedAt
		}
		items = append(items, admindomain.IntegrationMemberMapping{
			UserID: row.UserID, Name: stringValue(row.FullName), Email: row.Email, Role: row.Role,
			SlackUserID:    optionalNonBlankString(row.SlackUserID),
			SlackLinkedVia: optionalNonBlankString(row.SlackLinkedVia),
			SlackLinkedAt:  slackLinkedAt, GitHubUsername: row.GithubUsername,
			CalendarProvider: row.CalendarProvider, CalendarEmail: row.CalendarEmail,
			CalendarState: row.CalendarState,
		})
	}
	return items, nil
}

func getSlackIntegration(ctx context.Context, queries adminsql.Querier, workspaceID uuid.UUID) (*admindomain.SlackIntegrationDetail, error) {
	row, err := queries.GetAdminSlackIntegration(ctx, adminsql.GetAdminSlackIntegrationParams{WorkspaceID: workspaceID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get admin Slack integration: %w", err)
	}
	channels, err := integrationCount(row.ChannelCount, "Slack channels")
	if err != nil {
		return nil, err
	}
	mappings, err := integrationCount(row.ChannelMappingCount, "Slack channel mappings")
	if err != nil {
		return nil, err
	}
	return &admindomain.SlackIntegrationDetail{
		ID: row.ID, TeamID: row.SlackTeamID, TeamName: row.SlackTeamName,
		TeamDomain: row.SlackTeamDomain, InstalledByName: row.InstalledByName,
		InstalledByEmail: row.InstalledByEmail, ChannelCount: channels,
		ChannelMappingCount: mappings, LastSyncedAt: &row.LastSyncedAt, CreatedAt: row.CreatedAt,
	}, nil
}

func listGitHubInstallations(ctx context.Context, queries adminsql.Querier, workspaceID uuid.UUID) ([]admindomain.GitHubInstallationDetail, error) {
	rows, err := queries.ListAdminGitHubInstallations(ctx, adminsql.ListAdminGitHubInstallationsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list admin GitHub installations: %w", err)
	}
	items := make([]admindomain.GitHubInstallationDetail, 0, len(rows))
	for _, row := range rows {
		repositories, err := integrationCount(row.RepositoryCount, "GitHub repositories")
		if err != nil {
			return nil, err
		}
		mappings, err := integrationCount(row.TeamMappingCount, "GitHub team mappings")
		if err != nil {
			return nil, err
		}
		items = append(items, admindomain.GitHubInstallationDetail{
			ID: row.ID, AccountLogin: row.AccountLogin, AccountType: row.AccountType,
			RepositorySelection: row.RepositorySelection, State: row.State,
			InstalledByName: row.InstalledByName, InstalledByEmail: row.InstalledByEmail,
			RepositoryCount: repositories, TeamMappingCount: mappings,
			LastSyncedAt: &row.LastSyncedAt, CreatedAt: row.CreatedAt,
		})
	}
	return items, nil
}

func listCalendarConnections(ctx context.Context, queries adminsql.Querier, workspaceID uuid.UUID) ([]admindomain.CalendarConnectionDetail, error) {
	rows, err := queries.ListAdminCalendarConnections(ctx, adminsql.ListAdminCalendarConnectionsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, fmt.Errorf("list admin Calendar connections: %w", err)
	}
	items := make([]admindomain.CalendarConnectionDetail, 0, len(rows))
	for _, row := range rows {
		items = append(items, admindomain.CalendarConnectionDetail{
			ID: row.ConnectionID, UserID: row.UserID, UserName: stringValue(row.FullName),
			UserEmail: row.Email, Provider: row.Provider, ConnectedEmail: row.ConnectedEmail,
			State: row.SyncStatus, LastSyncedAt: row.LastSyncedAt, CreatedAt: row.CreatedAt,
		})
	}
	return items, nil
}

func getFigmaIntegration(ctx context.Context, queries adminsql.Querier, workspaceID uuid.UUID, now time.Time) (*admindomain.FigmaIntegrationDetail, error) {
	row, err := queries.GetAdminFigmaIntegration(ctx, adminsql.GetAdminFigmaIntegrationParams{WorkspaceID: workspaceID, NowAt: now})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get admin Figma integration: %w", err)
	}
	files, err := integrationCount(row.LinkedFileCount, "Figma linked files")
	if err != nil {
		return nil, err
	}
	webhooks, err := integrationCount(row.WebhookCount, "Figma webhooks")
	if err != nil {
		return nil, err
	}
	return &admindomain.FigmaIntegrationDetail{
		ID: row.ID, AccountLabel: row.AccountLabel, ConnectedByName: row.ConnectedByName,
		ConnectedByEmail: row.ConnectedByEmail, State: row.State, LinkedFileCount: files,
		WebhookCount: webhooks, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	}, nil
}

func integrationSummaries(integrations admindomain.WorkspaceIntegrations, memberCount int) []admindomain.IntegrationSummary {
	slackMappings, githubMappings, calendarMappings := 0, 0, 0
	for _, member := range integrations.MemberMappings {
		if member.SlackUserID != nil {
			slackMappings++
		}
		if member.GitHubUsername != nil {
			githubMappings++
		}
		if member.CalendarProvider != nil {
			calendarMappings++
		}
	}
	slack := admindomain.IntegrationSummary{Provider: "slack", State: "not_connected"}
	if integrations.Slack != nil {
		slack.State, slack.AccountLabel, slack.ConnectionCount = "connected", &integrations.Slack.TeamName, 1
		slack.MappingCount, slack.UnmappedMemberCount = slackMappings, unmappedCount(memberCount, slackMappings)
		slack.LastSyncedAt = integrations.Slack.LastSyncedAt
	}
	github := admindomain.IntegrationSummary{Provider: "github", State: "not_connected"}
	if len(integrations.GitHub) > 0 {
		github.State, github.ConnectionCount = "connected", len(integrations.GitHub)
		github.AccountLabel = &integrations.GitHub[0].AccountLogin
		github.MappingCount, github.UnmappedMemberCount = githubMappings, unmappedCount(memberCount, githubMappings)
		for _, installation := range integrations.GitHub {
			if installation.LastSyncedAt != nil && (github.LastSyncedAt == nil || installation.LastSyncedAt.After(*github.LastSyncedAt)) {
				github.LastSyncedAt = installation.LastSyncedAt
			}
			if installation.State == "suspended" {
				github.State = "suspended"
			}
		}
	}
	calendar := admindomain.IntegrationSummary{Provider: "calendar", State: "not_connected"}
	if len(integrations.Calendar) > 0 {
		calendar.State, calendar.ConnectionCount = "connected", len(integrations.Calendar)
		calendar.MappingCount, calendar.UnmappedMemberCount = calendarMappings, unmappedCount(memberCount, calendarMappings)
		for _, connection := range integrations.Calendar {
			if connection.LastSyncedAt != nil && (calendar.LastSyncedAt == nil || connection.LastSyncedAt.After(*calendar.LastSyncedAt)) {
				calendar.LastSyncedAt = connection.LastSyncedAt
			}
			if connection.State == "failed" {
				calendar.State = "degraded"
			}
		}
	}
	figma := admindomain.IntegrationSummary{Provider: "figma", State: "not_connected"}
	if integrations.Figma != nil {
		figma.State, figma.AccountLabel, figma.ConnectionCount = integrations.Figma.State, &integrations.Figma.AccountLabel, 1
		figma.LastSyncedAt = &integrations.Figma.CreatedAt
	}
	return []admindomain.IntegrationSummary{slack, github, calendar, figma}
}

func integrationCount(value int64, label string) (int, error) {
	count, err := safecast.Int64(value)
	if err != nil {
		return 0, fmt.Errorf("map %s count: %w", label, err)
	}
	return count, nil
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func unmappedCount(members, mappings int) int {
	if mappings >= members {
		return 0
	}
	return members - mappings
}

func optionalNonBlankString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
