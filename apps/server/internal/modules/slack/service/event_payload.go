package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	slackEventCallback    = "event_callback"
	slackEventAppMention  = "app_mention"
	slackEventMessage     = "message"
	slackEventUninstalled = "app_uninstalled"
	slackEventTokenRevoke = "tokens_revoked"
)

type slackEventEnvelope struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge"`
	TeamID    string          `json:"team_id"`
	APIAppID  string          `json:"api_app_id"`
	EventID   string          `json:"event_id"`
	EventTime int64           `json:"event_time"`
	Event     slackInnerEvent `json:"event"`
}

type slackInnerEvent struct {
	Type        string `json:"type"`
	Subtype     string `json:"subtype"`
	User        string `json:"user"`
	Text        string `json:"text"`
	Channel     string `json:"channel"`
	ChannelType string `json:"channel_type"`
	TS          string `json:"ts"`
	ThreadTS    string `json:"thread_ts"`
	BotID       string `json:"bot_id"`
	BotProfile  any    `json:"bot_profile"`
	Tokens      struct {
		OAuth []string `json:"oauth"`
		Bot   []string `json:"bot"`
	} `json:"tokens"`
}

type slackEventKind string

const (
	slackEventKindMention     slackEventKind = "mention"
	slackEventKindDirect      slackEventKind = "direct_message"
	slackEventKindUninstalled slackEventKind = "app_uninstalled"
	slackEventKindRevoked     slackEventKind = "tokens_revoked"
)

type normalizedSlackEvent struct {
	EventID             string
	TeamID              string
	Kind                slackEventKind
	UserID              string
	ChannelID           string
	MessageTS           string
	ThreadTS            string
	ReplyTS             string
	Text                string
	RevokedOAuthUserIDs []string
	RevokedBotUserIDs   []string
}

func decodeSlackEvent(rawBody []byte) (slackEventEnvelope, error) {
	decoder := json.NewDecoder(bytes.NewReader(rawBody))
	var envelope slackEventEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return slackEventEnvelope{}, fmt.Errorf("decode slack event: %w", err)
	}
	var trailing json.RawMessage
	err := decoder.Decode(&trailing)
	if err == nil {
		return slackEventEnvelope{}, errors.New("decode slack event: multiple JSON values")
	}
	if !errors.Is(err, io.EOF) {
		return slackEventEnvelope{}, fmt.Errorf("decode slack event trailing content: %w", err)
	}
	envelope.Type = strings.TrimSpace(envelope.Type)
	envelope.TeamID = strings.TrimSpace(envelope.TeamID)
	envelope.EventID = strings.TrimSpace(envelope.EventID)
	return envelope, nil
}

func normalizeSlackEvent(envelope slackEventEnvelope) (normalizedSlackEvent, bool) {
	if envelope.Type != slackEventCallback || envelope.EventID == "" || envelope.TeamID == "" {
		return normalizedSlackEvent{}, false
	}
	event := envelope.Event
	normalized := normalizedSlackEvent{
		EventID:             envelope.EventID,
		TeamID:              envelope.TeamID,
		UserID:              strings.TrimSpace(event.User),
		ChannelID:           strings.TrimSpace(event.Channel),
		MessageTS:           strings.TrimSpace(event.TS),
		ThreadTS:            strings.TrimSpace(event.ThreadTS),
		Text:                strings.TrimSpace(event.Text),
		RevokedOAuthUserIDs: append([]string(nil), event.Tokens.OAuth...),
		RevokedBotUserIDs:   append([]string(nil), event.Tokens.Bot...),
	}
	normalized.ReplyTS = normalized.ThreadTS
	if normalized.ThreadTS == "" {
		normalized.ThreadTS = normalized.MessageTS
	}
	switch strings.TrimSpace(event.Type) {
	case slackEventAppMention:
		normalized.Kind = slackEventKindMention
		normalized.ReplyTS = normalized.ThreadTS
	case slackEventMessage:
		if strings.TrimSpace(event.ChannelType) != "im" {
			return normalizedSlackEvent{}, false
		}
		normalized.Kind = slackEventKindDirect
	case slackEventUninstalled:
		normalized.Kind = slackEventKindUninstalled
		return normalized, true
	case slackEventTokenRevoke:
		normalized.Kind = slackEventKindRevoked
		return normalized, true
	default:
		return normalizedSlackEvent{}, false
	}
	if strings.TrimSpace(event.Subtype) != "" || strings.TrimSpace(event.BotID) != "" || event.BotProfile != nil {
		return normalizedSlackEvent{}, false
	}
	if normalized.UserID == "" || normalized.ChannelID == "" || normalized.MessageTS == "" || normalized.Text == "" {
		return normalizedSlackEvent{}, false
	}
	return normalized, true
}

func removeBotMention(text, botUserID string) string {
	text = strings.TrimSpace(text)
	botUserID = strings.TrimSpace(botUserID)
	if botUserID == "" {
		return text
	}
	return strings.TrimSpace(strings.ReplaceAll(text, "<@"+botUserID+">", ""))
}
