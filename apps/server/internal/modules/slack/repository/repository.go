package slackrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrActiveInstallationConflict  = errors.New("an active Slack installation conflicts with this connection")
	ErrWorkspaceAlreadyConnected   = errors.New("this FortyOne workspace is already connected to another Slack team")
	ErrSlackTeamAlreadyConnected   = errors.New("this Slack team is already connected to another FortyOne workspace")
	ErrUninstallInProgress         = errors.New("Slack uninstall is still processing")
	ErrUninstallResolutionRequired = errors.New("Slack uninstall requires operator resolution")
)

const (
	SlackUninstallMaxAttempts             = 8
	slackUninstallLease                   = 2 * time.Minute
	SlackInstallationLifecycleAdvisoryKey = int64(0x534c41434b)
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

type SlackUserLinkRecord struct {
	SlackUserID string    `db:"slack_user_id"`
	UserID      uuid.UUID `db:"user_id"`
	LinkedVia   string    `db:"linked_via"`
	LinkedAt    time.Time `db:"linked_at"`
}

type SlackWorkspaceRecord struct {
	ID                uuid.UUID  `db:"id"`
	WorkspaceID       uuid.UUID  `db:"workspace_id"`
	SlackTeamID       string     `db:"slack_team_id"`
	SlackTeamName     string     `db:"slack_team_name"`
	SlackTeamDomain   string     `db:"slack_team_domain"`
	BotUserID         *string    `db:"bot_user_id"`
	BotAccessToken    string     `db:"bot_access_token"`
	CredentialVersion int        `db:"credential_key_version"`
	InstallGeneration uuid.UUID  `db:"installation_generation"`
	AuthorizedAt      time.Time  `db:"installation_authorized_at"`
	SlackAppID        *string    `db:"slack_app_id"`
	EnterpriseID      *string    `db:"enterprise_id"`
	AuthedUserID      *string    `db:"authed_user_id"`
	Scope             *string    `db:"scope"`
	IsActive          bool       `db:"is_active"`
	InstalledByUserID *uuid.UUID `db:"installed_by_user_id"`
	RevokedAt         *time.Time `db:"revoked_at"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type LegacySlackCredentialRecord struct {
	SlackWorkspaceID uuid.UUID `db:"id"`
	Credential       string    `db:"credential"`
}

type SlackChannelRecord struct {
	ID                    uuid.UUID  `db:"id"`
	WorkspaceID           uuid.UUID  `db:"workspace_id"`
	SlackWorkspaceID      uuid.UUID  `db:"slack_workspace_id"`
	SlackChannelID        string     `db:"slack_channel_id"`
	Name                  string     `db:"name"`
	IsPrivate             bool       `db:"is_private"`
	IsArchived            bool       `db:"is_archived"`
	IsMember              bool       `db:"is_member"`
	IsActive              bool       `db:"is_active"`
	IsAssistantConfigured bool       `db:"is_assistant_configured"`
	LastSyncedAt          *time.Time `db:"last_synced_at"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

type OAuthInstallPayload struct {
	SlackTeamID       string
	SlackTeamName     string
	SlackTeamDomain   string
	BotUserID         *string
	BotAccessToken    string
	LegacyAccessToken string
	CredentialVersion int
	SlackAppID        *string
	EnterpriseID      *string
	AuthedUserID      *string
	Scope             *string
}

type SlackUninstallRecord struct {
	ID                   uuid.UUID  `db:"id"`
	SlackWorkspaceID     uuid.UUID  `db:"slack_workspace_id"`
	WorkspaceID          uuid.UUID  `db:"workspace_id"`
	InstallGeneration    uuid.UUID  `db:"installation_generation"`
	SlackTeamID          string     `db:"slack_team_id"`
	UninstallKind        string     `db:"uninstall_kind"`
	CredentialPayload    string     `db:"credential_payload"`
	CredentialKeyVersion int        `db:"credential_key_version"`
	Status               string     `db:"status"`
	AttemptCount         int        `db:"attempt_count"`
	LastError            *string    `db:"last_error"`
	NextAttemptAt        *time.Time `db:"next_attempt_at"`
	ProcessingStartedAt  *time.Time `db:"processing_started_at"`
	CompletedAt          *time.Time `db:"completed_at"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
}

type SlackUninstallInput struct {
	SlackWorkspaceID     uuid.UUID
	WorkspaceID          uuid.UUID
	InstallGeneration    uuid.UUID
	SlackTeamID          string
	UninstallKind        string
	CredentialPayload    string
	CredentialKeyVersion int
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
	payload.SlackTeamID = strings.TrimSpace(payload.SlackTeamID)
	if workspaceID == uuid.Nil || payload.SlackTeamID == "" {
		return SlackWorkspaceRecord{}, errors.New("workspace and Slack team are required")
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return SlackWorkspaceRecord{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err = lockSlackInstallationLifecycle(ctx, tx); err != nil {
		return SlackWorkspaceRecord{}, err
	}

	var activeInstallations []struct {
		WorkspaceID       uuid.UUID `db:"workspace_id"`
		SlackTeamID       string    `db:"slack_team_id"`
		InstallGeneration uuid.UUID `db:"installation_generation"`
	}
	if err = tx.SelectContext(ctx, &activeInstallations, `
		SELECT workspace_id, slack_team_id, installation_generation
		FROM slack_workspaces
		WHERE is_active = true
		  AND (workspace_id = $1 OR slack_team_id = $2)
		FOR UPDATE
	`, workspaceID, payload.SlackTeamID); err != nil {
		return SlackWorkspaceRecord{}, err
	}
	refreshingActiveInstallation := false
	previousInstallGeneration := uuid.Nil
	workspaceConflict := false
	teamConflict := false
	for _, installation := range activeInstallations {
		switch {
		case installation.WorkspaceID == workspaceID && installation.SlackTeamID == payload.SlackTeamID:
			refreshingActiveInstallation = true
			previousInstallGeneration = installation.InstallGeneration
		case installation.SlackTeamID == payload.SlackTeamID:
			teamConflict = true
		case installation.WorkspaceID == workspaceID:
			workspaceConflict = true
		}
	}
	if teamConflict {
		return SlackWorkspaceRecord{}, fmt.Errorf(
			"%w: %w; disconnect it from the other FortyOne workspace before reconnecting",
			ErrActiveInstallationConflict,
			ErrSlackTeamAlreadyConnected,
		)
	}
	if workspaceConflict {
		return SlackWorkspaceRecord{}, fmt.Errorf(
			"%w: %w; disconnect the current Slack team before installing another",
			ErrActiveInstallationConflict,
			ErrWorkspaceAlreadyConnected,
		)
	}

	var uninstallState struct {
		Processing         bool `db:"processing"`
		ResolutionRequired bool `db:"resolution_required"`
	}
	if err = tx.GetContext(ctx, &uninstallState, `
		SELECT EXISTS (
			SELECT 1 FROM slack_uninstall_outbox
			WHERE slack_team_id = $1 AND status = 'processing'
		) AS processing,
		EXISTS (
			SELECT 1 FROM slack_uninstall_outbox
			WHERE slack_team_id = $1 AND status = 'revocation_required'
		) AS resolution_required
	`, payload.SlackTeamID); err != nil {
		return SlackWorkspaceRecord{}, err
	}
	if uninstallState.Processing {
		return SlackWorkspaceRecord{}, fmt.Errorf("%w: retry the Slack installation shortly", ErrUninstallInProgress)
	}
	if uninstallState.ResolutionRequired {
		return SlackWorkspaceRecord{}, fmt.Errorf("%w: contact an administrator before reconnecting this Slack team", ErrUninstallResolutionRequired)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE slack_uninstall_outbox
		SET status = 'completed',
		    credential_payload = NULL,
		    last_error = 'superseded by Slack reinstall',
		    next_attempt_at = NULL,
		    processing_started_at = NULL,
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE slack_team_id = $1
		  AND status IN ('pending', 'failed')
	`, payload.SlackTeamID); err != nil {
		return SlackWorkspaceRecord{}, err
	}
	if _, err = tx.ExecContext(ctx, `
		DELETE FROM slack_workspaces
		WHERE is_active = false
		  AND (workspace_id = $1 OR slack_team_id = $2)
	`, workspaceID, payload.SlackTeamID); err != nil {
		return SlackWorkspaceRecord{}, err
	}
	if refreshingActiveInstallation {
		if err = cancelSlackMessagingTx(ctx, tx, payload.SlackTeamID, "Slack installation refreshed"); err != nil {
			return SlackWorkspaceRecord{}, err
		}
	}

	var row SlackWorkspaceRecord
	err = tx.GetContext(ctx, &row, `
		INSERT INTO slack_workspaces (
			workspace_id, slack_team_id, slack_team_name, slack_team_domain,
			bot_user_id, bot_access_token, credential_payload, credential_key_version,
			slack_app_id, enterprise_id, authed_user_id, scope, is_active, installed_by_user_id, revoked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true, $13, NULL)
		ON CONFLICT (workspace_id) DO UPDATE SET
			slack_team_id = EXCLUDED.slack_team_id,
			slack_team_name = EXCLUDED.slack_team_name,
			slack_team_domain = EXCLUDED.slack_team_domain,
			bot_user_id = EXCLUDED.bot_user_id,
			bot_access_token = EXCLUDED.bot_access_token,
			credential_payload = EXCLUDED.credential_payload,
			credential_key_version = EXCLUDED.credential_key_version,
			installation_generation = gen_random_uuid(),
			installation_authorized_at = NOW(),
			slack_app_id = EXCLUDED.slack_app_id,
			enterprise_id = EXCLUDED.enterprise_id,
			authed_user_id = EXCLUDED.authed_user_id,
			scope = EXCLUDED.scope,
			is_active = true,
			installed_by_user_id = EXCLUDED.installed_by_user_id,
			revoked_at = NULL,
			updated_at = NOW()
		RETURNING id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
		          bot_user_id, credential_payload AS bot_access_token, credential_key_version,
		          installation_generation, installation_authorized_at,
		          slack_app_id, enterprise_id, authed_user_id, scope, is_active, installed_by_user_id,
		          revoked_at, created_at, updated_at
	`,
		workspaceID,
		payload.SlackTeamID,
		payload.SlackTeamName,
		payload.SlackTeamDomain,
		payload.BotUserID,
		payload.LegacyAccessToken,
		payload.BotAccessToken,
		payload.CredentialVersion,
		payload.SlackAppID,
		payload.EnterpriseID,
		payload.AuthedUserID,
		payload.Scope,
		installedByUserID,
	)
	if err != nil {
		return SlackWorkspaceRecord{}, err
	}
	if refreshingActiveInstallation {
		if err = rebindSlackRequestThreadsTx(
			ctx,
			tx,
			workspaceID,
			payload.SlackTeamID,
			previousInstallGeneration,
			row.InstallGeneration,
		); err != nil {
			return SlackWorkspaceRecord{}, err
		}
	}

	authedUserID := ""
	if payload.AuthedUserID != nil {
		authedUserID = strings.TrimSpace(*payload.AuthedUserID)
	}
	if authedUserID != "" {
		result, linkErr := tx.ExecContext(ctx, `
			INSERT INTO slack_user_links (
				workspace_id,
				slack_workspace_id,
				slack_team_id,
				slack_user_id,
				user_id,
				linked_via,
				linked_at
			)
			SELECT $1, $2, $3, $4, $5, 'oauth_installer', NOW()
			FROM workspace_members wm
			JOIN users u ON u.user_id = wm.user_id
			WHERE wm.workspace_id = $1
			  AND wm.user_id = $5
			  AND u.is_active = true
			ON CONFLICT (workspace_id, slack_team_id, slack_user_id) DO UPDATE SET
				slack_workspace_id = EXCLUDED.slack_workspace_id,
				user_id = EXCLUDED.user_id,
				linked_via = EXCLUDED.linked_via,
				linked_at = NOW(),
				updated_at = NOW()
		`, workspaceID, row.ID, payload.SlackTeamID, authedUserID, installedByUserID)
		if linkErr != nil {
			return SlackWorkspaceRecord{}, linkErr
		}
		affected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return SlackWorkspaceRecord{}, rowsErr
		}
		if affected == 0 {
			return SlackWorkspaceRecord{}, sql.ErrNoRows
		}
	}
	if err := tx.Commit(); err != nil {
		return SlackWorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (SlackWorkspaceRecord, error) {
	var row SlackWorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
		       bot_user_id, COALESCE(NULLIF(credential_payload, ''), bot_access_token) AS bot_access_token,
		       CASE WHEN NULLIF(credential_payload, '') IS NULL THEN 0 ELSE credential_key_version END AS credential_key_version,
		       installation_generation, installation_authorized_at,
		       slack_app_id, enterprise_id, authed_user_id, scope, is_active, installed_by_user_id,
		       revoked_at, created_at, updated_at
		FROM slack_workspaces
		WHERE workspace_id = $1 AND is_active = true
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
		       bot_user_id, COALESCE(NULLIF(credential_payload, ''), bot_access_token) AS bot_access_token,
		       CASE WHEN NULLIF(credential_payload, '') IS NULL THEN 0 ELSE credential_key_version END AS credential_key_version,
		       installation_generation, installation_authorized_at,
		       slack_app_id, enterprise_id, authed_user_id, scope, is_active, installed_by_user_id,
		       revoked_at, created_at, updated_at
		FROM slack_workspaces
		WHERE slack_team_id = $1 AND is_active = true
		LIMIT 1
	`, slackTeamID)
	if err != nil {
		return SlackWorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) DisconnectSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (SlackUninstallRecord, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return SlackUninstallRecord{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err = lockSlackInstallationLifecycle(ctx, tx); err != nil {
		return SlackUninstallRecord{}, err
	}

	var installation struct {
		ID                   uuid.UUID `db:"id"`
		WorkspaceID          uuid.UUID `db:"workspace_id"`
		InstallGeneration    uuid.UUID `db:"installation_generation"`
		SlackTeamID          string    `db:"slack_team_id"`
		CredentialPayload    string    `db:"credential_payload"`
		CredentialKeyVersion int       `db:"credential_key_version"`
	}
	if err = tx.GetContext(ctx, &installation, `
		SELECT id, workspace_id, installation_generation, slack_team_id,
		       COALESCE(credential_payload, '') AS credential_payload, credential_key_version
		FROM slack_workspaces
		WHERE workspace_id = $1 AND is_active = true
		FOR UPDATE
	`, workspaceID); err != nil {
		return SlackUninstallRecord{}, err
	}
	if installation.CredentialKeyVersion <= 0 || strings.TrimSpace(installation.CredentialPayload) == "" {
		return SlackUninstallRecord{}, errors.New("Slack installation credential must be encrypted before disconnect")
	}

	var uninstall SlackUninstallRecord
	if err = tx.GetContext(ctx, &uninstall, `
		INSERT INTO slack_uninstall_outbox (
			slack_workspace_id, workspace_id, installation_generation, slack_team_id,
			credential_payload, credential_key_version, status, next_attempt_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending', NOW())
		RETURNING id, slack_workspace_id, workspace_id, installation_generation, slack_team_id,
		          uninstall_kind, credential_payload, credential_key_version, status, attempt_count, last_error,
		          next_attempt_at, processing_started_at, completed_at, created_at, updated_at
	`, installation.ID, installation.WorkspaceID, installation.InstallGeneration, installation.SlackTeamID,
		installation.CredentialPayload, installation.CredentialKeyVersion); err != nil {
		return SlackUninstallRecord{}, err
	}
	if err = cancelSlackMessagingTx(ctx, tx, installation.SlackTeamID, "Slack installation disconnected"); err != nil {
		return SlackUninstallRecord{}, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM slack_user_links WHERE workspace_id = $1`, workspaceID); err != nil {
		return SlackUninstallRecord{}, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM slack_workspaces WHERE id = $1`, installation.ID); err != nil {
		return SlackUninstallRecord{}, err
	}
	if err = tx.Commit(); err != nil {
		return SlackUninstallRecord{}, err
	}
	return uninstall, nil
}

func lockSlackInstallationLifecycle(ctx context.Context, tx *sqlx.Tx) error {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, SlackInstallationLifecycleAdvisoryKey); err != nil {
		return fmt.Errorf("lock Slack installation lifecycle: %w", err)
	}
	return nil
}

func cancelSlackMessagingTx(ctx context.Context, tx *sqlx.Tx, slackTeamID, reason string) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE messaging_inbound_events
		SET status = 'cancelled',
		    payload_encrypted = NULL,
		    last_error = $2,
		    recovery_enqueued_at = NULL,
		    processed_at = NOW(),
		    updated_at = NOW()
		WHERE provider = 'slack'
		  AND external_workspace_id = $1
		  AND status IN ('pending', 'processing', 'failed')
	`, slackTeamID, reason); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'cancelled',
		    content = NULL,
		    last_error = $2,
		    updated_at = NOW()
		WHERE provider = 'slack'
		  AND external_workspace_id = $1
		  AND status IN ('pending', 'delivering', 'failed')
	`, slackTeamID, reason); err != nil {
		return err
	}
	return nil
}

const rebindSlackRequestThreadsQuery = `
	UPDATE integration_request_threads
	SET installation_generation = $4,
	    updated_at = NOW()
	WHERE workspace_id = $1
	  AND provider = 'slack'
	  AND external_workspace_id = $2
	  AND installation_generation = $3
`

type slackRequestThreadRebindExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func rebindSlackRequestThreadsTx(
	ctx context.Context,
	execer slackRequestThreadRebindExecer,
	workspaceID uuid.UUID,
	slackTeamID string,
	previousGeneration uuid.UUID,
	currentGeneration uuid.UUID,
) error {
	slackTeamID = strings.TrimSpace(slackTeamID)
	if workspaceID == uuid.Nil || slackTeamID == "" || previousGeneration == uuid.Nil || currentGeneration == uuid.Nil {
		return errors.New("rebind Slack request threads requires workspace, team, and installation generations")
	}
	if previousGeneration == currentGeneration {
		return errors.New("rebind Slack request threads requires a rotated installation generation")
	}
	if _, err := execer.ExecContext(
		ctx,
		rebindSlackRequestThreadsQuery,
		workspaceID,
		slackTeamID,
		previousGeneration,
		currentGeneration,
	); err != nil {
		return fmt.Errorf("rebind Slack request threads to refreshed installation: %w", err)
	}
	return nil
}

func (r *Repo) EnqueueSlackUninstall(ctx context.Context, input SlackUninstallInput) (SlackUninstallRecord, error) {
	if input.SlackWorkspaceID == uuid.Nil {
		input.SlackWorkspaceID = uuid.New()
	}
	if input.InstallGeneration == uuid.Nil {
		input.InstallGeneration = uuid.New()
	}
	if strings.TrimSpace(input.UninstallKind) == "" {
		input.UninstallKind = "disconnect"
	}
	input.SlackTeamID = strings.TrimSpace(input.SlackTeamID)
	input.CredentialPayload = strings.TrimSpace(input.CredentialPayload)
	if input.WorkspaceID == uuid.Nil || input.SlackTeamID == "" || input.CredentialPayload == "" || input.CredentialKeyVersion <= 0 {
		return SlackUninstallRecord{}, errors.New("Slack uninstall requires workspace, team, and versioned encrypted credential")
	}
	var record SlackUninstallRecord
	err := r.db.GetContext(ctx, &record, `
		INSERT INTO slack_uninstall_outbox (
			slack_workspace_id, workspace_id, installation_generation, slack_team_id, uninstall_kind,
			credential_payload, credential_key_version, status, next_attempt_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW())
		RETURNING id, slack_workspace_id, workspace_id, installation_generation, slack_team_id,
		          uninstall_kind, credential_payload, credential_key_version, status, attempt_count,
		          last_error, next_attempt_at, processing_started_at, completed_at, created_at, updated_at
	`, input.SlackWorkspaceID, input.WorkspaceID, input.InstallGeneration, input.SlackTeamID,
		input.UninstallKind, input.CredentialPayload, input.CredentialKeyVersion)
	if err != nil {
		return SlackUninstallRecord{}, fmt.Errorf("enqueue Slack uninstall: %w", err)
	}
	return record, nil
}

func (r *Repo) ClaimSlackUninstall(ctx context.Context, id uuid.UUID) (SlackUninstallRecord, bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return SlackUninstallRecord{}, false, fmt.Errorf("begin Slack uninstall claim: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err = lockSlackInstallationLifecycle(ctx, tx); err != nil {
		return SlackUninstallRecord{}, false, err
	}
	var record SlackUninstallRecord
	err = tx.GetContext(ctx, &record, `
		UPDATE slack_uninstall_outbox
		SET status = 'processing',
		    attempt_count = attempt_count + 1,
		    last_error = NULL,
		    next_attempt_at = NULL,
		    processing_started_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		  AND attempt_count < $2
		  AND (
			(status IN ('pending', 'failed') AND COALESCE(next_attempt_at, NOW()) <= NOW())
			OR (status = 'processing' AND updated_at < NOW() - ($3 * INTERVAL '1 second'))
		  )
		RETURNING id, slack_workspace_id, workspace_id, installation_generation, slack_team_id,
		          uninstall_kind, credential_payload, credential_key_version, status, attempt_count, last_error,
		          next_attempt_at, processing_started_at, completed_at, created_at, updated_at
	`, id, SlackUninstallMaxAttempts, int64(slackUninstallLease/time.Second))
	if err == nil {
		if commitErr := tx.Commit(); commitErr != nil {
			return SlackUninstallRecord{}, false, fmt.Errorf("commit Slack uninstall claim: %w", commitErr)
		}
		return record, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		if commitErr := tx.Commit(); commitErr != nil {
			return SlackUninstallRecord{}, false, fmt.Errorf("commit empty Slack uninstall claim: %w", commitErr)
		}
		return SlackUninstallRecord{}, false, nil
	}
	return SlackUninstallRecord{}, false, fmt.Errorf("claim Slack uninstall: %w", err)
}

func (r *Repo) ClaimRecoverableSlackUninstalls(ctx context.Context, limit int) ([]SlackUninstallRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin recoverable Slack uninstall claim: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err = lockSlackInstallationLifecycle(ctx, tx); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE slack_uninstall_outbox
		SET status = 'revocation_required',
		    last_error = COALESCE(NULLIF(last_error, ''), 'Slack uninstall recovery lease expired after the final attempt'),
		    next_attempt_at = NULL,
		    processing_started_at = NULL,
		    updated_at = NOW()
		WHERE attempt_count >= $1
		  AND (
			status IN ('pending', 'failed')
			OR (status = 'processing' AND updated_at < NOW() - ($2 * INTERVAL '1 second'))
		  )
	`, SlackUninstallMaxAttempts, int64(slackUninstallLease/time.Second)); err != nil {
		return nil, fmt.Errorf("dead-letter exhausted Slack uninstalls: %w", err)
	}
	records := make([]SlackUninstallRecord, 0)
	err = tx.SelectContext(ctx, &records, `
		WITH candidates AS (
			SELECT id
			FROM slack_uninstall_outbox
			WHERE attempt_count < $2
			  AND (
				(status IN ('pending', 'failed') AND COALESCE(next_attempt_at, NOW()) <= NOW())
				OR (status = 'processing' AND updated_at < NOW() - ($3 * INTERVAL '1 second'))
			  )
			ORDER BY COALESCE(next_attempt_at, created_at), created_at
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE slack_uninstall_outbox suo
		SET status = 'processing',
		    attempt_count = suo.attempt_count + 1,
		    last_error = NULL,
		    next_attempt_at = NULL,
		    processing_started_at = NOW(),
		    updated_at = NOW()
		FROM candidates
		WHERE suo.id = candidates.id
		RETURNING suo.id, suo.slack_workspace_id, suo.workspace_id, suo.installation_generation,
		          suo.slack_team_id, suo.uninstall_kind, suo.credential_payload, suo.credential_key_version, suo.status,
		          suo.attempt_count, suo.last_error, suo.next_attempt_at, suo.processing_started_at,
		          suo.completed_at, suo.created_at, suo.updated_at
	`, limit, SlackUninstallMaxAttempts, int64(slackUninstallLease/time.Second))
	if err != nil {
		return nil, fmt.Errorf("claim recoverable Slack uninstalls: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit recoverable Slack uninstall claims: %w", err)
	}
	return records, nil
}

func (r *Repo) CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE slack_uninstall_outbox
		SET status = 'completed',
		    credential_payload = NULL,
		    last_error = NULLIF($2, ''),
		    next_attempt_at = NULL,
		    processing_started_at = NULL,
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, id, message)
	if err != nil {
		return fmt.Errorf("complete Slack uninstall: %w", err)
	}
	return nil
}

func (r *Repo) FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE slack_uninstall_outbox
		SET status = CASE WHEN $3 IS NULL THEN 'revocation_required' ELSE 'failed' END,
		    last_error = $2,
		    next_attempt_at = $3,
		    processing_started_at = NULL,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, id, message, nextAttemptAt)
	if err != nil {
		return fmt.Errorf("fail Slack uninstall: %w", err)
	}
	return nil
}

func (r *Repo) UpgradeSlackCredential(ctx context.Context, slackWorkspaceID uuid.UUID, encrypted string, version int) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE slack_workspaces
		SET credential_payload = $2,
		    credential_key_version = $3,
		    updated_at = NOW()
		WHERE id = $1
		  AND is_active = true
		  AND credential_key_version = 0
	`, slackWorkspaceID, encrypted, version)
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
	return nil
}

func (r *Repo) ScrubVersionedLegacySlackCredentials(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	result, err := r.db.ExecContext(ctx, `
		WITH candidates AS (
			SELECT id
			FROM slack_workspaces
			WHERE credential_key_version > 0
			  AND NULLIF(credential_payload, '') IS NOT NULL
			  AND NULLIF(bot_access_token, '') IS NOT NULL
			ORDER BY created_at ASC
			LIMIT $1
		)
		UPDATE slack_workspaces sw
		SET bot_access_token = '', updated_at = NOW()
		FROM candidates
		WHERE sw.id = candidates.id
	`, limit)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

func (r *Repo) ListLegacySlackCredentials(ctx context.Context, limit int) ([]LegacySlackCredentialRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows := make([]LegacySlackCredentialRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT id,
		       COALESCE(NULLIF(credential_payload, ''), bot_access_token) AS credential
		FROM slack_workspaces
		WHERE is_active = true
		  AND credential_key_version = 0
		  AND COALESCE(NULLIF(credential_payload, ''), bot_access_token) <> ''
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) DeactivateSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string, installGeneration uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err = lockSlackInstallationLifecycle(ctx, tx); err != nil {
		return err
	}

	var installation struct {
		ID          uuid.UUID `db:"id"`
		WorkspaceID uuid.UUID `db:"workspace_id"`
	}
	err = tx.GetContext(ctx, &installation, `
		SELECT id, workspace_id
		FROM slack_workspaces
		WHERE slack_team_id = $1
		  AND installation_generation = $2
		  AND is_active = true
		FOR UPDATE
	`, slackTeamID, installGeneration)
	if err != nil {
		return err
	}
	if err = cancelSlackMessagingTx(ctx, tx, slackTeamID, "Slack installation revoked"); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM slack_user_links WHERE workspace_id = $1`, installation.WorkspaceID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM slack_workspaces WHERE id = $1`, installation.ID); err != nil {
		return err
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

	if _, err = tx.ExecContext(ctx, `
		UPDATE slack_channels
		SET is_active = false, updated_at = NOW()
		WHERE workspace_id = $1
	`, workspaceID); err != nil {
		return err
	}

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
		       is_private, is_archived, is_member, is_active, is_assistant_configured,
		       last_synced_at, created_at, updated_at
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

const listWorkspaceTeamsForUserQuery = `
	SELECT t.team_id, t.code, t.name, t.color
	FROM teams t
	JOIN team_members tm ON tm.team_id = t.team_id
	JOIN workspace_members wm ON wm.workspace_id = t.workspace_id AND wm.user_id = tm.user_id
	JOIN users u ON u.user_id = tm.user_id
	LEFT JOIN user_team_orders uto ON uto.team_id = t.team_id
		AND uto.user_id = $2
		AND uto.workspace_id = $1
	WHERE t.workspace_id = $1
	  AND tm.user_id = $2
	  AND u.is_active = true
	ORDER BY
		CASE WHEN uto.order_index IS NOT NULL THEN 0 ELSE 1 END,
		uto.order_index ASC NULLS LAST,
		t.created_at DESC,
		t.team_id ASC
`

func (r *Repo) ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]TeamRecord, error) {
	rows := make([]TeamRecord, 0)
	err := r.db.SelectContext(ctx, &rows, listWorkspaceTeamsForUserQuery, workspaceID, userID)
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

func (r *Repo) CreateStoryLink(ctx context.Context, storyID uuid.UUID, sourceKey, title, linkURL string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO story_links (title, url, story_id, external_source_key)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (external_source_key)
		WHERE external_source_key IS NOT NULL
		DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			story_id = EXCLUDED.story_id,
			updated_at = NOW()
	`, title, linkURL, storyID, sourceKey)
	if err != nil {
		return fmt.Errorf("upsert Slack story source link: %w", err)
	}
	return nil
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
		SELECT sul.user_id
		FROM slack_user_links sul
		JOIN users u ON u.user_id = sul.user_id
		JOIN workspace_members wm ON wm.workspace_id = sul.workspace_id AND wm.user_id = sul.user_id
		WHERE sul.workspace_id = $1
		  AND sul.slack_team_id = $2
		  AND sul.slack_user_id = $3
		  AND u.is_active = true
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

func (r *Repo) FindSlackUserLinkByUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, userID uuid.UUID) (*SlackUserLinkRecord, error) {
	var row SlackUserLinkRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT sul.slack_user_id, sul.user_id, sul.linked_via, sul.linked_at
		FROM slack_user_links sul
		JOIN users u ON u.user_id = sul.user_id
		JOIN workspace_members wm ON wm.workspace_id = sul.workspace_id AND wm.user_id = sul.user_id
		WHERE sul.workspace_id = $1
		  AND sul.slack_team_id = $2
		  AND sul.user_id = $3
		  AND u.is_active = true
		LIMIT 1
	`, workspaceID, strings.TrimSpace(slackTeamID), userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) DeleteSlackUserLink(
	ctx context.Context,
	workspaceID uuid.UUID,
	slackTeamID, slackUserID string,
	userID uuid.UUID,
) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM slack_user_links
		WHERE workspace_id = $1
		  AND slack_team_id = $2
		  AND slack_user_id = $3
		  AND user_id = $4
	`, workspaceID, strings.TrimSpace(slackTeamID), strings.TrimSpace(slackUserID), userID)
	if err != nil {
		return false, err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return deleted > 0, nil
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CAST($10 AS jsonb), $11, $12, $13)
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
