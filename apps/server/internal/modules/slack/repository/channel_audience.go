package slackrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ChannelTeamAccessRecord is an administrator-defined Slack channel audience.
// An empty mapping means the channel is limited to the actor's public teams.
type ChannelTeamAccessRecord struct {
	SlackChannelID string    `db:"slack_channel_id"`
	TeamID         uuid.UUID `db:"team_id"`
}

func (r *Repo) ListChannelTeamAccess(ctx context.Context, workspaceID uuid.UUID) ([]ChannelTeamAccessRecord, error) {
	rows := make([]ChannelTeamAccessRecord, 0)
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT access.slack_channel_id, access.team_id
		FROM slack_channel_team_access access
		JOIN slack_workspaces installation
		  ON installation.id = access.slack_workspace_id
		 AND installation.workspace_id = access.workspace_id
		 AND installation.is_active = true
		JOIN teams team
		  ON team.team_id = access.team_id
		 AND team.workspace_id = access.workspace_id
		WHERE access.workspace_id = $1
		ORDER BY access.slack_channel_id ASC, team.name ASC
	`, workspaceID); err != nil {
		return nil, fmt.Errorf("list Slack channel team access: %w", err)
	}
	return rows, nil
}

func (r *Repo) ReplaceChannelTeamAccess(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	teamIDs []uuid.UUID,
	actorID uuid.UUID,
) (err error) {
	if workspaceID == uuid.Nil || slackWorkspaceID == uuid.Nil || actorID == uuid.Nil || slackChannelID == "" {
		return errors.New("workspace, installation, channel, and actor are required")
	}
	teamIDs = uniqueUUIDs(teamIDs)
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin Slack channel audience transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var channelExists bool
	if err = tx.GetContext(ctx, &channelExists, `
		SELECT EXISTS (
			SELECT 1
			FROM slack_channels channel_record
			JOIN slack_workspaces installation
			  ON installation.id = channel_record.slack_workspace_id
			 AND installation.workspace_id = channel_record.workspace_id
			 AND installation.is_active = true
			WHERE channel_record.workspace_id = $1
			  AND channel_record.slack_workspace_id = $2
			  AND channel_record.slack_channel_id = $3
			  AND channel_record.is_active = true
		)
	`, workspaceID, slackWorkspaceID, slackChannelID); err != nil {
		return fmt.Errorf("validate Slack channel audience target: %w", err)
	}
	if !channelExists {
		return sql.ErrNoRows
	}

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM slack_channel_team_access
		WHERE workspace_id = $1
		  AND slack_workspace_id = $2
		  AND slack_channel_id = $3
	`, workspaceID, slackWorkspaceID, slackChannelID); err != nil {
		return fmt.Errorf("clear Slack channel team access: %w", err)
	}

	for _, teamID := range teamIDs {
		result, insertErr := tx.ExecContext(ctx, `
			INSERT INTO slack_channel_team_access (
				workspace_id,
				slack_workspace_id,
				slack_channel_id,
				team_id,
				created_by_user_id
			)
			SELECT $1, $2, $3, team.team_id, $5
			FROM teams team
			WHERE team.workspace_id = $1
			  AND team.team_id = $4
		`, workspaceID, slackWorkspaceID, slackChannelID, teamID, actorID)
		if insertErr != nil {
			return fmt.Errorf("insert Slack channel team access: %w", insertErr)
		}
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return fmt.Errorf("inspect Slack channel team access insert: %w", rowsErr)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("team %s is not in workspace %s", teamID, workspaceID)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit Slack channel audience transaction: %w", err)
	}
	return nil
}

// ListAuthorizedChannelTeamIDs resolves the authoritative team boundary for a
// linked actor in a Slack channel. Explicit mappings are restrictive. Without
// a mapping, only public teams that the actor has joined are exposed.
func (r *Repo) ListAuthorizedChannelTeamIDs(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	teamIDs := make([]uuid.UUID, 0)
	if err := r.db.SelectContext(ctx, &teamIDs, `
		WITH configured_teams AS (
			SELECT access.team_id
			FROM slack_channel_team_access access
			WHERE access.workspace_id = $1
			  AND access.slack_workspace_id = $2
			  AND access.slack_channel_id = $3
		), configuration AS (
			SELECT EXISTS (SELECT 1 FROM configured_teams) AS is_configured
		)
		SELECT team.team_id
		FROM teams team
		JOIN team_members membership
		  ON membership.team_id = team.team_id
		 AND membership.user_id = $4
		JOIN workspace_members workspace_membership
		  ON workspace_membership.workspace_id = team.workspace_id
		 AND workspace_membership.user_id = membership.user_id
		JOIN users actor
		  ON actor.user_id = membership.user_id
		 AND actor.is_active = true
		CROSS JOIN configuration
		WHERE team.workspace_id = $1
		  AND (
			(configuration.is_configured AND team.team_id IN (SELECT team_id FROM configured_teams))
			OR
			(NOT configuration.is_configured AND team.is_private = false)
		  )
		ORDER BY team.name ASC
	`, workspaceID, slackWorkspaceID, slackChannelID, userID); err != nil {
		return nil, fmt.Errorf("list authorized Slack channel teams: %w", err)
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
