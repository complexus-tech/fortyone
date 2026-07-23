package mayahttp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (h *Handlers) validateRealtimeVoiceSession(ctx context.Context, workspaceID, userID, sessionID uuid.UUID) error {
	var active bool
	if err := h.db.GetContext(ctx, &active, `
		SELECT EXISTS (
			SELECT 1
			FROM maya_realtime_voice_sessions
			WHERE session_id = $1
				AND workspace_id = $2
				AND user_id = $3
				AND ended_at IS NULL
				AND started_at + ($4 * INTERVAL '1 second') > NOW()
		)
	`, sessionID, workspaceID, userID, durationSeconds(realtimeMaxSessionDuration)); err != nil {
		return fmt.Errorf("validate realtime voice session: %w", err)
	}
	if !active {
		return ErrMayaRealtimeSessionInactive
	}
	return nil
}

func (h *Handlers) claimRealtimeToolCall(ctx context.Context, req AppRealtimeToolRequest) (AppRealtimeToolResponse, bool, error) {
	requestHash := realtimeToolRequestHash(req)
	result, err := h.db.ExecContext(ctx, `
		INSERT INTO maya_realtime_voice_tool_calls (
			session_id,
			call_id,
			tool_name,
			request_hash
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (session_id, call_id) DO NOTHING
	`, req.SessionID, strings.TrimSpace(req.CallID), strings.TrimSpace(req.Name), requestHash)
	if err != nil {
		return AppRealtimeToolResponse{}, false, fmt.Errorf("claim realtime tool call: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return AppRealtimeToolResponse{}, false, fmt.Errorf("read realtime tool claim result: %w", err)
	}
	if rowsAffected == 1 {
		return AppRealtimeToolResponse{}, true, nil
	}

	var existing struct {
		RequestHash string `db:"request_hash"`
		Response    []byte `db:"response"`
	}
	if err := h.db.GetContext(ctx, &existing, `
		SELECT request_hash, response
		FROM maya_realtime_voice_tool_calls
		WHERE session_id = $1 AND call_id = $2
	`, req.SessionID, strings.TrimSpace(req.CallID)); err != nil {
		return AppRealtimeToolResponse{}, false, fmt.Errorf("read existing realtime tool call: %w", err)
	}
	if !hmac.Equal([]byte(existing.RequestHash), []byte(requestHash)) {
		return AppRealtimeToolResponse{}, false, ErrMayaRealtimeToolCallConflict
	}
	if len(existing.Response) == 0 {
		return AppRealtimeToolResponse{}, false, ErrMayaRealtimeToolCallInProgress
	}

	var response AppRealtimeToolResponse
	if err := json.Unmarshal(existing.Response, &response); err != nil {
		return AppRealtimeToolResponse{}, false, fmt.Errorf("decode existing realtime tool result: %w", err)
	}
	return response, false, nil
}

func (h *Handlers) completeRealtimeToolCall(ctx context.Context, req AppRealtimeToolRequest, response AppRealtimeToolResponse) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode realtime tool result: %w", err)
	}
	result, err := h.db.ExecContext(ctx, `
		UPDATE maya_realtime_voice_tool_calls
		SET response = $3::jsonb,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE session_id = $1
			AND call_id = $2
			AND response IS NULL
	`, req.SessionID, strings.TrimSpace(req.CallID), string(payload))
	if err != nil {
		return fmt.Errorf("complete realtime tool call: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read realtime tool completion result: %w", err)
	}
	if rowsAffected != 1 {
		return errors.New("realtime tool call was not available to complete")
	}
	return nil
}

func realtimeToolRequestHash(req AppRealtimeToolRequest) string {
	hash := sha256.New()
	hash.Write([]byte(strings.TrimSpace(req.Name)))
	hash.Write([]byte{0})
	hash.Write(req.Arguments)
	return hex.EncodeToString(hash.Sum(nil))
}

func (h *Handlers) confirmationToken(sessionID uuid.UUID, toolName string, normalizedArguments any) (string, error) {
	if strings.TrimSpace(h.secretKey) == "" {
		return "", errors.New("realtime confirmation signing is not configured")
	}
	payload, err := json.Marshal(normalizedArguments)
	if err != nil {
		return "", fmt.Errorf("encode realtime confirmation: %w", err)
	}
	mac := hmac.New(sha256.New, []byte(h.secretKey))
	mac.Write([]byte(sessionID.String()))
	mac.Write([]byte{0})
	mac.Write([]byte(toolName))
	mac.Write([]byte{0})
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (h *Handlers) validateConfirmationToken(sessionID uuid.UUID, toolName string, normalizedArguments any, token string) (bool, error) {
	expected, err := h.confirmationToken(sessionID, toolName, normalizedArguments)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(token))), nil
}
