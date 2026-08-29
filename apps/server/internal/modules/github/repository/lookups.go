package githubrepository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) FindRepositoryByExternalID(
	ctx context.Context,
	repositoryExternalID int64,
) (RepoByExternalRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return RepoByExternalRow{}, err
	}
	row, err := queries.FindGitHubRepositoryByExternalID(ctx, githubsql.FindGitHubRepositoryByExternalIDParams{
		GithubRepositoryID: repositoryExternalID,
	})
	if err != nil {
		return RepoByExternalRow{}, mapDatabaseError(err)
	}
	return RepoByExternalRow{
		ID:                   row.ID,
		WorkspaceID:          row.WorkspaceID,
		WorkspaceSlug:        row.WorkspaceSlug,
		FullName:             row.FullName,
		OwnerLogin:           row.OwnerLogin,
		RepositorySlug:       row.RepositorySlug,
		DefaultBranch:        row.DefaultBranch,
		GitHubInstallationID: row.GithubInstallationID,
	}, nil
}

func (r *Repo) FindRepositoryByID(
	ctx context.Context,
	workspaceID, repositoryID uuid.UUID,
) (RepoByExternalRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return RepoByExternalRow{}, err
	}
	row, err := queries.FindGitHubRepositoryByID(ctx, githubsql.FindGitHubRepositoryByIDParams{
		WorkspaceID:  workspaceID,
		RepositoryID: repositoryID,
	})
	if err != nil {
		return RepoByExternalRow{}, mapDatabaseError(err)
	}
	return RepoByExternalRow{
		ID:                   row.ID,
		WorkspaceID:          row.WorkspaceID,
		WorkspaceSlug:        row.WorkspaceSlug,
		FullName:             row.FullName,
		OwnerLogin:           row.OwnerLogin,
		RepositorySlug:       row.RepositorySlug,
		DefaultBranch:        row.DefaultBranch,
		GitHubInstallationID: row.GithubInstallationID,
	}, nil
}

func (r *Repo) FindIssueSyncLinkByRepositoryID(
	ctx context.Context,
	repositoryID uuid.UUID,
) (syncLinkRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return syncLinkRow{}, err
	}
	row, err := queries.FindGitHubIssueSyncLinkByRepositoryID(ctx, githubsql.FindGitHubIssueSyncLinkByRepositoryIDParams{
		RepositoryID: repositoryID,
	})
	if err != nil {
		return syncLinkRow{}, mapDatabaseError(err)
	}
	return syncLinkRow{
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
	}, nil
}

func (r *Repo) FindBidirectionalIssueSyncLinkByTeamID(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
) (bidirectionalLinkRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return bidirectionalLinkRow{}, err
	}
	row, err := queries.FindBidirectionalGitHubIssueSyncLinkByTeamID(ctx, githubsql.FindBidirectionalGitHubIssueSyncLinkByTeamIDParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return bidirectionalLinkRow{}, mapDatabaseError(err)
	}
	return bidirectionalLinkRow{
		ID:                   row.ID,
		RepositoryID:         row.RepositoryID,
		TeamID:               row.TeamID,
		SyncDirection:        row.SyncDirection,
		RepositoryName:       row.RepositoryName,
		OwnerLogin:           row.OwnerLogin,
		RepositorySlug:       row.RepositorySlug,
		RepositoryHTMLURL:    row.RepositoryHtmlURL,
		GitHubInstallationID: row.GithubInstallationID,
	}, nil
}

func (r *Repo) ResolveStoriesByRefs(
	ctx context.Context,
	workspaceID uuid.UUID,
	refs []string,
) ([]StoryMatch, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return nil, nil
	}

	matches := make([]StoryMatch, 0, len(refs))
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		normalized := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(ref)))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}

		letters := 0
		for index, character := range normalized {
			if character >= '0' && character <= '9' {
				letters = index
				break
			}
		}
		if letters == 0 || letters == len(normalized) {
			continue
		}
		sequenceID, err := strconv.ParseInt(normalized[letters:], 10, 32)
		if err != nil || sequenceID <= 0 {
			continue
		}
		sequenceID32 := int32(sequenceID)
		row, err := queries.ResolveGitHubStoryReference(ctx, githubsql.ResolveGitHubStoryReferenceParams{
			WorkspaceID: workspaceID,
			TeamCode:    normalized[:letters],
			SequenceID:  &sequenceID32,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		match, err := mapStoryMatch(row.StoryID, row.StatusID, row.TeamID, row.TeamCode, row.SequenceID, row.Title)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func (r *Repo) FindStoryLink(
	ctx context.Context,
	repositoryID uuid.UUID,
	externalType string,
	githubID int64,
	refName *string,
) (uuid.UUID, uuid.UUID, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	row, err := queries.FindGitHubStoryLink(ctx, githubsql.FindGitHubStoryLinkParams{
		RepositoryID: repositoryID,
		ExternalType: externalType,
		GithubID:     &githubID,
		RefName:      refName,
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, mapDatabaseError(err)
	}
	return row.ID, row.StoryID, nil
}

func (r *Repo) FindIssueStoryLinkByStoryID(
	ctx context.Context,
	workspaceID, storyID, repositoryID uuid.UUID,
) (issueStoryLinkRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return issueStoryLinkRow{}, err
	}
	row, err := queries.FindGitHubIssueStoryLinkByStoryID(ctx, githubsql.FindGitHubIssueStoryLinkByStoryIDParams{
		WorkspaceID:  workspaceID,
		StoryID:      storyID,
		RepositoryID: repositoryID,
	})
	if err != nil {
		return issueStoryLinkRow{}, mapDatabaseError(err)
	}
	return mapIssueStoryLink(row.ID, row.StoryID, row.RepositoryID, row.GithubID, row.GithubNumber,
		row.URL, row.Title, row.State, row.LastSyncedFrom, row.LastSyncHash)
}

func (r *Repo) FindIssueStoryLinkByExternalID(
	ctx context.Context,
	repositoryID uuid.UUID,
	githubID int64,
) (issueStoryLinkRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return issueStoryLinkRow{}, err
	}
	row, err := queries.FindGitHubIssueStoryLinkByExternalID(ctx, githubsql.FindGitHubIssueStoryLinkByExternalIDParams{
		RepositoryID: repositoryID,
		GithubID:     &githubID,
	})
	if err != nil {
		return issueStoryLinkRow{}, mapDatabaseError(err)
	}
	return mapIssueStoryLink(row.ID, row.StoryID, row.RepositoryID, row.GithubID, row.GithubNumber,
		row.URL, row.Title, row.State, row.LastSyncedFrom, row.LastSyncHash)
}

func (r *Repo) UpdateStoryLinkSyncState(
	ctx context.Context,
	linkID uuid.UUID,
	source, syncHash string,
) error {
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.UpdateGitHubStoryLinkSyncState(ctx, githubsql.UpdateGitHubStoryLinkSyncStateParams{
		SyncSource: &source,
		SyncHash:   &syncHash,
		LinkID:     linkID,
	})
}

func mapStoryMatch(
	storyID uuid.UUID,
	statusID *uuid.UUID,
	teamID uuid.UUID,
	teamCode string,
	sequenceID *int32,
	title string,
) (StoryMatch, error) {
	if statusID == nil || sequenceID == nil {
		return StoryMatch{}, fmt.Errorf("GitHub story match %s is missing workflow identity", storyID)
	}
	return StoryMatch{
		StoryID:    storyID,
		StatusID:   *statusID,
		TeamID:     teamID,
		TeamCode:   teamCode,
		SequenceID: int(*sequenceID),
		Title:      title,
	}, nil
}

func mapIssueStoryLink(
	id, storyID, repositoryID uuid.UUID,
	githubID *int64,
	githubNumber *int32,
	url string,
	title, state, lastSyncedFrom, lastSyncHash *string,
) (issueStoryLinkRow, error) {
	if githubID == nil || githubNumber == nil {
		return issueStoryLinkRow{}, fmt.Errorf("GitHub issue link %s is missing its external identity", id)
	}
	return issueStoryLinkRow{
		ID:             id,
		StoryID:        storyID,
		RepositoryID:   repositoryID,
		GitHubID:       *githubID,
		GitHubNumber:   int(*githubNumber),
		URL:            url,
		Title:          title,
		State:          state,
		LastSyncedFrom: lastSyncedFrom,
		LastSyncHash:   lastSyncHash,
	}, nil
}
