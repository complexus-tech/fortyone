package githubrepository

import (
	"context"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/google/uuid"
)

func (r *Repo) GetWorkspaceSettings(
	ctx context.Context,
	workspaceID uuid.UUID,
) (githubshared.CoreWorkspaceSettings, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, err
	}
	row, err := queries.GetOrCreateGitHubWorkspaceSettings(ctx, githubsql.GetOrCreateGitHubWorkspaceSettingsParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, mapDatabaseError(err)
	}
	return newCoreWorkspaceSettings(
		row.WorkspaceID,
		row.BranchFormat,
		row.LinkCommitsByMagicWords,
		row.SyncAssignees,
		row.SyncLabels,
		row.AutoPopulatePrBody,
		row.CloseOnCommitKeywords,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *Repo) GetWorkspaceSettingsByWorkspaceID(
	ctx context.Context,
	workspaceID uuid.UUID,
) (githubshared.CoreWorkspaceSettings, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, err
	}
	row, err := queries.GetGitHubWorkspaceSettings(ctx, githubsql.GetGitHubWorkspaceSettingsParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, mapDatabaseError(err)
	}
	return newCoreWorkspaceSettings(
		row.WorkspaceID,
		row.BranchFormat,
		row.LinkCommitsByMagicWords,
		row.SyncAssignees,
		row.SyncLabels,
		row.AutoPopulatePrBody,
		row.CloseOnCommitKeywords,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *Repo) UpdateWorkspaceSettings(
	ctx context.Context,
	workspaceID uuid.UUID,
	updates githubshared.CoreUpdateWorkspaceSettingsInput,
) (githubshared.CoreWorkspaceSettings, error) {
	if _, err := r.GetWorkspaceSettings(ctx, workspaceID); err != nil {
		return githubshared.CoreWorkspaceSettings{}, err
	}

	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, err
	}
	row, err := queries.UpdateGitHubWorkspaceSettings(ctx, githubsql.UpdateGitHubWorkspaceSettingsParams{
		BranchFormat:            updates.BranchFormat,
		LinkCommitsByMagicWords: updates.LinkCommitsByMagicWords,
		SyncAssignees:           updates.SyncAssignees,
		SyncLabels:              updates.SyncLabels,
		AutoPopulatePrBody:      updates.AutoPopulatePRBody,
		CloseOnCommitKeywords:   updates.CloseOnCommitKeywords,
		WorkspaceID:             workspaceID,
	})
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, mapDatabaseError(err)
	}
	return newCoreWorkspaceSettings(
		row.WorkspaceID,
		row.BranchFormat,
		row.LinkCommitsByMagicWords,
		row.SyncAssignees,
		row.SyncLabels,
		row.AutoPopulatePrBody,
		row.CloseOnCommitKeywords,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *Repo) ListInstallations(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]githubshared.CoreInstallation, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.ListGitHubInstallations(ctx, githubsql.ListGitHubInstallationsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, err
	}
	items := make([]githubshared.CoreInstallation, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreInstallation{
			ID:                   row.ID,
			GitHubInstallationID: row.GithubInstallationID,
			AccountID:            row.AccountID,
			AccountLogin:         row.AccountLogin,
			AccountType:          row.AccountType,
			AccountAvatarURL:     row.AccountAvatarURL,
			RepositorySelection:  row.RepositorySelection,
			IsActive:             row.IsActive,
			SuspendedAt:          row.SuspendedAt,
			DisconnectedAt:       row.DisconnectedAt,
			CreatedAt:            row.CreatedAt,
			UpdatedAt:            row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) ListRepositories(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]githubshared.CoreRepository, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.ListGitHubRepositories(ctx, githubsql.ListGitHubRepositoriesParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, err
	}
	items := make([]githubshared.CoreRepository, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreRepository{
			ID:                 row.ID,
			InstallationID:     row.InstallationID,
			GitHubRepositoryID: row.GithubRepositoryID,
			OwnerLogin:         row.OwnerLogin,
			Name:               row.Name,
			FullName:           row.FullName,
			Description:        row.Description,
			HTMLURL:            row.HtmlURL,
			DefaultBranch:      row.DefaultBranch,
			IsPrivate:          row.IsPrivate,
			IsArchived:         row.IsArchived,
			IsDisabled:         row.IsDisabled,
			IsActive:           row.IsActive,
			LastSyncedAt:       row.LastSyncedAt,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) ListIssueSyncLinks(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]githubshared.CoreIssueSyncLink, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.ListGitHubIssueSyncLinks(ctx, githubsql.ListGitHubIssueSyncLinksParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, err
	}
	items := make([]githubshared.CoreIssueSyncLink, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreIssueSyncLink{
			ID:             row.ID,
			RepositoryID:   row.RepositoryID,
			RepositoryName: row.RepositoryName,
			TeamID:         row.TeamID,
			TeamName:       row.TeamName,
			TeamColor:      row.TeamColor,
			SyncDirection:  row.SyncDirection,
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}
	return items, nil
}
