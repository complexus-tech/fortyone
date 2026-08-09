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
	slackEventCallback               = "event_callback"
	slackEventAppMention             = "app_mention"
	slackEventMessage                = "message"
	slackEventUninstalled            = "app_uninstalled"
	slackEventTokenRevoke            = "tokens_revoked"
	slackEventLinkShared             = "link_shared"
	slackEventEntityDetailsRequested = "entity_details_requested"
)

type slackEventEnvelope struct {
	Type               string          `json:"type"`
	Challenge          string          `json:"challenge"`
	TeamID             string          `json:"team_id"`
	APIAppID           string          `json:"api_app_id"`
	EventID            string          `json:"event_id"`
	EventTime          int64           `json:"event_time"`
	IsExtSharedChannel bool            `json:"is_ext_shared_channel"`
	Event              slackInnerEvent `json:"event"`
}

type slackInnerEvent struct {
	Type         string                `json:"type"`
	Subtype      string                `json:"subtype"`
	User         string                `json:"user"`
	Text         string                `json:"text"`
	Channel      string                `json:"channel"`
	ChannelType  string                `json:"channel_type"`
	TS           string                `json:"ts"`
	MessageTS    string                `json:"message_ts"`
	ThreadTS     string                `json:"thread_ts"`
	Source       string                `json:"source"`
	TriggerID    string                `json:"trigger_id"`
	EntityURL    string                `json:"entity_url"`
	AppUnfurlURL string                `json:"app_unfurl_url"`
	ExternalRef  slackEventExternalRef `json:"external_ref"`
	Links        []slackSharedLink     `json:"links"`
	BotID        string                `json:"bot_id"`
	BotProfile   any                   `json:"bot_profile"`
	Tokens       struct {
		OAuth []string `json:"oauth"`
		Bot   []string `json:"bot"`
	} `json:"tokens"`
}

type slackSharedLink struct {
	Domain string `json:"domain"`
	URL    string `json:"url"`
}

type slackEventExternalRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type slackEventKind string

const (
	slackEventKindMention       slackEventKind = "mention"
	slackEventKindDirect        slackEventKind = "direct_message"
	slackEventKindChannelThread slackEventKind = "channel_thread"
	slackEventKindUninstalled   slackEventKind = "app_uninstalled"
	slackEventKindRevoked       slackEventKind = "tokens_revoked"
	slackEventKindLinkShared    slackEventKind = "link_shared"
	slackEventKindEntityDetails slackEventKind = "entity_details_requested"
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
	Source              string
	TriggerID           string
	EntityURL           string
	AppUnfurlURL        string
	ExternalRef         slackEventExternalRef
	Links               []slackSharedLink
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
		EventID:      envelope.EventID,
		TeamID:       envelope.TeamID,
		UserID:       strings.TrimSpace(event.User),
		ChannelID:    strings.TrimSpace(event.Channel),
		MessageTS:    strings.TrimSpace(event.TS),
		ThreadTS:     strings.TrimSpace(event.ThreadTS),
		Text:         strings.TrimSpace(event.Text),
		Source:       strings.TrimSpace(event.Source),
		TriggerID:    strings.TrimSpace(event.TriggerID),
		EntityURL:    strings.TrimSpace(event.EntityURL),
		AppUnfurlURL: strings.TrimSpace(event.AppUnfurlURL),
		ExternalRef: slackEventExternalRef{
			ID:   strings.TrimSpace(event.ExternalRef.ID),
			Type: strings.TrimSpace(event.ExternalRef.Type),
		},
		RevokedOAuthUserIDs: append([]string(nil), event.Tokens.OAuth...),
		RevokedBotUserIDs:   append([]string(nil), event.Tokens.Bot...),
	}
	normalized.ReplyTS = normalized.ThreadTS
	if messageTS := strings.TrimSpace(event.MessageTS); messageTS != "" {
		normalized.MessageTS = messageTS
	}
	if normalized.ThreadTS == "" {
		normalized.ThreadTS = normalized.MessageTS
	}
	switch strings.TrimSpace(event.Type) {
	case slackEventAppMention:
		normalized.Kind = slackEventKindMention
		normalized.ReplyTS = normalized.ThreadTS
	case slackEventMessage:
		switch strings.TrimSpace(event.ChannelType) {
		case "im":
			normalized.Kind = slackEventKindDirect
		case "channel", "group":
			// Channel message subscriptions deliver every message visible to the
			// app. Only real thread replies are candidates for an existing Maya
			// conversation; a root message must never subscribe a thread.
			if normalized.ReplyTS == "" {
				return normalizedSlackEvent{}, false
			}
			normalized.Kind = slackEventKindChannelThread
		default:
			return normalizedSlackEvent{}, false
		}
	case slackEventUninstalled:
		normalized.Kind = slackEventKindUninstalled
		return normalized, true
	case slackEventTokenRevoke:
		normalized.Kind = slackEventKindRevoked
		return normalized, true
	case slackEventLinkShared:
		normalized.Kind = slackEventKindLinkShared
		if envelope.IsExtSharedChannel || normalized.UserID == "" || normalized.ChannelID == "" || normalized.MessageTS == "" {
			return normalizedSlackEvent{}, false
		}
		for _, link := range event.Links {
			link.Domain = strings.ToLower(strings.TrimSpace(link.Domain))
			link.URL = strings.TrimSpace(link.URL)
			if link.URL != "" {
				normalized.Links = append(normalized.Links, link)
			}
		}
		return normalized, len(normalized.Links) > 0
	case slackEventEntityDetailsRequested:
		normalized.Kind = slackEventKindEntityDetails
		if envelope.IsExtSharedChannel || normalized.UserID == "" || normalized.TriggerID == "" || normalized.ExternalRef.ID == "" {
			return normalizedSlackEvent{}, false
		}
		if normalized.EntityURL == "" {
			normalized.EntityURL = normalized.AppUnfurlURL
		}
		return normalized, normalized.EntityURL != ""
	default:
		return normalizedSlackEvent{}, false
	}
	if strings.TrimSpace(event.Subtype) != "" || strings.TrimSpace(event.BotID) != "" || event.BotProfile != nil {
		return normalizedSlackEvent{}, false
	}
	if envelope.IsExtSharedChannel && (normalized.Kind == slackEventKindMention || normalized.Kind == slackEventKindChannelThread) {
		// Channel audiences in Slack Connect can include users outside the
		// FortyOne workspace. Ignore these events until channel-level access
		// policy can be enforced authoritatively.
		return normalizedSlackEvent{}, false
	}
	if normalized.UserID == "" || normalized.ChannelID == "" || normalized.MessageTS == "" || normalized.Text == "" {
		return normalizedSlackEvent{}, false
	}
	return normalized, true
}

func containsSlackUserMention(text, userID string) bool {
	text = strings.TrimSpace(text)
	userID = strings.TrimSpace(userID)
	return userID != "" && strings.Contains(text, "<@"+userID+">")
}

func removeBotMention(text, botUserID string) string {
	text = strings.TrimSpace(text)
	botUserID = strings.TrimSpace(botUserID)
	if botUserID == "" {
		return text
	}
	return strings.TrimSpace(strings.ReplaceAll(text, "<@"+botUserID+">", ""))
}
