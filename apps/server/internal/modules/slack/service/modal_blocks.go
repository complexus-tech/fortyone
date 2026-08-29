package slack

import (
	"strings"
)

func selectInputBlock(blockID, actionID, label string, options []map[string]any, initialOption map[string]any, optional, dispatchAction bool) map[string]any {
	element := map[string]any{
		"type":      "static_select",
		"action_id": actionID,
		"options":   limitedSlackOptions(options),
	}
	if initialOption != nil {
		element["initial_option"] = initialOption
	}

	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	if dispatchAction {
		block["dispatch_action"] = true
	}
	return block
}

func externalSelectInputBlock(blockID, actionID, label string, initialOption map[string]any, optional bool, minQueryLength int, dispatchAction bool) map[string]any {
	element := map[string]any{
		"type":      "external_select",
		"action_id": actionID,
	}
	if initialOption != nil {
		element["initial_option"] = initialOption
	}
	if minQueryLength >= 0 {
		element["min_query_length"] = minQueryLength
	}
	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	if dispatchAction {
		block["dispatch_action"] = true
	}
	return block
}

func externalMultiSelectInputBlock(blockID, actionID, label string, initialOptions []map[string]any, optional bool, minQueryLength int) map[string]any {
	element := map[string]any{
		"type":      "multi_external_select",
		"action_id": actionID,
	}
	if len(initialOptions) > 0 {
		element["initial_options"] = limitedSlackOptions(initialOptions)
	}
	if minQueryLength > 0 {
		element["min_query_length"] = minQueryLength
	}
	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	return block
}

func toSlackOption(text, value string) map[string]any {
	trimmedText := strings.TrimSpace(text)
	trimmedValue := strings.TrimSpace(value)
	if trimmedText == "" {
		trimmedText = trimmedValue
	}
	if trimmedValue == "" {
		trimmedValue = trimmedText
	}
	trimmedText = truncateRunes(trimmedText, slackOptionTextMaxRunes)
	trimmedValue = truncateRunes(trimmedValue, slackOptionValueMaxRunes)
	return map[string]any{
		"text": map[string]string{
			"type": "plain_text",
			"text": trimmedText,
		},
		"value": trimmedValue,
	}
}

func slackPriorityOptions() []map[string]any {
	priorities := []string{slackPriorityNoPriority, "Low", "Medium", "High", "Urgent"}
	options := make([]map[string]any, 0, len(priorities))
	for _, priority := range priorities {
		options = append(options, toSlackOption(priority, priority))
	}
	return options
}

func plainInputBlock(blockID, actionID, label, initial string, multiline bool, placeholder string, optional bool) map[string]any {
	element := map[string]any{
		"type":      "plain_text_input",
		"action_id": actionID,
	}
	if multiline {
		element["multiline"] = true
	}
	switch blockID {
	case modalBlockTitle:
		element["max_length"] = modalTitleMaxRunes
	case modalBlockDescription:
		element["max_length"] = modalDescriptionMaxRunes
	}
	if initial != "" {
		element["initial_value"] = initial
	}
	if placeholder != "" {
		element["placeholder"] = map[string]string{
			"type": "plain_text",
			"text": placeholder,
		}
	}
	block := map[string]any{
		"type":     "input",
		"block_id": blockID,
		"label": map[string]string{
			"type": "plain_text",
			"text": label,
		},
		"element": element,
	}
	if optional {
		block["optional"] = true
	}
	return block
}
