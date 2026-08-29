package githubrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ResolveUserByGitHubID(ctx context.Context, githubUserID int64) (uuid.UUID, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return uuid.Nil, err
	}
	userID, err := queries.ResolveUserByGitHubID(ctx, githubsql.ResolveUserByGitHubIDParams{
		GithubUserID: &githubUserID,
	})
	return userID, mapDatabaseError(err)
}

type FortyOneUser = githubshared.FortyOneUser

func (r *Repo) ResolveFortyOneUsersByGitHubIDs(
	ctx context.Context,
	githubUserIDs []int64,
) (map[int64]FortyOneUser, error) {
	if len(githubUserIDs) == 0 {
		return map[int64]FortyOneUser{}, nil
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.ResolveFortyOneUsersByGitHubIDs(ctx, githubsql.ResolveFortyOneUsersByGitHubIDsParams{
		GithubUserIds: githubUserIDs,
	})
	if err != nil {
		return nil, err
	}
	users := make(map[int64]FortyOneUser, len(rows))
	for _, row := range rows {
		if row.GithubUserID == nil {
			return nil, errors.New("linked GitHub user is missing its external identity")
		}
		users[*row.GithubUserID] = FortyOneUser{
			Username:  row.Username,
			FullName:  row.FullName,
			AvatarURL: row.AvatarURL,
		}
	}
	return users, nil
}

func (r *Repo) ResolveFortyOneUserByFullName(
	ctx context.Context,
	fullName string,
) (FortyOneUser, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return FortyOneUser{}, err
	}
	row, err := queries.ResolveFortyOneUserByFullName(ctx, githubsql.ResolveFortyOneUserByFullNameParams{
		FullName: &fullName,
	})
	if err != nil {
		return FortyOneUser{}, mapDatabaseError(err)
	}
	return FortyOneUser{Username: row.Username, FullName: row.FullName, AvatarURL: row.AvatarURL}, nil
}

func (r *Repo) ResolveFortyOneUserByEmail(
	ctx context.Context,
	email string,
) (FortyOneUser, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return FortyOneUser{}, err
	}
	row, err := queries.ResolveFortyOneUserByEmail(ctx, githubsql.ResolveFortyOneUserByEmailParams{Email: email})
	if err != nil {
		return FortyOneUser{}, mapDatabaseError(err)
	}
	return FortyOneUser{Username: row.Username, FullName: row.FullName, AvatarURL: row.AvatarURL}, nil
}

func (r *Repo) UpdateStoryLinkReviewState(
	ctx context.Context,
	storyID, repositoryID uuid.UUID,
	prGitHubID int64,
	reviewState string,
	approved, changesRequested int,
) error {
	approved32, err := safecast.Int32(approved)
	if err != nil {
		return fmt.Errorf("convert approved GitHub review count: %w", err)
	}
	changesRequested32, err := safecast.Int32(changesRequested)
	if err != nil {
		return fmt.Errorf("convert requested GitHub review changes count: %w", err)
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.UpdateGitHubStoryLinkReviewState(ctx, githubsql.UpdateGitHubStoryLinkReviewStateParams{
		ReviewState:             &reviewState,
		ReviewsApproved:         approved32,
		ReviewsChangesRequested: changesRequested32,
		StoryID:                 storyID,
		RepositoryID:            repositoryID,
		GithubID:                &prGitHubID,
	})
}

func (r *Repo) UpdateStoryLinkCheckState(
	ctx context.Context,
	storyID, repositoryID uuid.UUID,
	prGitHubID int64,
	checkState string,
) error {
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.UpdateGitHubStoryLinkCheckState(ctx, githubsql.UpdateGitHubStoryLinkCheckStateParams{
		CheckState:   &checkState,
		StoryID:      storyID,
		RepositoryID: repositoryID,
		GithubID:     &prGitHubID,
	})
}

func (r *Repo) FindStoryLinksByPRNumber(
	ctx context.Context,
	repositoryID uuid.UUID,
	prNumber int,
) ([]StoryMatch, error) {
	prNumber32, err := safecast.Int32(prNumber)
	if err != nil {
		return nil, fmt.Errorf("convert GitHub pull request number: %w", err)
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.FindGitHubStoryLinksByPRNumber(ctx, githubsql.FindGitHubStoryLinksByPRNumberParams{
		RepositoryID: repositoryID,
		GithubNumber: &prNumber32,
	})
	if err != nil {
		return nil, err
	}
	matches := make([]StoryMatch, 0, len(rows))
	for _, row := range rows {
		match, err := mapStoryMatch(row.StoryID, row.StatusID, row.TeamID, row.TeamCode, row.SequenceID, row.Title)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func (r *Repo) ResolveOrCreateLabelsByName(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
	names []string,
) ([]uuid.UUID, error) {
	labelIDs := make([]uuid.UUID, 0, len(names))
	err := r.withinTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}, func(queries githubsql.Querier) error {
		for _, rawName := range names {
			name := strings.TrimSpace(rawName)
			if name == "" {
				continue
			}
			if err := queries.LockGitHubLabelName(ctx, githubsql.LockGitHubLabelNameParams{
				WorkspaceID: workspaceID,
				LabelName:   name,
			}); err != nil {
				return err
			}
			labelID, err := queries.FindGitHubLabelByName(ctx, githubsql.FindGitHubLabelByNameParams{
				WorkspaceID: &workspaceID,
				LabelName:   name,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				labelID, err = queries.InsertGitHubLabel(ctx, githubsql.InsertGitHubLabelParams{
					LabelName:   name,
					WorkspaceID: &workspaceID,
					TeamID:      &teamID,
				})
			}
			if err != nil {
				return err
			}
			labelIDs = append(labelIDs, labelID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return labelIDs, nil
}
