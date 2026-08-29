package chatsessions

func message(id, role string, parts ...any) map[string]any {
	return map[string]any{"id": id, "parts": parts, "role": role}
}

func textPart(text string) map[string]any {
	return map[string]any{"text": text, "type": "text"}
}

func approvalPart(state string, approval any) map[string]any {
	part := map[string]any{
		"input":      map[string]any{"teamId": "team-1", "title": "Launch"},
		"state":      state,
		"toolCallId": "call-1",
		"type":       "tool-createStory",
	}
	if approval != nil {
		part["approval"] = approval
	}
	return part
}

func lastPart(messages []any) map[string]any {
	message := messages[len(messages)-1].(map[string]any)
	parts := message["parts"].([]any)
	return parts[len(parts)-1].(map[string]any)
}
