package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) RecordRequestLog(ctx context.Context, input CoreRequestLogInput) {
	statusCode := input.ResponseCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	requestDetails := parseRequestLogDetails(input.RequestType, input.RawBody)
	workspaceID := s.resolveWorkspaceIDFromLog(ctx, requestDetails.SlackTeamID)

	entry := slackRequestLogInsert{
		RequestType:  input.RequestType,
		Endpoint:     strings.TrimSpace(input.Endpoint),
		WorkspaceID:  workspaceID,
		SlackTeamID:  optionalString(requestDetails.SlackTeamID),
		SlackUserID:  optionalString(requestDetails.SlackUserID),
		SlackChannel: optionalString(requestDetails.SlackChannelID),
		Command:      optionalString(requestDetails.Command),
		Headers:      safeRequestLogHeaders(input.Headers),
		ResponseCode: statusCode,
		Outcome:      truncateForLog(strings.TrimSpace(input.Outcome), 120),
		ErrorMessage: optionalString(truncateForLog(input.ErrorMessage, 1000)),
	}

	if err := s.repo.InsertRequestLog(ctx, entry); err != nil {
		s.log.Warn(ctx, "failed to insert slack request log", "error", err, "request_type", input.RequestType)
	}
}

func safeRequestLogHeaders(headers map[string]string) []byte {
	allowed := map[string]struct{}{
		"X-Slack-Retry-Num":    {},
		"X-Slack-Retry-Reason": {},
		"User-Agent":           {},
		"Content-Type":         {},
	}
	filtered := make(map[string]string, len(allowed))
	for key, value := range headers {
		canonicalKey := http.CanonicalHeaderKey(strings.TrimSpace(key))
		if _, ok := allowed[canonicalKey]; !ok {
			continue
		}
		if value = truncateForLog(strings.TrimSpace(value), 250); value != "" {
			filtered[canonicalKey] = value
		}
	}
	payload, err := json.Marshal(filtered)
	if err != nil {
		return []byte("{}")
	}
	return payload
}

type requestLogDetails struct {
	SlackTeamID    string
	SlackUserID    string
	SlackChannelID string
	Command        string
	TriggerID      string
}

func parseRequestLogDetails(requestType string, rawBody []byte) requestLogDetails {
	switch strings.TrimSpace(requestType) {
	case "commands":
		return parseCommandLogDetails(rawBody)
	case "interactivity":
		return parseInteractivityLogDetails(rawBody)
	case "events":
		return parseEventsLogDetails(rawBody)
	default:
		return requestLogDetails{}
	}
}

func parseCommandLogDetails(rawBody []byte) requestLogDetails {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return requestLogDetails{}
	}
	return requestLogDetails{
		SlackTeamID:    strings.TrimSpace(values.Get("team_id")),
		SlackUserID:    strings.TrimSpace(values.Get("user_id")),
		SlackChannelID: strings.TrimSpace(values.Get("channel_id")),
		Command:        strings.TrimSpace(values.Get("command")),
		TriggerID:      strings.TrimSpace(values.Get("trigger_id")),
	}
}

func parseInteractivityLogDetails(rawBody []byte) requestLogDetails {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return requestLogDetails{}
	}
	payloadText := strings.TrimSpace(values.Get("payload"))
	if payloadText == "" {
		return requestLogDetails{}
	}
	var payload interactionPayload
	if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
		return requestLogDetails{}
	}

	teamID := strings.TrimSpace(payload.Team.ID)
	if teamID == "" {
		if metadata, metadataErr := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata); metadataErr == nil {
			teamID = strings.TrimSpace(metadata.Source.SlackTeamID)
		}
	}
	actionID := suggestionActionID(payload)
	if actionID == "" && len(payload.Actions) > 0 {
		actionID = strings.TrimSpace(payload.Actions[0].ActionID)
	}
	userID := strings.TrimSpace(payload.User.ID)
	if userID == "" {
		userID = strings.TrimSpace(payload.Message.User)
	}

	return requestLogDetails{
		SlackTeamID:    teamID,
		SlackUserID:    userID,
		SlackChannelID: strings.TrimSpace(payload.Channel.ID),
		Command:        actionID,
		TriggerID:      strings.TrimSpace(payload.TriggerID),
	}
}

func parseEventsLogDetails(rawBody []byte) requestLogDetails {
	var payload struct {
		TeamID string `json:"team_id"`
		Event  struct {
			Channel string `json:"channel"`
			User    string `json:"user"`
		} `json:"event"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return requestLogDetails{}
	}

	return requestLogDetails{
		SlackTeamID:    strings.TrimSpace(payload.TeamID),
		SlackUserID:    strings.TrimSpace(payload.Event.User),
		SlackChannelID: strings.TrimSpace(payload.Event.Channel),
	}
}

func (s *Service) resolveWorkspaceIDFromLog(ctx context.Context, slackTeamID string) *uuid.UUID {
	if strings.TrimSpace(slackTeamID) == "" {
		return nil
	}
	workspace, err := s.repo.GetWorkspaceBySlackTeamID(ctx, slackTeamID)
	if err != nil {
		return nil
	}
	return &workspace.ID
}

func truncateForLog(value string, maxLength int) string {
	trimmed := strings.TrimSpace(value)
	if maxLength <= 0 || len(trimmed) <= maxLength {
		return trimmed
	}
	return trimmed[:maxLength]
}
