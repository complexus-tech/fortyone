package slack

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (p *EventProcessor) logAssistantResponseError(
	ctx context.Context,
	err error,
	workspaceID, userID, inboundEventID uuid.UUID,
	attemptCount int,
	event normalizedSlackEvent,
) {
	if p.log == nil || err == nil {
		return
	}

	classification := "retryable"
	switch {
	case errors.Is(err, errAssistantNotConfigured):
		classification = "not_configured"
	case isPermanentAssistantProviderError(err):
		classification = "permanent_provider_error"
	}
	fields := []any{
		"error", err,
		"classification", classification,
		"workspace_id", workspaceID,
		"user_id", userID,
		"inbound_event_id", inboundEventID,
		"attempt_count", attemptCount,
		"slack_event_id", event.EventID,
		"slack_team_id", event.TeamID,
		"slack_channel_id", event.ChannelID,
	}
	var apiError *assistantAPIError
	if errors.As(err, &apiError) && apiError != nil {
		fields = append(fields,
			"openai_status_code", apiError.StatusCode,
			"openai_error_code", strings.TrimSpace(apiError.Code),
			"openai_request_id", strings.TrimSpace(apiError.RequestID),
		)
	}
	p.log.Error(context.WithoutCancel(ctx), "Slack Maya assistant response failed", fields...)
}
