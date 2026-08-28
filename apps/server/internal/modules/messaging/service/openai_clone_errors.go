package messaging

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func cloneToolDefinitions(definitions []ToolDefinition) []ToolDefinition {
	cloned := make([]ToolDefinition, len(definitions))
	for index, definition := range definitions {
		cloned[index] = definition
		cloned[index].Parameters = cloneStringAnyMap(definition.Parameters)
	}
	return cloned
}

func cloneStringAnyMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		switch typed := item.(type) {
		case map[string]any:
			cloned[key] = cloneStringAnyMap(typed)
		case []string:
			if typed == nil {
				cloned[key] = []string(nil)
				break
			}
			items := make([]string, len(typed))
			copy(items, typed)
			cloned[key] = items
		case []any:
			items := make([]any, len(typed))
			for index, nested := range typed {
				if object, ok := nested.(map[string]any); ok {
					items[index] = cloneStringAnyMap(object)
				} else {
					items[index] = nested
				}
			}
			cloned[key] = items
		default:
			cloned[key] = item
		}
	}
	return cloned
}

func cloneRawMessages(messages []json.RawMessage) []json.RawMessage {
	cloned := make([]json.RawMessage, len(messages))
	for index, message := range messages {
		cloned[index] = cloneRawMessage(message)
	}
	return cloned
}

func cloneRawMessage(message json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), message...)
}

func safetyIdentifier(userID string) string {
	hash := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(hash[:])
}

func decodeAPIError(statusCode int, requestID string, body []byte) error {
	var payload struct {
		Error *openAIError `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Error != nil {
		return &APIError{
			StatusCode: statusCode,
			Code:       payload.Error.Code,
			Message:    payload.Error.message(),
			RequestID:  requestID,
		}
	}
	return &APIError{StatusCode: statusCode, RequestID: requestID}
}

func addUsage(usage *Usage, value responsesUsage) {
	usage.InputTokens += value.InputTokens
	usage.OutputTokens += value.OutputTokens
	usage.TotalTokens += value.TotalTokens
}
