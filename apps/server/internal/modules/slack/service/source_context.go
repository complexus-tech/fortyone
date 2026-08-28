package slack

import (
	"fmt"
	"strings"
)

func messageToTitle(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "New task"
	}
	firstLine := strings.Split(trimmed, "\n")[0]
	firstLine = strings.TrimSpace(firstLine)
	if len([]rune(firstLine)) <= 120 {
		return firstLine
	}
	return strings.TrimSpace(truncateRunes(firstLine, 120))
}

func buildConnectSlackAccountMessage(connectURL string) string {
	if strings.TrimSpace(connectURL) == "" {
		return "Connect your FortyOne account before creating tasks from Slack."
	}
	return fmt.Sprintf("Connect your FortyOne account to continue: <%s|Connect FortyOne account>", connectURL)
}

func buildPrefilledDescription(source requestSourceContext) string {
	message := strings.TrimSpace(source.SlackText)
	if message == "" {
		return ""
	}
	identity := strings.TrimSpace(source.SlackUsername)
	if identity == "" {
		identity = strings.TrimSpace(source.SlackUserID)
	}
	if identity == "" {
		return truncateRunes("> "+message, modalDescriptionMaxRunes)
	}
	if strings.TrimSpace(source.SlackUserID) == "" {
		return truncateRunes(fmt.Sprintf("@%s said:\n> %s", identity, message), modalDescriptionMaxRunes)
	}
	return truncateRunes(fmt.Sprintf("@[%s](%s) said:\n> %s", identity, strings.TrimSpace(source.SlackUserID), message), modalDescriptionMaxRunes)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func buildSourceExternalID(source requestSourceContext) string {
	parts := []string{}
	if source.SlackTeamID != "" {
		parts = append(parts, source.SlackTeamID)
	}
	if source.SlackChannelID != "" {
		parts = append(parts, source.SlackChannelID)
	}
	if source.SlackMessageTS != "" {
		parts = append(parts, source.SlackMessageTS)
	}
	if source.SlackThreadTS != "" {
		parts = append(parts, source.SlackThreadTS)
	}
	return strings.Join(parts, ":")
}

func permalinkFromContext(source requestSourceContext) string {
	if source.SlackTeamDomain == "" || source.SlackChannelID == "" || source.SlackMessageTS == "" {
		return ""
	}
	messageTS := strings.ReplaceAll(source.SlackMessageTS, ".", "")
	return fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", source.SlackTeamDomain, source.SlackChannelID, messageTS)
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func isZeroRequestSource(source requestSourceContext) bool {
	return strings.TrimSpace(source.SlackTeamID) == "" &&
		strings.TrimSpace(source.SlackTeamDomain) == "" &&
		strings.TrimSpace(source.SlackChannelID) == "" &&
		strings.TrimSpace(source.SlackChannel) == "" &&
		strings.TrimSpace(source.SlackMessageTS) == "" &&
		strings.TrimSpace(source.SlackThreadTS) == "" &&
		strings.TrimSpace(source.SlackUserID) == "" &&
		strings.TrimSpace(source.SlackUsername) == "" &&
		strings.TrimSpace(source.SlackText) == ""
}
