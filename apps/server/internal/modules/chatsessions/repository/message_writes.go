package chatsessionsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type dbMessageWrite struct {
	Messages         json.RawMessage `db:"messages"`
	WriteGeneration  int64           `db:"write_generation"`
	WriteToken       uuid.NullUUID   `db:"write_token"`
	WriteOperation   sql.NullString  `db:"write_operation"`
	WriteFinalizedAt sql.NullTime    `db:"write_finalized_at"`
}

type dbDurableApprovalReceipt struct {
	Status       string `db:"status"`
	Output       []byte `db:"output"`
	LeaseExpired bool   `db:"lease_expired"`
}

const (
	uncertainApprovalOutputMessage = "Maya could not verify whether this approved change finished. Check the workspace before trying it again; an identical change is blocked until this execution is reconciled."
	expiredApprovalOutputMessage   = "This approved change expired before execution and was not run. Ask Maya to prepare it again."
)

// BeginMessageWrite serializes transcript transitions on the chat_messages
// row. A new opaque token supersedes abandoned model reservations only after
// the incoming request has been validated against the committed transcript.
func (r *repo) BeginMessageWrite(ctx context.Context, params chatsessions.BeginMessageWriteParams) (chatsessions.CoreMessageWriteReservation, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.BeginMessageWrite")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, fmt.Errorf("begin message write transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := upsertMessageWriteSession(ctx, tx, params.Session); err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO chat_messages (session_id, messages)
		VALUES ($1, CAST('[]' AS jsonb))
		ON CONFLICT (session_id) DO NOTHING
	`, params.Session.ID); err != nil {
		return chatsessions.CoreMessageWriteReservation{}, fmt.Errorf("initialize chat messages: %w", err)
	}

	write, err := getMessageWriteForUpdate(ctx, tx, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	current, err := decodeMessages(write.Messages)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	requestMessages := params.Messages

	if err := expireStaleExecutingMutationApprovals(ctx, tx, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID); err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	if params.Operation == chatsessions.MessageWriteApproval {
		current, err = chatsessions.PrepareMutationApprovalRetries(
			current,
			params.Messages,
			func(intent chatsessions.MutationApprovalRetryIntent) (bool, error) {
				return prepareMutationApprovalRetry(
					ctx,
					tx,
					params.Session.ID,
					params.Session.UserID,
					params.Session.WorkspaceID,
					intent,
				)
			},
		)
		if err != nil {
			return chatsessions.CoreMessageWriteReservation{}, err
		}
	}
	current, err = chatsessions.RecoverDurableApprovalReceipts(
		current,
		func(toolCallID string) (chatsessions.DurableApprovalReceipt, bool, error) {
			return getDurableApprovalReceipt(ctx, tx, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID, toolCallID)
		},
	)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}

	if params.Operation != chatsessions.MessageWriteApproval {
		pending, err := hasPendingMutationExecution(ctx, tx, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID)
		if err != nil {
			return chatsessions.CoreMessageWriteReservation{}, err
		}
		if pending {
			return chatsessions.CoreMessageWriteReservation{}, chatsessions.ErrMessageWriteApprovalOpen
		}
	} else {
		var reconciledIncoming []any
		current, reconciledIncoming, err = chatsessions.ReconcileCompletedApprovalReservation(
			current,
			params.Messages,
			func(toolCallID string) (any, bool, error) {
				return getCompletedApprovalOutput(ctx, tx, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID, toolCallID)
			},
		)
		if err != nil {
			return chatsessions.CoreMessageWriteReservation{}, err
		}
		params.Messages = reconciledIncoming
	}

	reservedMessages, err := chatsessions.ReserveMessageWriteForTarget(current, params.Messages, params.Operation, params.TargetMessageID)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	canonicalMessages, repaired, err := chatsessions.CanonicalMessageWriteResponse(reservedMessages, requestMessages)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	encoded, err := json.Marshal(reservedMessages)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, fmt.Errorf("encode reserved chat messages: %w", err)
	}

	reservation := chatsessions.CoreMessageWriteReservation{
		Generation: write.WriteGeneration + 1,
		Token:      uuid.New(),
	}
	if repaired {
		reservation.Messages = canonicalMessages
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE chat_messages
		SET messages = $2,
			write_generation = $3,
			write_token = $4,
			write_operation = $5,
			write_finalized_at = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
	`, params.Session.ID, encoded, reservation.Generation, reservation.Token, params.Operation)
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, fmt.Errorf("reserve chat message write: %w", err)
	}
	if err := requireOneMessageWriteRow(result); err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	if err := tx.Commit(); err != nil {
		return chatsessions.CoreMessageWriteReservation{}, fmt.Errorf("commit message write reservation: %w", err)
	}
	return reservation, nil
}

// FinalizeMessageWrite is a token-and-generation CAS. A response from an older
// request is deliberately acknowledged as an unapplied no-op.
func (r *repo) FinalizeMessageWrite(ctx context.Context, params chatsessions.FinalizeMessageWriteParams) (chatsessions.CoreMessageWriteResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.FinalizeMessageWrite")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("begin message write finalization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	write, err := getMessageWriteForUpdate(ctx, tx, params.SessionID, params.UserID, params.WorkspaceID)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	if write.WriteGeneration != params.Generation || !write.WriteToken.Valid || write.WriteToken.UUID != params.Token {
		return chatsessions.CoreMessageWriteResult{Applied: false}, nil
	}
	if !write.WriteOperation.Valid {
		return chatsessions.CoreMessageWriteResult{}, chatsessions.ErrMessageWriteConflict
	}

	current, err := decodeMessages(write.Messages)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	merged, err := chatsessions.FinalizeMessageWriteTransition(current, params.Messages, chatsessions.MessageWriteOperation(write.WriteOperation.String))
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	if write.WriteFinalizedAt.Valid {
		if reflect.DeepEqual(current, merged) {
			return chatsessions.CoreMessageWriteResult{Applied: true}, nil
		}
		return chatsessions.CoreMessageWriteResult{}, chatsessions.ErrMessageWriteConflict
	}

	encoded, err := json.Marshal(merged)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("encode finalized chat messages: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE chat_messages
		SET messages = $4,
			write_finalized_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND write_generation = $2
			AND write_token = $3
	`, params.SessionID, params.Generation, params.Token, encoded)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("finalize chat message write: %w", err)
	}
	if err := requireOneMessageWriteRow(result); err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("commit chat message finalization: %w", err)
	}
	return chatsessions.CoreMessageWriteResult{Applied: true}, nil
}

// RecoverMutationApprovalOutput performs a surgical projection of a completed
// ledger output. It neither changes nor invalidates an unrelated active write
// reservation, so a later model finalization preserves the recovered prefix.
func (r *repo) RecoverMutationApprovalOutput(ctx context.Context, params chatsessions.RecoverMutationApprovalOutputParams) (chatsessions.CoreMessageWriteResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.RecoverMutationApprovalOutput")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("begin mutation output recovery: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var output json.RawMessage
	if err := tx.GetContext(ctx, &output, `
		SELECT execution.output
		FROM chat_mutation_approval_executions AS execution
		JOIN chat_sessions AS session ON session.id = execution.session_id
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.status = 'completed'
			AND session.user_id = $2
			AND session.workspace_id = $3
			AND session.deleted_at IS NULL
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID, params.Fingerprint); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chatsessions.CoreMessageWriteResult{}, chatsessions.ErrNotFound
		}
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("get completed mutation approval output: %w", err)
	}
	if !json.Valid(output) {
		return chatsessions.CoreMessageWriteResult{}, errors.New("completed mutation approval output is invalid")
	}

	write, err := getMessageWriteForUpdate(ctx, tx, params.SessionID, params.UserID, params.WorkspaceID)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	messages, err := decodeMessages(write.Messages)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	var decodedOutput any
	if err := json.Unmarshal(output, &decodedOutput); err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("decode completed mutation approval output: %w", err)
	}
	applied, changed, err := mergeCompletedToolOutput(messages, params.ToolCallID, decodedOutput)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	if !applied || !changed {
		return chatsessions.CoreMessageWriteResult{Applied: applied}, nil
	}

	encoded, err := json.Marshal(messages)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("encode recovered mutation output: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE chat_messages
		SET messages = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
	`, params.SessionID, encoded)
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("recover mutation approval output: %w", err)
	}
	if err := requireOneMessageWriteRow(result); err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return chatsessions.CoreMessageWriteResult{}, fmt.Errorf("commit mutation output recovery: %w", err)
	}
	return chatsessions.CoreMessageWriteResult{Applied: true}, nil
}

func upsertMessageWriteSession(ctx context.Context, tx *sqlx.Tx, session chatsessions.CoreChatSession) error {
	const query = `
		INSERT INTO chat_sessions (id, user_id, workspace_id, title)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET updated_at = CURRENT_TIMESTAMP
		WHERE chat_sessions.user_id = EXCLUDED.user_id
			AND chat_sessions.workspace_id = EXCLUDED.workspace_id
			AND chat_sessions.deleted_at IS NULL
		RETURNING id
	`
	var id string
	if err := tx.GetContext(ctx, &id, query, session.ID, session.UserID, session.WorkspaceID, session.Title); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chatsessions.ErrNotFound
		}
		return fmt.Errorf("upsert message write session: %w", err)
	}
	return nil
}

func getMessageWriteForUpdate(ctx context.Context, tx *sqlx.Tx, sessionID string, userID, workspaceID uuid.UUID) (dbMessageWrite, error) {
	const query = `
		SELECT messages.messages,
		       messages.write_generation,
		       messages.write_token,
		       messages.write_operation,
		       messages.write_finalized_at
		FROM chat_messages AS messages
		JOIN chat_sessions AS session ON session.id = messages.session_id
		WHERE messages.session_id = $1
			AND session.user_id = $2
			AND session.workspace_id = $3
			AND session.deleted_at IS NULL
		FOR UPDATE OF messages
	`
	var write dbMessageWrite
	if err := tx.GetContext(ctx, &write, query, sessionID, userID, workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbMessageWrite{}, chatsessions.ErrNotFound
		}
		return dbMessageWrite{}, fmt.Errorf("get message write: %w", err)
	}
	return write, nil
}

func hasPendingMutationExecution(ctx context.Context, tx *sqlx.Tx, sessionID string, userID, workspaceID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM chat_mutation_approval_executions
			WHERE session_id = $1
				AND user_id = $2
					AND workspace_id = $3
					AND (
						status = 'retry_ready'
						OR (
							status IN ('ready', 'executing')
							AND lease_expires_at > CURRENT_TIMESTAMP
						)
					)
			)
	`
	var pending bool
	if err := tx.GetContext(ctx, &pending, query, sessionID, userID, workspaceID); err != nil {
		return false, fmt.Errorf("check pending mutation approval: %w", err)
	}
	return pending, nil
}

func prepareMutationApprovalRetry(
	ctx context.Context,
	tx *sqlx.Tx,
	sessionID string,
	userID, workspaceID uuid.UUID,
	intent chatsessions.MutationApprovalRetryIntent,
) (bool, error) {
	evidence, err := json.Marshal(map[string]any{
		"kind":     "server_verified_idempotent_retry",
		"toolName": intent.ToolName,
	})
	if err != nil {
		return false, fmt.Errorf("marshal safe mutation retry evidence: %w", err)
	}

	const query = `
		WITH prepared AS (
			UPDATE chat_mutation_approval_executions AS execution
			SET status = 'retry_ready',
				lease_token = NULL,
				lease_expires_at = NULL,
				started_at = NULL,
				last_reconciliation_resolution = $6,
				last_reconciliation_evidence = CAST($7 AS jsonb),
				last_reconciled_at = CURRENT_TIMESTAMP,
				reconciliation_count = reconciliation_count + 1,
				updated_at = CURRENT_TIMESTAMP
			FROM chat_sessions AS session
			WHERE execution.session_id = $1
				AND execution.user_id = $2
				AND execution.workspace_id = $3
				AND execution.tool_call_id = $4
				AND execution.fingerprint = $5
				AND execution.status = 'failed_uncertain'
				AND execution.reconciliation_count = 0
				AND session.id = execution.session_id
				AND session.user_id = $2
				AND session.workspace_id = $3
				AND session.deleted_at IS NULL
			RETURNING TRUE AS prepared
		), existing AS (
			SELECT TRUE AS prepared
			FROM chat_mutation_approval_executions AS execution
			JOIN chat_sessions AS session ON session.id = execution.session_id
			WHERE execution.session_id = $1
				AND execution.user_id = $2
				AND execution.workspace_id = $3
				AND execution.tool_call_id = $4
				AND execution.fingerprint = $5
				AND execution.status = 'retry_ready'
				AND execution.reconciliation_count = 1
				AND execution.last_reconciliation_resolution = $6
				AND session.user_id = $2
				AND session.workspace_id = $3
				AND session.deleted_at IS NULL
		)
		SELECT prepared FROM prepared
		UNION ALL
		SELECT prepared FROM existing
		LIMIT 1
	`
	var prepared bool
	if err := tx.GetContext(
		ctx,
		&prepared,
		query,
		sessionID,
		userID,
		workspaceID,
		intent.ToolCallID,
		intent.Fingerprint,
		mutationApprovalSafeRetryResolution,
		string(evidence),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("prepare safe mutation approval retry: %w", err)
	}
	return prepared, nil
}

func expireStaleExecutingMutationApprovals(ctx context.Context, tx *sqlx.Tx, sessionID string, userID, workspaceID uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE chat_mutation_approval_executions
		SET status = 'failed_uncertain',
			failure_code = $4,
			lease_token = NULL,
			lease_expires_at = NULL,
			failed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND status = 'executing'
			AND lease_expires_at <= CURRENT_TIMESTAMP
	`, sessionID, userID, workspaceID, mutationApprovalLeaseExpiredFailure); err != nil {
		return fmt.Errorf("expire stale mutation approval executions: %w", err)
	}
	return nil
}

func getDurableApprovalReceipt(ctx context.Context, tx *sqlx.Tx, sessionID string, userID, workspaceID uuid.UUID, toolCallID string) (chatsessions.DurableApprovalReceipt, bool, error) {
	const query = `
		SELECT status,
		       output,
		       COALESCE(lease_expires_at <= CURRENT_TIMESTAMP, FALSE) AS lease_expired
		FROM chat_mutation_approval_executions
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
			AND status IN ('completed', 'failed_uncertain', 'ready', 'retry_ready')
	`
	var stored dbDurableApprovalReceipt
	if err := tx.GetContext(ctx, &stored, query, sessionID, userID, workspaceID, toolCallID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chatsessions.DurableApprovalReceipt{}, false, nil
		}
		return chatsessions.DurableApprovalReceipt{}, false, fmt.Errorf("get durable approval receipt: %w", err)
	}
	return toDurableApprovalReceipt(stored)
}

func toDurableApprovalReceipt(stored dbDurableApprovalReceipt) (chatsessions.DurableApprovalReceipt, bool, error) {
	switch stored.Status {
	case "completed":
		if !json.Valid(stored.Output) {
			return chatsessions.DurableApprovalReceipt{}, false, errors.New("completed mutation approval output is invalid")
		}
		var output any
		if err := json.Unmarshal(stored.Output, &output); err != nil {
			return chatsessions.DurableApprovalReceipt{}, false, fmt.Errorf("decode completed approval receipt: %w", err)
		}
		return chatsessions.DurableApprovalReceipt{
			HaltsFollowing: durableApprovalOutputHaltsFollowing(output),
			Output:         output,
		}, true, nil
	case "failed_uncertain":
		return chatsessions.DurableApprovalReceipt{
			HaltsFollowing: true,
			Output: map[string]any{
				"error":   uncertainApprovalOutputMessage,
				"success": false,
			},
		}, true, nil
	case "ready":
		if !stored.LeaseExpired {
			return chatsessions.DurableApprovalReceipt{}, false, nil
		}
		return chatsessions.DurableApprovalReceipt{
			HaltsFollowing: true,
			Output: map[string]any{
				"error":   expiredApprovalOutputMessage,
				"success": false,
			},
		}, true, nil
	case "retry_ready":
		// The exact persisted approval was deliberately reopened in the same
		// transaction. It is pending execution, not a terminal receipt.
		return chatsessions.DurableApprovalReceipt{}, false, nil
	default:
		return chatsessions.DurableApprovalReceipt{}, false, fmt.Errorf("unsupported durable approval receipt status %q", stored.Status)
	}
}

func durableApprovalOutputHaltsFollowing(output any) bool {
	object, ok := output.(map[string]any)
	if !ok {
		return false
	}
	if success, exists := object["success"]; exists && success == false {
		return true
	}
	errorValue, exists := object["error"]
	return exists && errorValue != nil
}

func getCompletedApprovalOutput(ctx context.Context, tx *sqlx.Tx, sessionID string, userID, workspaceID uuid.UUID, toolCallID string) (any, bool, error) {
	const query = `
		SELECT output
		FROM chat_mutation_approval_executions
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
			AND status = 'completed'
	`
	var output json.RawMessage
	if err := tx.GetContext(ctx, &output, query, sessionID, userID, workspaceID, toolCallID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get completed approval output for reservation: %w", err)
	}
	var decoded any
	if err := json.Unmarshal(output, &decoded); err != nil {
		return nil, false, fmt.Errorf("decode completed approval output for reservation: %w", err)
	}
	return decoded, true, nil
}

func decodeMessages(data json.RawMessage) ([]any, error) {
	var messages []any
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("decode chat messages: %w", err)
	}
	return messages, nil
}

func requireOneMessageWriteRow(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read chat message write result: %w", err)
	}
	if rows != 1 {
		return chatsessions.ErrMessageWriteConflict
	}
	return nil
}

func mergeCompletedToolOutput(messages []any, toolCallID string, output any) (applied, changed bool, err error) {
	var matched map[string]any
	for _, rawMessage := range messages {
		message, ok := rawMessage.(map[string]any)
		if !ok {
			continue
		}
		parts, ok := message["parts"].([]any)
		if !ok {
			continue
		}
		for _, rawPart := range parts {
			part, ok := rawPart.(map[string]any)
			if !ok || part["toolCallId"] != toolCallID {
				continue
			}
			typeName, typeOK := part["type"].(string)
			if !typeOK || len(typeName) < len("tool-") || typeName[:len("tool-")] != "tool-" || matched != nil {
				return false, false, chatsessions.ErrMessageWriteConflict
			}
			matched = part
		}
	}
	if matched == nil {
		return false, false, nil
	}

	state, _ := matched["state"].(string)
	if state == "output-available" {
		if reflect.DeepEqual(matched["output"], output) {
			return true, false, nil
		}
		// The verified completed ledger is authoritative over a generic
		// uncertainty/pending receipt persisted after a response-loss race.
		matched["output"] = output
		delete(matched, "errorText")
		return true, true, nil
	}
	if state != "approval-requested" && state != "approval-responded" {
		return false, false, chatsessions.ErrMessageWriteConflict
	}
	matched["state"] = "output-available"
	matched["output"] = output
	delete(matched, "errorText")
	return true, true, nil
}
