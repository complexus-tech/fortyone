package taskhandlers

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// latestTaskDigestItems selects by event time, with the repository's notification
// ID ordering as a deterministic tie-breaker. Distinct tasks with the same title
// remain separate; non-task notifications retain their existing behavior.
func latestTaskDigestItems(items []NotificationEmailDigestItem) []NotificationEmailDigestItem {
	type taskKey struct {
		storyID       uuid.UUID
		scheduleIssue bool
	}
	keyFor := func(item NotificationEmailDigestItem) (taskKey, bool) {
		if item.NotificationType != "story_update" || item.EntityType != "story" || item.EntityID == uuid.Nil {
			return taskKey{}, false
		}
		var metadata struct {
			ScheduleIssue json.RawMessage `json:"scheduleIssue"`
		}
		if len(item.Message) > 0 {
			if err := json.Unmarshal(item.Message, &metadata); err != nil {
				// Preserve invalid items so the normal renderer reports the error.
				return taskKey{}, false
			}
		}
		return taskKey{storyID: item.EntityID, scheduleIssue: len(metadata.ScheduleIssue) > 0 && string(metadata.ScheduleIssue) != "null"}, true
	}
	latest := make(map[taskKey]int)
	for index, item := range items {
		key, coalesce := keyFor(item)
		if !coalesce {
			continue
		}
		previous, exists := latest[key]
		if !exists || item.CreatedAt.After(items[previous].CreatedAt) ||
			(item.CreatedAt.Equal(items[previous].CreatedAt) && item.NotificationID.String() > items[previous].NotificationID.String()) {
			latest[key] = index
		}
	}
	result := make([]NotificationEmailDigestItem, 0, len(items))
	for index, item := range items {
		if key, coalesce := keyFor(item); coalesce && latest[key] != index {
			continue
		}
		result = append(result, item)
	}
	return result
}

func taskOnlyDigest(items []NotificationEmailDigestItem) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item.EntityType != "story" {
			return false
		}
	}
	return true
}

func notificationVariableValues(variables map[string]Variable) []string {
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		variable := variables[key]
		values = append(values, notificationVariableText(variable, maxNotificationMessageRunes))
	}
	return nonEmptyStrings(values...)
}

// notificationSemanticProtectedTokens keeps the factual roles in activity
// messages bound together. Requiring only the individual variable values
// would allow generated copy to swap an actor, field, assignee, or status while
// still passing literal-token validation. For long comment-style messages, the
// author/action prefix stays exact while the user-authored body remains free to
// be summarized.
func notificationSemanticProtectedTokens(message NotificationMessage, parsedText string) []string {
	values := notificationVariableValues(message.Variables)
	if len(values) == 0 {
		return values
	}

	const maxProtectedActivityRunes = 300
	activity := notificationLiteralText(parsedText, 0)
	if len([]rune(activity)) <= maxProtectedActivityRunes {
		return nonEmptyStrings(activity)
	}
	if separator := strings.Index(activity, ":"); separator > 0 {
		semanticPrefix := strings.TrimSpace(activity[:separator])
		if len([]rune(semanticPrefix)) <= maxProtectedActivityRunes {
			return nonEmptyStrings(semanticPrefix)
		}
	}
	return values
}
