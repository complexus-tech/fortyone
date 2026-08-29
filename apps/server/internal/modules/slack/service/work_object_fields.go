package slack

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

func appendSlackWorkObjectCustomDateField(fields *[]SlackWorkObjectCustomField, key, label string, value *time.Time) bool {
	if value == nil || value.IsZero() {
		return false
	}
	*fields = append(*fields, SlackWorkObjectCustomField{
		Key:   key,
		Label: label,
		Value: value.UTC().Format(time.DateOnly),
		Type:  slackDateFieldType,
	})
	return true
}

func slackWorkObjectCreatorLabel(creatorName string) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.Join(strings.Fields(creatorName), " "))
	if creatorLabel == "" {
		return "A team member"
	}
	return creatorLabel
}

func slackStoryTitleEdit(enabled bool) *SlackWorkObjectEdit {
	if !enabled {
		return nil
	}
	return &SlackWorkObjectEdit{
		Enabled: true,
		Text: &SlackWorkObjectEditText{
			MinLength: 1,
			MaxLength: modalTitleMaxRunes,
		},
	}
}

func slackOptionalWorkObjectEdit(enabled bool) *SlackWorkObjectEdit {
	if !enabled {
		return nil
	}
	return &SlackWorkObjectEdit{Enabled: true, Optional: true}
}

func slackWorkObjectPriorityOptions() []SlackWorkObjectSelectOption {
	priorities := []string{slackPriorityNoPriority, "Low", "Medium", "High", "Urgent"}
	options := make([]SlackWorkObjectSelectOption, 0, len(priorities))
	for _, priority := range priorities {
		options = append(options, newSlackWorkObjectSelectOption(priority, priority))
	}
	return options
}

func newSlackWorkObjectSelectOption(value, label string) SlackWorkObjectSelectOption {
	return SlackWorkObjectSelectOption{
		Value: strings.TrimSpace(value),
		Text: SlackWorkObjectOptionText{
			Type: "plain_text",
			Text: truncateSlackWorkObjectText(label, 75),
		},
	}
}

func validSlackWorkObjectSelectOptions(options []SlackWorkObjectSelectOption) bool {
	if len(options) == 0 || len(options) > slackWorkObjectSelectLimit {
		return false
	}
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		value := strings.TrimSpace(option.Value)
		label := strings.TrimSpace(option.Text.Text)
		if value == "" || label == "" || len(value) > 150 {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func cloneSlackWorkObjectSelectOptions(options []SlackWorkObjectSelectOption) []SlackWorkObjectSelectOption {
	return append([]SlackWorkObjectSelectOption(nil), options...)
}

func slackStoryExternalRefID(link FortyOneStoryLink, externalID string) string {
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		externalID = link.StoryReference
	}
	return strings.ToLower(link.WorkspaceSlug) + ":" + externalID
}

func slackStoryExternalRefMatches(link FortyOneStoryLink, externalID, actual string) bool {
	actual = strings.TrimSpace(actual)
	return actual == slackStoryExternalRefID(link, externalID) ||
		actual == slackStoryExternalRefID(link, "") ||
		strings.EqualFold(actual, link.StoryReference)
}

func slackRequestExternalRefID(link FortyOneRequestLink) string {
	return strings.ToLower(link.WorkspaceSlug) + ":" + link.TeamID.String() + ":" + link.RequestID.String()
}

func validSlackRequestExternalRef(link FortyOneRequestLink, externalRefID string) bool {
	return strings.TrimSpace(externalRefID) == slackRequestExternalRefID(link)
}

func slackObjectiveExternalRefID(link FortyOneObjectiveLink, externalID string) string {
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		externalID = link.ObjectiveID.String()
	}
	return strings.ToLower(link.WorkspaceSlug) + ":" + link.TeamID.String() + ":" + externalID
}

func validSlackObjectiveExternalRef(link FortyOneObjectiveLink, externalRefID string) bool {
	return strings.TrimSpace(externalRefID) == slackObjectiveExternalRefID(link, "")
}

func slackSprintExternalRefID(link FortyOneSprintLink, externalID string) string {
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		externalID = link.SprintID.String()
	}
	return strings.ToLower(link.WorkspaceSlug) + ":" + link.TeamID.String() + ":" + externalID
}

func validSlackSprintExternalRef(link FortyOneSprintLink, externalRefID string) bool {
	return strings.TrimSpace(externalRefID) == slackSprintExternalRefID(link, "")
}

func validateSlackUnfurlDestination(channelID, messageTS string) error {
	channelID = strings.TrimSpace(channelID)
	messageTS = strings.TrimSpace(messageTS)
	if channelID == "" || messageTS == "" || strings.ContainsAny(channelID+messageTS, " \t\r\n") {
		return errors.New("slack unfurl channel and timestamp are required")
	}
	return nil
}

func validateSlackUnfurlRequestDestination(request SlackChatUnfurlRequest) error {
	channelID := strings.TrimSpace(request.Channel)
	messageTS := strings.TrimSpace(request.TS)
	unfurlID := strings.TrimSpace(request.UnfurlID)
	source := strings.TrimSpace(request.Source)

	hasConversationDestination := channelID != "" || messageTS != ""
	hasEventDestination := unfurlID != "" || source != ""
	if hasConversationDestination == hasEventDestination {
		return errors.New("slack unfurl requires exactly one destination pair")
	}
	if hasConversationDestination {
		return validateSlackUnfurlDestination(channelID, messageTS)
	}
	if unfurlID == "" || (source != "composer" && source != "conversations_history") {
		return errors.New("slack unfurl ID and valid source are required")
	}
	return nil
}

func isSafeFortyOneHTTPSURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || parsed.Port() != "" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	return host == "fortyone.app" || strings.HasSuffix(host, ".fortyone.app")
}

func slackWorkObjectUser(slackUserID, displayName string) *SlackWorkObjectUser {
	slackUserID = strings.ToUpper(strings.TrimSpace(slackUserID))
	if slackUserIDPattern.MatchString(slackUserID) {
		return &SlackWorkObjectUser{UserID: slackUserID}
	}
	displayName = truncateSlackWorkObjectText(displayName, 255)
	if displayName == "" {
		return nil
	}
	return &SlackWorkObjectUser{Text: displayName}
}

func normalizeSlackTagColor(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "red", "yellow", "green", "gray", "blue":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}
