package chatsessionsdomain

import (
	"fmt"
	"reflect"
)

// ReserveMessageWrite validates one of the three request-side transcript
// transitions and returns a detached transcript suitable for persistence.
// Existing messages always win over equivalent client copies, which keeps
// omitted historical attachments and already-durable tool outputs intact.
func ReserveMessageWrite(current, incoming []any, operation MessageWriteOperation) ([]any, error) {
	return ReserveMessageWriteForTarget(current, incoming, operation, "")
}

// CanonicalMessageWriteResponse returns only request-safe history while
// projecting terminal repairs made by the server. Stored file payloads never
// replace historical attachment placeholders in the response.
func CanonicalMessageWriteResponse(persisted, request []any) ([]any, bool, error) {
	if len(persisted) != len(request) {
		return nil, false, ErrMessageWriteConflict
	}
	canonicalValue, err := cloneJSONValue(request)
	if err != nil {
		return nil, false, err
	}
	canonical, ok := canonicalValue.([]any)
	if !ok {
		return nil, false, fmt.Errorf("%w: transcript must be an array", ErrMessageWriteInvalid)
	}

	repaired := false
	for messageIndex := range persisted {
		persistedMessage, persistedOK := asObject(persisted[messageIndex])
		requestMessage, requestOK := asObject(request[messageIndex])
		canonicalMessage, canonicalOK := asObject(canonical[messageIndex])
		if !persistedOK || !requestOK || !canonicalOK || !sameMessageIdentity(persistedMessage, requestMessage) {
			return nil, false, ErrMessageWriteConflict
		}
		persistedParts, persistedOK := asArray(persistedMessage["parts"])
		requestParts, requestOK := asArray(requestMessage["parts"])
		canonicalParts, canonicalOK := asArray(canonicalMessage["parts"])
		if !persistedOK || !requestOK || !canonicalOK || len(persistedParts) != len(requestParts) {
			return nil, false, ErrMessageWriteConflict
		}

		for partIndex := range persistedParts {
			if reflect.DeepEqual(persistedParts[partIndex], requestParts[partIndex]) {
				continue
			}
			persistedPart, persistedObject := asObject(persistedParts[partIndex])
			requestPart, requestObject := asObject(requestParts[partIndex])
			if !persistedObject || !requestObject {
				return nil, false, ErrMessageWriteConflict
			}
			if isHistoricalFilePlaceholder(persistedPart, requestPart) {
				continue
			}
			if !sameToolPartIdentity(persistedPart, requestPart) ||
				!sameToolPartBase(persistedPart, requestPart) ||
				!isTerminalToolState(toolState(persistedPart)) {
				return nil, false, ErrMessageWriteConflict
			}
			copiedPart, err := cloneJSONValue(persistedPart)
			if err != nil {
				return nil, false, err
			}
			canonicalParts[partIndex] = copiedPart
			repaired = true
		}
	}
	return canonical, repaired, nil
}

// ReserveMessageWriteForTarget binds regeneration to the exact message
// boundary selected by AI SDK. A shorter array alone is never sufficient
// authority to truncate committed history.
func ReserveMessageWriteForTarget(current, incoming []any, operation MessageWriteOperation, targetMessageID string) ([]any, error) {
	if err := validateMessageSequence(incoming); err != nil {
		return nil, err
	}

	switch operation {
	case MessageWriteAppend:
		if targetMessageID != "" {
			return nil, fmt.Errorf("%w: editing an existing user message is not supported", ErrMessageWriteInvalid)
		}
		if hasUnresolvedApproval(current) {
			return nil, ErrMessageWriteApprovalOpen
		}
		if len(incoming) == 0 || len(incoming) < len(current) || messageRole(incoming[len(incoming)-1]) != "user" {
			return nil, fmt.Errorf("%w: append must end with a user message", ErrMessageWriteInvalid)
		}
		if err := validateNewMessageIdentities(current, incoming[len(current):]); err != nil {
			return nil, err
		}
		return mergePrefixAndSuffix(current, incoming, true)

	case MessageWriteRegenerate:
		if hasUnresolvedApproval(current) {
			return nil, ErrMessageWriteApprovalOpen
		}
		// AI SDK exposes a single regenerate action for retrying a failed turn.
		// When the original request failed before its reservation was committed,
		// the client still holds the latest user message while durable history
		// does not. Recover that exact one-message append here; otherwise a safe
		// retry of a transient setup failure is rejected as a stale regeneration.
		if len(incoming) == len(current)+1 && messageRole(incoming[len(incoming)-1]) == "user" {
			incomingMessage, _ := asObject(incoming[len(incoming)-1])
			incomingMessageID, _ := incomingMessage["id"].(string)
			if targetMessageID == "" || targetMessageID == incomingMessageID {
				merged, err := mergePrefixAndSuffix(current, incoming, true)
				if err != nil {
					return nil, err
				}
				if err := validateNewMessageIdentities(current, incoming[len(current):]); err != nil {
					return nil, err
				}
				return merged, nil
			}
		}
		expectedLength, err := regenerationPrefixLength(current, targetMessageID)
		if err != nil {
			return nil, err
		}
		if len(incoming) == 0 || len(incoming) != expectedLength || messageRole(incoming[len(incoming)-1]) != "user" {
			return nil, fmt.Errorf("%w: regeneration must retain a user-message prefix", ErrMessageWriteInvalid)
		}
		return mergePrefixAndSuffix(current[:len(incoming)], incoming, true)

	case MessageWriteApproval:
		return reserveApprovalTransition(current, incoming)

	default:
		return nil, fmt.Errorf("%w: unsupported write operation %q", ErrMessageWriteInvalid, operation)
	}
}

func regenerationPrefixLength(current []any, targetMessageID string) (int, error) {
	if len(current) == 0 {
		return 0, ErrMessageWriteConflict
	}

	targetIndex := len(current) - 1
	if targetMessageID != "" {
		targetIndex = -1
		for index, raw := range current {
			message, ok := asObject(raw)
			if ok && message["id"] == targetMessageID {
				targetIndex = index
				break
			}
		}
		if targetIndex < 0 {
			return 0, ErrMessageWriteConflict
		}
	}
	if messageRole(current[targetIndex]) == "assistant" {
		return targetIndex, nil
	}
	return targetIndex + 1, nil
}
