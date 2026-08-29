package slack

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	slackAssistantStatusUsername = "Maya"
	slackAssistantThinkingStatus = "is thinking..."
	slackAssistantStatusTimeout  = 2 * time.Second
)

// SlackAssistantStatusSetter exposes Slack's native assistant status without
// coupling event processing tests to the Web API client.
type SlackAssistantStatusSetter interface {
	SetStatus(ctx context.Context, botToken, channelID, threadTS, status string) error
}

type slackAssistantStatusSetter struct {
	client *slackWebClient
}

func (s *slackAssistantStatusSetter) SetStatus(
	ctx context.Context,
	botToken, channelID, threadTS, status string,
) error {
	if s == nil || s.client == nil {
		return errors.New("slack assistant status client is not configured")
	}
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	if channelID == "" || threadTS == "" {
		return errors.New("slack assistant status requires a channel and thread")
	}
	return s.client.callJSON(ctx, botToken, "assistant.threads.setStatus", map[string]any{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"status":     status,
		"username":   slackAssistantStatusUsername,
	}, nil)
}

func (p *EventProcessor) startAssistantThinkingStatus(
	ctx context.Context,
	event normalizedSlackEvent,
	botToken string,
) bool {
	if p == nil || p.statusSetter == nil || strings.TrimSpace(event.ChannelID) == "" || strings.TrimSpace(event.ThreadTS) == "" {
		return false
	}
	statusCtx, cancel := context.WithTimeout(ctx, slackAssistantStatusTimeout)
	defer cancel()
	if err := p.statusSetter.SetStatus(statusCtx, botToken, event.ChannelID, event.ThreadTS, slackAssistantThinkingStatus); err != nil {
		if p.log != nil {
			p.log.Warn(ctx, "failed setting Slack Maya thinking status",
				"error", err,
				"slack_event_id", event.EventID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
			)
		}
		return false
	}
	return true
}

func (p *EventProcessor) clearAssistantThinkingStatus(ctx context.Context, event normalizedSlackEvent, botToken string) bool {
	if p == nil || p.statusSetter == nil || strings.TrimSpace(event.ChannelID) == "" || strings.TrimSpace(event.ThreadTS) == "" {
		return false
	}
	statusCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), slackAssistantStatusTimeout)
	defer cancel()
	if err := p.statusSetter.SetStatus(statusCtx, botToken, event.ChannelID, event.ThreadTS, ""); err != nil {
		if p.log != nil {
			p.log.Warn(statusCtx, "failed clearing Slack Maya thinking status",
				"error", err,
				"slack_event_id", event.EventID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
			)
		}
		return false
	}
	return true
}
