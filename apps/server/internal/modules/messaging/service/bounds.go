package messaging

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	// MaximumMessageBytes bounds each current prompt and retained conversation
	// turn before it can reach an assistant provider. Bytes are intentionally
	// used instead of characters because they form a conservative, tokenizer-
	// independent cost bound for UTF-8 input.
	MaximumMessageBytes = 16 << 10

	// MaximumConversationBytes bounds retained history, excluding the current
	// prompt. The newest coherent history is preserved when older turns exceed
	// either this aggregate bound or MaximumConversationTurns.
	MaximumConversationBytes = 48 << 10

	MaximumConversationTurns = 20

	// MaximumGuidanceRunes bounds trusted workspace-admin instructions appended
	// to the assistant's base instructions for a single request.
	MaximumGuidanceRunes = 4_000
)

// NormalizeRequest validates the authoritative identity and current prompt,
// then returns a copy with history bounded from newest to oldest. Oversized
// historical turns are treated as a context boundary: the turn and everything
// older than it are omitted. The current prompt is never silently truncated.
func NormalizeRequest(request Request) (Request, error) {
	if request.WorkspaceID == uuid.Nil || request.UserID == uuid.Nil {
		return Request{}, fmt.Errorf("%w: workspace and user are required", ErrInvalidRequest)
	}
	if err := ValidatePrompt(request.Prompt); err != nil {
		return Request{}, err
	}

	for index, turn := range request.Conversation {
		if turn.Role != RoleUser && turn.Role != RoleAssistant {
			return Request{}, fmt.Errorf("%w: turn %d has unsupported role %q", ErrInvalidRequest, index, turn.Role)
		}
		if strings.TrimSpace(turn.Text) == "" {
			return Request{}, fmt.Errorf("%w: turn %d has empty text", ErrInvalidRequest, index)
		}
	}
	allowedTeamIDs, err := normalizedAllowedTeamIDs(request.AllowedTeamIDs)
	if err != nil {
		return Request{}, err
	}
	sharedTeamIDs, err := normalizedSharedTeamIDs(request.SharedTeamIDs, allowedTeamIDs)
	if err != nil {
		return Request{}, err
	}
	runtimeContext, err := normalizeRuntimeContext(request.RuntimeContext)
	if err != nil {
		return Request{}, err
	}

	request.AllowedTeamIDs = allowedTeamIDs
	request.SharedTeamIDs = sharedTeamIDs
	request.RuntimeContext = runtimeContext
	request.Guidance = strings.TrimSpace(request.Guidance)
	if guidanceRunes := len([]rune(request.Guidance)); guidanceRunes > MaximumGuidanceRunes {
		return Request{}, fmt.Errorf("%w: workspace guidance is %d characters; maximum is %d", ErrInvalidRequest, guidanceRunes, MaximumGuidanceRunes)
	}
	request.Conversation = newestBoundedConversation(request.Conversation)
	return request, nil
}

func normalizedSharedTeamIDs(teamIDs, allowedTeamIDs []uuid.UUID) ([]uuid.UUID, error) {
	if teamIDs == nil {
		return nil, nil
	}

	result := make([]uuid.UUID, 0, len(teamIDs))
	seen := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == uuid.Nil {
			return nil, fmt.Errorf("%w: shared team IDs must not contain a nil UUID", ErrInvalidRequest)
		}
		if _, duplicate := seen[teamID]; duplicate {
			continue
		}
		seen[teamID] = struct{}{}
		result = append(result, teamID)
	}
	if len(result) == 0 {
		return result, nil
	}

	allowed := make(map[uuid.UUID]struct{}, len(allowedTeamIDs))
	for _, teamID := range allowedTeamIDs {
		allowed[teamID] = struct{}{}
	}
	for _, teamID := range result {
		if _, ok := allowed[teamID]; !ok {
			return nil, fmt.Errorf("%w: shared team %s is outside the allowed team scope", ErrInvalidRequest, teamID)
		}
	}
	return result, nil
}

func normalizedAllowedTeamIDs(teamIDs []uuid.UUID) ([]uuid.UUID, error) {
	if teamIDs == nil {
		return nil, nil
	}

	result := make([]uuid.UUID, 0, len(teamIDs))
	seen := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == uuid.Nil {
			return nil, fmt.Errorf("%w: allowed team IDs must not contain a nil UUID", ErrInvalidRequest)
		}
		if _, duplicate := seen[teamID]; duplicate {
			continue
		}
		seen[teamID] = struct{}{}
		result = append(result, teamID)
	}
	return result, nil
}

func cloneOptionalUUIDs(values []uuid.UUID) []uuid.UUID {
	if values == nil {
		return nil
	}
	return append(make([]uuid.UUID, 0, len(values)), values...)
}

// ValidatePrompt validates a current provider message before adapters persist
// it or call an assistant. Adapters can use this to produce a durable friendly
// response for oversized input without retrying a deterministic failure.
func ValidatePrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidRequest)
	}
	if len(prompt) > MaximumMessageBytes {
		return fmt.Errorf(
			"%w: %w: prompt is %d bytes; maximum is %d",
			ErrInvalidRequest,
			ErrMessageTooLarge,
			len(prompt),
			MaximumMessageBytes,
		)
	}
	return nil
}

func newestBoundedConversation(turns []ConversationTurn) []ConversationTurn {
	if len(turns) == 0 {
		return nil
	}

	start := len(turns)
	totalBytes := 0
	for index := len(turns) - 1; index >= 0; index-- {
		turnBytes := len(turns[index].Text)
		if turnBytes > MaximumMessageBytes || totalBytes+turnBytes > MaximumConversationBytes {
			break
		}
		if len(turns)-index > MaximumConversationTurns {
			break
		}
		totalBytes += turnBytes
		start = index
	}

	for start < len(turns) && turns[start].Role == RoleAssistant {
		start++
	}
	if start == len(turns) {
		return nil
	}
	return append([]ConversationTurn(nil), turns[start:]...)
}
