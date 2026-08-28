package maya

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	"github.com/google/uuid"
)

const (
	RealtimeMonthlyVoiceLimit  = 10 * time.Minute
	RealtimeMaxSessionDuration = 5 * time.Minute
)

var ErrRealtimeNotConfigured = mayadomain.ErrRealtimeNotConfigured
var ErrRealtimeMonthlyLimitExceeded = mayadomain.ErrRealtimeMonthlyLimitExceeded
var ErrRealtimeSessionInactive = mayadomain.ErrRealtimeSessionInactive
var ErrRealtimeToolCallConflict = mayadomain.ErrRealtimeToolCallConflict
var ErrRealtimeToolCallInProgress = mayadomain.ErrRealtimeToolCallInProgress
var ErrRealtimeToolCallCompletionConflict = mayadomain.ErrRealtimeToolCallCompletionConflict

type RealtimeVoiceSessionLease = mayadomain.RealtimeVoiceSessionLease

type RealtimeToolCallInput struct {
	SessionID uuid.UUID
	CallID    string
	ToolName  string
	Arguments json.RawMessage
}

type RealtimeToolCallClaim = mayadomain.RealtimeToolCallClaim

type RealtimeRepository interface {
	ReserveRealtimeVoiceSession(
		ctx context.Context,
		workspaceID, userID uuid.UUID,
		monthlyLimit, maxSessionDuration time.Duration,
	) (RealtimeVoiceSessionLease, error)
	EndRealtimeVoiceSession(
		ctx context.Context,
		workspaceID, userID, sessionID uuid.UUID,
		maxSessionDuration time.Duration,
	) error
	RealtimeVoiceSessionIsActive(
		ctx context.Context,
		workspaceID, userID, sessionID uuid.UUID,
		maxSessionDuration time.Duration,
	) (bool, error)
	ClaimRealtimeToolCall(
		ctx context.Context,
		sessionID uuid.UUID,
		callID, toolName, requestHash string,
	) (RealtimeToolCallClaim, error)
	CompleteRealtimeToolCall(
		ctx context.Context,
		sessionID uuid.UUID,
		callID string,
		response json.RawMessage,
	) error
}

func (s *Service) WorkspaceCanUseMaya(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	repository, err := s.scheduleRepository()
	if err != nil {
		return false, err
	}
	return repository.WorkspaceCanUseMaya(ctx, workspaceID)
}

func (s *Service) BeginRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
) (RealtimeVoiceSessionLease, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return RealtimeVoiceSessionLease{}, fmt.Errorf("%w: workspace and user are required", ErrInvalidPlanInput)
	}
	if s.realtime == nil {
		return RealtimeVoiceSessionLease{}, ErrRealtimeNotConfigured
	}
	return s.realtime.ReserveRealtimeVoiceSession(
		ctx,
		workspaceID,
		userID,
		RealtimeMonthlyVoiceLimit,
		RealtimeMaxSessionDuration,
	)
}

func (s *Service) EndRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID, sessionID uuid.UUID,
) error {
	if workspaceID == uuid.Nil || userID == uuid.Nil || sessionID == uuid.Nil {
		return fmt.Errorf("%w: workspace, user, and session are required", ErrInvalidPlanInput)
	}
	if s.realtime == nil {
		return ErrRealtimeNotConfigured
	}
	return s.realtime.EndRealtimeVoiceSession(
		ctx,
		workspaceID,
		userID,
		sessionID,
		RealtimeMaxSessionDuration,
	)
}

func (s *Service) ValidateRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID, sessionID uuid.UUID,
) error {
	if workspaceID == uuid.Nil || userID == uuid.Nil || sessionID == uuid.Nil {
		return ErrRealtimeSessionInactive
	}
	if s.realtime == nil {
		return ErrRealtimeNotConfigured
	}
	active, err := s.realtime.RealtimeVoiceSessionIsActive(
		ctx,
		workspaceID,
		userID,
		sessionID,
		RealtimeMaxSessionDuration,
	)
	if err != nil {
		return err
	}
	if !active {
		return ErrRealtimeSessionInactive
	}
	return nil
}

func (s *Service) ClaimRealtimeToolCall(
	ctx context.Context,
	input RealtimeToolCallInput,
) (RealtimeToolCallClaim, error) {
	callID := strings.TrimSpace(input.CallID)
	toolName := strings.TrimSpace(input.ToolName)
	if input.SessionID == uuid.Nil || callID == "" || toolName == "" {
		return RealtimeToolCallClaim{}, fmt.Errorf("%w: session, call, and tool are required", ErrInvalidPlanInput)
	}
	if s.realtime == nil {
		return RealtimeToolCallClaim{}, ErrRealtimeNotConfigured
	}
	return s.realtime.ClaimRealtimeToolCall(
		ctx,
		input.SessionID,
		callID,
		toolName,
		realtimeToolRequestHash(toolName, input.Arguments),
	)
}

func (s *Service) CompleteRealtimeToolCall(
	ctx context.Context,
	input RealtimeToolCallInput,
	response json.RawMessage,
) error {
	callID := strings.TrimSpace(input.CallID)
	if input.SessionID == uuid.Nil || callID == "" || !json.Valid(response) {
		return fmt.Errorf("%w: session, call, and valid response are required", ErrInvalidPlanInput)
	}
	if s.realtime == nil {
		return ErrRealtimeNotConfigured
	}
	return s.realtime.CompleteRealtimeToolCall(ctx, input.SessionID, callID, response)
}

func realtimeToolRequestHash(toolName string, arguments json.RawMessage) string {
	digest := sha256.New()
	_, _ = digest.Write([]byte(strings.TrimSpace(toolName)))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write(arguments)
	return hex.EncodeToString(digest.Sum(nil))
}
