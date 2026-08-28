package chatsessionsdomain

import (
	"encoding/json"
	"fmt"
)

func validApprovalTransition(currentRaw, incomingRaw any) bool {
	current, currentOK := asObject(currentRaw)
	incoming, incomingOK := asObject(incomingRaw)
	if !currentOK || !incomingOK {
		return false
	}
	currentID, currentIDOK := current["id"].(string)
	incomingID, incomingIDOK := incoming["id"].(string)
	_, approved := incoming["approved"].(bool)
	return currentIDOK && incomingIDOK && currentID != "" && currentID == incomingID && approved
}

func approvalDenied(raw any) bool {
	approval, ok := asObject(raw)
	if !ok {
		return false
	}
	approved, ok := approval["approved"].(bool)
	return ok && !approved
}

func toolState(part map[string]any) string {
	state, _ := part["state"].(string)
	return state
}

func isTerminalToolState(state string) bool {
	return state == "output-available" || state == "output-denied"
}

func toolOutputIndicatesFailure(output any) bool {
	object, ok := asObject(output)
	if !ok {
		return false
	}
	if success, exists := object["success"]; exists && success == false {
		return true
	}
	errorValue, exists := object["error"]
	return exists && errorValue != nil
}

func messageRole(raw any) string {
	message, ok := asObject(raw)
	if !ok {
		return ""
	}
	role, _ := message["role"].(string)
	return role
}

func asObject(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	return object, ok
}

func asArray(value any) ([]any, bool) {
	array, ok := value.([]any)
	return array, ok
}

func cloneObject(value map[string]any) (map[string]any, error) {
	cloned, err := cloneJSONValue(value)
	if err != nil {
		return nil, err
	}
	object, ok := asObject(cloned)
	if !ok {
		return nil, fmt.Errorf("%w: expected object", ErrMessageWriteInvalid)
	}
	return object, nil
}

func cloneJSONValue(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: encode transcript: %v", ErrMessageWriteInvalid, err)
	}
	var cloned any
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil, fmt.Errorf("%w: decode transcript: %v", ErrMessageWriteInvalid, err)
	}
	return cloned, nil
}

// CloneJSONValue returns a deep JSON-compatible copy of a transcript value.
func CloneJSONValue(value any) (any, error) {
	return cloneJSONValue(value)
}
