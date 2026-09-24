package slackrepository

import (
	"context"
	"fmt"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repo) RegisterSlackFileImport(ctx context.Context, input slackdomain.RegisterFileImport) (uuid.UUID, error) {
	id, err := repository.queries.RegisterSlackFileImport(ctx, slacksql.RegisterSlackFileImportParams{
		SlackUserID: input.SlackUserID, SlackChannelID: input.ChannelID,
		SlackThreadTs: input.ThreadTS, SlackMessageTs: input.MessageTS,
		SlackFileID: input.FileID, IdempotencyKey: input.IdempotencyKey,
		ActorID: input.ActorID, StoryID: input.StoryID,
		SlackWorkspaceID: input.InstallationID, WorkspaceID: input.WorkspaceID,
		InstallationGeneration: input.InstallGeneration, SlackTeamID: input.SlackTeamID,
	})
	return id, mapDatabaseError(err)
}

func (repository *Repo) ClaimSlackFileImport(ctx context.Context, id uuid.UUID) (slackdomain.FileImport, bool, error) {
	row, err := repository.queries.ClaimSlackFileImport(ctx, slacksql.ClaimSlackFileImportParams{ImportID: id})
	if err == pgx.ErrNoRows {
		return slackdomain.FileImport{}, false, nil
	}
	if err != nil {
		return slackdomain.FileImport{}, false, mapDatabaseError(err)
	}
	return slackdomain.FileImport{
		ID: row.ID, WorkspaceID: row.WorkspaceID, SlackWorkspaceID: row.SlackWorkspaceID,
		InstallGeneration: row.InstallationGeneration, SlackTeamID: row.SlackTeamID,
		SlackUserID: row.SlackUserID, ChannelID: row.SlackChannelID,
		ThreadTS: row.SlackThreadTs, MessageTS: row.SlackMessageTs, FileID: row.SlackFileID,
		StoryID: row.StoryID, ActorID: row.ActorID, AttemptCount: row.AttemptCount,
	}, true, nil
}

func (repository *Repo) AuthorizeSlackFileImport(ctx context.Context, id uuid.UUID) (bool, error) {
	allowed, err := repository.queries.AuthorizeSlackFileImport(ctx, slacksql.AuthorizeSlackFileImportParams{ImportID: id})
	return allowed, mapDatabaseError(err)
}

func (repository *Repo) CompleteSlackFileImport(ctx context.Context, id uuid.UUID, attempt int32, attachmentID uuid.UUID) error {
	rows, err := repository.queries.CompleteSlackFileImport(ctx, slacksql.CompleteSlackFileImportParams{
		AttachmentID: attachmentID, ImportID: id, AttemptCount: attempt,
	})
	if err != nil {
		return mapDatabaseError(err)
	}
	if rows == 0 {
		return fmt.Errorf("Slack file import %s was no longer claimed: %w", id, slackdomain.ErrConflict)
	}
	return nil
}

func (repository *Repo) FailSlackFileImport(ctx context.Context, id uuid.UUID, attempt int32) error {
	_, err := repository.queries.FailSlackFileImport(ctx, slacksql.FailSlackFileImportParams{
		ImportID: id, AttemptCount: attempt,
	})
	return mapDatabaseError(err)
}

func (repository *Repo) CancelSlackFileImport(ctx context.Context, id uuid.UUID, attempt int32) error {
	_, err := repository.queries.CancelSlackFileImport(ctx, slacksql.CancelSlackFileImportParams{
		ImportID: id, AttemptCount: attempt,
	})
	return mapDatabaseError(err)
}

func (repository *Repo) ListRecoverableSlackFileImports(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit < 1 || limit > 100 {
		limit = 100
	}
	ids, err := repository.queries.ListRecoverableSlackFileImports(ctx, slacksql.ListRecoverableSlackFileImportsParams{ResultLimit: int32(limit)})
	return ids, mapDatabaseError(err)
}
