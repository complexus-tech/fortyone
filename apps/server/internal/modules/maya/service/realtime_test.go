package maya

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRealtimeRepository struct {
	reserve  func(context.Context, uuid.UUID, uuid.UUID, time.Duration, time.Duration) (RealtimeVoiceSessionLease, error)
	end      func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Duration) error
	active   func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Duration) (bool, error)
	claim    func(context.Context, uuid.UUID, string, string, string) (RealtimeToolCallClaim, error)
	complete func(context.Context, uuid.UUID, string, json.RawMessage) error
}

func (repository fakeRealtimeRepository) ReserveRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	monthlyLimit, maxSessionDuration time.Duration,
) (RealtimeVoiceSessionLease, error) {
	return repository.reserve(ctx, workspaceID, userID, monthlyLimit, maxSessionDuration)
}

func (repository fakeRealtimeRepository) EndRealtimeVoiceSession(
	ctx context.Context,
	workspaceID, userID, sessionID uuid.UUID,
	maxSessionDuration time.Duration,
) error {
	if repository.end == nil {
		return nil
	}
	return repository.end(ctx, workspaceID, userID, sessionID, maxSessionDuration)
}

func (repository fakeRealtimeRepository) RealtimeVoiceSessionIsActive(
	ctx context.Context,
	workspaceID, userID, sessionID uuid.UUID,
	maxSessionDuration time.Duration,
) (bool, error) {
	return repository.active(ctx, workspaceID, userID, sessionID, maxSessionDuration)
}

func (repository fakeRealtimeRepository) ClaimRealtimeToolCall(
	ctx context.Context,
	sessionID uuid.UUID,
	callID, toolName, requestHash string,
) (RealtimeToolCallClaim, error) {
	return repository.claim(ctx, sessionID, callID, toolName, requestHash)
}

func (repository fakeRealtimeRepository) CompleteRealtimeToolCall(
	ctx context.Context,
	sessionID uuid.UUID,
	callID string,
	response json.RawMessage,
) error {
	return repository.complete(ctx, sessionID, callID, response)
}

func TestBeginRealtimeVoiceSessionUsesDomainQuota(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	want := RealtimeVoiceSessionLease{
		SessionID: uuid.New(), MaxDuration: 2 * time.Minute, RemainingDuration: 7 * time.Minute,
	}
	service := &Service{realtime: fakeRealtimeRepository{
		reserve: func(
			_ context.Context,
			gotWorkspaceID, gotUserID uuid.UUID,
			monthlyLimit, maxSessionDuration time.Duration,
		) (RealtimeVoiceSessionLease, error) {
			if gotWorkspaceID != workspaceID || gotUserID != userID {
				t.Fatalf("reservation subject = %s/%s, want %s/%s", gotWorkspaceID, gotUserID, workspaceID, userID)
			}
			if monthlyLimit != RealtimeMonthlyVoiceLimit || maxSessionDuration != RealtimeMaxSessionDuration {
				t.Fatalf("reservation quota = %s/%s", monthlyLimit, maxSessionDuration)
			}
			return want, nil
		},
	}}

	got, err := service.BeginRealtimeVoiceSession(t.Context(), workspaceID, userID)
	if err != nil || got != want {
		t.Fatalf("begin realtime session = %+v, %v; want %+v", got, err, want)
	}
}

func TestValidateRealtimeVoiceSessionMapsInactiveState(t *testing.T) {
	t.Parallel()

	service := &Service{realtime: fakeRealtimeRepository{
		active: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Duration) (bool, error) {
			return false, nil
		},
	}}
	err := service.ValidateRealtimeVoiceSession(t.Context(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRealtimeSessionInactive) {
		t.Fatalf("validate inactive realtime session error = %v", err)
	}
}

func TestRealtimeToolCallCanonicalizesIdentityAndHashesExactArguments(t *testing.T) {
	t.Parallel()

	sessionID := uuid.New()
	arguments := json.RawMessage(`{"teamId":"alpha"}`)
	var observedHash string
	service := &Service{realtime: fakeRealtimeRepository{
		claim: func(
			_ context.Context,
			gotSessionID uuid.UUID,
			callID, toolName, requestHash string,
		) (RealtimeToolCallClaim, error) {
			if gotSessionID != sessionID || callID != "call-1" || toolName != "list_teams" {
				t.Fatalf("canonical claim = %s/%q/%q", gotSessionID, callID, toolName)
			}
			observedHash = requestHash
			return RealtimeToolCallClaim{Claimed: true}, nil
		},
	}}

	claim, err := service.ClaimRealtimeToolCall(t.Context(), RealtimeToolCallInput{
		SessionID: sessionID,
		CallID:    " call-1 ",
		ToolName:  " list_teams ",
		Arguments: arguments,
	})
	if err != nil || !claim.Claimed {
		t.Fatalf("claim realtime tool call = %+v, %v", claim, err)
	}
	if observedHash == "" || observedHash != realtimeToolRequestHash("list_teams", arguments) {
		t.Fatalf("request hash = %q", observedHash)
	}
	if observedHash == realtimeToolRequestHash("list_teams", json.RawMessage(`{"teamId":"beta"}`)) {
		t.Fatal("request hash must bind the exact arguments")
	}
}

func TestCompleteRealtimeToolCallRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	service := &Service{realtime: fakeRealtimeRepository{
		complete: func(context.Context, uuid.UUID, string, json.RawMessage) error {
			t.Fatal("invalid response must not reach persistence")
			return nil
		},
	}}
	err := service.CompleteRealtimeToolCall(
		t.Context(),
		RealtimeToolCallInput{SessionID: uuid.New(), CallID: "call-1"},
		json.RawMessage(`{"broken":`),
	)
	if !errors.Is(err, ErrInvalidPlanInput) {
		t.Fatalf("complete invalid realtime tool call error = %v", err)
	}
}
