package chatsessionsdomain

import (
	"fmt"
	"reflect"
)

func validateMessageSequence(messages []any) error {
	for _, raw := range messages {
		message, ok := asObject(raw)
		if !ok {
			return fmt.Errorf("%w: message must be an object", ErrMessageWriteInvalid)
		}
		role, roleOK := message["role"].(string)
		if !roleOK || (role != "user" && role != "assistant" && role != "system") {
			return fmt.Errorf("%w: message role is invalid", ErrMessageWriteInvalid)
		}
		if _, ok := asArray(message["parts"]); !ok {
			return fmt.Errorf("%w: message parts are invalid", ErrMessageWriteInvalid)
		}
	}
	return nil
}

// Historical transcripts predate the unique-ID invariant and may contain
// empty or duplicate IDs. They remain valid only as a positionally matching
// prefix; every newly appended message must carry a fresh, unique identity.
func validateNewMessageIdentities(current, suffix []any) error {
	seen := make(map[string]struct{}, len(current)+len(suffix))
	for _, raw := range current {
		message, ok := asObject(raw)
		if !ok {
			continue
		}
		id, ok := message["id"].(string)
		if ok && id != "" {
			seen[id] = struct{}{}
		}
	}
	for _, raw := range suffix {
		message, ok := asObject(raw)
		if !ok {
			return fmt.Errorf("%w: message must be an object", ErrMessageWriteInvalid)
		}
		id, ok := message["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("%w: new message id is required", ErrMessageWriteInvalid)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("%w: new message id must be unique", ErrMessageWriteInvalid)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func messagesCompatible(currentRaw, incomingRaw any, allowMonotonicToolOutput bool) bool {
	current, currentOK := asObject(currentRaw)
	incoming, incomingOK := asObject(incomingRaw)
	if !currentOK || !incomingOK || !sameMessageIdentity(current, incoming) {
		return false
	}
	currentParts, currentOK := asArray(current["parts"])
	incomingParts, incomingOK := asArray(incoming["parts"])
	if !currentOK || !incomingOK || len(currentParts) != len(incomingParts) {
		return false
	}
	for index := range currentParts {
		if reflect.DeepEqual(currentParts[index], incomingParts[index]) {
			continue
		}
		currentPart, currentObject := asObject(currentParts[index])
		incomingPart, incomingObject := asObject(incomingParts[index])
		if !currentObject || !incomingObject {
			return false
		}
		if isHistoricalFilePlaceholder(currentPart, incomingPart) {
			continue
		}
		if allowMonotonicToolOutput && sameToolPartIdentity(currentPart, incomingPart) && sameToolPartBase(currentPart, incomingPart) && isTerminalToolState(toolState(currentPart)) {
			incomingState := toolState(incomingPart)
			if incomingState == "approval-requested" || incomingState == "approval-responded" {
				continue
			}
		}
		return false
	}
	return true
}

func hasUnresolvedApproval(messages []any) bool {
	if len(messages) == 0 || messageRole(messages[len(messages)-1]) != "assistant" {
		return false
	}
	message, ok := asObject(messages[len(messages)-1])
	if !ok {
		return false
	}
	parts, ok := asArray(message["parts"])
	if !ok {
		return false
	}
	for _, rawPart := range parts {
		part, ok := asObject(rawPart)
		if !ok {
			continue
		}
		state := toolState(part)
		if state == "approval-requested" || state == "approval-responded" {
			return true
		}
	}
	return false
}

func sameMessageIdentity(left, right map[string]any) bool {
	return left["id"] == right["id"] && left["role"] == right["role"]
}

func sameToolPartIdentity(left, right map[string]any) bool {
	leftType, leftTypeOK := left["type"].(string)
	rightType, rightTypeOK := right["type"].(string)
	leftCallID, leftCallIDOK := left["toolCallId"].(string)
	rightCallID, rightCallIDOK := right["toolCallId"].(string)
	return leftTypeOK && rightTypeOK && leftCallIDOK && rightCallIDOK && leftType == rightType && leftCallID == rightCallID && len(leftCallID) > 0
}

func sameToolPartBase(left, right map[string]any) bool {
	return reflect.DeepEqual(left["input"], right["input"])
}

func isHistoricalFilePlaceholder(current, incoming map[string]any) bool {
	return current["type"] == "file" && incoming["type"] == "text" && incoming["text"] == HistoricalAttachmentPlaceholder
}
