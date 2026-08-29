package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type dbMutationApprovalExecution = chatsessionssql.ChatMutationApprovalExecution

func reclaimExpiredReadyMutationApproval(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
	leaseToken uuid.UUID,
) (chatsessionssql.ChatMutationApprovalExecution, error) {
	return queries.ReclaimExpiredReadyMutationApproval(ctx, chatsessionssql.ReclaimExpiredReadyMutationApprovalParams{
		LeaseToken:  &leaseToken,
		SessionID:   params.SessionID,
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		ToolCallID:  params.ToolCallID,
		Fingerprint: params.Fingerprint,
	})
}

func claimRetryReadyMutationApproval(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
	leaseToken uuid.UUID,
) (chatsessionssql.ChatMutationApprovalExecution, error) {
	return queries.ClaimRetryReadyMutationApproval(ctx, chatsessionssql.ClaimRetryReadyMutationApprovalParams{
		LeaseToken:  &leaseToken,
		SessionID:   params.SessionID,
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		ToolCallID:  params.ToolCallID,
		Fingerprint: params.Fingerprint,
	})
}

func failExpiredExecutingMutationApproval(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
) (chatsessionssql.ChatMutationApprovalExecution, error) {
	failureCode := mutationApprovalLeaseExpiredFailure
	return queries.FailExpiredExecutingMutationApproval(ctx, chatsessionssql.FailExpiredExecutingMutationApprovalParams{
		FailureCode: &failureCode,
		SessionID:   params.SessionID,
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		ToolCallID:  params.ToolCallID,
		Fingerprint: params.Fingerprint,
	})
}

func lockMutationApproval(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
) (chatsessionssql.ChatMutationApprovalExecution, error) {
	execution, err := queries.LockMutationApproval(ctx, chatsessionssql.LockMutationApprovalParams{
		SessionID:   params.SessionID,
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		ToolCallID:  params.ToolCallID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return chatsessionssql.ChatMutationApprovalExecution{}, chatsessions.ErrNotFound
	}
	if err != nil {
		return chatsessionssql.ChatMutationApprovalExecution{}, fmt.Errorf("lock mutation approval execution: %w", err)
	}
	return execution, nil
}

func claimUnresolvedMutationApprovalFingerprint(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
	leaseToken uuid.UUID,
) (chatsessions.CoreMutationApprovalExecution, error) {
	execution, err := lockUnresolvedMutationApprovalFingerprint(ctx, queries, params)
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}

	switch execution.Status {
	case "ready":
		// A ready row has never crossed Start. Preserve its identity as a
		// terminal known-not-run tombstone, then create a separate lease for the
		// newer approval. Rewriting the old identity would let a stale browser
		// replay it after the replacement completes.
		current := execution
		execution, err = terminalizeExpiredReadyAndClaimReplacement(
			ctx,
			queries,
			params,
			execution.SessionID,
			execution.ToolCallID,
			leaseToken,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return toCoreMutationApprovalExecution(current)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("replace expired ready mutation approval: %w", err)
		}
		return toClaimedMutationApprovalExecution(execution)
	case "executing":
		current := execution
		existingParams := params
		existingParams.SessionID = execution.SessionID
		existingParams.ToolCallID = execution.ToolCallID
		execution, err = failExpiredExecutingMutationApproval(ctx, queries, existingParams)
		if errors.Is(err, pgx.ErrNoRows) {
			return toCoreMutationApprovalExecution(current)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		return toCoreMutationApprovalExecution(execution)
	case "retry_ready":
		return chatsessions.CoreMutationApprovalExecution{
			State:       chatsessions.MutationApprovalExecutionFailed,
			FailureCode: mutationApprovalWrongOriginFailure,
		}, nil
	default:
		return toCoreMutationApprovalExecution(execution)
	}
}

func lockUnresolvedMutationApprovalFingerprint(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
) (chatsessionssql.ChatMutationApprovalExecution, error) {
	// Quarantine follows the user and workspace rather than one chat. This
	// prevents a new conversation from replaying an identical mutation whose
	// outcome is still uncertain.
	execution, err := queries.LockUnresolvedMutationApprovalFingerprint(ctx, chatsessionssql.LockUnresolvedMutationApprovalFingerprintParams{
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		SessionID:   params.SessionID,
		ToolCallID:  params.ToolCallID,
		Fingerprint: params.Fingerprint,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return chatsessionssql.ChatMutationApprovalExecution{}, chatsessions.ErrNotFound
	}
	if err != nil {
		return chatsessionssql.ChatMutationApprovalExecution{}, fmt.Errorf("lock unresolved mutation approval fingerprint: %w", err)
	}
	return execution, nil
}

func terminalizeExpiredReadyAndClaimReplacement(
	ctx context.Context,
	queries chatsessionssql.Querier,
	params chatsessions.MutationApprovalExecutionParams,
	previousSessionID string,
	previousToolCallID string,
	leaseToken uuid.UUID,
) (chatsessionssql.ChatMutationApprovalExecution, error) {
	expiredOutput, err := json.Marshal(map[string]any{
		"error":   expiredApprovalOutputMessage,
		"success": false,
	})
	if err != nil {
		return chatsessionssql.ChatMutationApprovalExecution{}, fmt.Errorf("marshal expired approval output: %w", err)
	}
	return queries.TerminalizeExpiredReadyAndClaimReplacement(ctx, chatsessionssql.TerminalizeExpiredReadyAndClaimReplacementParams{
		ToolCallID:         params.ToolCallID,
		Fingerprint:        params.Fingerprint,
		LeaseToken:         &leaseToken,
		SessionID:          params.SessionID,
		UserID:             params.UserID,
		WorkspaceID:        params.WorkspaceID,
		ExpiredOutput:      expiredOutput,
		PreviousSessionID:  previousSessionID,
		PreviousToolCallID: previousToolCallID,
	})
}

func toClaimedMutationApprovalExecution(execution chatsessionssql.ChatMutationApprovalExecution) (chatsessions.CoreMutationApprovalExecution, error) {
	if (execution.Status != "ready" && execution.Status != "retry_ready") || execution.LeaseToken == nil || execution.LeaseExpiresAt == nil {
		return chatsessions.CoreMutationApprovalExecution{}, errors.New("claimed mutation approval has an invalid lease")
	}
	return chatsessions.CoreMutationApprovalExecution{
		State:          chatsessions.MutationApprovalExecutionClaimed,
		LeaseToken:     execution.LeaseToken,
		LeaseExpiresAt: execution.LeaseExpiresAt,
	}, nil
}

func toCoreMutationApprovalExecution(execution chatsessionssql.ChatMutationApprovalExecution) (chatsessions.CoreMutationApprovalExecution, error) {
	switch execution.Status {
	case "ready", "retry_ready":
		if execution.LeaseExpiresAt == nil {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("ready mutation approval has no lease expiry")
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:          chatsessions.MutationApprovalExecutionReady,
			LeaseExpiresAt: execution.LeaseExpiresAt,
		}, nil
	case "executing":
		if execution.LeaseExpiresAt == nil {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("executing mutation approval has no lease expiry")
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:          chatsessions.MutationApprovalExecutionExecuting,
			LeaseExpiresAt: execution.LeaseExpiresAt,
		}, nil
	case "completed":
		if !json.Valid(execution.Output) {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("completed mutation approval has invalid output")
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:  chatsessions.MutationApprovalExecutionCompleted,
			Output: append(json.RawMessage(nil), execution.Output...),
		}, nil
	case "failed_uncertain":
		if execution.FailureCode == nil {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("uncertain mutation approval has no failure code")
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:       chatsessions.MutationApprovalExecutionFailed,
			FailureCode: *execution.FailureCode,
		}, nil
	default:
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("unknown mutation approval status %q", execution.Status)
	}
}
