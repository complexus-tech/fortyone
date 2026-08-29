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

func (repository *Repo) CreateStoryLink(ctx context.Context, storyID uuid.UUID, sourceKey, title, linkURL string) error {
	if err := repository.queries.CreateStoryLink(ctx, slacksql.CreateStoryLinkParams{
		StoryID: storyID, SourceKey: strings.TrimSpace(sourceKey),
		Title: strings.TrimSpace(title), URL: strings.TrimSpace(linkURL),
	}); err != nil {
		return fmt.Errorf("upsert Slack story source link: %w", mapDatabaseError(err))
	}
	return nil
}

func (repository *Repo) ListWorkspaceMembersForSlackLinking(ctx context.Context, workspaceID uuid.UUID) ([]WorkspaceMemberRecord, error) {
	rows, err := repository.queries.ListWorkspaceMembersForSlackLinking(ctx, slacksql.ListWorkspaceMembersForSlackLinkingParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]WorkspaceMemberRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.WorkspaceMember{UserID: row.UserID, Email: row.Email})
	}
	return result, nil
}

func (repository *Repo) UpsertSlackUserLinks(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackTeamID string,
	links []SlackUserLinkUpsert,
) error {
	if len(links) == 0 {
		return nil
	}
	return repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		for _, link := range links {
			slackUserID := strings.TrimSpace(link.SlackUserID)
			if slackUserID == "" || link.UserID == uuid.Nil {
				continue
			}
			linkedVia := strings.TrimSpace(link.LinkedVia)
			if linkedVia == "" {
				linkedVia = "email_match"
			}
			affected, err := queries.UpsertSlackUserLink(ctx, slacksql.UpsertSlackUserLinkParams{
				WorkspaceID: workspaceID, SlackWorkspaceID: slackWorkspaceID,
				SlackTeamID: strings.TrimSpace(slackTeamID), SlackUserID: slackUserID,
				UserID: link.UserID, LinkedVia: linkedVia,
			})
			if err != nil {
				return fmt.Errorf("upsert Slack user link: %w", err)
			}
			if affected != 1 {
				return slackdomain.ErrConflict
			}
		}
		return nil
	})
}

func (repository *Repo) FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error) {
	id, err := repository.queries.FindLinkedUserIDBySlackUser(ctx, slacksql.FindLinkedUserIDBySlackUserParams{
		WorkspaceID: workspaceID, SlackTeamID: strings.TrimSpace(slackTeamID),
		SlackUserID: strings.TrimSpace(slackUserID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return &id, nil
}

func (repository *Repo) FindSlackUserLinkByUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, userID uuid.UUID) (*SlackUserLinkRecord, error) {
	row, err := repository.queries.FindSlackUserLinkByUser(ctx, slacksql.FindSlackUserLinkByUserParams{
		WorkspaceID: workspaceID, SlackTeamID: strings.TrimSpace(slackTeamID), UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return &slackdomain.UserLink{SlackUserID: row.SlackUserID, UserID: row.UserID, LinkedVia: row.LinkedVia, LinkedAt: row.LinkedAt}, nil
}

func (repository *Repo) DeleteSlackUserLink(
	ctx context.Context,
	workspaceID uuid.UUID,
	slackTeamID, slackUserID string,
	userID uuid.UUID,
) (bool, error) {
	affected, err := repository.queries.DeleteSlackUserLink(ctx, slacksql.DeleteSlackUserLinkParams{
		WorkspaceID: workspaceID, SlackTeamID: strings.TrimSpace(slackTeamID),
		SlackUserID: strings.TrimSpace(slackUserID), UserID: userID,
	})
	if err != nil {
		return false, mapDatabaseError(err)
	}
	return affected == 1, nil
}
