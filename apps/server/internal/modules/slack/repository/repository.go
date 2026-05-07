package slackrepository

import (
	"context"
	"database/sql"
	"errors"
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

type SlackWorkspaceSettingsRecord struct {
	WorkspaceID       uuid.UUID `db:"workspace_id"`
	DefaultCreateMode string    `db:"default_create_mode"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
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

type SlackChannelLinkRecord struct {
	ID             uuid.UUID `db:"id"`
	WorkspaceID    uuid.UUID `db:"workspace_id"`
	SlackChannelID string    `db:"slack_channel_id"`
	TeamID         uuid.UUID `db:"team_id"`
	TeamCode       string    `db:"team_code"`
	TeamName       string    `db:"team_name"`
	TeamColor      string    `db:"team_color"`
	IsActive       bool      `db:"is_active"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
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

func (r *Repo) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (SlackWorkspaceSettingsRecord, error) {
	var row SlackWorkspaceSettingsRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT workspace_id, default_create_mode, created_at, updated_at
		FROM slack_workspace_settings
		WHERE workspace_id = $1
	`, workspaceID)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return SlackWorkspaceSettingsRecord{}, err
	}
	err = r.db.GetContext(ctx, &row, `
		INSERT INTO slack_workspace_settings (workspace_id)
		VALUES ($1)
		RETURNING workspace_id, default_create_mode, created_at, updated_at
	`, workspaceID)
	if err != nil {
		return SlackWorkspaceSettingsRecord{}, err
	}
	return row, nil
}

func (r *Repo) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, defaultCreateMode string) (SlackWorkspaceSettingsRecord, error) {
	var row SlackWorkspaceSettingsRecord
	err := r.db.GetContext(ctx, &row, `
		UPDATE slack_workspace_settings
		SET default_create_mode = $2,
		    updated_at = NOW()
		WHERE workspace_id = $1
		RETURNING workspace_id, default_create_mode, created_at, updated_at
	`, workspaceID, defaultCreateMode)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return SlackWorkspaceSettingsRecord{}, err
	}
	err = r.db.GetContext(ctx, &row, `
		INSERT INTO slack_workspace_settings (workspace_id, default_create_mode)
		VALUES ($1, $2)
		RETURNING workspace_id, default_create_mode, created_at, updated_at
	`, workspaceID, defaultCreateMode)
	if err != nil {
		return SlackWorkspaceSettingsRecord{}, err
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

func (r *Repo) UpsertChannelLink(ctx context.Context, workspaceID uuid.UUID, slackChannelID string, teamID, createdByUserID uuid.UUID) (SlackChannelLinkRecord, error) {
	var row SlackChannelLinkRecord
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO slack_team_channel_links (
			workspace_id, slack_channel_id, team_id, is_active, created_by_user_id
		) VALUES ($1, $2, $3, true, $4)
		ON CONFLICT (workspace_id, slack_channel_id) WHERE is_active = true
		DO UPDATE SET
			team_id = EXCLUDED.team_id,
			is_active = true,
			updated_at = NOW()
		RETURNING id, workspace_id, slack_channel_id, team_id, is_active, created_at, updated_at
	`, workspaceID, slackChannelID, teamID, createdByUserID)
	if err != nil {
		return SlackChannelLinkRecord{}, err
	}
	return r.GetChannelLinkByID(ctx, workspaceID, row.ID)
}

func (r *Repo) GetChannelLinkByID(ctx context.Context, workspaceID, linkID uuid.UUID) (SlackChannelLinkRecord, error) {
	var row SlackChannelLinkRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT l.id,
		       l.workspace_id,
		       l.slack_channel_id,
		       l.team_id,
		       t.code AS team_code,
		       t.name AS team_name,
		       t.color AS team_color,
		       l.is_active,
		       l.created_at,
		       l.updated_at
		FROM slack_team_channel_links l
		JOIN teams t ON t.team_id = l.team_id
		WHERE l.workspace_id = $1 AND l.id = $2
	`, workspaceID, linkID)
	if err != nil {
		return SlackChannelLinkRecord{}, err
	}
	return row, nil
}

func (r *Repo) ListChannelLinks(ctx context.Context, workspaceID uuid.UUID) ([]SlackChannelLinkRecord, error) {
	rows := make([]SlackChannelLinkRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT l.id,
		       l.workspace_id,
		       l.slack_channel_id,
		       l.team_id,
		       t.code AS team_code,
		       t.name AS team_name,
		       t.color AS team_color,
		       l.is_active,
		       l.created_at,
		       l.updated_at
		FROM slack_team_channel_links l
		JOIN teams t ON t.team_id = l.team_id
		WHERE l.workspace_id = $1
		ORDER BY t.name ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) UpdateChannelLinkActive(ctx context.Context, workspaceID, linkID uuid.UUID, isActive bool) (SlackChannelLinkRecord, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE slack_team_channel_links
		SET is_active = $3, updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2
	`, workspaceID, linkID, isActive)
	if err != nil {
		return SlackChannelLinkRecord{}, err
	}
	return r.GetChannelLinkByID(ctx, workspaceID, linkID)
}

func (r *Repo) DeleteChannelLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM slack_team_channel_links
		WHERE workspace_id = $1 AND id = $2
	`, workspaceID, linkID)
	return err
}

func (r *Repo) FindTeamByChannel(ctx context.Context, workspaceID uuid.UUID, slackChannelID string) (TeamRecord, error) {
	var row TeamRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT t.team_id, t.code, t.name, t.color
		FROM slack_team_channel_links l
		JOIN teams t ON t.team_id = l.team_id
		WHERE l.workspace_id = $1
		  AND l.slack_channel_id = $2
		  AND l.is_active = true
		LIMIT 1
	`, workspaceID, slackChannelID)
	if err != nil {
		return TeamRecord{}, err
	}
	return row, nil
}

func (r *Repo) ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]TeamRecord, error) {
	rows := make([]TeamRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT team_id, code, name, color
		FROM teams
		WHERE workspace_id = $1
		  AND deleted_at IS NULL
		ORDER BY name ASC
	`, workspaceID)
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
		  AND deleted_at IS NULL
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

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	var statusID uuid.UUID
	err := r.db.GetContext(ctx, &statusID, `
		SELECT status_id
		FROM statuses
		WHERE team_id = $1
		  AND category = $2
		  AND deleted_at IS NULL
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

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
