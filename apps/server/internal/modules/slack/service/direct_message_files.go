package slack

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const slackDirectMessageFilePageLimit = 15

// loadSlackDirectMessageFiles looks only at recent messages through the current
// unthreaded DM. Each file keeps its own source timestamp for import checks.
func (p *EventProcessor) loadSlackDirectMessageFiles(
	ctx context.Context,
	botToken string,
	event normalizedSlackEvent,
) ([]slackMessageFile, error) {
	if p == nil || p.webClient == nil {
		return nil, errors.New("slack direct message reader is not configured")
	}
	if event.Kind != slackEventKindDirect || event.ReplyTS != "" ||
		strings.TrimSpace(event.ChannelID) == "" || strings.TrimSpace(event.MessageTS) == "" ||
		strings.TrimSpace(event.UserID) == "" {
		return nil, errors.New("slack direct message source is incomplete")
	}

	query := url.Values{
		"channel":   {event.ChannelID},
		"latest":    {event.MessageTS},
		"inclusive": {"true"},
		"limit":     {strconv.Itoa(slackDirectMessageFilePageLimit)},
	}
	var response struct {
		Messages []slackThreadMessage `json:"messages"`
	}
	if err := p.webClient.callJSON(ctx, botToken, "conversations.history?"+query.Encode(), nil, &response); err != nil {
		return nil, fmt.Errorf("read recent Slack direct message files: %w", err)
	}

	files := make([]slackMessageFile, 0)
	for _, message := range response.Messages {
		message.TS = strings.TrimSpace(message.TS)
		message.ThreadTS = strings.TrimSpace(message.ThreadTS)
		message.UserID = strings.TrimSpace(message.UserID)
		message.Subtype = strings.TrimSpace(message.Subtype)
		message.BotID = strings.TrimSpace(message.BotID)
		message.AppID = strings.TrimSpace(message.AppID)
		if message.UserID != event.UserID || len(message.Files) == 0 ||
			!supportedSlackThreadMessage(message, "") {
			continue
		}
		threadTS := message.ThreadTS
		if threadTS == "" {
			threadTS = message.TS
		}
		for _, file := range message.Files {
			file.MessageTS = message.TS
			file.ThreadTS = threadTS
			files = append(files, file)
		}
	}
	return files, nil
}
