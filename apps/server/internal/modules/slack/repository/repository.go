package slackrepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{log: log, db: db}
}

type WorkspaceRecord struct {
	ID   uuid.UUID `db:"workspace_id"`
	Slug string    `db:"slug"`
	Name string    `db:"name"`
}

type TeamRecord struct {
	ID    uuid.UUID `db:"team_id"`
	Code  string    `db:"code"`
	Name  string    `db:"name"`
	Color string    `db:"color"`
}

type StatusRecord struct {
	ID       uuid.UUID `db:"status_id"`
	Name     string    `db:"name"`
	Category string    `db:"category"`
}

type TeamMemberRecord struct {
	UserID   uuid.UUID `db:"user_id"`
	Username string    `db:"username"`
	FullName string    `db:"full_name"`
	Email    string    `db:"email"`
}

type LabelRecord struct {
	ID   uuid.UUID `db:"label_id"`
	Name string    `db:"name"`
}

type ObjectiveRecord struct {
	ID   uuid.UUID `db:"objective_id"`
	Name string    `db:"name"`
}

type WorkspaceMemberRecord struct {
	UserID uuid.UUID `db:"user_id"`
	Email  string    `db:"email"`
}

type SlackUserLinkUpsert struct {
	SlackUserID string
	UserID      uuid.UUID
	LinkedVia   string
}

type SlackWorkspaceRecord struct {
	ID                uuid.UUID  `db:"id"`
	WorkspaceID       uuid.UUID  `db:"workspace_id"`
	SlackTeamID       string     `db:"slack_team_id"`
	SlackTeamName     string     `db:"slack_team_name"`
	SlackTeamDomain   string     `db:"slack_team_domain"`
	BotUserID         *string    `db:"bot_user_id"`
	BotAccessToken    string     `db:"bot_access_token"`
	Scope             *string    `db:"scope"`
	IsActive          bool       `db:"is_active"`
	InstalledByUserID *uuid.UUID `db:"installed_by_user_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type SlackChannelRecord struct {
	ID               uuid.UUID  `db:"id"`
	WorkspaceID      uuid.UUID  `db:"workspace_id"`
	SlackWorkspaceID uuid.UUID  `db:"slack_workspace_id"`
	SlackChannelID   string     `db:"slack_channel_id"`
	Name             string     `db:"name"`
	IsPrivate        bool       `db:"is_private"`
	IsArchived       bool       `db:"is_archived"`
	IsMember         bool       `db:"is_member"`
	IsActive         bool       `db:"is_active"`
	LastSyncedAt     *time.Time `db:"last_synced_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

type OAuthInstallPayload struct {
	SlackTeamID     string
	SlackTeamName   string
	SlackTeamDomain string
	BotUserID       *string
	BotAccessToken  string
	Scope           *string
}

type SlackChannelPayload struct {
	SlackChannelID string
	Name           string
	IsPrivate      bool
	IsArchived     bool
	IsMember       bool
}

type SlackRequestLogInsert struct {
	RequestType  string
	Endpoint     string
	WorkspaceID  *uuid.UUID
	SlackTeamID  *string
	SlackUserID  *string
	SlackChannel *string
	Command      *string
	TriggerID    *string
	RequestBody  *string
	Headers      []byte
	ResponseCode int
	Outcome      string
	ErrorMessage *string
}

type SlackRequestLogRecord struct {
	ID           uuid.UUID  `db:"id"`
	RequestType  string     `db:"request_type"`
	Endpoint     string     `db:"endpoint"`
	WorkspaceID  *uuid.UUID `db:"workspace_id"`
	SlackTeamID  *string    `db:"slack_team_id"`
	SlackUserID  *string    `db:"slack_user_id"`
	SlackChannel *string    `db:"slack_channel_id"`
	Command      *string    `db:"command"`
	TriggerID    *string    `db:"trigger_id"`
	RequestBody  *string    `db:"request_body"`
	Headers      []byte     `db:"headers"`
	ResponseCode int        `db:"response_code"`
	Outcome      string     `db:"outcome"`
	ErrorMessage *string    `db:"error_message"`
	CreatedAt    time.Time  `db:"created_at"`
}

func (r *Repo) FindWorkspaceBySlug(ctx context.Context, slug string) (WorkspaceRecord, error) {
	var row WorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT workspace_id, slug, name
		FROM workspaces
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug)
	if err != nil {
		return WorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (WorkspaceRecord, error) {
	var row WorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT workspace_id, slug, name
		FROM workspaces
		WHERE workspace_id = $1 AND deleted_at IS NULL
	`, workspaceID)
	if err != nil {
		return WorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (TeamRecord, error) {
	var row TeamRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT team_id, code, name, color
		FROM teams
		WHERE workspace_id = $1 AND LOWER(code) = LOWER($2)
		LIMIT 1
	`, workspaceID, code)
	if err != nil {
		return TeamRecord{}, err
	}
	return row, nil
}

func (r *Repo) FindTeamByID(ctx context.Context, workspaceID, teamID uuid.UUID) (TeamRecord, error) {
	var row TeamRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT team_id, code, name, color
		FROM teams
		WHERE workspace_id = $1 AND team_id = $2
		LIMIT 1
	`, workspaceID, teamID)
	if err != nil {
		return TeamRecord{}, err
	}
	return row, nil
}

func (r *Repo) GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (WorkspaceRecord, error) {
	var row WorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT w.workspace_id, w.slug, w.name
		FROM slack_workspaces sw
		JOIN workspaces w ON w.workspace_id = sw.workspace_id
		WHERE sw.slack_team_id = $1 AND sw.is_active = true
		LIMIT 1
	`, slackTeamID)
	if err != nil {
		return WorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) UpsertSlackWorkspace(ctx context.Context, workspaceID, installedByUserID uuid.UUID, payload OAuthInstallPayload) (SlackWorkspaceRecord, error) {
	var row SlackWorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO slack_workspaces (
			workspace_id, slack_team_id, slack_team_name, slack_team_domain,
			bot_user_id, bot_access_token, scope, is_active, installed_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8)
		ON CONFLICT (workspace_id) DO UPDATE SET
			slack_team_id = EXCLUDED.slack_team_id,
			slack_team_name = EXCLUDED.slack_team_name,
			slack_team_domain = EXCLUDED.slack_team_domain,
			bot_user_id = EXCLUDED.bot_user_id,
			bot_access_token = EXCLUDED.bot_access_token,
			scope = EXCLUDED.scope,
			is_active = true,
			installed_by_user_id = EXCLUDED.installed_by_user_id,
			updated_at = NOW()
		RETURNING id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
		          bot_user_id, bot_access_token, scope, is_active, installed_by_user_id, created_at, updated_at
	`,
		workspaceID,
		payload.SlackTeamID,
		payload.SlackTeamName,
		payload.SlackTeamDomain,
		payload.BotUserID,
		payload.BotAccessToken,
		payload.Scope,
		installedByUserID,
	)
	if err != nil {
		return SlackWorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (SlackWorkspaceRecord, error) {
	var row SlackWorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
		       bot_user_id, bot_access_token, scope, is_active, installed_by_user_id, created_at, updated_at
		FROM slack_workspaces
		WHERE workspace_id = $1
		LIMIT 1
	`, workspaceID)
	if err != nil {
		return SlackWorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (SlackWorkspaceRecord, error) {
	var row SlackWorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
		       bot_user_id, bot_access_token, scope, is_active, installed_by_user_id, created_at, updated_at
		FROM slack_workspaces
		WHERE slack_team_id = $1 AND is_active = true
		LIMIT 1
	`, slackTeamID)
	if err != nil {
		return SlackWorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) DisconnectSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		DELETE FROM slack_workspaces
		WHERE workspace_id = $1
	`, workspaceID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (r *Repo) UpsertChannels(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, channels []SlackChannelPayload) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, channel := range channels {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO slack_channels (
				workspace_id, slack_workspace_id, slack_channel_id, name,
				is_private, is_archived, is_member, is_active, last_synced_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW())
			ON CONFLICT (workspace_id, slack_channel_id) DO UPDATE SET
				slack_workspace_id = EXCLUDED.slack_workspace_id,
				name = EXCLUDED.name,
				is_private = EXCLUDED.is_private,
				is_archived = EXCLUDED.is_archived,
				is_member = EXCLUDED.is_member,
				is_active = true,
				last_synced_at = NOW(),
				updated_at = NOW()
		`,
			workspaceID,
			slackWorkspaceID,
			channel.SlackChannelID,
			channel.Name,
			channel.IsPrivate,
			channel.IsArchived,
			channel.IsMember,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]SlackChannelRecord, error) {
	rows := make([]SlackChannelRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id, workspace_id, slack_workspace_id, slack_channel_id, name,
		       is_private, is_archived, is_member, is_active, last_synced_at, created_at, updated_at
		FROM slack_channels
		WHERE workspace_id = $1 AND is_active = true
		ORDER BY name ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]TeamRecord, error) {
	rows := make([]TeamRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT team_id, code, name, color
		FROM teams
		WHERE workspace_id = $1
		ORDER BY name ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]TeamRecord, error) {
	rows := make([]TeamRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT t.team_id, t.code, t.name, t.color
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.team_id
		WHERE t.workspace_id = $1
		  AND tm.user_id = $2
		ORDER BY t.name ASC
	`, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]StatusRecord, error) {
	rows := make([]StatusRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT status_id, name, category
		FROM statuses
		WHERE team_id = $1
		ORDER BY order_index ASC, name ASC
	`, teamID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]TeamMemberRecord, error) {
	rows := make([]TeamMemberRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT u.user_id, u.username, COALESCE(u.full_name, '') AS full_name, u.email
		FROM team_members tm
		JOIN users u ON u.user_id = tm.user_id
		WHERE tm.team_id = $1
		  AND u.is_active = true
		ORDER BY COALESCE(NULLIF(TRIM(u.full_name), ''), NULLIF(TRIM(u.username), ''), u.email) ASC
	`, teamID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) ListTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID) ([]LabelRecord, error) {
	rows := make([]LabelRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT label_id, name
		FROM labels
		WHERE workspace_id = $1
		  AND (team_id = $2 OR team_id IS NULL)
		ORDER BY name ASC
	`, workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (TeamMemberRecord, error) {
	var row TeamMemberRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT u.user_id, u.username, COALESCE(u.full_name, '') AS full_name, u.email
		FROM team_members tm
		JOIN users u ON u.user_id = tm.user_id
		WHERE tm.team_id = $1
		  AND tm.user_id = $2
		  AND u.is_active = true
		LIMIT 1
	`, teamID, userID)
	if err != nil {
		return TeamMemberRecord{}, err
	}
	return row, nil
}

func (r *Repo) FindTeamLabelByID(ctx context.Context, workspaceID, teamID, labelID uuid.UUID) (LabelRecord, error) {
	var row LabelRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT label_id, name
		FROM labels
		WHERE workspace_id = $1
		  AND label_id = $2
		  AND (team_id = $3 OR team_id IS NULL)
		LIMIT 1
	`, workspaceID, labelID, teamID)
	if err != nil {
		return LabelRecord{}, err
	}
	return row, nil
}

func (r *Repo) FindTeamObjectiveByID(ctx context.Context, workspaceID, teamID, objectiveID uuid.UUID) (ObjectiveRecord, error) {
	var row ObjectiveRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT objective_id, name
		FROM objectives
		WHERE workspace_id = $1
		  AND team_id = $2
		  AND objective_id = $3
		LIMIT 1
	`, workspaceID, teamID, objectiveID)
	if err != nil {
		return ObjectiveRecord{}, err
	}
	return row, nil
}

func (r *Repo) SearchTeamMembers(ctx context.Context, teamID uuid.UUID, query string, limit int) ([]TeamMemberRecord, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	rows := make([]TeamMemberRecord, 0)
	searchQuery := "%" + query + "%"
	err := r.db.SelectContext(ctx, &rows, `
		SELECT u.user_id, u.username, COALESCE(u.full_name, '') AS full_name, u.email
		FROM team_members tm
		JOIN users u ON u.user_id = tm.user_id
		WHERE tm.team_id = $1
		  AND u.is_active = true
		  AND (
			LOWER(COALESCE(u.full_name, '')) LIKE LOWER($2)
			OR LOWER(COALESCE(u.username, '')) LIKE LOWER($2)
			OR LOWER(COALESCE(u.email, '')) LIKE LOWER($2)
		  )
		ORDER BY COALESCE(NULLIF(TRIM(u.full_name), ''), NULLIF(TRIM(u.username), ''), u.email) ASC
		LIMIT $3
	`, teamID, searchQuery, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) SearchTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]LabelRecord, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	rows := make([]LabelRecord, 0)
	searchQuery := "%" + query + "%"
	err := r.db.SelectContext(ctx, &rows, `
		SELECT label_id, name
		FROM labels
		WHERE workspace_id = $1
		  AND (team_id = $2 OR team_id IS NULL)
		  AND LOWER(name) LIKE LOWER($3)
		ORDER BY name ASC
		LIMIT $4
	`, workspaceID, teamID, searchQuery, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) SearchTeamObjectives(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]ObjectiveRecord, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	rows := make([]ObjectiveRecord, 0)
	searchQuery := "%" + query + "%"
	err := r.db.SelectContext(ctx, &rows, `
		SELECT objective_id, name
		FROM objectives
		WHERE workspace_id = $1
		  AND team_id = $2
		  AND LOWER(name) LIKE LOWER($3)
		ORDER BY name ASC
		LIMIT $4
	`, workspaceID, teamID, searchQuery, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) CreateStoryLink(ctx context.Context, storyID uuid.UUID, title, linkURL string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO story_links (title, url, story_id)
		VALUES ($1, $2, $3)
	`, title, linkURL, storyID)
	return err
}

func (r *Repo) ListWorkspaceMembersForSlackLinking(ctx context.Context, workspaceID uuid.UUID) ([]WorkspaceMemberRecord, error) {
	rows := make([]WorkspaceMemberRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT u.user_id, u.email
		FROM workspace_members wm
		JOIN users u ON u.user_id = wm.user_id
		WHERE wm.workspace_id = $1
		  AND u.is_active = true
		  AND u.is_system = false
		  AND TRIM(COALESCE(u.email, '')) <> ''
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) UpsertSlackUserLinks(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, slackTeamID string, links []SlackUserLinkUpsert) error {
	if len(links) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, link := range links {
		slackUserID := strings.TrimSpace(link.SlackUserID)
		if slackUserID == "" || link.UserID == uuid.Nil {
			continue
		}
		linkedVia := strings.TrimSpace(link.LinkedVia)
		if linkedVia == "" {
			linkedVia = "email_match"
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO slack_user_links (
				workspace_id,
				slack_workspace_id,
				slack_team_id,
				slack_user_id,
				user_id,
				linked_via,
				linked_at
			) VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (workspace_id, slack_team_id, slack_user_id) DO UPDATE SET
				slack_workspace_id = EXCLUDED.slack_workspace_id,
				user_id = EXCLUDED.user_id,
				linked_via = EXCLUDED.linked_via,
				linked_at = NOW(),
				updated_at = NOW()
		`, workspaceID, slackWorkspaceID, slackTeamID, slackUserID, link.UserID, linkedVia)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repo) FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.GetContext(ctx, &userID, `
		SELECT user_id
		FROM slack_user_links
		WHERE workspace_id = $1
		  AND slack_team_id = $2
		  AND slack_user_id = $3
		LIMIT 1
	`, workspaceID, strings.TrimSpace(slackTeamID), strings.TrimSpace(slackUserID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &userID, nil
}

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	var statusID uuid.UUID
	err := r.db.GetContext(ctx, &statusID, `
		SELECT status_id
		FROM statuses
		WHERE team_id = $1
		  AND category = $2
		ORDER BY order_index ASC
		LIMIT 1
	`, teamID, category)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &statusID, nil
}

func (r *Repo) InsertRequestLog(ctx context.Context, entry SlackRequestLogInsert) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO slack_request_logs (
			request_type,
			endpoint,
			workspace_id,
			slack_team_id,
			slack_user_id,
			slack_channel_id,
			command,
			trigger_id,
			request_body,
			headers,
			response_code,
			outcome,
			error_message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13)
	`,
		entry.RequestType,
		entry.Endpoint,
		entry.WorkspaceID,
		entry.SlackTeamID,
		entry.SlackUserID,
		entry.SlackChannel,
		entry.Command,
		entry.TriggerID,
		entry.RequestBody,
		string(entry.Headers),
		entry.ResponseCode,
		entry.Outcome,
		entry.ErrorMessage,
	)
	return err
}

func (r *Repo) ListRequestLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]SlackRequestLogRecord, error) {
	rows := make([]SlackRequestLogRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT
			id,
			request_type,
			endpoint,
			workspace_id,
			slack_team_id,
			slack_user_id,
			slack_channel_id,
			command,
			trigger_id,
			request_body,
			headers,
			response_code,
			outcome,
			error_message,
			created_at
		FROM slack_request_logs
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
