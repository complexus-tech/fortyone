package mayadomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMayaAccessDenied                   = errors.New("workspace does not have access to Maya auto-scheduling")
	ErrRealtimeInvalidInput               = errors.New("invalid Maya realtime persistence input")
	ErrRealtimeNotConfigured              = errors.New("maya realtime persistence is not configured")
	ErrRealtimeMonthlyLimitExceeded       = errors.New("monthly realtime voice limit reached")
	ErrRealtimeSessionInactive            = errors.New("realtime voice session is not active")
	ErrRealtimeToolCallConflict           = errors.New("realtime tool call conflicts with an existing call")
	ErrRealtimeToolCallInProgress         = errors.New("realtime tool call is already processing")
	ErrRealtimeToolCallCompletionConflict = errors.New("realtime tool call was not available to complete")
)

type RealtimeVoiceSessionLease struct {
	SessionID         uuid.UUID
	MaxDuration       time.Duration
	RemainingDuration time.Duration
}

type RealtimeToolCallClaim struct {
	Response json.RawMessage
	Claimed  bool
}
