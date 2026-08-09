package integrationrequestsrepository

import (
	"context"
	"os"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestCommentDeliveryPostgresContract exercises the caller-idempotency,
// provider-payload, nullable external-author, and retained-status invariants
// against a fully migrated PostgreSQL database when one is supplied.
func TestCommentDeliveryPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_REQUEST_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_REQUEST_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	ctx := context.Background()
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	threadID := uuid.New()
	authorID := uuid.New()
	installGeneration := uuid.New()
	clientKey := uuid.New()

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, authorID, "comment-author-"+authorID.String(), "comment-author-"+authorID.String()+"@example.com")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, 'Comment delivery test', $2)
	`, workspaceID, "comment-delivery-"+workspaceID.String())
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Comment team', 'CMT', '#111827')
	`, teamID, workspaceID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`, teamID, authorID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO integration_requests (
			id, workspace_id, team_id, provider, source_type,
			source_external_id, title, priority
		) VALUES ($1, $2, $3, 'slack', 'message', $4, 'Durable comment request', 'High')
	`, requestID, workspaceID, teamID, "event-"+requestID.String())
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO integration_request_threads (
			id, workspace_id, integration_request_id, provider,
			external_workspace_id, installation_generation,
			external_channel_id, external_thread_id
		) VALUES ($1, $2, $3, 'slack', 'T-comment', $4, 'C-comment', '1710000000.001')
	`, threadID, workspaceID, requestID, installGeneration)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := db.Exec(`DELETE FROM workspaces WHERE workspace_id = $1`, workspaceID)
		require.NoError(t, cleanupErr)
		_, cleanupErr = db.Exec(`DELETE FROM users WHERE user_id = $1`, authorID)
		require.NoError(t, cleanupErr)
	})

	repo := New(nil, db)
	input := integrationrequests.CoreCreateCommentInput{
		WorkspaceID:          workspaceID,
		RequestID:            requestID,
		AuthorID:             authorID,
		ClientIdempotencyKey: clientKey,
		Body:                 "Ship this reply",
	}
	prepared := integrationrequests.CorePreparedProviderComment{
		ProviderPayload: []byte(`{"opaque_provider_authorization":{"team":"selected"}}`),
	}
	_, first, err := repo.CreateOutboundComment(ctx, input, prepared)
	require.NoError(t, err)
	require.Equal(t, &clientKey, first.ClientIdempotencyKey)
	require.Equal(t, "sending", valueOrEmpty(first.DeliveryStatus))

	_, duplicate, err := repo.CreateOutboundComment(ctx, input, prepared)
	require.NoError(t, err)
	require.Equal(t, first.ID, duplicate.ID)

	conflict := input
	conflict.Body = "Different content"
	_, _, err = repo.CreateOutboundComment(ctx, conflict, prepared)
	require.ErrorIs(t, err, integrationrequests.ErrIdempotencyConflict)

	var commentCount, deliveryCount int
	require.NoError(t, db.GetContext(ctx, &commentCount, `
		SELECT COUNT(*) FROM integration_request_comments
		WHERE workspace_id = $1 AND client_idempotency_key = $2
	`, workspaceID, clientKey))
	require.NoError(t, db.GetContext(ctx, &deliveryCount, `
		SELECT COUNT(*) FROM messaging_outbound_deliveries
		WHERE workspace_id = $1 AND idempotency_key = $2
	`, workspaceID, "integration-request-comment:"+clientKey.String()))
	require.Equal(t, 1, commentCount)
	require.Equal(t, 1, deliveryCount)

	_, err = db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'failed', attempt_count = 2
		WHERE workspace_id = $1 AND idempotency_key = $2
	`, workspaceID, "integration-request-comment:"+clientKey.String())
	require.NoError(t, err)
	retrying, err := repo.GetCommentForUser(ctx, workspaceID, first.ID, authorID)
	require.NoError(t, err)
	require.Equal(t, "retrying", valueOrEmpty(retrying.DeliveryStatus))

	_, err = db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'failed', attempt_count = 20
		WHERE workspace_id = $1 AND idempotency_key = $2
	`, workspaceID, "integration-request-comment:"+clientKey.String())
	require.NoError(t, err)
	failed, err := repo.GetCommentForUser(ctx, workspaceID, first.ID, authorID)
	require.NoError(t, err)
	require.Equal(t, "failed", valueOrEmpty(failed.DeliveryStatus))

	_, err = db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'cancelled'
		WHERE workspace_id = $1 AND idempotency_key = $2
	`, workspaceID, "integration-request-comment:"+clientKey.String())
	require.NoError(t, err)
	notSent, err := repo.GetCommentForUser(ctx, workspaceID, first.ID, authorID)
	require.NoError(t, err)
	require.Equal(t, "not-sent", valueOrEmpty(notSent.DeliveryStatus))

	_, err = db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'delivered', external_message_id = '1710000000.002', delivered_at = NOW()
		WHERE workspace_id = $1 AND idempotency_key = $2
	`, workspaceID, "integration-request-comment:"+clientKey.String())
	require.NoError(t, err)
	sent, err := repo.GetCommentForUser(ctx, workspaceID, first.ID, authorID)
	require.NoError(t, err)
	require.Equal(t, "sent", valueOrEmpty(sent.DeliveryStatus))

	_, err = db.ExecContext(ctx, `
		DELETE FROM messaging_outbound_deliveries
		WHERE workspace_id = $1 AND idempotency_key = $2
	`, workspaceID, "integration-request-comment:"+clientKey.String())
	require.NoError(t, err)
	retained, err := repo.GetCommentForUser(ctx, workspaceID, first.ID, authorID)
	require.NoError(t, err)
	require.Equal(t, "sent", valueOrEmpty(retained.DeliveryStatus))

	handled, err := repo.IngestInboundProviderComment(ctx, integrationrequests.CoreInboundProviderCommentInput{
		Provider:               integrationrequests.ProviderSlack,
		ExternalWorkspaceID:    "T-comment",
		InstallationGeneration: installGeneration,
		ExternalChannelID:      "C-comment",
		ExternalThreadID:       "1710000000.001",
		ExternalMessageID:      "1710000000.003",
		ExternalAuthorID:       "U-external",
		AuthorUserID:           nil,
		Body:                   "Reply from an unlinked participant",
		CreatedAt:              time.Now().UTC(),
	})
	require.NoError(t, err)
	require.True(t, handled)
	comments, err := repo.listThreadComments(ctx, threadID)
	require.NoError(t, err)
	require.Len(t, comments, 2)
	require.Nil(t, comments[1].AuthorUserID)
	require.Equal(t, "U-external", valueOrEmpty(comments[1].ExternalAuthorID))

	handled, err = repo.IngestInboundProviderComment(ctx, integrationrequests.CoreInboundProviderCommentInput{
		Provider:               integrationrequests.ProviderSlack,
		ExternalWorkspaceID:    "T-comment",
		InstallationGeneration: uuid.New(),
		ExternalChannelID:      "C-comment",
		ExternalThreadID:       "1710000000.001",
		ExternalMessageID:      "1710000000.004",
		ExternalAuthorID:       "U-external",
		Body:                   "Stale installation reply",
	})
	require.NoError(t, err)
	require.False(t, handled)
}

func valueOrEmpty[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
