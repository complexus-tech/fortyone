package figmarepository

import (
	"context"
	"encoding/json"
	"time"

	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repo struct{ db *sqlx.DB }

func New(db *sqlx.DB) *Repo { return &Repo{db: db} }

func (r *Repo) SaveOAuthState(ctx context.Context, state figma.OAuthState) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO figma_oauth_states (state_hash, workspace_id, user_id, workspace_slug, code_verifier, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, state.StateHash, state.WorkspaceID, state.UserID, state.WorkspaceSlug, state.CodeVerifier, state.ExpiresAt)
	return err
}

func (r *Repo) ConsumeOAuthState(ctx context.Context, stateHash string, now time.Time) (figma.OAuthState, error) {
	var row struct {
		StateHash     string    `db:"state_hash"`
		WorkspaceID   uuid.UUID `db:"workspace_id"`
		UserID        uuid.UUID `db:"user_id"`
		WorkspaceSlug string    `db:"workspace_slug"`
		CodeVerifier  string    `db:"code_verifier"`
		ExpiresAt     time.Time `db:"expires_at"`
	}
	err := r.db.GetContext(ctx, &row, `
		UPDATE figma_oauth_states SET consumed_at = $2
		WHERE state_hash = $1 AND consumed_at IS NULL AND expires_at > $2
		RETURNING state_hash, workspace_id, user_id, workspace_slug, code_verifier, expires_at
	`, stateHash, now)
	return figma.OAuthState{StateHash: row.StateHash, WorkspaceID: row.WorkspaceID, UserID: row.UserID, WorkspaceSlug: row.WorkspaceSlug, CodeVerifier: row.CodeVerifier, ExpiresAt: row.ExpiresAt}, err
}

func (r *Repo) UpsertConnection(ctx context.Context, connection figma.Connection) (figma.Connection, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return figma.Connection{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE figma_webhooks SET is_active=false,updated_at=now() WHERE connection_id IN (SELECT id FROM figma_connections WHERE workspace_id=$1 AND is_active)`, connection.WorkspaceID); err != nil {
		return figma.Connection{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE figma_connections SET is_active=false,disconnected_at=now(),updated_at=now() WHERE workspace_id=$1 AND is_active`, connection.WorkspaceID); err != nil {
		return figma.Connection{}, err
	}
	var row connectionRow
	err = tx.GetContext(ctx, &row, `
		INSERT INTO figma_connections (workspace_id, figma_user_id, figma_email, figma_handle, token_payload, scopes, expires_at, connected_by_user_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, workspace_id, figma_user_id, figma_email, figma_handle, token_payload, scopes, expires_at, connected_by_user_id, is_active, created_at, updated_at
	`, connection.WorkspaceID, connection.FigmaUserID, connection.Email, connection.Handle, connection.TokenPayload, pq.StringArray(connection.Scopes), connection.ExpiresAt, connection.ConnectedByUserID)
	if err != nil {
		return figma.Connection{}, err
	}
	if err := tx.Commit(); err != nil {
		return figma.Connection{}, err
	}
	return row.core(), nil
}

func (r *Repo) GetConnection(ctx context.Context, workspaceID uuid.UUID) (figma.Connection, error) {
	var row connectionRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, figma_user_id, figma_email, figma_handle, token_payload, scopes, expires_at, connected_by_user_id, is_active, created_at, updated_at
		FROM figma_connections WHERE workspace_id = $1 AND is_active
	`, workspaceID)
	return row.core(), err
}

func (r *Repo) UpdateConnectionToken(ctx context.Context, id uuid.UUID, payload string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE figma_connections SET token_payload=$2, expires_at=$3, updated_at=now() WHERE id=$1 AND is_active`, id, payload, expiresAt)
	return err
}

func (r *Repo) Disconnect(ctx context.Context, workspaceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE figma_connections SET is_active=false, disconnected_at=now(), updated_at=now() WHERE workspace_id=$1 AND is_active`, workspaceID)
	return err
}

func (r *Repo) ListStoryLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]figma.StoryLink, error) {
	var rows []storyLinkRow
	err := r.db.SelectContext(ctx, &rows, storyLinkSelect+` WHERE workspace_id=$1 AND story_id=$2 ORDER BY created_at`, workspaceID, storyID)
	return coreLinks(rows), err
}

func (r *Repo) ListLinksByFile(ctx context.Context, workspaceID uuid.UUID, fileKey string) ([]figma.StoryLink, error) {
	var rows []storyLinkRow
	err := r.db.SelectContext(ctx, &rows, storyLinkSelect+` WHERE workspace_id=$1 AND file_key=$2 ORDER BY created_at`, workspaceID, fileKey)
	return coreLinks(rows), err
}

func (r *Repo) UpsertStoryLink(ctx context.Context, link figma.StoryLink) (figma.StoryLink, error) {
	metadata := []byte(link.Artifact.Metadata)
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return figma.StoryLink{}, err
	}
	defer tx.Rollback()
	var storyLinkID uuid.UUID
	externalKey := "figma:" + link.StoryID.String() + ":" + link.Artifact.FileKey + ":"
	if link.Artifact.NodeID != nil {
		externalKey += *link.Artifact.NodeID
	}
	if err := tx.GetContext(ctx, &storyLinkID, `
		INSERT INTO story_links (title,url,story_id,external_source_key) VALUES ($1,$2,$3,$4)
		ON CONFLICT (external_source_key) WHERE external_source_key IS NOT NULL
		DO UPDATE SET title=excluded.title,url=excluded.url,updated_at=now()
		RETURNING link_id
	`, link.Artifact.NodeName, link.Artifact.CanonicalURL, link.StoryID, externalKey); err != nil {
		return figma.StoryLink{}, err
	}
	var row storyLinkRow
	err = tx.GetContext(ctx, &row, `
		INSERT INTO story_figma_links (workspace_id,story_id,created_by_user_id,story_link_id,file_key,node_id,original_url,canonical_url,file_name,node_name,node_type,thumbnail_url,version,last_modified,metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (story_id,file_key,COALESCE(node_id,'')) DO UPDATE SET
			story_link_id=excluded.story_link_id, original_url=excluded.original_url, canonical_url=excluded.canonical_url,
			file_name=excluded.file_name,node_name=excluded.node_name,node_type=excluded.node_type,thumbnail_url=excluded.thumbnail_url,
			version=excluded.version,last_modified=excluded.last_modified,metadata=excluded.metadata,last_synced_at=now(),unavailable_at=NULL,updated_at=now()
		RETURNING `+storyLinkColumns, link.WorkspaceID, link.StoryID, link.CreatedByUserID, storyLinkID, link.Artifact.FileKey, link.Artifact.NodeID, link.Artifact.OriginalURL, link.Artifact.CanonicalURL, link.Artifact.FileName, link.Artifact.NodeName, link.Artifact.NodeType, link.Artifact.ThumbnailURL, link.Artifact.Version, link.Artifact.LastModified, metadata)
	if err != nil {
		return figma.StoryLink{}, err
	}
	if err := tx.Commit(); err != nil {
		return figma.StoryLink{}, err
	}
	return row.core(), nil
}

func (r *Repo) UpdateStoryLink(ctx context.Context, link figma.StoryLink) error {
	metadata := []byte(link.Artifact.Metadata)
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE story_figma_links SET file_name=$3,node_name=$4,node_type=$5,thumbnail_url=$6,version=$7,last_modified=$8,
			dev_status=$9,dev_resource_id=$10,metadata=$11,last_synced_at=now(),unavailable_at=$12,updated_at=now()
		WHERE id=$1 AND workspace_id=$2
	`, link.ID, link.WorkspaceID, link.Artifact.FileName, link.Artifact.NodeName, link.Artifact.NodeType, link.Artifact.ThumbnailURL, link.Artifact.Version, link.Artifact.LastModified, link.DevStatus, link.DevResourceID, metadata, link.UnavailableAt)
	return err
}

func (r *Repo) GetStoryLink(ctx context.Context, workspaceID, linkID uuid.UUID) (figma.StoryLink, error) {
	var row storyLinkRow
	err := r.db.GetContext(ctx, &row, storyLinkSelect+` WHERE workspace_id=$1 AND id=$2`, workspaceID, linkID)
	return row.core(), err
}

func (r *Repo) DeleteStoryLink(ctx context.Context, workspaceID, linkID uuid.UUID) (figma.StoryLink, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return figma.StoryLink{}, err
	}
	defer tx.Rollback()
	var row storyLinkRow
	if err := tx.GetContext(ctx, &row, `DELETE FROM story_figma_links WHERE id=$1 AND workspace_id=$2 RETURNING `+storyLinkColumns, linkID, workspaceID); err != nil {
		return figma.StoryLink{}, err
	}
	if row.StoryLinkID != nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM story_links WHERE link_id=$1`, *row.StoryLinkID); err != nil {
			return figma.StoryLink{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return figma.StoryLink{}, err
	}
	return row.core(), nil
}

func (r *Repo) SaveWebhook(ctx context.Context, webhook figma.Webhook) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO figma_webhooks (connection_id,file_key,event_type,figma_webhook_id,passcode_hash)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (connection_id,file_key,event_type) DO UPDATE SET figma_webhook_id=excluded.figma_webhook_id,passcode_hash=excluded.passcode_hash,is_active=true,updated_at=now()
	`, webhook.ConnectionID, webhook.FileKey, webhook.EventType, webhook.FigmaWebhookID, webhook.PasscodeHash)
	return err
}

func (r *Repo) GetWebhook(ctx context.Context, id int64) (figma.Webhook, error) {
	var row struct {
		ID             uuid.UUID `db:"id"`
		ConnectionID   uuid.UUID `db:"connection_id"`
		WorkspaceID    uuid.UUID `db:"workspace_id"`
		FileKey        string    `db:"file_key"`
		EventType      string    `db:"event_type"`
		FigmaWebhookID int64     `db:"figma_webhook_id"`
		PasscodeHash   string    `db:"passcode_hash"`
		IsActive       bool      `db:"is_active"`
	}
	err := r.db.GetContext(ctx, &row, `
		SELECT w.id,w.connection_id,c.workspace_id,w.file_key,w.event_type,w.figma_webhook_id,w.passcode_hash,w.is_active
		FROM figma_webhooks w
		JOIN figma_connections c ON c.id=w.connection_id AND c.is_active
		WHERE w.figma_webhook_id=$1 AND w.is_active
	`, id)
	return figma.Webhook{ID: row.ID, ConnectionID: row.ConnectionID, WorkspaceID: row.WorkspaceID, FileKey: row.FileKey, EventType: row.EventType, FigmaWebhookID: row.FigmaWebhookID, PasscodeHash: row.PasscodeHash, IsActive: row.IsActive}, err
}

func (r *Repo) RecordWebhookEvent(ctx context.Context, eventKey string, event figma.WebhookEvent) (bool, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return false, err
	}
	result, err := r.db.ExecContext(ctx, `INSERT INTO figma_webhook_events (figma_webhook_id,event_type,event_key,payload,processed_at) VALUES ($1,$2,$3,$4,now()) ON CONFLICT (event_key) DO NOTHING`, int64(event.WebhookID), event.EventType, eventKey, payload)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (r *Repo) FindWebhook(ctx context.Context, connectionID uuid.UUID, fileKey, eventType string) (figma.Webhook, error) {
	var row webhookRow
	err := r.db.GetContext(ctx, &row, `SELECT id,connection_id,file_key,event_type,figma_webhook_id,passcode_hash,is_active FROM figma_webhooks WHERE connection_id=$1 AND file_key=$2 AND event_type=$3 AND is_active`, connectionID, fileKey, eventType)
	return row.core(), err
}

func (r *Repo) ListWebhooks(ctx context.Context, connectionID uuid.UUID) ([]figma.Webhook, error) {
	var rows []webhookRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT id,connection_id,file_key,event_type,figma_webhook_id,passcode_hash,is_active FROM figma_webhooks WHERE connection_id=$1 AND is_active`, connectionID); err != nil {
		return nil, err
	}
	webhooks := make([]figma.Webhook, 0, len(rows))
	for _, row := range rows {
		webhooks = append(webhooks, row.core())
	}
	return webhooks, nil
}

func (r *Repo) DeactivateWebhook(ctx context.Context, figmaWebhookID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE figma_webhooks SET is_active=false,updated_at=now() WHERE figma_webhook_id=$1`, figmaWebhookID)
	return err
}

type connectionRow struct {
	ID                uuid.UUID      `db:"id"`
	WorkspaceID       uuid.UUID      `db:"workspace_id"`
	FigmaUserID       string         `db:"figma_user_id"`
	Email             *string        `db:"figma_email"`
	Handle            *string        `db:"figma_handle"`
	TokenPayload      string         `db:"token_payload"`
	Scopes            pq.StringArray `db:"scopes"`
	ExpiresAt         time.Time      `db:"expires_at"`
	ConnectedByUserID uuid.UUID      `db:"connected_by_user_id"`
	IsActive          bool           `db:"is_active"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
}

func (r connectionRow) core() figma.Connection {
	return figma.Connection{ID: r.ID, WorkspaceID: r.WorkspaceID, FigmaUserID: r.FigmaUserID, Email: r.Email, Handle: r.Handle, TokenPayload: r.TokenPayload, Scopes: []string(r.Scopes), ExpiresAt: r.ExpiresAt, ConnectedByUserID: r.ConnectedByUserID, IsActive: r.IsActive, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

const storyLinkColumns = `id,workspace_id,story_id,created_by_user_id,story_link_id,file_key,node_id,original_url,canonical_url,file_name,node_name,node_type,thumbnail_url,version,last_modified,dev_status,dev_resource_id,metadata,last_synced_at,unavailable_at,created_at,updated_at`
const storyLinkSelect = `SELECT ` + storyLinkColumns + ` FROM story_figma_links`

type storyLinkRow struct {
	ID              uuid.UUID  `db:"id"`
	WorkspaceID     uuid.UUID  `db:"workspace_id"`
	StoryID         uuid.UUID  `db:"story_id"`
	CreatedByUserID uuid.UUID  `db:"created_by_user_id"`
	StoryLinkID     *uuid.UUID `db:"story_link_id"`
	FileKey         string     `db:"file_key"`
	NodeID          *string    `db:"node_id"`
	OriginalURL     string     `db:"original_url"`
	CanonicalURL    string     `db:"canonical_url"`
	FileName        string     `db:"file_name"`
	NodeName        *string    `db:"node_name"`
	NodeType        *string    `db:"node_type"`
	ThumbnailURL    *string    `db:"thumbnail_url"`
	Version         *string    `db:"version"`
	LastModified    *time.Time `db:"last_modified"`
	DevStatus       *string    `db:"dev_status"`
	DevResourceID   *string    `db:"dev_resource_id"`
	Metadata        []byte     `db:"metadata"`
	LastSyncedAt    time.Time  `db:"last_synced_at"`
	UnavailableAt   *time.Time `db:"unavailable_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

func (r storyLinkRow) core() figma.StoryLink {
	return figma.StoryLink{ID: r.ID, WorkspaceID: r.WorkspaceID, StoryID: r.StoryID, CreatedByUserID: r.CreatedByUserID, StoryLinkID: r.StoryLinkID, Artifact: figma.Artifact{FileKey: r.FileKey, NodeID: r.NodeID, OriginalURL: r.OriginalURL, CanonicalURL: r.CanonicalURL, FileName: r.FileName, NodeName: r.NodeName, NodeType: r.NodeType, ThumbnailURL: r.ThumbnailURL, Version: r.Version, LastModified: r.LastModified, Metadata: r.Metadata}, DevStatus: r.DevStatus, DevResourceID: r.DevResourceID, LastSyncedAt: r.LastSyncedAt, UnavailableAt: r.UnavailableAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
func coreLinks(rows []storyLinkRow) []figma.StoryLink {
	links := make([]figma.StoryLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, row.core())
	}
	return links
}

var _ figma.Repository = (*Repo)(nil)

type webhookRow struct {
	ID             uuid.UUID `db:"id"`
	ConnectionID   uuid.UUID `db:"connection_id"`
	FileKey        string    `db:"file_key"`
	EventType      string    `db:"event_type"`
	FigmaWebhookID int64     `db:"figma_webhook_id"`
	PasscodeHash   string    `db:"passcode_hash"`
	IsActive       bool      `db:"is_active"`
}

func (r webhookRow) core() figma.Webhook {
	return figma.Webhook{ID: r.ID, ConnectionID: r.ConnectionID, FileKey: r.FileKey, EventType: r.EventType, FigmaWebhookID: r.FigmaWebhookID, PasscodeHash: r.PasscodeHash, IsActive: r.IsActive}
}
