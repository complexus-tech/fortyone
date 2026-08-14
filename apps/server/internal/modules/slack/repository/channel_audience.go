package slackrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ChannelTeamAccessRecord is a public team mapping attached to a Slack
// channel. Assistant disclosure uses it only while the channel's assistant
// configuration marker is enabled.
type ChannelTeamAccessRecord struct {
	SlackChannelID string    `db:"slack_channel_id"`
	TeamID         uuid.UUID `db:"team_id"`
}

// AssistantChannelTeamScope separates the teams available to actor-scoped
// tools from the stricter subset available to cross-user tools in a Slack
// channel. Both slices contain only public teams the actor has joined.
type AssistantChannelTeamScope struct {
	AllowedTeamIDs []uuid.UUID
	SharedTeamIDs  []uuid.UUID
}

type assistantChannelTeamScopeRow struct {
	TeamID           uuid.UUID `db:"team_id"`
	ExplicitlyMapped bool      `db:"explicitly_mapped"`
}

const authorizedAssistantChannelTeamScopeQuery = `
	WITH configured_public_teams AS (
		SELECT access.team_id
		FROM slack_channel_team_access access
		JOIN slack_channels configured_channel
		  ON configured_channel.workspace_id = access.workspace_id
		 AND configured_channel.slack_workspace_id = access.slack_workspace_id
		 AND configured_channel.slack_channel_id = access.slack_channel_id
		 AND configured_channel.is_active = true
		 AND configured_channel.is_assistant_configured = true
		JOIN teams mapped_team
		  ON mapped_team.team_id = access.team_id
		 AND mapped_team.workspace_id = access.workspace_id
		 AND mapped_team.is_private = false
		WHERE access.workspace_id = $1
		  AND access.slack_workspace_id = $2
		  AND access.slack_channel_id = $3
	), configuration AS (
		SELECT EXISTS (SELECT 1 FROM configured_public_teams) AS is_configured
	)
	SELECT team.team_id, configuration.is_configured AS explicitly_mapped
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
	  AND team.is_private = false
	  AND (
		(configuration.is_configured AND team.team_id IN (SELECT team_id FROM configured_public_teams))
		OR
		NOT configuration.is_configured
	  )
	ORDER BY team.name ASC
`

const listAssistantChannelTeamAccessQuery = `
	SELECT access.slack_channel_id, access.team_id
	FROM slack_channel_team_access access
	JOIN slack_workspaces installation
	  ON installation.id = access.slack_workspace_id
	 AND installation.workspace_id = access.workspace_id
	 AND installation.is_active = true
	JOIN teams team
	  ON team.team_id = access.team_id
	 AND team.workspace_id = access.workspace_id
	 AND team.is_private = false
	WHERE access.workspace_id = $1
	ORDER BY access.slack_channel_id ASC, team.name ASC
`

const deleteAssistantPublicChannelTeamAccessQuery = `
	DELETE FROM slack_channel_team_access access
	USING teams team
	WHERE access.workspace_id = $1
	  AND access.slack_workspace_id = $2
	  AND access.slack_channel_id = $3
	  AND team.team_id = access.team_id
	  AND team.workspace_id = access.workspace_id
	  AND team.is_private = false
`

const insertAssistantChannelTeamAccessQuery = `
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
	  AND team.is_private = false
`

type assistantChannelTeamAccessTx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

func (r *Repo) ListAssistantChannelTeamAccess(ctx context.Context, workspaceID uuid.UUID) ([]ChannelTeamAccessRecord, error) {
	rows := make([]ChannelTeamAccessRecord, 0)
	if err := r.db.SelectContext(ctx, &rows, listAssistantChannelTeamAccessQuery, workspaceID); err != nil {
		return nil, fmt.Errorf("list Slack assistant channel team access: %w", err)
	}
	return rows, nil
}

// ReplaceAssistantChannelTeamAccess updates the public mappings in the unified
// channel audience used by assistant and non-assistant Slack flows. Explicitly
// unconfiguring the assistant changes only the marker and preserves every
// mapping. A configured channel may intentionally have no public mappings,
// which represents personal-only assistant access.
func (r *Repo) ReplaceAssistantChannelTeamAccess(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	isConfigured bool,
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

	result, err := tx.ExecContext(ctx, `
		UPDATE slack_channels channel_record
		SET is_assistant_configured = $4,
		    updated_at = NOW()
		FROM slack_workspaces installation
		WHERE channel_record.workspace_id = $1
		  AND channel_record.slack_workspace_id = $2
		  AND channel_record.slack_channel_id = $3
		  AND channel_record.is_active = true
		  AND installation.id = channel_record.slack_workspace_id
		  AND installation.workspace_id = channel_record.workspace_id
		  AND installation.is_active = true
	`, workspaceID, slackWorkspaceID, slackChannelID, isConfigured)
	if err != nil {
		return fmt.Errorf("update Slack assistant channel configuration: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect Slack assistant channel configuration update: %w", err)
	}
	if rowsAffected != 1 {
		return sql.ErrNoRows
	}

	if err = replaceAssistantPublicChannelTeamAccessTx(
		ctx,
		tx,
		workspaceID,
		slackWorkspaceID,
		slackChannelID,
		isConfigured,
		teamIDs,
		actorID,
	); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit Slack channel audience transaction: %w", err)
	}
	return nil
}

func replaceAssistantPublicChannelTeamAccessTx(
	ctx context.Context,
	tx assistantChannelTeamAccessTx,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	isConfigured bool,
	teamIDs []uuid.UUID,
	actorID uuid.UUID,
) error {
	if !isConfigured {
		return nil
	}

	if _, err := tx.ExecContext(
		ctx,
		deleteAssistantPublicChannelTeamAccessQuery,
		workspaceID,
		slackWorkspaceID,
		slackChannelID,
	); err != nil {
		return fmt.Errorf("clear public Slack assistant channel team access: %w", err)
	}

	for _, teamID := range teamIDs {
		var isPrivate bool
		if err := tx.GetContext(ctx, &isPrivate, `
			SELECT team.is_private
			FROM teams team
			WHERE team.workspace_id = $1
			  AND team.team_id = $2
		`, workspaceID, teamID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("team %s is not in workspace %s", teamID, workspaceID)
			}
			return fmt.Errorf("validate Slack assistant channel team: %w", err)
		}
		if isPrivate {
			continue
		}

		result, insertErr := tx.ExecContext(
			ctx,
			insertAssistantChannelTeamAccessQuery,
			workspaceID,
			slackWorkspaceID,
			slackChannelID,
			teamID,
			actorID,
		)
		if insertErr != nil {
			return fmt.Errorf("insert Slack assistant channel team access: %w", insertErr)
		}
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return fmt.Errorf("inspect Slack assistant channel team access insert: %w", rowsErr)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("team %s is not in workspace %s", teamID, workspaceID)
		}
	}
	return nil
}

// GetAuthorizedAssistantChannelTeamScope resolves the safe v1 disclosure
// boundary for an assistant turn in a Slack channel. Actor-scoped tools retain
// access to the actor's joined public teams when no mapping exists. Cross-user
// tools receive teams only when an administrator explicitly mapped the channel.
// Private FortyOne teams are excluded from both scopes for channel delivery;
// a legacy private-only mapping is therefore treated as no assistant mapping.
func (r *Repo) GetAuthorizedAssistantChannelTeamScope(
	ctx context.Context,
	workspaceID, slackWorkspaceID uuid.UUID,
	slackChannelID string,
	userID uuid.UUID,
) (AssistantChannelTeamScope, error) {
	rows := make([]assistantChannelTeamScopeRow, 0)
	if err := r.db.SelectContext(
		ctx,
		&rows,
		authorizedAssistantChannelTeamScopeQuery,
		workspaceID,
		slackWorkspaceID,
		slackChannelID,
		userID,
	); err != nil {
		return AssistantChannelTeamScope{}, fmt.Errorf("get authorized Slack assistant channel team scope: %w", err)
	}

	return assistantChannelTeamScope(rows), nil
}

func assistantChannelTeamScope(rows []assistantChannelTeamScopeRow) AssistantChannelTeamScope {
	scope := AssistantChannelTeamScope{
		AllowedTeamIDs: make([]uuid.UUID, 0, len(rows)),
		SharedTeamIDs:  make([]uuid.UUID, 0, len(rows)),
	}
	for _, row := range rows {
		if row.TeamID == uuid.Nil {
			continue
		}
		scope.AllowedTeamIDs = append(scope.AllowedTeamIDs, row.TeamID)
		if row.ExplicitlyMapped {
			scope.SharedTeamIDs = append(scope.SharedTeamIDs, row.TeamID)
		}
	}
	return scope
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
