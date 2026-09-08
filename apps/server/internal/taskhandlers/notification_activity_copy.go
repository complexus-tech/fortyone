package taskhandlers

import "strings"

// Activity copy stays bound to persisted event values. Generated introductions
// can vary, but must not repeat the task title, obscure a date, or rewrite a
// status transition into a different change.
func notificationActivityCopy(message NotificationMessage) (string, string, []string) {
	var detail string
	if reason, ok := message.Variables["reason"]; ok && strings.HasSuffix(message.Template, ": {reason}") {
		message.Template = strings.TrimSuffix(message.Template, ": {reason}")
		detail = notificationPlainText(reason.Value, maxNotificationMessageRunes)
		// Already-persisted scheduling notifications retain the old explanation.
		// Their scheduled_for value supplies the actual slot in the main line.
		if detail == "The assignee's availability or this story's scheduling constraints changed, so Maya moved it to the next safe slot." {
			detail = "The assignee's availability or the task's scheduling constraints changed."
		}
	}
	var highlights []string
	for key, variable := range message.Variables {
		switch key {
		case "actor", "reason", "content", "body", "comment", "title", "description":
			continue
		}
		switch variable.Type {
		case "value", "date", "assignee", "status", "priority", "field":
			value := notificationPlainText(variable.Value, 0)
			if value != "" && len([]rune(value)) <= 120 {
				highlights = append(highlights, value)
			}
		}
	}
	return parseNotificationMessage(message).Text, detail, highlights
}

func notificationFactHighlights(values []string) []string {
	var highlights []string
	for _, value := range values {
		if _, date, ok := strings.Cut(value, " starts on "); ok {
			value = date
		}
		for _, prefix := range []string{
			"objective health is ", "objective status is ", "current value is ", "target value is ",
			"last updated on ", "generated on ", "missing elements are ", "strategy foundation ",
			"health is ", "status is ", "measurement is ", "ends on ", "including ", "has ", "in ",
		} {
			if strings.HasPrefix(value, prefix) {
				value = strings.TrimPrefix(value, prefix)
				break
			}
		}
		if value = strings.TrimSpace(value); value != "" && len([]rune(value)) <= 100 {
			highlights = append(highlights, value)
		}
	}
	return highlights
}
