package slackrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const MaxSlackAgentGuidanceRunes = slackdomain.MaxAgentGuidanceRunes

type (
	AgentSettingsRecord = slackdomain.AgentSettings
	AgentSettingsInput  = slackdomain.UpdateAgentSettings
)

// GetAgentSettings is the provider-worker read. Human dashboard callers use
// GetAgentSettingsForAdmin so they cannot bypass the current actor predicate.
func (repository *Repo) GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (AgentSettingsRecord, error) {
	if workspaceID == uuid.Nil {
		return AgentSettingsRecord{}, errors.Join(slackdomain.ErrInvalidInput, errors.New("workspace is required"))
	}
	row, err := repository.queries.GetSlackAgentSettingsTrusted(ctx, slacksql.GetSlackAgentSettingsTrustedParams{WorkspaceID: workspaceID})
	if err != nil {
		return AgentSettingsRecord{}, fmt.Errorf("get trusted Slack agent settings: %w", mapDatabaseError(err))
	}
	return slackdomain.AgentSettings{WorkspaceID: row.WorkspaceID, Guidance: row.Guidance, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (repository *Repo) GetAgentSettingsForAdmin(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
) (AgentSettingsRecord, error) {
	if err := query.Validate(); err != nil {
		return AgentSettingsRecord{}, err
	}
	row, err := repository.queries.GetSlackAgentSettingsForAdmin(ctx, slacksql.GetSlackAgentSettingsForAdminParams{
		WorkspaceID: query.WorkspaceID,
		ActorID:     query.ActorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentSettingsRecord{}, slackdomain.ErrForbidden
	}
	if err != nil {
		return AgentSettingsRecord{}, fmt.Errorf("get Slack agent settings for admin: %w", mapDatabaseError(err))
	}
	return slackdomain.AgentSettings{WorkspaceID: row.WorkspaceID, Guidance: row.Guidance, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (repository *Repo) UpsertAgentSettingsForAdmin(
	ctx context.Context,
	command slackdomain.UpdateAgentSettingsCommand,
) (AgentSettingsRecord, error) {
	command.Guidance = strings.TrimSpace(command.Guidance)
	if err := command.Validate(); err != nil {
		return AgentSettingsRecord{}, err
	}
	row, err := repository.queries.UpsertSlackAgentSettingsForAdmin(ctx, slacksql.UpsertSlackAgentSettingsForAdminParams{
		WorkspaceID: command.WorkspaceID,
		ActorID:     command.ActorID,
		Guidance:    command.Guidance,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentSettingsRecord{}, slackdomain.ErrForbidden
	}
	if err != nil {
		return AgentSettingsRecord{}, fmt.Errorf("upsert Slack agent settings for admin: %w", mapDatabaseError(err))
	}
	return slackdomain.AgentSettings{WorkspaceID: row.WorkspaceID, Guidance: row.Guidance, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}
