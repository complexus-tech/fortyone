package chatsessionsdomain

import (
	"fmt"
	"strings"
)

// ValidateLegacyMessageInitialization permits only safe text/file history on
// legacy create/save endpoints. Tool state must enter through the reservation
// and approval protocols.
func ValidateLegacyMessageInitialization(messages []any) error {
	for _, rawMessage := range messages {
		message, ok := rawMessage.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: legacy message must be an object", ErrMessageWriteInvalid)
		}
		parts, ok := message["parts"].([]any)
		if !ok {
			return fmt.Errorf("%w: legacy message parts are invalid", ErrMessageWriteInvalid)
		}
		for _, rawPart := range parts {
			part, ok := rawPart.(map[string]any)
			if !ok {
				return fmt.Errorf("%w: legacy message part must be an object", ErrMessageWriteInvalid)
			}
			partType, _ := part["type"].(string)
			_, hasToolCallID := part["toolCallId"]
			_, hasApproval := part["approval"]
			if strings.HasPrefix(partType, "tool-") || partType == "dynamic-tool" ||
				partType == "tool-invocation" || hasToolCallID || hasApproval {
				return fmt.Errorf("%w: legacy initialization cannot contain tool state", ErrMessageWriteInvalid)
			}
		}
	}
	return nil
}
