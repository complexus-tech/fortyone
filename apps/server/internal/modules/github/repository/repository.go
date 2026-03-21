package githubrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{db: db, log: log}
}

type installationRow struct {
	ID                   uuid.UUID  `db:"id"`
	GitHubInstallationID int64      `db:"github_installation_id"`
	AccountID            int64      `db:"account_id"`
	AccountLogin         string     `db:"account_login"`
	AccountType          string     `db:"account_type"`
	AccountAvatarURL     *string    `db:"account_avatar_url"`
	RepositorySelection  string     `db:"repository_selection"`
	IsActive             bool       `db:"is_active"`
	SuspendedAt          *time.Time `db:"suspended_at"`
	DisconnectedAt       *time.Time `db:"disconnected_at"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
}

type repositoryRow struct {
	ID                 uuid.UUID  `db:"id"`
	InstallationID     uuid.UUID  `db:"installation_id"`
	GitHubRepositoryID int64      `db:"github_repository_id"`
	OwnerLogin         string     `db:"owner_login"`
	Name               string     `db:"name"`
	FullName           string     `db:"full_name"`
	Description        *string    `db:"description"`
	HTMLURL            string     `db:"html_url"`
	DefaultBranch      string     `db:"default_branch"`
	IsPrivate          bool       `db:"is_private"`
	IsArchived         bool       `db:"is_archived"`
	IsDisabled         bool       `db:"is_disabled"`
	IsActive           bool       `db:"is_active"`
	LastSyncedAt       *time.Time `db:"last_synced_at"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

type syncLinkRow struct {
	ID             uuid.UUID `db:"id"`
	RepositoryID   uuid.UUID `db:"repository_id"`
	RepositoryName string    `db:"repository_name"`
	TeamID         uuid.UUID `db:"team_id"`
	TeamName       string    `db:"team_name"`
	TeamColor      string    `db:"team_color"`
	SyncDirection  string    `db:"sync_direction"`
	IsActive       bool      `db:"is_active"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type bidirectionalLinkRow struct {
	ID                   uuid.UUID `db:"id"`
	RepositoryID         uuid.UUID `db:"repository_id"`
	TeamID               uuid.UUID `db:"team_id"`
	SyncDirection        string    `db:"sync_direction"`
	RepositoryName       string    `db:"repository_name"`
	OwnerLogin           string    `db:"owner_login"`
	RepositorySlug       string    `db:"repository_slug"`
	RepositoryHTMLURL    string    `db:"repository_html_url"`
	GitHubInstallationID int64     `db:"github_installation_id"`
}

type workflowRuleRow struct {
	ID                uuid.UUID  `db:"id"`
	EventKey          string     `db:"event_key"`
	TargetStatusID    *uuid.UUID `db:"target_status_id"`
	BaseBranchPattern *string    `db:"base_branch_pattern"`
	IsActive          bool       `db:"is_active"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type workspaceSettingsRow struct {
	WorkspaceID             uuid.UUID `db:"workspace_id"`
	BranchFormat            string    `db:"branch_format"`
	LinkCommitsByMagicWords bool      `db:"link_commits_by_magic_words"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`
}

type statusRow struct {
	ID       uuid.UUID `db:"status_id"`
	Name     string    `db:"name"`
	Category string    `db:"category"`
	Color    string    `db:"color"`
}

type StoryMatch struct {
	StoryID    uuid.UUID `db:"id"`
	StatusID   uuid.UUID `db:"status_id"`
	TeamID     uuid.UUID `db:"team_id"`
	TeamCode   string    `db:"team_code"`
	SequenceID int       `db:"sequence_id"`
	Title      string    `db:"title"`
}

type RepoByExternalRow struct {
	ID                   uuid.UUID `db:"id"`
	WorkspaceID          uuid.UUID `db:"workspace_id"`
	WorkspaceSlug        string    `db:"workspace_slug"`
	FullName             string    `db:"full_name"`
	OwnerLogin           string    `db:"owner_login"`
	RepositorySlug       string    `db:"repository_slug"`
	DefaultBranch        string    `db:"default_branch"`
	GitHubInstallationID int64     `db:"github_installation_id"`
}

type issueStoryLinkRow struct {
	ID           uuid.UUID `db:"id"`
	StoryID      uuid.UUID `db:"story_id"`
	RepositoryID uuid.UUID `db:"repository_id"`
	GitHubID     int64     `db:"github_id"`
	GitHubNumber int       `db:"github_number"`
	URL          string    `db:"url"`
	Title        *string   `db:"title"`
	State        *string   `db:"state"`
}

func (r *Repo) EnsureStoryLink(ctx context.Context, storyID uuid.UUID, title *string, url string) error {
	if strings.TrimSpace(url) == "" {
		return nil
	}

	var existingID uuid.UUID
	err := r.db.GetContext(ctx, &existingID, `
		SELECT link_id
		FROM story_links
		WHERE story_id = $1 AND url = $2
		LIMIT 1
	`, storyID, url)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO story_links (title, url, story_id)
		VALUES ($1, $2, $3)
	`, title, url, storyID)
	return err
}

func (r *Repo) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (githubshared.CoreWorkspaceSettings, error) {
	var row workspaceSettingsRow
	query := `
		SELECT workspace_id, branch_format, link_commits_by_magic_words, created_at, updated_at
		FROM github_workspace_settings
		WHERE workspace_id = $1
	`
	err := r.db.GetContext(ctx, &row, query, workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			insertQuery := `
				INSERT INTO github_workspace_settings (workspace_id)
				VALUES ($1)
				RETURNING workspace_id, branch_format, link_commits_by_magic_words, created_at, updated_at
			`
			if err := r.db.GetContext(ctx, &row, insertQuery, workspaceID); err != nil {
				return githubshared.CoreWorkspaceSettings{}, err
			}
			return toCoreWorkspaceSettings(row), nil
		}
		return githubshared.CoreWorkspaceSettings{}, err
	}
	return toCoreWorkspaceSettings(row), nil
}

func (r *Repo) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, updates githubshared.CoreUpdateWorkspaceSettingsInput) (githubshared.CoreWorkspaceSettings, error) {
	current, err := r.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return githubshared.CoreWorkspaceSettings{}, err
	}
	if updates.BranchFormat != nil {
		current.BranchFormat = *updates.BranchFormat
	}
	if updates.LinkCommitsByMagicWords != nil {
		current.LinkCommitsByMagicWords = *updates.LinkCommitsByMagicWords
	}
	query := `
		UPDATE github_workspace_settings
		SET branch_format = $2, link_commits_by_magic_words = $3, updated_at = NOW()
		WHERE workspace_id = $1
	`
	if _, err := r.db.ExecContext(ctx, query, workspaceID, current.BranchFormat, current.LinkCommitsByMagicWords); err != nil {
		return githubshared.CoreWorkspaceSettings{}, err
	}
	return r.GetWorkspaceSettings(ctx, workspaceID)
}

func (r *Repo) ListInstallations(ctx context.Context, workspaceID uuid.UUID) ([]githubshared.CoreInstallation, error) {
	var rows []installationRow
	query := `
		SELECT id, github_installation_id, account_id, account_login, account_type, account_avatar_url, repository_selection,
		       is_active, suspended_at, disconnected_at, created_at, updated_at
		FROM github_installations
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID); err != nil {
		return nil, err
	}
	items := make([]githubshared.CoreInstallation, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreInstallation{
			ID: row.ID, GitHubInstallationID: row.GitHubInstallationID, AccountID: row.AccountID,
			AccountLogin: row.AccountLogin, AccountType: row.AccountType, AccountAvatarURL: row.AccountAvatarURL,
			RepositorySelection: row.RepositorySelection, IsActive: row.IsActive, SuspendedAt: row.SuspendedAt,
			DisconnectedAt: row.DisconnectedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) ListRepositories(ctx context.Context, workspaceID uuid.UUID) ([]githubshared.CoreRepository, error) {
	var rows []repositoryRow
	query := `
		SELECT id, installation_id, github_repository_id, owner_login, name, full_name, description, html_url, default_branch,
		       is_private, is_archived, is_disabled, is_active, last_synced_at, created_at, updated_at
		FROM github_repositories
		WHERE workspace_id = $1
		ORDER BY full_name ASC
	`
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID); err != nil {
		return nil, err
	}
	items := make([]githubshared.CoreRepository, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreRepository{
			ID: row.ID, InstallationID: row.InstallationID, GitHubRepositoryID: row.GitHubRepositoryID,
			OwnerLogin: row.OwnerLogin, Name: row.Name, FullName: row.FullName, Description: row.Description,
			HTMLURL: row.HTMLURL, DefaultBranch: row.DefaultBranch, IsPrivate: row.IsPrivate,
			IsArchived: row.IsArchived, IsDisabled: row.IsDisabled, IsActive: row.IsActive,
			LastSyncedAt: row.LastSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) ListIssueSyncLinks(ctx context.Context, workspaceID uuid.UUID) ([]githubshared.CoreIssueSyncLink, error) {
	var rows []syncLinkRow
	query := `
		SELECT l.id, l.repository_id, gr.full_name AS repository_name, l.team_id, t.name AS team_name, t.color AS team_color,
		       l.sync_direction, l.is_active, l.created_at, l.updated_at
		FROM github_issue_sync_links l
		INNER JOIN github_repositories gr ON gr.id = l.repository_id
		INNER JOIN teams t ON t.team_id = l.team_id
		WHERE l.workspace_id = $1
		ORDER BY gr.full_name ASC
	`
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID); err != nil {
		return nil, err
	}
	items := make([]githubshared.CoreIssueSyncLink, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreIssueSyncLink{
			ID: row.ID, RepositoryID: row.RepositoryID, RepositoryName: row.RepositoryName, TeamID: row.TeamID,
			TeamName: row.TeamName, TeamColor: row.TeamColor, SyncDirection: row.SyncDirection,
			IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) UpsertInstallationWithRepositories(
	ctx context.Context,
	workspaceID uuid.UUID,
	installedByUserID uuid.UUID,
	appID int64,
	installation GithubInstallationPayload,
	repositories []GithubRepositoryPayload,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	permissionsJSON, _ := json.Marshal(installation.Permissions)
	eventsJSON, _ := json.Marshal(installation.Events)
	var installationID uuid.UUID
	installationQuery := `
		INSERT INTO github_installations (
			workspace_id, github_app_id, github_installation_id, account_id, account_login, account_type,
			account_avatar_url, repository_selection, permissions, events, installed_by_user_id,
			installed_by_github_user_id, is_active, disconnected_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, true, NULL, NOW()
		)
		ON CONFLICT (github_installation_id) DO UPDATE SET
			workspace_id = EXCLUDED.workspace_id,
			account_id = EXCLUDED.account_id,
			account_login = EXCLUDED.account_login,
			account_type = EXCLUDED.account_type,
			account_avatar_url = EXCLUDED.account_avatar_url,
			repository_selection = EXCLUDED.repository_selection,
			permissions = EXCLUDED.permissions,
			events = EXCLUDED.events,
			installed_by_user_id = EXCLUDED.installed_by_user_id,
			installed_by_github_user_id = EXCLUDED.installed_by_github_user_id,
			is_active = true,
			disconnected_at = NULL,
			updated_at = NOW()
		RETURNING id
	`
	if err := tx.GetContext(
		ctx,
		&installationID,
		installationQuery,
		workspaceID,
		appID,
		installation.ID,
		installation.Account.ID,
		installation.Account.Login,
		installation.Account.Type,
		installation.Account.AvatarURL,
		installation.RepositorySelection,
		permissionsJSON,
		eventsJSON,
		installedByUserID,
		installation.Sender.ID,
	); err != nil {
		return err
	}

	repoIDs := make([]int64, 0, len(repositories))
	for _, repository := range repositories {
		repoIDs = append(repoIDs, repository.ID)
		repoQuery := `
			INSERT INTO github_repositories (
				workspace_id, installation_id, github_repository_id, owner_id, owner_login, name, full_name, description,
				html_url, clone_url, ssh_url, default_branch, is_private, is_archived, is_disabled, is_active, last_synced_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8,
				$9, $10, $11, $12, $13, $14, $15, true, NOW(), NOW()
			)
			ON CONFLICT (installation_id, github_repository_id) DO UPDATE SET
				workspace_id = EXCLUDED.workspace_id,
				owner_id = EXCLUDED.owner_id,
				owner_login = EXCLUDED.owner_login,
				name = EXCLUDED.name,
				full_name = EXCLUDED.full_name,
				description = EXCLUDED.description,
				html_url = EXCLUDED.html_url,
				clone_url = EXCLUDED.clone_url,
				ssh_url = EXCLUDED.ssh_url,
				default_branch = EXCLUDED.default_branch,
				is_private = EXCLUDED.is_private,
				is_archived = EXCLUDED.is_archived,
				is_disabled = EXCLUDED.is_disabled,
				is_active = true,
				last_synced_at = NOW(),
				updated_at = NOW()
		`
		if _, err := tx.ExecContext(
			ctx,
			repoQuery,
			workspaceID,
			installationID,
			repository.ID,
			repository.Owner.ID,
			repository.Owner.Login,
			repository.Name,
			repository.FullName,
			repository.Description,
			repository.HTMLURL,
			repository.CloneURL,
			repository.SSHURL,
			repository.DefaultBranch,
			repository.Private,
			repository.Archived,
			repository.Disabled,
		); err != nil {
			return err
		}
	}

	if len(repoIDs) > 0 {
		query, args, err := sqlx.In(`
			UPDATE github_repositories
			SET is_active = false, updated_at = NOW()
			WHERE installation_id = ? AND github_repository_id NOT IN (?)
		`, installationID, repoIDs)
		if err != nil {
			return err
		}
		query = tx.Rebind(query)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repo) CreateIssueSyncLink(ctx context.Context, workspaceID, userID uuid.UUID, input githubshared.CoreIssueSyncLinkInput) (githubshared.CoreIssueSyncLink, error) {
	var id uuid.UUID
	query := `
		INSERT INTO github_issue_sync_links (workspace_id, repository_id, team_id, sync_direction, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if err := r.db.GetContext(ctx, &id, query, workspaceID, input.RepositoryID, input.TeamID, input.SyncDirection, userID); err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	link, err := r.GetIssueSyncLink(ctx, workspaceID, id)
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	return link, nil
}

func (r *Repo) GetIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID) (githubshared.CoreIssueSyncLink, error) {
	var row syncLinkRow
	query := `
		SELECT l.id, l.repository_id, gr.full_name AS repository_name, l.team_id, t.name AS team_name, t.color AS team_color,
		       l.sync_direction, l.is_active, l.created_at, l.updated_at
		FROM github_issue_sync_links l
		INNER JOIN github_repositories gr ON gr.id = l.repository_id
		INNER JOIN teams t ON t.team_id = l.team_id
		WHERE l.workspace_id = $1 AND l.id = $2
	`
	if err := r.db.GetContext(ctx, &row, query, workspaceID, linkID); err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	return githubshared.CoreIssueSyncLink{
		ID: row.ID, RepositoryID: row.RepositoryID, RepositoryName: row.RepositoryName, TeamID: row.TeamID,
		TeamName: row.TeamName, TeamColor: row.TeamColor, SyncDirection: row.SyncDirection,
		IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *Repo) UpdateIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID, input githubshared.CoreUpdateIssueSyncLinkInput) (githubshared.CoreIssueSyncLink, error) {
	current, err := r.GetIssueSyncLink(ctx, workspaceID, linkID)
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	syncDirection := current.SyncDirection
	isActive := current.IsActive
	if input.SyncDirection != nil {
		syncDirection = *input.SyncDirection
	}
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	query := `
		UPDATE github_issue_sync_links
		SET sync_direction = $3, is_active = $4, updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2
	`
	if _, err := r.db.ExecContext(ctx, query, workspaceID, linkID, syncDirection, isActive); err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	return r.GetIssueSyncLink(ctx, workspaceID, linkID)
}

func (r *Repo) DeleteIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM github_issue_sync_links WHERE workspace_id = $1 AND id = $2`, workspaceID, linkID)
	return err
}

func (r *Repo) GetTeamWorkflowSettings(ctx context.Context, workspaceID, teamID uuid.UUID) (githubshared.CoreTeamGitHubSettings, error) {
	var rows []workflowRuleRow
	query := `
		SELECT id, event_key, target_status_id, base_branch_pattern, is_active, created_at, updated_at
		FROM github_team_workflow_rules
		WHERE workspace_id = $1 AND team_id = $2
		ORDER BY created_at ASC
	`
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID, teamID); err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	items := make([]githubshared.CoreWorkflowRule, 0, len(rows))
	for _, row := range rows {
		items = append(items, githubshared.CoreWorkflowRule{
			ID: row.ID, EventKey: row.EventKey, TargetStatusID: row.TargetStatusID,
			BaseBranchPattern: row.BaseBranchPattern, IsActive: row.IsActive,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return githubshared.CoreTeamGitHubSettings{TeamID: teamID, Rules: items}, nil
}

func (r *Repo) ReplaceTeamWorkflowSettings(ctx context.Context, workspaceID, teamID uuid.UUID, rules []githubshared.CoreWorkflowRuleInput) (githubshared.CoreTeamGitHubSettings, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(ctx, `DELETE FROM github_team_workflow_rules WHERE workspace_id = $1 AND team_id = $2`, workspaceID, teamID); err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	for _, rule := range rules {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO github_team_workflow_rules (workspace_id, team_id, event_key, target_status_id, base_branch_pattern, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			workspaceID,
			teamID,
			rule.EventKey,
			rule.TargetStatusID,
			rule.BaseBranchPattern,
			rule.IsActive,
		); err != nil {
			return githubshared.CoreTeamGitHubSettings{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	return r.GetTeamWorkflowSettings(ctx, workspaceID, teamID)
}

func (r *Repo) ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]statusRow, error) {
	var rows []statusRow
	query := `SELECT status_id, name, category, color FROM statuses WHERE team_id = $1 ORDER BY order_index ASC`
	if err := r.db.SelectContext(ctx, &rows, query, teamID); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) FindRepositoryByExternalID(ctx context.Context, repositoryExternalID int64) (RepoByExternalRow, error) {
	var row RepoByExternalRow
	query := `
		SELECT
			gr.id,
			gr.workspace_id,
			w.slug AS workspace_slug,
			gr.full_name,
			gr.owner_login,
			gr.name AS repository_slug,
			gr.default_branch,
			gi.github_installation_id
		FROM github_repositories gr
		INNER JOIN github_installations gi ON gi.id = gr.installation_id
		INNER JOIN workspaces w ON w.workspace_id = gr.workspace_id
		WHERE gr.github_repository_id = $1 AND gr.is_active = true
	`
	err := r.db.GetContext(ctx, &row, query, repositoryExternalID)
	return row, err
}

func (r *Repo) FindIssueSyncLinkByRepositoryID(ctx context.Context, repositoryID uuid.UUID) (syncLinkRow, error) {
	var row syncLinkRow
	query := `
		SELECT l.id, l.repository_id, gr.full_name AS repository_name, l.team_id, t.name AS team_name, t.color AS team_color,
		       l.sync_direction, l.is_active, l.created_at, l.updated_at
		FROM github_issue_sync_links l
		INNER JOIN github_repositories gr ON gr.id = l.repository_id
		INNER JOIN teams t ON t.team_id = l.team_id
		WHERE l.repository_id = $1 AND l.is_active = true
	`
	err := r.db.GetContext(ctx, &row, query, repositoryID)
	return row, err
}

func (r *Repo) FindBidirectionalIssueSyncLinkByTeamID(ctx context.Context, workspaceID, teamID uuid.UUID) (bidirectionalLinkRow, error) {
	var row bidirectionalLinkRow
	query := `
		SELECT l.id, l.repository_id, l.team_id, l.sync_direction,
		       gr.full_name AS repository_name, gr.owner_login, gr.name AS repository_slug,
		       gr.html_url AS repository_html_url, gi.github_installation_id
		FROM github_issue_sync_links l
		INNER JOIN github_repositories gr ON gr.id = l.repository_id
		INNER JOIN github_installations gi ON gi.id = gr.installation_id
		WHERE l.workspace_id = $1
		  AND l.team_id = $2
		  AND l.is_active = true
		  AND l.sync_direction = 'bidirectional'
		  AND gr.is_active = true
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &row, query, workspaceID, teamID)
	return row, err
}

func (r *Repo) ResolveStoriesByRefs(ctx context.Context, workspaceID uuid.UUID, refs []string) ([]StoryMatch, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	matches := make([]StoryMatch, 0)
	seen := map[string]struct{}{}
	for _, ref := range refs {
		normalized := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(ref), "-", ""), " ", ""))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		letters := 0
		for i, ch := range normalized {
			if ch >= '0' && ch <= '9' {
				letters = i
				break
			}
		}
		if letters == 0 || letters == len(normalized) {
			continue
		}
		teamCode := normalized[:letters]
		sequenceID := normalized[letters:]
		var match StoryMatch
		query := `
			SELECT s.id, s.status_id, s.team_id, t.code AS team_code, s.sequence_id, s.title
			FROM stories s
			INNER JOIN teams t ON t.team_id = s.team_id
			WHERE s.workspace_id = $1 AND UPPER(t.code) = $2 AND s.sequence_id = $3
		`
		if err := r.db.GetContext(ctx, &match, query, workspaceID, teamCode, sequenceID); err == nil {
			matches = append(matches, match)
		}
	}
	return matches, nil
}

func (r *Repo) FindStoryLink(ctx context.Context, repositoryID uuid.UUID, externalType string, githubID int64, refName *string) (uuid.UUID, uuid.UUID, error) {
	var row struct {
		ID      uuid.UUID `db:"id"`
		StoryID uuid.UUID `db:"story_id"`
	}
	query := `
		SELECT id, story_id
		FROM github_story_links
		WHERE repository_id = $1 AND external_type = $2
		  AND COALESCE(github_id, 0) = $3
		  AND COALESCE(ref_name, '') = COALESCE($4, '')
	`
	err := r.db.GetContext(ctx, &row, query, repositoryID, externalType, githubID, refName)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return row.ID, row.StoryID, nil
}

func (r *Repo) FindIssueStoryLinkByStoryID(ctx context.Context, storyID, repositoryID uuid.UUID) (issueStoryLinkRow, error) {
	var row issueStoryLinkRow
	query := `
		SELECT id, story_id, repository_id, github_id, github_number, url, title, state
		FROM github_story_links
		WHERE story_id = $1
		  AND repository_id = $2
		  AND external_type = 'issue'
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &row, query, storyID, repositoryID)
	return row, err
}

func (r *Repo) CreateOrUpdateExternalStory(ctx context.Context, workspaceID, teamID, reporterID, repositoryID uuid.UUID, title, description string, externalType string, githubID int64, githubNumber int, url string) (uuid.UUID, error) {
	if _, storyID, err := r.FindStoryLink(ctx, repositoryID, externalType, githubID, nil); err == nil {
		updateQuery := `UPDATE stories SET title = $2, description = $3, description_html = $3, updated_at = NOW() WHERE id = $1`
		if _, err := r.db.ExecContext(ctx, updateQuery, storyID, title, description); err != nil {
			return uuid.Nil, err
		}
		return storyID, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var statusID uuid.UUID
	if err := tx.GetContext(ctx, &statusID, `SELECT status_id FROM statuses WHERE team_id = $1 AND category = 'unstarted' ORDER BY order_index ASC LIMIT 1`, teamID); err != nil {
		return uuid.Nil, err
	}

	var sequenceID int
	if err := tx.GetContext(
		ctx,
		&sequenceID,
		`INSERT INTO team_story_sequences (workspace_id, team_id, current_sequence)
		 VALUES ($1, $2, 0)
		 ON CONFLICT (workspace_id, team_id)
		 DO UPDATE SET current_sequence = team_story_sequences.current_sequence + 1
		 RETURNING current_sequence`,
		workspaceID,
		teamID,
	); err != nil {
		return uuid.Nil, err
	}

	var storyID uuid.UUID
	insertStoryQuery := `
		INSERT INTO stories (
			sequence_id, title, description, description_html, status_id, priority, estimate_unit,
			team_id, workspace_id, reporter_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $3, $4, 'No Priority', NULL,
			$5, $6, $7, NOW(), NOW()
		)
		RETURNING id
	`
	if err := tx.GetContext(ctx, &storyID, insertStoryQuery, sequenceID+1, title, description, statusID, teamID, workspaceID, reporterID); err != nil {
		return uuid.Nil, err
	}

	linkQuery := `
		INSERT INTO github_story_links (
			workspace_id, story_id, repository_id, external_type, github_id, github_number, url, title, state
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'open')
	`
	if _, err := tx.ExecContext(ctx, linkQuery, workspaceID, storyID, repositoryID, externalType, githubID, githubNumber, url, title); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return storyID, nil
}

func (r *Repo) UpsertStoryLink(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID, externalType string, githubID int64, githubNumber int, refName *string, url, title, state string, metadata any) error {
	metadataJSON, _ := json.Marshal(metadata)
	if linkID, _, err := r.FindStoryLink(ctx, repositoryID, externalType, githubID, refName); err == nil {
		_, err = r.db.ExecContext(
			ctx,
			`UPDATE github_story_links
			 SET url = $2, title = $3, state = $4, metadata = $5, last_seen_at = NOW(), updated_at = NOW()
			 WHERE id = $1`,
			linkID,
			url,
			title,
			state,
			metadataJSON,
		)
		return err
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO github_story_links (
			workspace_id, story_id, repository_id, external_type, github_id, github_number, ref_name, url, title, state, metadata, last_seen_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())`,
		workspaceID,
		storyID,
		repositoryID,
		externalType,
		githubID,
		githubNumber,
		refName,
		url,
		title,
		state,
		metadataJSON,
	)
	return err
}

func (r *Repo) MoveStoryToStatus(ctx context.Context, storyID, statusID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE stories SET status_id = $2, updated_at = NOW() WHERE id = $1`, storyID, statusID)
	return err
}

func (r *Repo) GetStatusCategory(ctx context.Context, statusID uuid.UUID) (string, error) {
	var category string
	err := r.db.GetContext(ctx, &category, `SELECT category FROM statuses WHERE status_id = $1`, statusID)
	return category, err
}

func (r *Repo) UpdateStoryDescription(ctx context.Context, storyID uuid.UUID, description *string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE stories
		 SET description = $2, description_html = $2, updated_at = NOW()
		 WHERE id = $1`,
		storyID,
		description,
	)
	return err
}

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	var statusID uuid.UUID
	err := r.db.GetContext(ctx, &statusID, `SELECT status_id FROM statuses WHERE team_id = $1 AND category = $2 ORDER BY order_index ASC LIMIT 1`, teamID, category)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &statusID, nil
}

func (r *Repo) RecordWebhookEvent(ctx context.Context, deliveryID, eventName, action string, installationExternalID, repositoryExternalID, senderExternalID *int64, payload []byte) (bool, error) {
	query := `
		INSERT INTO github_webhook_events (delivery_id, event_name, action, installation_external_id, repository_external_id, sender_external_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		ON CONFLICT (delivery_id) DO NOTHING
	`
	result, err := r.db.ExecContext(ctx, query, deliveryID, eventName, action, installationExternalID, repositoryExternalID, senderExternalID, string(payload))
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *Repo) MarkWebhookProcessed(ctx context.Context, deliveryID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE github_webhook_events SET processing_state = 'processed', processed_at = NOW(), attempts = attempts + 1 WHERE delivery_id = $1`, deliveryID)
	return err
}

func (r *Repo) MarkWebhookFailed(ctx context.Context, deliveryID string, errMessage string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE github_webhook_events SET processing_state = 'failed', processed_at = NOW(), attempts = attempts + 1, error_message = $2 WHERE delivery_id = $1`, deliveryID, errMessage)
	return err
}

func toCoreWorkspaceSettings(row workspaceSettingsRow) githubshared.CoreWorkspaceSettings {
	return githubshared.CoreWorkspaceSettings{
		WorkspaceID:             row.WorkspaceID,
		BranchFormat:            row.BranchFormat,
		LinkCommitsByMagicWords: row.LinkCommitsByMagicWords,
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}

type GithubInstallationPayload struct {
	ID                  int64                            `json:"id"`
	Account             GithubInstallationAccountPayload `json:"account"`
	RepositorySelection string                           `json:"repository_selection"`
	Permissions         map[string]string                `json:"permissions"`
	Events              []string                         `json:"events"`
	Sender              GithubInstallationSenderPayload  `json:"sender"`
}

type GithubInstallationAccountPayload struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Type      string  `json:"type"`
	AvatarURL *string `json:"avatar_url"`
}

type GithubInstallationSenderPayload struct {
	ID int64 `json:"id"`
}

type GithubRepositoryPayload struct {
	ID            int64                        `json:"id"`
	Name          string                       `json:"name"`
	FullName      string                       `json:"full_name"`
	Description   *string                      `json:"description"`
	HTMLURL       string                       `json:"html_url"`
	CloneURL      string                       `json:"clone_url"`
	SSHURL        string                       `json:"ssh_url"`
	DefaultBranch string                       `json:"default_branch"`
	Private       bool                         `json:"private"`
	Archived      bool                         `json:"archived"`
	Disabled      bool                         `json:"disabled"`
	Owner         GithubRepositoryOwnerPayload `json:"owner"`
}

type GithubRepositoryOwnerPayload struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

func (r *Repo) EnsureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	ctx, _ = web.AddSpan(ctx, "business.repository.github")
	return ctx
}
