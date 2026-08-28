package slackrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChannelTeamAccessRecord = slackdomain.ChannelTeamAccess
type AssistantChannelTeamScope = slackdomain.AssistantChannelTeamScope

func (repository *Repo) ListAssistantChannelTeamAccessForAdmin(
	ctx context.Context,
	query slackdomain.WorkspaceActorQuery,
) ([]ChannelTeamAccessRecord, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListAssistantChannelTeamAccessForAdmin(ctx, slacksql.ListAssistantChannelTeamAccessForAdminParams{
		WorkspaceID: query.WorkspaceID,
		ActorID:     query.ActorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list Slack assistant channel team access for admin: %w", mapDatabaseError(err))
	}
	if len(rows) == 0 {
		if err := repository.requireCurrentWorkspaceRole(ctx, query, authorization.WorkspaceRoleAdmin); err != nil {
			return nil, err
		}
	}
	result := make([]ChannelTeamAccessRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.ChannelTeamAccess{SlackChannelID: row.SlackChannelID, TeamID: row.TeamID})
	}
	return result, nil
}

func (repository *Repo) ReplaceAssistantChannelTeamAccess(
	ctx context.Context,
	command slackdomain.ReplaceChannelAudienceCommand,
) error {
	command.SlackChannelID = strings.TrimSpace(command.SlackChannelID)
	command.TeamIDs = uniqueUUIDs(command.TeamIDs)
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
			return fmt.Errorf("authorize Slack channel audience mutation: %w", err)
		}
		affected, err := queries.UpdateSlackAssistantChannelConfiguration(ctx, slacksql.UpdateSlackAssistantChannelConfigurationParams{
			WorkspaceID: command.WorkspaceID, SlackWorkspaceID: command.InstallationID,
			SlackChannelID: command.SlackChannelID, IsConfigured: command.Configured,
			InstallationGeneration: command.InstallationGeneration,
		})
		if err != nil {
			return fmt.Errorf("update Slack assistant channel configuration: %w", err)
		}
		if affected != 1 {
			return slackdomain.ErrConflict
		}
		if !command.Configured {
			return nil
		}
		if _, err := queries.DeleteAssistantPublicChannelTeamAccess(ctx, slacksql.DeleteAssistantPublicChannelTeamAccessParams{
			WorkspaceID: command.WorkspaceID, SlackWorkspaceID: command.InstallationID,
			SlackChannelID: command.SlackChannelID,
		}); err != nil {
			return fmt.Errorf("clear public Slack assistant channel team access: %w", err)
		}
		for _, teamID := range command.TeamIDs {
			affected, err := queries.InsertAssistantChannelTeamAccess(ctx, slacksql.InsertAssistantChannelTeamAccessParams{
				WorkspaceID: command.WorkspaceID, SlackWorkspaceID: command.InstallationID,
				SlackChannelID: command.SlackChannelID, TeamID: teamID, ActorID: command.ActorID,
			})
			if err != nil {
				return fmt.Errorf("insert Slack assistant channel team access: %w", err)
			}
			if affected != 1 {
				return errors.Join(slackdomain.ErrInvalidInput, fmt.Errorf("team %s is not a public team in this workspace", teamID))
			}
		}
		return nil
	})
}

func (repository *Repo) GetAuthorizedAssistantChannelTeamScope(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	userID uuid.UUID,
) (AssistantChannelTeamScope, error) {
	rows, err := repository.queries.GetAuthorizedAssistantChannelTeamScope(ctx, slacksql.GetAuthorizedAssistantChannelTeamScopeParams{
		WorkspaceID: workspaceID, SlackWorkspaceID: slackWorkspaceID,
		SlackChannelID: strings.TrimSpace(slackChannelID), UserID: userID,
	})
	if err != nil {
		return AssistantChannelTeamScope{}, fmt.Errorf("get authorized Slack assistant channel team scope: %w", mapDatabaseError(err))
	}
	scope := AssistantChannelTeamScope{AllowedTeamIDs: make([]uuid.UUID, 0, len(rows)), SharedTeamIDs: make([]uuid.UUID, 0, len(rows))}
	for _, row := range rows {
		if row.TeamID == uuid.Nil {
			continue
		}
		scope.AllowedTeamIDs = append(scope.AllowedTeamIDs, row.TeamID)
		if row.ExplicitlyMapped {
			scope.SharedTeamIDs = append(scope.SharedTeamIDs, row.TeamID)
		}
	}
	return scope, nil
}

func (repository *Repo) ListAuthorizedChannelTeamIDs(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	teamIDs, err := repository.queries.ListAuthorizedChannelTeamIDs(ctx, slacksql.ListAuthorizedChannelTeamIDsParams{
		WorkspaceID: workspaceID, SlackWorkspaceID: slackWorkspaceID,
		SlackChannelID: strings.TrimSpace(slackChannelID), UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list authorized Slack channel teams: %w", mapDatabaseError(err))
	}
	return teamIDs, nil
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
