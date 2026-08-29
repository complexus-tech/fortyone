package mayarepository

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ReserveRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	monthlyLimit, maxSessionDuration time.Duration,
) (mayadomain.RealtimeVoiceSessionLease, error) {
	if err := r.validateRealtimeConfiguration(); err != nil {
		return mayadomain.RealtimeVoiceSessionLease{}, err
	}
	if workspaceID == uuid.Nil || userID == uuid.Nil || monthlyLimit <= 0 || maxSessionDuration <= 0 {
		return mayadomain.RealtimeVoiceSessionLease{}, mayadomain.ErrRealtimeInvalidInput
	}

	var lease mayadomain.RealtimeVoiceSessionLease
	err := r.withinTransaction(
		ctx,
		pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite},
		func(queries mayasql.Querier) error {
			if _, err := queries.LockWorkspaceForRealtimeVoice(ctx, mayasql.LockWorkspaceForRealtimeVoiceParams{
				WorkspaceID: workspaceID,
			}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return mayadomain.ErrRealtimeSessionInactive
				}
				return fmt.Errorf("lock workspace for realtime voice session: %w", err)
			}
			hasAccess, err := queries.WorkspaceCanUseMaya(ctx, mayasql.WorkspaceCanUseMayaParams{
				WorkspaceID: workspaceID,
			})
			if err != nil {
				return fmt.Errorf("check Maya access for realtime voice session: %w", err)
			}
			if !hasAccess {
				return mayadomain.ErrMayaAccessDenied
			}

			usedSeconds, err := queries.RealtimeVoiceMonthlyUsageSeconds(
				ctx,
				mayasql.RealtimeVoiceMonthlyUsageSecondsParams{
					MaxSessionSeconds: maxSessionDuration.Seconds(),
					WorkspaceID:       workspaceID,
				},
			)
			if err != nil {
				return fmt.Errorf("read realtime voice monthly usage: %w", err)
			}
			if usedSeconds < 0 {
				return errors.New("realtime voice monthly usage is negative")
			}
			monthlyLimitSeconds := int64(monthlyLimit / time.Second)
			if usedSeconds >= monthlyLimitSeconds {
				return mayadomain.ErrRealtimeMonthlyLimitExceeded
			}

			remainingDuration := time.Duration(monthlyLimitSeconds-usedSeconds) * time.Second
			sessionID, err := queries.CreateRealtimeVoiceSession(ctx, mayasql.CreateRealtimeVoiceSessionParams{
				WorkspaceID: workspaceID,
				UserID:      userID,
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return mayadomain.ErrMayaAccessDenied
				}
				return fmt.Errorf("create realtime voice session record: %w", err)
			}
			lease = mayadomain.RealtimeVoiceSessionLease{
				SessionID:         sessionID,
				MaxDuration:       min(remainingDuration, maxSessionDuration),
				RemainingDuration: remainingDuration,
			}
			return nil
		},
	)
	if err != nil {
		return mayadomain.RealtimeVoiceSessionLease{}, fmt.Errorf("reserve realtime voice session transaction: %w", err)
	}
	return lease, nil
}

func (r *Repo) EndRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID, sessionID uuid.UUID,
	maxSessionDuration time.Duration,
) error {
	if err := r.validateRealtimeConfiguration(); err != nil {
		return err
	}
	_, err := r.queries.EndRealtimeVoiceSession(ctx, mayasql.EndRealtimeVoiceSessionParams{
		MaxSessionSeconds: maxSessionDuration.Seconds(),
		WorkspaceID:       workspaceID,
		UserID:            userID,
		SessionID:         sessionID,
	})
	if err != nil {
		return fmt.Errorf("end realtime voice session: %w", err)
	}
	return nil
}

func (r *Repo) RealtimeVoiceSessionIsActive(
	ctx context.Context,
	workspaceID, userID, sessionID uuid.UUID,
	maxSessionDuration time.Duration,
) (bool, error) {
	if err := r.validateRealtimeConfiguration(); err != nil {
		return false, err
	}
	active, err := r.queries.RealtimeVoiceSessionIsActive(
		ctx,
		mayasql.RealtimeVoiceSessionIsActiveParams{
			SessionID:         sessionID,
			WorkspaceID:       workspaceID,
			UserID:            userID,
			MaxSessionSeconds: maxSessionDuration.Seconds(),
		},
	)
	if err != nil {
		return false, fmt.Errorf("validate realtime voice session: %w", err)
	}
	return active, nil
}

func (r *Repo) ClaimRealtimeToolCall(
	ctx context.Context,
	sessionID uuid.UUID,
	callID, toolName, requestHash string,
) (mayadomain.RealtimeToolCallClaim, error) {
	if err := r.validateRealtimeConfiguration(); err != nil {
		return mayadomain.RealtimeToolCallClaim{}, err
	}
	rowsAffected, err := r.queries.ClaimRealtimeToolCall(ctx, mayasql.ClaimRealtimeToolCallParams{
		SessionID:   sessionID,
		CallID:      callID,
		ToolName:    toolName,
		RequestHash: requestHash,
	})
	if err != nil {
		return mayadomain.RealtimeToolCallClaim{}, fmt.Errorf("claim realtime tool call: %w", err)
	}
	if rowsAffected == 1 {
		return mayadomain.RealtimeToolCallClaim{Claimed: true}, nil
	}

	existing, err := r.queries.GetRealtimeToolCall(ctx, mayasql.GetRealtimeToolCallParams{
		SessionID: sessionID,
		CallID:    callID,
	})
	if err != nil {
		return mayadomain.RealtimeToolCallClaim{}, fmt.Errorf("read existing realtime tool call: %w", err)
	}
	if !hmac.Equal([]byte(existing.RequestHash), []byte(requestHash)) {
		return mayadomain.RealtimeToolCallClaim{}, mayadomain.ErrRealtimeToolCallConflict
	}
	if len(existing.Response) == 0 {
		return mayadomain.RealtimeToolCallClaim{}, mayadomain.ErrRealtimeToolCallInProgress
	}
	return mayadomain.RealtimeToolCallClaim{
		Response: append(json.RawMessage(nil), existing.Response...),
	}, nil
}

func (r *Repo) CompleteRealtimeToolCall(
	ctx context.Context,
	sessionID uuid.UUID,
	callID string,
	response json.RawMessage,
) error {
	if err := r.validateRealtimeConfiguration(); err != nil {
		return err
	}
	rowsAffected, err := r.queries.CompleteRealtimeToolCall(ctx, mayasql.CompleteRealtimeToolCallParams{
		Response:  response,
		SessionID: sessionID,
		CallID:    callID,
	})
	if err != nil {
		return fmt.Errorf("complete realtime tool call: %w", err)
	}
	if rowsAffected != 1 {
		return mayadomain.ErrRealtimeToolCallCompletionConflict
	}
	return nil
}

func (r *Repo) validateRealtimeConfiguration() error {
	if r == nil || r.queries == nil || r.withinTransaction == nil {
		return mayadomain.ErrRealtimeNotConfigured
	}
	return nil
}
