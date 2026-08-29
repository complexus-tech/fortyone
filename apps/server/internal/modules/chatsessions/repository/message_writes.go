package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type dbDurableApprovalReceipt struct {
	Status       string
	Output       []byte
	LeaseExpired bool
}

const (
	uncertainApprovalOutputMessage = "Maya could not verify whether this approved change finished. Check the workspace before trying it again; an identical change is blocked until this execution is reconciled."
	expiredApprovalOutputMessage   = "This approved change expired before execution and was not run. Ask Maya to prepare it again."
)

// BeginMessageWrite serializes transcript transitions on the chat_messages
// row. A new opaque token supersedes abandoned model reservations only after
// the incoming request has been validated against the committed transcript.
func (r *repo) BeginMessageWrite(ctx context.Context, params chatsessions.BeginMessageWriteParams) (chatsessions.CoreMessageWriteReservation, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.BeginMessageWrite")
	defer span.End()

	var reservation chatsessions.CoreMessageWriteReservation
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		if _, err := queries.UpsertChatSession(ctx, chatsessionssql.UpsertChatSessionParams{
			SessionID:   params.Session.ID,
			UserID:      params.Session.UserID,
			WorkspaceID: params.Session.WorkspaceID,
			Title:       params.Session.Title,
		}); errors.Is(err, pgx.ErrNoRows) {
			return chatsessions.ErrNotFound
		} else if err != nil {
			return fmt.Errorf("upsert message write session: %w", err)
		}
		if _, err := queries.InitializeChatMessages(ctx, chatsessionssql.InitializeChatMessagesParams{
			SessionID: params.Session.ID,
		}); err != nil {
			return fmt.Errorf("initialize chat messages: %w", err)
		}

		write, err := lockMessageWrite(ctx, queries, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID)
		if err != nil {
			return err
		}
		current, err := decodeMessages(write.Messages)
		if err != nil {
			return err
		}
		requestMessages := params.Messages

		if err := expireStaleExecutingMutationApprovals(ctx, queries, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID); err != nil {
			return err
		}
		if params.Operation == chatsessions.MessageWriteApproval {
			current, err = chatsessions.PrepareMutationApprovalRetries(
				current,
				params.Messages,
				func(intent chatsessions.MutationApprovalRetryIntent) (bool, error) {
					return prepareMutationApprovalRetry(
						ctx,
						queries,
						params.Session.ID,
						params.Session.UserID,
						params.Session.WorkspaceID,
						intent,
					)
				},
			)
			if err != nil {
				return err
			}
		}
		current, err = chatsessions.RecoverDurableApprovalReceipts(
			current,
			func(toolCallID string) (chatsessions.DurableApprovalReceipt, bool, error) {
				return getDurableApprovalReceipt(ctx, queries, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID, toolCallID)
			},
		)
		if err != nil {
			return err
		}

		if params.Operation != chatsessions.MessageWriteApproval {
			pending, err := hasPendingMutationExecution(ctx, queries, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID)
			if err != nil {
				return err
			}
			if pending {
				return chatsessions.ErrMessageWriteApprovalOpen
			}
		} else {
			var reconciledIncoming []any
			current, reconciledIncoming, err = chatsessions.ReconcileCompletedApprovalReservation(
				current,
				params.Messages,
				func(toolCallID string) (any, bool, error) {
					return getCompletedApprovalOutput(ctx, queries, params.Session.ID, params.Session.UserID, params.Session.WorkspaceID, toolCallID)
				},
			)
			if err != nil {
				return err
			}
			params.Messages = reconciledIncoming
		}

		reservedMessages, err := chatsessions.ReserveMessageWriteForTarget(current, params.Messages, params.Operation, params.TargetMessageID)
		if err != nil {
			return err
		}
		canonicalMessages, repaired, err := chatsessions.CanonicalMessageWriteResponse(reservedMessages, requestMessages)
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(reservedMessages)
		if err != nil {
			return fmt.Errorf("encode reserved chat messages: %w", err)
		}

		reservation = chatsessions.CoreMessageWriteReservation{
			Generation: write.WriteGeneration + 1,
			Token:      uuid.New(),
		}
		if repaired {
			reservation.Messages = canonicalMessages
		}
		operation := string(params.Operation)
		rows, err := queries.ReserveMessageWrite(ctx, chatsessionssql.ReserveMessageWriteParams{
			Messages:        encoded,
			WriteGeneration: reservation.Generation,
			WriteToken:      &reservation.Token,
			WriteOperation:  &operation,
			SessionID:       params.Session.ID,
		})
		if err != nil {
			return fmt.Errorf("reserve chat message write: %w", err)
		}
		return requireOneMessageWriteRow(rows)
	})
	if err != nil {
		return chatsessions.CoreMessageWriteReservation{}, err
	}
	return reservation, nil
}

// FinalizeMessageWrite is a token-and-generation CAS. A response from an older
// request is deliberately acknowledged as an unapplied no-op.
func (r *repo) FinalizeMessageWrite(ctx context.Context, params chatsessions.FinalizeMessageWriteParams) (chatsessions.CoreMessageWriteResult, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.FinalizeMessageWrite")
	defer span.End()

	result := chatsessions.CoreMessageWriteResult{}
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		write, err := lockMessageWrite(ctx, queries, params.SessionID, params.UserID, params.WorkspaceID)
		if err != nil {
			return err
		}
		if write.WriteGeneration != params.Generation || write.WriteToken == nil || *write.WriteToken != params.Token {
			result.Applied = false
			return nil
		}
		if write.WriteOperation == nil {
			return chatsessions.ErrMessageWriteConflict
		}

		current, err := decodeMessages(write.Messages)
		if err != nil {
			return err
		}
		merged, err := chatsessions.FinalizeMessageWriteTransition(current, params.Messages, chatsessions.MessageWriteOperation(*write.WriteOperation))
		if err != nil {
			return err
		}
		if write.WriteFinalizedAt != nil {
			if reflect.DeepEqual(current, merged) {
				result.Applied = true
				return nil
			}
			return chatsessions.ErrMessageWriteConflict
		}

		encoded, err := json.Marshal(merged)
		if err != nil {
			return fmt.Errorf("encode finalized chat messages: %w", err)
		}
		rows, err := queries.FinalizeMessageWriteCAS(ctx, chatsessionssql.FinalizeMessageWriteCASParams{
			Messages:        encoded,
			SessionID:       params.SessionID,
			WriteGeneration: params.Generation,
			WriteToken:      &params.Token,
		})
		if err != nil {
			return fmt.Errorf("finalize chat message write: %w", err)
		}
		if err := requireOneMessageWriteRow(rows); err != nil {
			return err
		}
		result.Applied = true
		return nil
	})
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	return result, nil
}

// RecoverMutationApprovalOutput performs a surgical projection of a completed
// ledger output. It neither changes nor invalidates an unrelated active write
// reservation, so a later model finalization preserves the recovered prefix.
func (r *repo) RecoverMutationApprovalOutput(ctx context.Context, params chatsessions.RecoverMutationApprovalOutputParams) (chatsessions.CoreMessageWriteResult, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.RecoverMutationApprovalOutput")
	defer span.End()

	result := chatsessions.CoreMessageWriteResult{}
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		output, err := queries.GetCompletedMutationApprovalOutputByFingerprint(ctx, chatsessionssql.GetCompletedMutationApprovalOutputByFingerprintParams{
			SessionID:   params.SessionID,
			UserID:      params.UserID,
			WorkspaceID: params.WorkspaceID,
			ToolCallID:  params.ToolCallID,
			Fingerprint: params.Fingerprint,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return chatsessions.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("get completed mutation approval output: %w", err)
		}
		if !json.Valid(output) {
			return errors.New("completed mutation approval output is invalid")
		}

		write, err := lockMessageWrite(ctx, queries, params.SessionID, params.UserID, params.WorkspaceID)
		if err != nil {
			return err
		}
		messages, err := decodeMessages(write.Messages)
		if err != nil {
			return err
		}
		var decodedOutput any
		if err := json.Unmarshal(output, &decodedOutput); err != nil {
			return fmt.Errorf("decode completed mutation approval output: %w", err)
		}
		applied, changed, err := mergeCompletedToolOutput(messages, params.ToolCallID, decodedOutput)
		if err != nil {
			return err
		}
		if !applied || !changed {
			result.Applied = applied
			return nil
		}

		encoded, err := json.Marshal(messages)
		if err != nil {
			return fmt.Errorf("encode recovered mutation output: %w", err)
		}
		rows, err := queries.RecoverMessageWrite(ctx, chatsessionssql.RecoverMessageWriteParams{
			Messages:  encoded,
			SessionID: params.SessionID,
		})
		if err != nil {
			return fmt.Errorf("recover mutation approval output: %w", err)
		}
		if err := requireOneMessageWriteRow(rows); err != nil {
			return err
		}
		result.Applied = true
		return nil
	})
	if err != nil {
		return chatsessions.CoreMessageWriteResult{}, err
	}
	return result, nil
}

func lockMessageWrite(
	ctx context.Context,
	queries chatsessionssql.Querier,
	sessionID string,
	userID, workspaceID uuid.UUID,
) (chatsessionssql.ChatMessage, error) {
	write, err := queries.LockMessageWrite(ctx, chatsessionssql.LockMessageWriteParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return chatsessionssql.ChatMessage{}, chatsessions.ErrNotFound
	}
	if err != nil {
		return chatsessionssql.ChatMessage{}, fmt.Errorf("lock chat message write: %w", err)
	}
	return write, nil
}

func hasPendingMutationExecution(
	ctx context.Context,
	queries chatsessionssql.Querier,
	sessionID string,
	userID, workspaceID uuid.UUID,
) (bool, error) {
	pending, err := queries.HasPendingMutationApproval(ctx, chatsessionssql.HasPendingMutationApprovalParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("check pending mutation approval: %w", err)
	}
	return pending, nil
}

func prepareMutationApprovalRetry(
	ctx context.Context,
	queries chatsessionssql.Querier,
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
	resolution := mutationApprovalSafeRetryResolution
	prepared, err := queries.PrepareMutationApprovalRetry(ctx, chatsessionssql.PrepareMutationApprovalRetryParams{
		Resolution:  &resolution,
		Evidence:    evidence,
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
		ToolCallID:  intent.ToolCallID,
		Fingerprint: intent.Fingerprint,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("prepare safe mutation approval retry: %w", err)
	}
	return prepared, nil
}

func expireStaleExecutingMutationApprovals(
	ctx context.Context,
	queries chatsessionssql.Querier,
	sessionID string,
	userID, workspaceID uuid.UUID,
) error {
	failureCode := mutationApprovalLeaseExpiredFailure
	if _, err := queries.ExpireStaleMutationApprovals(ctx, chatsessionssql.ExpireStaleMutationApprovalsParams{
		FailureCode: &failureCode,
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	}); err != nil {
		return fmt.Errorf("expire stale mutation approval executions: %w", err)
	}
	return nil
}

func getDurableApprovalReceipt(
	ctx context.Context,
	queries chatsessionssql.Querier,
	sessionID string,
	userID, workspaceID uuid.UUID,
	toolCallID string,
) (chatsessions.DurableApprovalReceipt, bool, error) {
	stored, err := queries.GetDurableApprovalReceipt(ctx, chatsessionssql.GetDurableApprovalReceiptParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
		ToolCallID:  toolCallID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return chatsessions.DurableApprovalReceipt{}, false, nil
	}
	if err != nil {
		return chatsessions.DurableApprovalReceipt{}, false, fmt.Errorf("get durable approval receipt: %w", err)
	}
	return toDurableApprovalReceipt(dbDurableApprovalReceipt{
		Status:       stored.Status,
		Output:       stored.Output,
		LeaseExpired: stored.LeaseExpired,
	})
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

func getCompletedApprovalOutput(
	ctx context.Context,
	queries chatsessionssql.Querier,
	sessionID string,
	userID, workspaceID uuid.UUID,
	toolCallID string,
) (any, bool, error) {
	output, err := queries.GetCompletedApprovalOutput(ctx, chatsessionssql.GetCompletedApprovalOutputParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
		ToolCallID:  toolCallID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get completed approval output for reservation: %w", err)
	}
	var decoded any
	if err := json.Unmarshal(output, &decoded); err != nil {
		return nil, false, fmt.Errorf("decode completed approval output for reservation: %w", err)
	}
	return decoded, true, nil
}

func decodeMessages(data []byte) ([]any, error) {
	var messages []any
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("decode chat messages: %w", err)
	}
	return messages, nil
}

func requireOneMessageWriteRow(rows int64) error {
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
