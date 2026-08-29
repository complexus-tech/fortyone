package githubrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func (r *Repo) UpsertStoryLink(
	ctx context.Context,
	workspaceID, storyID, repositoryID uuid.UUID,
	externalType string,
	githubID int64,
	githubNumber int,
	refName *string,
	url, title, state string,
	metadata any,
) error {
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	githubNumber32, err := safecast.Int32(githubNumber)
	if err != nil {
		return fmt.Errorf("convert GitHub external number: %w", err)
	}
	return queries.UpsertGitHubStoryLink(ctx, githubsql.UpsertGitHubStoryLinkParams{
		WorkspaceID:  workspaceID,
		StoryID:      storyID,
		RepositoryID: repositoryID,
		ExternalType: externalType,
		GithubID:     &githubID,
		GithubNumber: &githubNumber32,
		RefName:      refName,
		URL:          url,
		Title:        &title,
		State:        &state,
		Metadata:     metadataJSON,
	})
}

func (r *Repo) GetStatusCategory(ctx context.Context, statusID uuid.UUID) (string, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return "", err
	}
	category, err := queries.GetGitHubStatusCategory(ctx, githubsql.GetGitHubStatusCategoryParams{
		StatusID: statusID,
	})
	if err != nil {
		return "", mapDatabaseError(err)
	}
	if category == nil {
		return "", errors.New("GitHub status category is missing")
	}
	return *category, nil
}
