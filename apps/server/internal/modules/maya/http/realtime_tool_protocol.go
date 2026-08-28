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

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
)

func (h *Handlers) validateRealtimeVoiceSession(ctx context.Context, workspaceID, userID, sessionID uuid.UUID) error {
	return h.service.ValidateRealtimeVoiceSession(ctx, workspaceID, userID, sessionID)
}

func (h *Handlers) claimRealtimeToolCall(ctx context.Context, req AppRealtimeToolRequest) (AppRealtimeToolResponse, bool, error) {
	claim, err := h.service.ClaimRealtimeToolCall(ctx, maya.RealtimeToolCallInput{
		SessionID: req.SessionID,
		CallID:    req.CallID,
		ToolName:  req.Name,
		Arguments: req.Arguments,
	})
	if err != nil {
		return AppRealtimeToolResponse{}, false, err
	}
	if claim.Claimed {
		return AppRealtimeToolResponse{}, true, nil
	}

	var response AppRealtimeToolResponse
	if err := json.Unmarshal(claim.Response, &response); err != nil {
		return AppRealtimeToolResponse{}, false, fmt.Errorf("decode existing realtime tool result: %w", err)
	}
	return response, false, nil
}

func (h *Handlers) completeRealtimeToolCall(ctx context.Context, req AppRealtimeToolRequest, response AppRealtimeToolResponse) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode realtime tool result: %w", err)
	}
	return h.service.CompleteRealtimeToolCall(ctx, maya.RealtimeToolCallInput{
		SessionID: req.SessionID,
		CallID:    req.CallID,
		ToolName:  req.Name,
		Arguments: req.Arguments,
	}, payload)
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
