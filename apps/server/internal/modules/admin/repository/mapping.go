package adminrepository

import (
	"fmt"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

func userFromGetRow(row adminsql.GetAdminUserRow) (admindomain.UserSummary, error) {
	workspaceCount, err := safecast.Int64(row.WorkspaceCount)
	if err != nil {
		return admindomain.UserSummary{}, fmt.Errorf("map user workspace count: %w", err)
	}
	return admindomain.UserSummary{
		ID: row.UserID, Username: row.Username, Email: row.Email,
		FullName: stringValue(row.FullName), AvatarURL: stringValue(row.AvatarURL),
		IsActive: row.IsActive, IsSystem: row.IsSystem, IsInternal: row.IsInternal,
		LoginReactivationPolicy: admindomain.LoginReactivationPolicy(row.LoginReactivationPolicy),
		LastLoginAt:             row.LastLoginAt, LastUsedWorkspaceID: row.LastUsedWorkspaceID,
		LastUsedWorkspace: row.LastUsedWorkspace, GitHubUsername: row.GithubUsername,
		WorkspaceCount: workspaceCount, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func userFromListRow(row adminsql.ListAdminUsersRow) (admindomain.UserSummary, error) {
	return userFromGetRow(adminsql.GetAdminUserRow{
		UserID: row.UserID, Username: row.Username, Email: row.Email,
		FullName: row.FullName, AvatarURL: row.AvatarURL, IsActive: row.IsActive,
		IsSystem: row.IsSystem, IsInternal: row.IsInternal, LastLoginAt: row.LastLoginAt,
		LoginReactivationPolicy: row.LoginReactivationPolicy,
		LastUsedWorkspaceID:     row.LastUsedWorkspaceID, LastUsedWorkspace: row.LastUsedWorkspace,
		GithubUsername: row.GithubUsername, WorkspaceCount: row.WorkspaceCount,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	})
}

func workspaceFromGetRow(row adminsql.GetAdminWorkspaceRow) (admindomain.WorkspaceSummary, error) {
	memberCount, teamCount, storyCount, seats, err := workspaceCounts(
		row.MemberCount, row.TeamCount, row.StoryCount, row.SubscriptionSeats,
	)
	if err != nil {
		return admindomain.WorkspaceSummary{}, err
	}
	return admindomain.WorkspaceSummary{
		ID: row.WorkspaceID, Name: row.Name, Slug: row.Slug, AvatarURL: row.AvatarURL,
		Color: row.WorkspaceColor, TeamSize: row.WorkspaceTeamSize,
		TrialEndsOn: row.TrialEndsOn, DeletedAt: row.DeletedAt,
		LastAccessedAt: row.LastAccessedAt, CreatedByUserID: row.CreatedBy,
		CreatedByEmail: row.CreatedByEmail, CreatedByName: row.CreatedByName,
		MemberCount: memberCount, TeamCount: teamCount, StoryCount: storyCount,
		SubscriptionTier: enumString(row.SubscriptionTier), SubscriptionStatus: row.SubscriptionStatus,
		SubscriptionSeats: seats, StripeCustomerID: row.StripeCustomerID,
		StripeSubscriptionID: row.StripeSubscriptionID, SlackInstalled: row.SlackInstalled,
		GitHubInstalled: row.GithubInstalled, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func workspaceFromListRow(row adminsql.ListAdminWorkspacesRow) (admindomain.WorkspaceSummary, error) {
	memberCount, teamCount, storyCount, seats, err := workspaceCounts(
		row.MemberCount, row.TeamCount, row.StoryCount, row.SubscriptionSeats,
	)
	if err != nil {
		return admindomain.WorkspaceSummary{}, err
	}
	return admindomain.WorkspaceSummary{
		ID: row.WorkspaceID, Name: row.Name, Slug: row.Slug, AvatarURL: row.AvatarURL,
		Color: row.Color, TeamSize: row.TeamSize, TrialEndsOn: row.TrialEndsOn,
		DeletedAt: row.DeletedAt, LastAccessedAt: row.LastAccessedAt,
		CreatedByUserID: row.CreatedBy, CreatedByEmail: row.CreatedByEmail,
		CreatedByName: row.CreatedByName, MemberCount: memberCount, TeamCount: teamCount,
		StoryCount: storyCount, SubscriptionTier: enumString(row.SubscriptionTier),
		SubscriptionStatus: row.SubscriptionStatus, SubscriptionSeats: seats,
		StripeCustomerID: row.StripeCustomerID, StripeSubscriptionID: row.StripeSubscriptionID,
		SlackInstalled: row.SlackInstalled, GitHubInstalled: row.GithubInstalled,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func workspaceCounts(
	members, teams, stories int64,
	seats *int32,
) (int, int, int, *int, error) {
	memberCount, err := safecast.Int64(members)
	if err != nil {
		return 0, 0, 0, nil, fmt.Errorf("map workspace member count: %w", err)
	}
	teamCount, err := safecast.Int64(teams)
	if err != nil {
		return 0, 0, 0, nil, fmt.Errorf("map workspace team count: %w", err)
	}
	storyCount, err := safecast.Int64(stories)
	if err != nil {
		return 0, 0, 0, nil, fmt.Errorf("map workspace story count: %w", err)
	}
	if seats == nil {
		return memberCount, teamCount, storyCount, nil, nil
	}
	seatCount, err := safecast.Int64(int64(*seats))
	if err != nil {
		return 0, 0, 0, nil, fmt.Errorf("map subscription seat count: %w", err)
	}
	return memberCount, teamCount, storyCount, &seatCount, nil
}

func enumString(value *adminsql.SubscriptionTierEnum) *string {
	if value == nil {
		return nil
	}
	mapped := string(*value)
	return &mapped
}

func dashboardFromRow(row adminsql.GetAdminDashboardSummaryRow) (admindomain.DashboardSummary, error) {
	values := []int64{
		row.TotalWorkspaces, row.ActiveTrials, row.ExpiredTrials, row.PaidWorkspaces,
		row.DeletedWorkspaces, row.TotalUsers, row.InternalUsers, row.ActiveSubscriptions,
		row.SlackInstallations, row.GithubInstallations, row.RecentAdminAuditLogs,
	}
	converted := make([]int, len(values))
	for index, value := range values {
		mapped, err := safecast.Int64(value)
		if err != nil {
			return admindomain.DashboardSummary{}, fmt.Errorf("map dashboard count: %w", err)
		}
		converted[index] = mapped
	}
	return admindomain.DashboardSummary{
		TotalWorkspaces: converted[0], ActiveTrials: converted[1],
		ExpiredTrials: converted[2], PaidWorkspaces: converted[3],
		DeletedWorkspaces: converted[4], TotalUsers: converted[5],
		InternalUsers: converted[6], ActiveSubscriptions: converted[7],
		SlackInstallations: converted[8], GitHubInstallations: converted[9],
		RecentAdminAuditLogs: converted[10],
	}, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
