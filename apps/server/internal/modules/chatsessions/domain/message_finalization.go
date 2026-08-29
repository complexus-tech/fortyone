package chatsessionsdomain

import (
	"fmt"
	"reflect"
)

// FinalizeMessageWriteTransition applies only the response suffix or approval
// state transition owned by a reservation. It never accepts a rewritten or
// reordered history prefix.
func FinalizeMessageWriteTransition(current, incoming []any, operation MessageWriteOperation) ([]any, error) {
	if err := validateMessageSequence(incoming); err != nil {
		return nil, err
	}

	switch operation {
	case MessageWriteAppend, MessageWriteRegenerate:
		if len(incoming) < len(current) {
			return nil, ErrMessageWriteConflict
		}
		merged, err := mergePrefixAndSuffix(current, incoming, true)
		if err != nil {
			return nil, err
		}
		for _, message := range incoming[len(current):] {
			if messageRole(message) != "assistant" {
				return nil, fmt.Errorf("%w: model finalization may only append assistant messages", ErrMessageWriteInvalid)
			}
		}
		if err := validateNewMessageIdentities(current, incoming[len(current):]); err != nil {
			return nil, err
		}
		return merged, nil

	case MessageWriteApproval:
		return finalizeApprovalTransition(current, incoming)

	default:
		return nil, fmt.Errorf("%w: unsupported write operation %q", ErrMessageWriteInvalid, operation)
	}
}

func mergePrefixAndSuffix(current, incoming []any, allowMonotonicToolOutput bool) ([]any, error) {
	if len(incoming) < len(current) {
		return nil, ErrMessageWriteConflict
	}

	merged := make([]any, 0, len(incoming))
	for index := range current {
		if !messagesCompatible(current[index], incoming[index], allowMonotonicToolOutput) {
			return nil, ErrMessageWriteConflict
		}
		copy, err := cloneJSONValue(current[index])
		if err != nil {
			return nil, err
		}
		merged = append(merged, copy)
	}
	for _, message := range incoming[len(current):] {
		copy, err := cloneJSONValue(message)
		if err != nil {
			return nil, err
		}
		merged = append(merged, copy)
	}
	return merged, nil
}

func reserveApprovalTransition(current, incoming []any) ([]any, error) {
	if len(current) == 0 || len(current) != len(incoming) {
		return nil, ErrMessageWriteConflict
	}
	for index := 0; index < len(current)-1; index++ {
		if !messagesCompatible(current[index], incoming[index], true) {
			return nil, ErrMessageWriteConflict
		}
	}

	currentMessage, ok := asObject(current[len(current)-1])
	if !ok || messageRole(currentMessage) != "assistant" {
		return nil, ErrMessageWriteConflict
	}
	incomingMessage, ok := asObject(incoming[len(incoming)-1])
	if !ok || !sameMessageIdentity(currentMessage, incomingMessage) {
		return nil, ErrMessageWriteConflict
	}

	currentParts, ok := asArray(currentMessage["parts"])
	if !ok {
		return nil, fmt.Errorf("%w: assistant message parts are missing", ErrMessageWriteInvalid)
	}
	incomingParts, ok := asArray(incomingMessage["parts"])
	if !ok || len(currentParts) != len(incomingParts) {
		return nil, ErrMessageWriteConflict
	}

	mergedParts := make([]any, len(currentParts))
	hasApprovalState := false
	for index := range currentParts {
		currentPart, currentIsObject := asObject(currentParts[index])
		incomingPart, incomingIsObject := asObject(incomingParts[index])
		if !currentIsObject || !incomingIsObject {
			if !reflect.DeepEqual(currentParts[index], incomingParts[index]) {
				return nil, ErrMessageWriteConflict
			}
			mergedParts[index] = currentParts[index]
			continue
		}

		if isHistoricalFilePlaceholder(currentPart, incomingPart) || reflect.DeepEqual(currentPart, incomingPart) {
			mergedPart := currentPart
			if toolState(currentPart) == "approval-responded" {
				if !validApprovalTransition(currentPart["approval"], incomingPart["approval"]) {
					return nil, ErrMessageWriteConflict
				}
				if approvalDenied(currentPart["approval"]) {
					clonedPart, cloneErr := cloneObject(currentPart)
					if cloneErr != nil {
						return nil, cloneErr
					}
					mergedPart = clonedPart
					mergedPart["state"] = "output-denied"
				}
				hasApprovalState = true
			} else if isTerminalToolState(toolState(currentPart)) {
				hasApprovalState = true
			}
			mergedParts[index] = mergedPart
			continue
		}
		if !sameToolPartIdentity(currentPart, incomingPart) || !sameToolPartBase(currentPart, incomingPart) {
			return nil, ErrMessageWriteConflict
		}
		if toolState(currentPart) == "output-denied" && toolState(incomingPart) == "approval-responded" && validApprovalTransition(currentPart["approval"], incomingPart["approval"]) && approvalDenied(incomingPart["approval"]) {
			mergedParts[index] = currentPart
			hasApprovalState = true
			continue
		}
		if toolState(currentPart) != "approval-requested" || toolState(incomingPart) != "approval-responded" || !validApprovalTransition(currentPart["approval"], incomingPart["approval"]) {
			return nil, ErrMessageWriteConflict
		}
		copiedPart, err := cloneObject(incomingPart)
		if err != nil {
			return nil, err
		}
		if approvalDenied(incomingPart["approval"]) {
			copiedPart["state"] = "output-denied"
		}
		mergedParts[index] = copiedPart
		hasApprovalState = true
	}
	if !hasApprovalState {
		return nil, fmt.Errorf("%w: approval write contains no response", ErrMessageWriteInvalid)
	}

	mergedLast, err := cloneObject(currentMessage)
	if err != nil {
		return nil, err
	}
	mergedLast["parts"] = mergedParts
	merged := make([]any, len(current))
	for index := 0; index < len(current)-1; index++ {
		merged[index], err = cloneJSONValue(current[index])
		if err != nil {
			return nil, err
		}
	}
	merged[len(merged)-1] = mergedLast
	return merged, nil
}

func finalizeApprovalTransition(current, incoming []any) ([]any, error) {
	if len(current) == 0 || len(current) != len(incoming) {
		return nil, ErrMessageWriteConflict
	}
	for index := 0; index < len(current)-1; index++ {
		if !messagesCompatible(current[index], incoming[index], true) {
			return nil, ErrMessageWriteConflict
		}
	}

	currentMessage, currentOK := asObject(current[len(current)-1])
	incomingMessage, incomingOK := asObject(incoming[len(incoming)-1])
	if !currentOK || !incomingOK || !sameMessageIdentity(currentMessage, incomingMessage) || messageRole(currentMessage) != "assistant" {
		return nil, ErrMessageWriteConflict
	}
	currentParts, currentOK := asArray(currentMessage["parts"])
	incomingParts, incomingOK := asArray(incomingMessage["parts"])
	if !currentOK || !incomingOK || len(currentParts) != len(incomingParts) {
		return nil, ErrMessageWriteConflict
	}

	mergedParts := make([]any, len(currentParts))
	for index := range currentParts {
		currentPart, currentIsObject := asObject(currentParts[index])
		incomingPart, incomingIsObject := asObject(incomingParts[index])
		if !currentIsObject || !incomingIsObject {
			if !reflect.DeepEqual(currentParts[index], incomingParts[index]) {
				return nil, ErrMessageWriteConflict
			}
			mergedParts[index] = currentParts[index]
			continue
		}
		if reflect.DeepEqual(currentPart, incomingPart) || isHistoricalFilePlaceholder(currentPart, incomingPart) {
			mergedParts[index] = currentPart
			continue
		}
		if !sameToolPartIdentity(currentPart, incomingPart) || !sameToolPartBase(currentPart, incomingPart) {
			return nil, ErrMessageWriteConflict
		}

		currentState := toolState(currentPart)
		incomingState := toolState(incomingPart)
		if isTerminalToolState(currentState) {
			// A terminal receipt can only be replayed byte-for-byte. In
			// particular, a stale approval response cannot roll it backward.
			return nil, ErrMessageWriteConflict
		}
		if currentState == "approval-requested" && incomingState == "approval-requested" {
			mergedParts[index] = currentPart
			continue
		}
		if currentState != "approval-responded" || (incomingState != "output-available" && incomingState != "output-denied") {
			return nil, ErrMessageWriteConflict
		}
		if incomingState == "output-available" {
			if _, exists := incomingPart["output"]; !exists {
				return nil, fmt.Errorf("%w: tool output is missing", ErrMessageWriteInvalid)
			}
		} else if approval, exists := currentPart["approval"]; !exists || !approvalDenied(approval) {
			return nil, ErrMessageWriteConflict
		}

		mergedPart, err := cloneObject(currentPart)
		if err != nil {
			return nil, err
		}
		mergedPart["state"] = incomingState
		delete(mergedPart, "errorText")
		if incomingState == "output-available" {
			mergedPart["output"], err = cloneJSONValue(incomingPart["output"])
			if err != nil {
				return nil, err
			}
		} else {
			delete(mergedPart, "output")
		}
		mergedParts[index] = mergedPart
	}

	mergedLast, err := cloneObject(currentMessage)
	if err != nil {
		return nil, err
	}
	mergedLast["parts"] = mergedParts
	merged := make([]any, len(current))
	for index := 0; index < len(current)-1; index++ {
		merged[index], err = cloneJSONValue(current[index])
		if err != nil {
			return nil, err
		}
	}
	merged[len(merged)-1] = mergedLast
	return merged, nil
}
