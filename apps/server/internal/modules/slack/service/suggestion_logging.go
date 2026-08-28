package slack

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type suggestionDebugInput struct {
	Outcome      string
	Reason       string
	Query        string
	ActionID     string
	SlackTeamID  string
	WorkspaceID  uuid.UUID
	ResolvedTeam uuid.UUID
	ResultCount  int
}

func (s *Service) recordSuggestionDebug(ctx context.Context, payload interactionPayload, input suggestionDebugInput) {
	var workspaceIDPtr *uuid.UUID
	if input.WorkspaceID != uuid.Nil {
		workspaceIDPtr = &input.WorkspaceID
	}
	slackTeamID := optionalString(input.SlackTeamID)
	slackUserID := optionalString(payload.User.ID)
	slackChannelID := optionalString(payload.Channel.ID)
	errorMessage := optionalString(input.Reason)
	logFields := []any{
		"outcome", strings.TrimSpace(input.Outcome),
		"reason", strings.TrimSpace(input.Reason),
		"action_id", strings.TrimSpace(input.ActionID),
		"query_length", len([]rune(strings.TrimSpace(input.Query))),
		"result_count", input.ResultCount,
		"slack_team_id", strings.TrimSpace(input.SlackTeamID),
		"slack_user_id", strings.TrimSpace(payload.User.ID),
		"workspace_id", input.WorkspaceID,
		"resolved_team_id", input.ResolvedTeam,
	}
	if s.log != nil {
		switch {
		case strings.Contains(input.Outcome, "_error_"):
			s.log.Error(ctx, "Slack modal suggestion search failed", logFields...)
		case strings.Contains(input.Outcome, "_skipped_"):
			s.log.Warn(ctx, "Slack modal suggestion search was rejected", logFields...)
		case input.ResultCount == 0:
			s.log.Info(ctx, "Slack modal suggestion search returned no options", logFields...)
		}
	}

	if insertErr := s.repo.InsertRequestLog(ctx, slackRequestLogInsert{
		RequestType:  "suggestion_search",
		Endpoint:     "/integrations/slack/interactivity",
		WorkspaceID:  workspaceIDPtr,
		SlackTeamID:  slackTeamID,
		SlackUserID:  slackUserID,
		SlackChannel: slackChannelID,
		Headers:      []byte("{}"),
		ResponseCode: http.StatusOK,
		Outcome:      truncateForLog(strings.TrimSpace(input.Outcome), 120),
		ErrorMessage: errorMessage,
	}); insertErr != nil {
		s.log.Warn(ctx, "failed writing suggestion diagnostic log", "error", insertErr)
	}
}

func errorString(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	if message := strings.TrimSpace(err.Error()); message != "" {
		return message
	}
	return fallback
}

func suggestionActionID(payload interactionPayload) string {
	if actionID := strings.TrimSpace(payload.ActionID); actionID != "" {
		return actionID
	}
	if len(payload.Actions) > 0 {
		return strings.TrimSpace(payload.Actions[0].ActionID)
	}
	return ""
}

func suggestionBlockID(payload interactionPayload) string {
	if blockID := strings.TrimSpace(payload.BlockID); blockID != "" {
		return blockID
	}
	if len(payload.Actions) > 0 {
		return strings.TrimSpace(payload.Actions[0].BlockID)
	}
	return ""
}

func suggestionQuery(payload interactionPayload) string {
	if query := strings.TrimSpace(payload.Value); query != "" {
		return query
	}
	if len(payload.Actions) > 0 {
		if query := strings.TrimSpace(payload.Actions[0].Value); query != "" {
			return query
		}
	}
	return ""
}
