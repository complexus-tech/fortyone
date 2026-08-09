package integrationrequestsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type providerThreadRow struct {
	ID                      uuid.UUID  `db:"id"`
	WorkspaceID             uuid.UUID  `db:"workspace_id"`
	IntegrationRequestID    uuid.UUID  `db:"integration_request_id"`
	TeamID                  uuid.UUID  `db:"team_id"`
	AcceptedStoryID         *uuid.UUID `db:"accepted_story_id"`
	Provider                string     `db:"provider"`
	ExternalWorkspaceID     string     `db:"external_workspace_id"`
	InstallationGeneration  *uuid.UUID `db:"installation_generation"`
	ExternalChannelID       string     `db:"external_channel_id"`
	ExternalThreadID        string     `db:"external_thread_id"`
	ExternalSourceMessageID *string    `db:"external_source_message_id"`
	SourceURL               *string    `db:"source_url"`
	RequestTitle            string     `db:"request_title"`
	CreatedAt               time.Time  `db:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at"`
}

type commentRow struct {
	ID                     uuid.UUID  `db:"id"`
	WorkspaceID            uuid.UUID  `db:"workspace_id"`
	ThreadID               uuid.UUID  `db:"thread_id"`
	Direction              string     `db:"direction"`
	AuthorUserID           *uuid.UUID `db:"author_user_id"`
	AuthorName             string     `db:"author_name"`
	AuthorAvatar           *string    `db:"author_avatar"`
	ExternalAuthorID       *string    `db:"external_author_id"`
	ExternalMessageID      *string    `db:"external_message_id"`
	ClientIdempotencyKey   *uuid.UUID `db:"client_idempotency_key"`
	OutboundIdempotencyKey *string    `db:"outbound_idempotency_key"`
	DeliveryStatus         *string    `db:"delivery_status"`
	Body                   string     `db:"body"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
}

const providerThreadSelect = `
	SELECT irt.id, irt.workspace_id, irt.integration_request_id,
	       ir.team_id, ir.accepted_story_id, irt.provider,
	       irt.external_workspace_id, irt.installation_generation,
	       irt.external_channel_id, irt.external_thread_id,
	       irt.external_source_message_id, COALESCE(irt.source_url, ir.source_url) AS source_url,
	       ir.title AS request_title, irt.created_at, irt.updated_at
	FROM integration_request_threads irt
	JOIN integration_requests ir ON ir.id = irt.integration_request_id
`

func (r *Repo) BindProviderThread(ctx context.Context, input integrationrequests.CoreBindProviderThreadInput) (integrationrequests.CoreProviderThread, error) {
	var row providerThreadRow
	err := r.db.GetContext(ctx, &row, `
		WITH bound AS (
			INSERT INTO integration_request_threads (
				workspace_id, integration_request_id, provider,
				external_workspace_id, installation_generation,
				external_channel_id, external_thread_id,
				external_source_message_id, source_url
			)
			SELECT $1, ir.id, $3, $4, $5, $6, $7, NULLIF($8, ''), $9
			FROM integration_requests ir
			WHERE ir.id = $2 AND ir.workspace_id = $1 AND ir.provider = $3
			ON CONFLICT (integration_request_id, provider) DO UPDATE SET
				external_workspace_id = EXCLUDED.external_workspace_id,
				installation_generation = EXCLUDED.installation_generation,
				external_channel_id = EXCLUDED.external_channel_id,
				external_thread_id = EXCLUDED.external_thread_id,
				external_source_message_id = COALESCE(EXCLUDED.external_source_message_id, integration_request_threads.external_source_message_id),
				source_url = COALESCE(EXCLUDED.source_url, integration_request_threads.source_url),
				updated_at = NOW()
			RETURNING *
		)
		SELECT bound.id, bound.workspace_id, bound.integration_request_id,
		       ir.team_id, ir.accepted_story_id, bound.provider,
		       bound.external_workspace_id, bound.installation_generation,
		       bound.external_channel_id, bound.external_thread_id,
		       bound.external_source_message_id, COALESCE(bound.source_url, ir.source_url) AS source_url,
		       ir.title AS request_title, bound.created_at, bound.updated_at
		FROM bound
		JOIN integration_requests ir ON ir.id = bound.integration_request_id
	`, input.WorkspaceID, input.IntegrationRequestID, strings.TrimSpace(input.Provider), strings.TrimSpace(input.ExternalWorkspaceID), input.InstallationGeneration, strings.TrimSpace(input.ExternalChannelID), strings.TrimSpace(input.ExternalThreadID), optionalString(input.ExternalSourceMessageID), input.SourceURL)
	if err != nil {
		return integrationrequests.CoreProviderThread{}, err
	}
	return toCoreProviderThread(row), nil
}

func (r *Repo) HasAuthorizedProviderThread(ctx context.Context, input integrationrequests.CoreProviderThreadMatchInput) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1
			FROM integration_request_threads irt
			JOIN integration_requests ir ON ir.id = irt.integration_request_id
			WHERE irt.workspace_id = $1
			  AND irt.provider = $2
			  AND irt.external_workspace_id = $3
			  AND irt.installation_generation = $4
			  AND irt.external_channel_id = $5
			  AND irt.external_thread_id = $6
			  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$7")+`
		)
	`, input.WorkspaceID, strings.TrimSpace(input.Provider), strings.TrimSpace(input.ExternalWorkspaceID), input.InstallationGeneration, strings.TrimSpace(input.ExternalChannelID), strings.TrimSpace(input.ExternalThreadID), input.UserID)
	return exists, err
}

func (r *Repo) HasCurrentProviderThread(ctx context.Context, input integrationrequests.CoreProviderThreadLookupInput) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1
			FROM integration_request_threads irt
			WHERE irt.workspace_id = $1
			  AND irt.provider = $2
			  AND irt.external_workspace_id = $3
			  AND irt.installation_generation = $4
			  AND irt.external_channel_id = $5
			  AND irt.external_thread_id = $6
		)
	`, input.WorkspaceID, strings.TrimSpace(input.Provider), strings.TrimSpace(input.ExternalWorkspaceID), input.InstallationGeneration, strings.TrimSpace(input.ExternalChannelID), strings.TrimSpace(input.ExternalThreadID))
	return exists, err
}

func (r *Repo) FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (integrationrequests.CoreProviderThread, error) {
	var row providerThreadRow
	err := r.db.GetContext(ctx, &row, providerThreadSelect+`
		WHERE irt.workspace_id = $1
		  AND irt.integration_request_id = $2
		  AND irt.provider = $3
	`, workspaceID, requestID, strings.TrimSpace(provider))
	if errors.Is(err, sql.ErrNoRows) {
		return integrationrequests.CoreProviderThread{}, integrationrequests.ErrProviderThreadNotFound
	}
	if err != nil {
		return integrationrequests.CoreProviderThread{}, err
	}
	return toCoreProviderThread(row), nil
}

func (r *Repo) GetThreadActivityForRequest(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequests.CoreThreadActivity, error) {
	var thread providerThreadRow
	err := r.db.GetContext(ctx, &thread, providerThreadSelect+`
		WHERE irt.workspace_id = $1
		  AND irt.integration_request_id = $2
		  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$3")+`
	`, workspaceID, requestID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return integrationrequests.CoreThreadActivity{}, integrationrequests.ErrProviderThreadNotFound
	}
	if err != nil {
		return integrationrequests.CoreThreadActivity{}, err
	}
	comments, err := r.listThreadComments(ctx, thread.ID)
	if err != nil {
		return integrationrequests.CoreThreadActivity{}, err
	}
	return integrationrequests.CoreThreadActivity{Thread: toCoreProviderThread(thread), Comments: comments}, nil
}

func (r *Repo) ListProviderThreadsForStory(ctx context.Context, workspaceID, storyID, userID uuid.UUID) ([]integrationrequests.CoreProviderThread, error) {
	rows := make([]providerThreadRow, 0)
	err := r.db.SelectContext(ctx, &rows, providerThreadSelect+`
		WHERE irt.workspace_id = $1
		  AND ir.accepted_story_id = $2
		  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$3")+`
		ORDER BY irt.created_at ASC, irt.id ASC
	`, workspaceID, storyID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]integrationrequests.CoreProviderThread, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreProviderThread(row))
	}
	return result, nil
}

func (r *Repo) CreateOutboundComment(ctx context.Context, input integrationrequests.CoreCreateCommentInput, prepared integrationrequests.CorePreparedProviderComment) (integrationrequests.CoreProviderThread, integrationrequests.CoreIntegrationRequestComment, error) {
	providerPayload := strings.TrimSpace(string(prepared.ProviderPayload))
	if providerPayload != "" && !json.Valid([]byte(providerPayload)) {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, errors.New("provider comment payload must be valid JSON")
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var existing commentRow
	err = tx.GetContext(ctx, &existing, commentSelect+`
		WHERE irc.workspace_id = $1
		  AND irc.client_idempotency_key = $2
	`, input.WorkspaceID, input.ClientIdempotencyKey)
	if err == nil {
		var thread providerThreadRow
		threadErr := tx.GetContext(ctx, &thread, providerThreadSelect+`
			JOIN integration_request_comments irc ON irc.thread_id = irt.id
			WHERE irc.id = $1
			  AND irt.integration_request_id = $2
			  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$3")+`
		`, existing.ID, input.RequestID, input.AuthorID)
		if threadErr != nil || existing.AuthorUserID == nil || *existing.AuthorUserID != input.AuthorID || existing.Body != strings.TrimSpace(input.Body) {
			return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, integrationrequests.ErrIdempotencyConflict
		}
		if err := tx.Commit(); err != nil {
			return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
		}
		return toCoreProviderThread(thread), toCoreComment(existing), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
	}

	var thread providerThreadRow
	err = tx.GetContext(ctx, &thread, providerThreadSelect+`
		WHERE irt.workspace_id = $1
		  AND irt.integration_request_id = $2
		  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$3")+`
		FOR UPDATE OF irt
	`, input.WorkspaceID, input.RequestID, input.AuthorID)
	if errors.Is(err, sql.ErrNoRows) {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, integrationrequests.ErrProviderThreadNotFound
	}
	if err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
	}

	commentID := uuid.New()
	idempotencyKey := "integration-request-comment:" + input.ClientIdempotencyKey.String()
	var comment commentRow
	err = tx.GetContext(ctx, &comment, `
		INSERT INTO integration_request_comments (
			id, workspace_id, thread_id, direction, author_user_id,
			client_idempotency_key, outbound_idempotency_key, delivery_status, body
		) VALUES ($1, $2, $3, 'outbound', $4, $5, $6, 'sending', $7)
		ON CONFLICT (workspace_id, client_idempotency_key)
		WHERE client_idempotency_key IS NOT NULL
		DO NOTHING
		RETURNING id, workspace_id, thread_id, direction, author_user_id,
		          COALESCE((SELECT COALESCE(full_name, email) FROM users WHERE user_id = $4), 'FortyOne user') AS author_name,
		          (SELECT avatar_url FROM users WHERE user_id = $4) AS author_avatar,
		          external_author_id, external_message_id, client_idempotency_key,
		          outbound_idempotency_key, delivery_status, body, created_at, updated_at
	`, commentID, input.WorkspaceID, thread.ID, input.AuthorID, input.ClientIdempotencyKey, idempotencyKey, strings.TrimSpace(input.Body))
	if errors.Is(err, sql.ErrNoRows) {
		if existingThread, existingComment, conflictErr := r.resolveOutboundCommentConflict(ctx, tx, input); conflictErr == nil {
			if commitErr := tx.Commit(); commitErr != nil {
				return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, commitErr
			}
			return existingThread, existingComment, nil
		}
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, integrationrequests.ErrIdempotencyConflict
	}
	if err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO messaging_outbound_deliveries (
			provider, workspace_id, user_id, installation_generation,
			external_workspace_id, external_recipient_user_id, idempotency_key,
			external_channel_id, external_thread_id,
			content, provider_payload, purpose, status, attempt_count
		) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10,
		          CAST(NULLIF($11, '') AS jsonb), 'provider_message', 'pending', 0)
	`, thread.Provider, input.WorkspaceID, input.AuthorID, thread.InstallationGeneration, thread.ExternalWorkspaceID, strings.TrimSpace(prepared.ExternalRecipientUserID), idempotencyKey, thread.ExternalChannelID, thread.ExternalThreadID, strings.TrimSpace(input.Body), providerPayload)
	if err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
	}
	if err := tx.Commit(); err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, err
	}
	return toCoreProviderThread(thread), toCoreComment(comment), nil
}

func (r *Repo) resolveOutboundCommentConflict(ctx context.Context, tx *sqlx.Tx, input integrationrequests.CoreCreateCommentInput) (integrationrequests.CoreProviderThread, integrationrequests.CoreIntegrationRequestComment, error) {
	var comment commentRow
	if err := tx.GetContext(ctx, &comment, commentSelect+`
		WHERE irc.workspace_id = $1
		  AND irc.client_idempotency_key = $2
	`, input.WorkspaceID, input.ClientIdempotencyKey); err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, integrationrequests.ErrIdempotencyConflict
	}
	if comment.AuthorUserID == nil || *comment.AuthorUserID != input.AuthorID || comment.Body != strings.TrimSpace(input.Body) {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, integrationrequests.ErrIdempotencyConflict
	}
	var thread providerThreadRow
	if err := tx.GetContext(ctx, &thread, providerThreadSelect+`
		JOIN integration_request_comments irc ON irc.thread_id = irt.id
		WHERE irc.id = $1
		  AND irt.integration_request_id = $2
		  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$3")+`
	`, comment.ID, input.RequestID, input.AuthorID); err != nil {
		return integrationrequests.CoreProviderThread{}, integrationrequests.CoreIntegrationRequestComment{}, integrationrequests.ErrIdempotencyConflict
	}
	return toCoreProviderThread(thread), toCoreComment(comment), nil
}

func (r *Repo) IngestInboundProviderComment(ctx context.Context, input integrationrequests.CoreInboundProviderCommentInput) (bool, error) {
	if strings.TrimSpace(input.Provider) == "" || strings.TrimSpace(input.ExternalWorkspaceID) == "" || strings.TrimSpace(input.ExternalChannelID) == "" || strings.TrimSpace(input.ExternalThreadID) == "" || strings.TrimSpace(input.ExternalMessageID) == "" || strings.TrimSpace(input.ExternalAuthorID) == "" || strings.TrimSpace(input.Body) == "" {
		return false, nil
	}
	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO integration_request_comments (
			workspace_id, thread_id, direction, author_user_id,
			external_author_id, external_message_id, body, created_at, updated_at
		)
		SELECT irt.workspace_id, irt.id, 'inbound', $7, NULLIF($6, ''), $5, $8, $9, $9
		FROM integration_request_threads irt
		JOIN integration_requests ir ON ir.id = irt.integration_request_id
		WHERE irt.provider = $1
		  AND irt.external_workspace_id = $2
		  AND irt.external_channel_id = $3
		  AND irt.external_thread_id = $4
		  AND irt.installation_generation = $10
		ON CONFLICT (thread_id, external_message_id)
		WHERE external_message_id IS NOT NULL
		DO NOTHING
	`, strings.TrimSpace(input.Provider), strings.TrimSpace(input.ExternalWorkspaceID), strings.TrimSpace(input.ExternalChannelID), strings.TrimSpace(input.ExternalThreadID), strings.TrimSpace(input.ExternalMessageID), strings.TrimSpace(input.ExternalAuthorID), input.AuthorUserID, strings.TrimSpace(input.Body), createdAt, input.InstallationGeneration)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, nil
	}
	var known bool
	err = r.db.GetContext(ctx, &known, `
		SELECT EXISTS (
			SELECT 1
			FROM integration_request_threads irt
			JOIN integration_requests ir ON ir.id = irt.integration_request_id
			WHERE irt.provider = $1
			  AND irt.external_workspace_id = $2
			  AND irt.external_channel_id = $3
			  AND irt.external_thread_id = $4
			  AND irt.installation_generation = $5
		)
	`, strings.TrimSpace(input.Provider), strings.TrimSpace(input.ExternalWorkspaceID), strings.TrimSpace(input.ExternalChannelID), strings.TrimSpace(input.ExternalThreadID), input.InstallationGeneration)
	return known, err
}

const commentSelect = `
	SELECT irc.id, irc.workspace_id, irc.thread_id, irc.direction,
	       irc.author_user_id,
	       COALESCE(u.full_name, u.email, NULLIF(irc.external_author_id, ''), 'Slack user') AS author_name,
	       u.avatar_url AS author_avatar,
	       irc.external_author_id, irc.external_message_id,
	       irc.client_idempotency_key, irc.outbound_idempotency_key,
	       irc.delivery_status, irc.body, irc.created_at, irc.updated_at
	FROM integration_request_comments irc
	LEFT JOIN users u ON u.user_id = irc.author_user_id
`

func (r *Repo) GetCommentForUser(ctx context.Context, workspaceID, commentID, userID uuid.UUID) (integrationrequests.CoreIntegrationRequestComment, error) {
	var row commentRow
	err := r.db.GetContext(ctx, &row, commentSelect+`
		JOIN integration_request_threads irt ON irt.id = irc.thread_id
		JOIN integration_requests ir ON ir.id = irt.integration_request_id
		WHERE irc.workspace_id = $1
		  AND irc.id = $2
		  AND `+teamAccessPredicate("ir.team_id", "ir.workspace_id", "$3")+`
	`, workspaceID, commentID, userID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequestComment{}, err
	}
	return toCoreComment(row), nil
}

func (r *Repo) listThreadComments(ctx context.Context, threadID uuid.UUID) ([]integrationrequests.CoreIntegrationRequestComment, error) {
	rows := make([]commentRow, 0)
	err := r.db.SelectContext(ctx, &rows, commentSelect+`
		WHERE irc.thread_id = $1
		ORDER BY irc.created_at ASC, irc.id ASC
	`, threadID)
	if err != nil {
		return nil, err
	}
	result := make([]integrationrequests.CoreIntegrationRequestComment, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCoreComment(row))
	}
	return result, nil
}

func toCoreProviderThread(row providerThreadRow) integrationrequests.CoreProviderThread {
	return integrationrequests.CoreProviderThread{
		ID: row.ID, WorkspaceID: row.WorkspaceID, IntegrationRequestID: row.IntegrationRequestID,
		TeamID: row.TeamID, AcceptedStoryID: row.AcceptedStoryID, Provider: row.Provider,
		ExternalWorkspaceID: row.ExternalWorkspaceID, InstallationGeneration: row.InstallationGeneration,
		ExternalChannelID: row.ExternalChannelID, ExternalThreadID: row.ExternalThreadID,
		ExternalSourceMessageID: row.ExternalSourceMessageID, SourceURL: row.SourceURL,
		RequestTitle: row.RequestTitle, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func toCoreComment(row commentRow) integrationrequests.CoreIntegrationRequestComment {
	return integrationrequests.CoreIntegrationRequestComment{
		ID: row.ID, WorkspaceID: row.WorkspaceID, ThreadID: row.ThreadID, Direction: row.Direction,
		AuthorUserID: row.AuthorUserID, AuthorName: row.AuthorName, AuthorAvatar: row.AuthorAvatar,
		ExternalAuthorID: row.ExternalAuthorID, ExternalMessageID: row.ExternalMessageID,
		ClientIdempotencyKey:   row.ClientIdempotencyKey,
		OutboundIdempotencyKey: row.OutboundIdempotencyKey, DeliveryStatus: row.DeliveryStatus,
		Body: row.Body, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
