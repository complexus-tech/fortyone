// Package githubrepository implements GitHub persistence with module-local
// sqlc queries and native pgx transactions.
package githubrepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryUnavailable = errors.New("GitHub repository is not configured")

type Repo struct {
	queries        githubsql.Querier
	runTransaction func(context.Context, pgx.TxOptions, func(githubsql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repo {
	if pool == nil {
		return &Repo{}
	}

	queries := githubsql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	return &Repo{
		queries: queries,
		runTransaction: func(
			ctx context.Context,
			options pgx.TxOptions,
			operation func(githubsql.Querier) error,
		) error {
			return transactor.WithinTransaction(ctx, options, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
	}
}

func (r *Repo) configuredQueries() (githubsql.Querier, error) {
	if r == nil || r.queries == nil {
		return nil, errRepositoryUnavailable
	}
	return r.queries, nil
}

func (r *Repo) withinTransaction(
	ctx context.Context,
	options pgx.TxOptions,
	operation func(githubsql.Querier) error,
) error {
	if r == nil || r.queries == nil || r.runTransaction == nil {
		return errRepositoryUnavailable
	}
	return mapDatabaseError(r.runTransaction(ctx, options, operation))
}

// mapDatabaseError preserves the existing repository contract while callers
// migrate independently from database/sql sentinels to domain errors.
func mapDatabaseError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return sql.ErrNoRows
	}
	return err
}

func (r *Repo) GetWorkspaceRole(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
) (authorization.WorkspaceRole, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return "", err
	}
	role, err := queries.GetGitHubWorkspaceRole(ctx, githubsql.GetGitHubWorkspaceRoleParams{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return "", mapDatabaseError(err)
	}
	return authorization.WorkspaceRole(role), nil
}

// Compatibility aliases keep existing repository callers source-compatible
// while the service consumes GitHub-owned neutral records.
type StoryMatch = githubshared.StoryMatch
type RepoByExternalRow = githubshared.RepositoryRecord
type syncLinkRow = githubshared.IssueSyncLinkRecord
type bidirectionalLinkRow = githubshared.BidirectionalIssueSyncLink
type issueStoryLinkRow = githubshared.IssueStoryLink
type statusRow = githubshared.TeamStatus

func (r *Repo) EnsureStoryLink(
	ctx context.Context,
	storyID uuid.UUID,
	title *string,
	url string,
) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil
	}

	return r.withinTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}, func(queries githubsql.Querier) error {
		lockParams := githubsql.LockStoryURLParams{StoryID: storyID, URL: url}
		if err := queries.LockStoryURL(ctx, lockParams); err != nil {
			return err
		}
		exists, err := queries.GitHubStoryURLExists(ctx, githubsql.GitHubStoryURLExistsParams{
			StoryID: storyID,
			URL:     url,
		})
		if err != nil || exists {
			return err
		}
		return queries.InsertGitHubStoryURL(ctx, githubsql.InsertGitHubStoryURLParams{
			Title:   title,
			URL:     url,
			StoryID: storyID,
		})
	})
}
