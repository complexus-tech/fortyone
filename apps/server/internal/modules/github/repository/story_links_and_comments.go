package githubrepository

import (
	"context"
	"fmt"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/google/uuid"
)

type StoryGitHubLink = githubshared.StoryGitHubLink
type StoryIssueWithInstallation = githubshared.StoryIssueWithInstallation

func (r *Repo) GetStoryLinkedIssues(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
) ([]StoryIssueWithInstallation, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.GetStoryLinkedGitHubIssues(ctx, githubsql.GetStoryLinkedGitHubIssuesParams{
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return nil, err
	}
	issues := make([]StoryIssueWithInstallation, 0, len(rows))
	for _, row := range rows {
		if row.GithubNumber == nil {
			return nil, fmt.Errorf("GitHub issue link in repository %s is missing its issue number", row.RepositoryID)
		}
		issues = append(issues, StoryIssueWithInstallation{
			RepositoryID:         row.RepositoryID,
			GitHubNumber:         int(*row.GithubNumber),
			OwnerLogin:           row.OwnerLogin,
			RepositorySlug:       row.RepositorySlug,
			GitHubInstallationID: row.GithubInstallationID,
		})
	}
	return issues, nil
}

func (r *Repo) GetStoryGitHubLinks(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
) ([]StoryGitHubLink, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.GetStoryGitHubLinks(ctx, githubsql.GetStoryGitHubLinksParams{
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return nil, err
	}
	links := make([]StoryGitHubLink, 0, len(rows))
	for _, row := range rows {
		var githubNumber *int
		if row.GithubNumber != nil {
			value := int(*row.GithubNumber)
			githubNumber = &value
		}
		links = append(links, StoryGitHubLink{
			ID:                 row.ID,
			ExternalType:       row.ExternalType,
			GitHubNumber:       githubNumber,
			URL:                row.URL,
			Title:              row.Title,
			State:              row.State,
			ReviewState:        row.ReviewState,
			ReviewsApproved:    int(row.ReviewsApproved),
			ReviewsChangesReq:  int(row.ReviewsChangesRequested),
			CheckState:         row.CheckState,
			RepositoryFullName: row.RepositoryFullName,
			RefName:            row.RefName,
			CreatedAt:          row.CreatedAt,
		})
	}
	return links, nil
}

func (r *Repo) DeleteStoryGitHubLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.DeleteGitHubStoryLink(ctx, githubsql.DeleteGitHubStoryLinkParams{
		WorkspaceID: workspaceID,
		LinkID:      linkID,
	})
}

func (r *Repo) RecordOutboundGitHubComment(
	ctx context.Context,
	workspaceID, storyID, repositoryID uuid.UUID,
	githubCommentID int64,
	localCommentID *uuid.UUID,
	createdByUserID uuid.UUID,
) error {
	if githubCommentID == 0 {
		return nil
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.RecordOutboundGitHubComment(ctx, githubsql.RecordOutboundGitHubCommentParams{
		WorkspaceID:     workspaceID,
		StoryID:         storyID,
		RepositoryID:    repositoryID,
		LocalCommentID:  localCommentID,
		GithubCommentID: githubCommentID,
		CreatedByUserID: &createdByUserID,
	})
}

func (r *Repo) ReserveInboundGitHubComment(
	ctx context.Context,
	workspaceID, storyID, repositoryID uuid.UUID,
	githubCommentID int64,
	createdByUserID uuid.UUID,
) (bool, error) {
	if githubCommentID == 0 {
		return false, nil
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return false, err
	}
	affected, err := queries.ReserveInboundGitHubComment(ctx, githubsql.ReserveInboundGitHubCommentParams{
		WorkspaceID:     workspaceID,
		StoryID:         storyID,
		RepositoryID:    repositoryID,
		GithubCommentID: githubCommentID,
		CreatedByUserID: &createdByUserID,
	})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *Repo) CompleteInboundGitHubComment(
	ctx context.Context,
	repositoryID uuid.UUID,
	githubCommentID int64,
	localCommentID uuid.UUID,
) error {
	if githubCommentID == 0 {
		return nil
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.CompleteInboundGitHubComment(ctx, githubsql.CompleteInboundGitHubCommentParams{
		LocalCommentID:  &localCommentID,
		RepositoryID:    repositoryID,
		GithubCommentID: githubCommentID,
	})
}

func (r *Repo) DeleteGitHubCommentLink(
	ctx context.Context,
	repositoryID uuid.UUID,
	githubCommentID int64,
) error {
	if githubCommentID == 0 {
		return nil
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.DeleteGitHubCommentLink(ctx, githubsql.DeleteGitHubCommentLinkParams{
		RepositoryID:    repositoryID,
		GithubCommentID: githubCommentID,
	})
}

func (r *Repo) IsOutboundGitHubComment(
	ctx context.Context,
	repositoryID uuid.UUID,
	githubCommentID int64,
) (bool, error) {
	if githubCommentID == 0 {
		return false, nil
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return false, err
	}
	return queries.IsOutboundGitHubComment(ctx, githubsql.IsOutboundGitHubCommentParams{
		RepositoryID:    repositoryID,
		GithubCommentID: githubCommentID,
	})
}
