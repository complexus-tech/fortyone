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

func (repository *Repo) UpsertChannels(ctx context.Context, command slackdomain.SyncChannelsCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}
	return repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		if _, err := queries.LockSlackWorkspaceAdmin(ctx, slacksql.LockSlackWorkspaceAdminParams{
			WorkspaceID: command.WorkspaceID,
			ActorID:     command.ActorID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return slackdomain.ErrForbidden
		} else if err != nil {
			return fmt.Errorf("authorize Slack channel sync: %w", err)
		}
		if _, err := queries.LockSlackInstallationForChannelSync(ctx, slacksql.LockSlackInstallationForChannelSyncParams{
			SlackWorkspaceID:       command.InstallationID,
			WorkspaceID:            command.WorkspaceID,
			InstallationGeneration: command.InstallationGeneration,
		}); errors.Is(err, pgx.ErrNoRows) {
			return slackdomain.ErrConflict
		} else if err != nil {
			return fmt.Errorf("lock Slack installation for channel sync: %w", err)
		}
		if _, err := queries.MarkSlackChannelsInactive(ctx, slacksql.MarkSlackChannelsInactiveParams{WorkspaceID: command.WorkspaceID}); err != nil {
			return fmt.Errorf("mark Slack channels inactive: %w", err)
		}
		for _, channel := range command.Channels {
			if err := queries.UpsertSlackChannel(ctx, slacksql.UpsertSlackChannelParams{
				WorkspaceID: command.WorkspaceID, SlackWorkspaceID: command.InstallationID,
				SlackChannelID: strings.TrimSpace(channel.SlackChannelID), Name: strings.TrimSpace(channel.Name),
				IsPrivate: channel.IsPrivate, IsArchived: channel.IsArchived, IsMember: channel.IsMember,
			}); err != nil {
				return fmt.Errorf("upsert Slack channel: %w", err)
			}
		}
		return nil
	})
}

func (repository *Repo) ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]SlackChannelRecord, error) {
	rows, err := repository.queries.ListChannels(ctx, slacksql.ListChannelsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]SlackChannelRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Channel{
			ID: row.ID, WorkspaceID: row.WorkspaceID, SlackWorkspaceID: row.SlackWorkspaceID,
			SlackChannelID: row.SlackChannelID, Name: row.Name, IsPrivate: row.IsPrivate,
			IsArchived: row.IsArchived, IsMember: row.IsMember, IsActive: row.IsActive,
			IsAssistantConfigured: row.IsAssistantConfigured, LastSyncedAt: row.LastSyncedAt,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return result, nil
}

func (repository *Repo) DeactivateSlackWorkspaceByTeamID(
	ctx context.Context,
	slackTeamID string,
	installGeneration uuid.UUID,
) error {
	slackTeamID = strings.TrimSpace(slackTeamID)
	if slackTeamID == "" || installGeneration == uuid.Nil {
		return errors.Join(slackdomain.ErrInvalidInput, errors.New("slack team and generation are required"))
	}
	return repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		if err := lockInstallationLifecycle(ctx, queries); err != nil {
			return err
		}
		installation, err := queries.LockSlackInstallationByTeamID(ctx, slacksql.LockSlackInstallationByTeamIDParams{
			SlackTeamID: slackTeamID, InstallationGeneration: installGeneration,
		})
		if err != nil {
			return err
		}
		if err := cancelSlackMessaging(ctx, queries, slackTeamID, "Slack installation revoked"); err != nil {
			return err
		}
		if _, err := queries.DeleteSlackUserLinksByWorkspace(ctx, slacksql.DeleteSlackUserLinksByWorkspaceParams{WorkspaceID: installation.WorkspaceID}); err != nil {
			return err
		}
		if affected, err := queries.DeleteSlackInstallationByID(ctx, slacksql.DeleteSlackInstallationByIDParams{SlackWorkspaceID: installation.ID}); err != nil {
			return err
		} else if affected != 1 {
			return slackdomain.ErrConflict
		}
		return nil
	})
}
