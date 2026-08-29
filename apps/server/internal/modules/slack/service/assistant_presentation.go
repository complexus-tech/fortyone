package slack

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

func isAssistantBudgetNotice(content string) bool {
	switch strings.TrimSpace(content) {
	case assistantUserRateLimitReply, assistantWorkspaceRateReply, assistantDailyLimitReply:
		return true
	default:
		return false
	}
}

func conversationThreadID(event normalizedSlackEvent) string {
	if event.Kind == slackEventKindDirect && event.ReplyTS == "" {
		return "dm:" + event.ChannelID
	}
	return event.ThreadTS
}

func assistantSurfaceForSlackEvent(event normalizedSlackEvent) assistantRuntimeSurface {
	kind := assistantSurfaceThread
	if event.Kind == slackEventKindDirect {
		kind = assistantSurfaceDirect
	}
	return assistantRuntimeSurface{
		Provider: "slack",
		Kind:     kind,
	}
}

func deterministicSlackMessageID(value string) string {
	digest := sha256.Sum256([]byte(value))
	bytes := digest[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(bytes)
	if err != nil {
		return hex.EncodeToString(digest[:16])
	}
	return id.String()
}

func truncateSlackText(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= slackMessageTextLimit {
		return value
	}
	return strings.TrimSpace(string(runes[:slackMessageTextLimit-1])) + "…"
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if len(message) <= 1000 {
		return message
	}
	return message[:1000]
}
