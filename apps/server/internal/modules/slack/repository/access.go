package slackrepository

import (
	"context"
	"errors"
	"fmt"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repo) GetWorkspaceRole(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
) (authorization.WorkspaceRole, error) {
	if workspaceID == uuid.Nil || actorID == uuid.Nil || repository == nil || repository.queries == nil {
		return "", slackdomain.ErrForbidden
	}
	role, err := repository.queries.GetWorkspaceRole(ctx, slacksql.GetWorkspaceRoleParams{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", slackdomain.ErrForbidden
	}
	if err != nil {
		return "", fmt.Errorf("get Slack workspace actor role: %w", mapDatabaseError(err))
	}
	result := authorization.WorkspaceRole(role)
	if err := authorization.ValidateWorkspaceRole(result); err != nil {
		return "", errors.Join(slackdomain.ErrForbidden, err)
	}
	return result, nil
}

func (repository *Repo) GetSlackWorkspaceForMember(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
) (SlackWorkspaceRecord, error) {
	if err := query.Validate(); err != nil {
		return SlackWorkspaceRecord{}, err
	}
	row, err := repository.queries.GetSlackWorkspaceForMember(ctx, slacksql.GetSlackWorkspaceForMemberParams{
		WorkspaceID: query.WorkspaceID,
		ActorID:     query.ActorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		if authErr := repository.requireCurrentWorkspaceRole(ctx, query, authorization.WorkspaceRoleMember); authErr != nil {
			return SlackWorkspaceRecord{}, authErr
		}
		return SlackWorkspaceRecord{}, slackdomain.ErrNotFound
	}
	if err != nil {
		return SlackWorkspaceRecord{}, fmt.Errorf("get Slack installation for member: %w", mapDatabaseError(err))
	}
	return mapInstallation(installationFields{
		id: row.ID, workspaceID: row.WorkspaceID, slackTeamID: row.SlackTeamID,
		slackTeamName: row.SlackTeamName, slackTeamDomain: row.SlackTeamDomain,
		botUserID: row.BotUserID, credentialPayload: row.CredentialPayload,
		credentialKeyVersion:     row.CredentialKeyVersion,
		installationGeneration:   row.InstallationGeneration,
		installationAuthorizedAt: row.InstallationAuthorizedAt,
		slackAppID:               row.SlackAppID, enterpriseID: row.EnterpriseID,
		authedUserID: row.AuthedUserID, scope: row.Scope, isActive: row.IsActive,
		installedByUserID: row.InstalledByUserID, revokedAt: row.RevokedAt,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	}), nil
}

func (repository *Repo) FindSlackUserLinkForMember(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
	slackTeamID string,
) (*SlackUserLinkRecord, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	row, err := repository.queries.FindSlackUserLinkForMember(ctx, slacksql.FindSlackUserLinkForMemberParams{
		WorkspaceID: query.WorkspaceID,
		ActorID:     query.ActorID,
		SlackTeamID: slackTeamID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		if authErr := repository.requireCurrentWorkspaceRole(ctx, query, authorization.WorkspaceRoleMember); authErr != nil {
			return nil, authErr
		}
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find Slack user link for member: %w", mapDatabaseError(err))
	}
	return &slackdomain.UserLink{
		SlackUserID: row.SlackUserID,
		UserID:      row.UserID,
		LinkedVia:   row.LinkedVia,
		LinkedAt:    row.LinkedAt,
	}, nil
}

func (repository *Repo) ListChannelsForMember(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
) ([]SlackChannelRecord, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListChannelsForMember(ctx, slacksql.ListChannelsForMemberParams{
		WorkspaceID: query.WorkspaceID,
		ActorID:     query.ActorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list Slack channels for member: %w", mapDatabaseError(err))
	}
	if len(rows) == 0 {
		if err := repository.requireCurrentWorkspaceRole(ctx, query, authorization.WorkspaceRoleMember); err != nil {
			return nil, err
		}
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

func (repository *Repo) ListRequestLogsForAdmin(
	ctx context.Context,
	query slackdomain.ListRequestLogsQuery,
) ([]SlackRequestLogRecord, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListRequestLogsForAdmin(ctx, slacksql.ListRequestLogsForAdminParams{
		WorkspaceID: query.WorkspaceID,
		ActorID:     query.ActorID,
		ResultLimit: query.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list Slack request logs for admin: %w", mapDatabaseError(err))
	}
	if len(rows) == 0 {
		if err := repository.requireCurrentWorkspaceRole(
			ctx,
			slackdomain.WorkspaceActorQuery{WorkspaceID: query.WorkspaceID, ActorID: query.ActorID},
			authorization.WorkspaceRoleAdmin,
		); err != nil {
			return nil, err
		}
	}
	result := make([]SlackRequestLogRecord, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapRequestLog(row)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}

func (repository *Repo) requireCurrentWorkspaceRole(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
	minimum authorization.WorkspaceRole,
) error {
	role, err := repository.GetWorkspaceRole(ctx, query.WorkspaceID, query.ActorID)
	if err != nil {
		return err
	}
	if err := authorization.RequireMinimumWorkspaceRole(role, minimum); err != nil {
		return errors.Join(slackdomain.ErrForbidden, err)
	}
	return nil
}
