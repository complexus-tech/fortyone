package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func chatTestMessage(id, role string, parts ...any) map[string]any {
	return map[string]any{
		"id":    id,
		"parts": parts,
		"role":  role,
	}
}

type chatSessionTestActor struct {
	userID      uuid.UUID
	workspaceID uuid.UUID
}

func seedChatSessionActor(t *testing.T, db *pgxpool.Pool) chatSessionTestActor {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "chat-test-"+suffix, "chat-test-"+suffix+"@example.invalid")
	require.NoError(t, err)
	_, err = db.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, "Chat session test", "chat-test-"+suffix, userID)
	require.NoError(t, err)
	return chatSessionTestActor{userID: userID, workspaceID: workspaceID}
}

func chatSessionTestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
}

func mustGetChatMessages(t *testing.T, repository *repo, sessionID string, actor chatSessionTestActor) []any {
	t.Helper()
	messages, err := repository.GetMessages(context.Background(), sessionID, actor.userID, actor.workspaceID)
	require.NoError(t, err)
	return messages
}

func mutationApprovalTestParams(sessionID string, actor chatSessionTestActor, toolCallID, fingerprint string) chatsessions.MutationApprovalExecutionParams {
	return chatsessions.MutationApprovalExecutionParams{
		SessionID:   sessionID,
		UserID:      actor.userID,
		WorkspaceID: actor.workspaceID,
		ToolCallID:  toolCallID,
		Fingerprint: fingerprint,
	}
}

func expireMutationApprovalLease(ctx context.Context, db *pgxpool.Pool, params chatsessions.MutationApprovalExecutionParams) error {
	result, err := db.Exec(ctx, `
		UPDATE chat_mutation_approval_executions
		SET lease_expires_at = CURRENT_TIMESTAMP - INTERVAL '1 second'
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New("mutation approval lease was not found")
	}
	return nil
}

func assertMutationApprovalReconciliationAudit(t *testing.T, db *pgxpool.Pool, params chatsessions.MutationApprovalExecutionParams, expectedResolution string) {
	t.Helper()
	var resolution string
	var evidence json.RawMessage
	var count int
	err := db.QueryRow(context.Background(), `
		SELECT last_reconciliation_resolution, last_reconciliation_evidence, reconciliation_count
		FROM chat_mutation_approval_executions
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(&resolution, &evidence, &count)
	require.NoError(t, err)
	require.Equal(t, expectedResolution, resolution)
	require.JSONEq(t, `{"kind":"idempotency_lookup","reference":"story-create:call-expired","summary":"The idempotency receipt proves that the story was created."}`, string(evidence))
	require.Equal(t, 1, count)
}

func assertMutationApprovalReset(t *testing.T, db *pgxpool.Pool, params chatsessions.MutationApprovalExecutionParams) {
	t.Helper()
	var status string
	var startedAt *time.Time
	var leaseExpiresAt time.Time
	var resolution string
	err := db.QueryRow(context.Background(), `
		SELECT status, started_at, lease_expires_at, last_reconciliation_resolution
		FROM chat_mutation_approval_executions
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(
		&status,
		&startedAt,
		&leaseExpiresAt,
		&resolution,
	)
	require.NoError(t, err)
	require.Equal(t, "ready", status)
	require.Nil(t, startedAt)
	require.LessOrEqual(t, leaseExpiresAt, time.Now())
	require.Equal(t, "verified_not_applied", resolution)
}

func assertMutationApprovalRetryReady(t *testing.T, db *pgxpool.Pool, params chatsessions.MutationApprovalExecutionParams) {
	t.Helper()
	var status string
	var resolution string
	var count int
	var leaseToken *uuid.UUID
	err := db.QueryRow(context.Background(), `
		SELECT status, last_reconciliation_resolution, reconciliation_count, lease_token
		FROM chat_mutation_approval_executions
		WHERE session_id = $1 AND user_id = $2 AND workspace_id = $3 AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(&status, &resolution, &count, &leaseToken)
	require.NoError(t, err)
	require.Equal(t, "retry_ready", status)
	require.Equal(t, mutationApprovalSafeRetryResolution, resolution)
	require.Equal(t, 1, count)
	require.Nil(t, leaseToken)
}

func assertMutationApprovalFailedAfterRetry(t *testing.T, db *pgxpool.Pool, params chatsessions.MutationApprovalExecutionParams) {
	t.Helper()
	var status string
	var resolution string
	var count int
	err := db.QueryRow(context.Background(), `
		SELECT status, last_reconciliation_resolution, reconciliation_count
		FROM chat_mutation_approval_executions
		WHERE session_id = $1 AND user_id = $2 AND workspace_id = $3 AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(&status, &resolution, &count)
	require.NoError(t, err)
	require.Equal(t, "failed_uncertain", status)
	require.Equal(t, mutationApprovalSafeRetryResolution, resolution)
	require.Equal(t, 1, count)
}
