package chatsessions

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const historicalAttachmentPlaceholder = "[Historical attachment omitted from this request.]"

type CompletedApprovalOutputLookup func(toolCallID string) (output any, found bool, err error)

type DurableApprovalReceipt struct {
	Output         any
	HaltsFollowing bool
}

type DurableApprovalReceiptLookup func(toolCallID string) (receipt DurableApprovalReceipt, found bool, err error)

const skippedApprovalOutputMessage = "Maya did not run this approved change because an earlier approved change was unresolved. Review the earlier result, then ask Maya to prepare this change again."

// RecoverDurableApprovalOutputs repairs any approval-responded part whose
// mutation already has a completed ledger receipt. This lets a later normal
// turn or regeneration self-heal after both transcript finalization attempts
// failed, without ever replaying the mutation.
func RecoverDurableApprovalOutputs(current []any, lookup CompletedApprovalOutputLookup) ([]any, error) {
	return RecoverDurableApprovalReceipts(current, func(toolCallID string) (DurableApprovalReceipt, bool, error) {
		output, found, err := lookup(toolCallID)
		return DurableApprovalReceipt{Output: output}, found, err
	})
}

// RecoverDurableApprovalReceipts repairs terminal ledger results before a new
// turn. Durable state is authoritative even when a prior stream persisted a
// generic terminal error after losing the completion response. An uncertain,
// expired, or failed completed result also terminalizes later approved calls
// in the same assistant message because ordered execution must stop there.
func RecoverDurableApprovalReceipts(current []any, lookup DurableApprovalReceiptLookup) ([]any, error) {
	recoveredValue, err := cloneJSONValue(current)
	if err != nil {
		return nil, err
	}
	recovered, ok := recoveredValue.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: transcript must be an array", ErrMessageWriteInvalid)
	}
	for _, rawMessage := range recovered {
		message, ok := asObject(rawMessage)
		if !ok {
			continue
		}
		parts, ok := asArray(message["parts"])
		if !ok {
			continue
		}
		haltsFollowing := false
		for _, rawPart := range parts {
			part, ok := asObject(rawPart)
			if !ok {
				continue
			}
			state := toolState(part)
			if state != "approval-responded" && state != "output-available" {
				continue
			}
			if approvalDenied(part["approval"]) {
				part["state"] = "output-denied"
				delete(part, "output")
				delete(part, "errorText")
				continue
			}
			toolCallID, ok := part["toolCallId"].(string)
			if !ok || toolCallID == "" {
				continue
			}
			if state == "output-available" && !toolOutputIndicatesFailure(part["output"]) {
				// Successful terminal receipts already came through the reservation
				// CAS. Only failure-shaped terminal output can be an internal
				// uncertainty that must be replaced or a halt that must propagate.
				continue
			}
			receipt, found, err := lookup(toolCallID)
			if err != nil {
				return nil, err
			}
			if state == "output-available" {
				if found {
					part["output"] = receipt.Output
					haltsFollowing = haltsFollowing || receipt.HaltsFollowing
					delete(part, "errorText")
				}
				continue
			}
			if !found && !haltsFollowing {
				continue
			}
			part["state"] = "output-available"
			if found {
				part["output"] = receipt.Output
				haltsFollowing = haltsFollowing || receipt.HaltsFollowing
			} else {
				part["output"] = map[string]any{
					"error":   skippedApprovalOutputMessage,
					"success": false,
				}
			}
			delete(part, "errorText")
		}
	}
	return recovered, nil
}

// RecoverCompletedApprovalOutputsForReservation accepts a partially delivered
// multi-approval retry only when every newly terminal part exactly matches the
// durable ledger output. The returned transcript remains subject to the normal
// approval-transition validator before a new generation is reserved.
func RecoverCompletedApprovalOutputsForReservation(current, incoming []any, lookup CompletedApprovalOutputLookup) ([]any, error) {
	recovered, _, err := ReconcileCompletedApprovalReservation(current, incoming, lookup)
	return recovered, err
}

// ReconcileCompletedApprovalReservation canonicalizes either side of a
// response-loss retry to the exact completed ledger output. This handles both
// a client that is ahead of storage after a partial stream and storage that is
// ahead of a client whose finalization response was lost.
func ReconcileCompletedApprovalReservation(current, incoming []any, lookup CompletedApprovalOutputLookup) ([]any, []any, error) {
	recoveredValue, err := cloneJSONValue(current)
	if err != nil {
		return nil, nil, err
	}
	recovered, ok := recoveredValue.([]any)
	reconciledIncomingValue, err := cloneJSONValue(incoming)
	if err != nil {
		return nil, nil, err
	}
	reconciledIncoming, incomingArray := reconciledIncomingValue.([]any)
	if !ok || !incomingArray || len(recovered) == 0 || len(recovered) != len(incoming) {
		return current, incoming, nil
	}
	currentMessage, currentOK := asObject(recovered[len(recovered)-1])
	incomingMessage, incomingOK := asObject(reconciledIncoming[len(reconciledIncoming)-1])
	if !currentOK || !incomingOK || !sameMessageIdentity(currentMessage, incomingMessage) {
		return recovered, reconciledIncoming, nil
	}
	currentParts, currentOK := asArray(currentMessage["parts"])
	incomingParts, incomingOK := asArray(incomingMessage["parts"])
	if !currentOK || !incomingOK || len(currentParts) != len(incomingParts) {
		return recovered, reconciledIncoming, nil
	}

	for index := range currentParts {
		currentPart, currentObject := asObject(currentParts[index])
		incomingPart, incomingObject := asObject(incomingParts[index])
		if !currentObject || !incomingObject {
			continue
		}
		currentState := toolState(currentPart)
		incomingState := toolState(incomingPart)
		clientAhead := currentState == "approval-responded" && incomingState == "output-available"
		serverAhead := currentState == "output-available" && incomingState == "approval-responded"
		if !clientAhead && !serverAhead {
			continue
		}
		if !sameToolPartIdentity(currentPart, incomingPart) || !sameToolPartBase(currentPart, incomingPart) {
			return nil, nil, ErrMessageWriteConflict
		}
		terminalPart := incomingPart
		if serverAhead {
			// Storage is authoritative over a stale approval response. This also
			// preserves server-generated uncertain, expired, and skipped receipts
			// that intentionally have no completed ledger output of their own.
			copy, err := cloneJSONValue(currentPart)
			if err != nil {
				return nil, nil, err
			}
			incomingParts[index] = copy
			continue
		}
		terminalOutput, exists := terminalPart["output"]
		if !exists {
			return nil, nil, fmt.Errorf("%w: tool output is missing", ErrMessageWriteInvalid)
		}
		toolCallID := currentPart["toolCallId"].(string)
		durableOutput, found, err := lookup(toolCallID)
		if err != nil {
			return nil, nil, err
		}
		if !found || !reflect.DeepEqual(durableOutput, terminalOutput) {
			return nil, nil, ErrMessageWriteConflict
		}
		currentPart["state"] = "output-available"
		currentPart["output"] = durableOutput
		delete(currentPart, "errorText")
	}
	return recovered, reconciledIncoming, nil
}

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
	return current["type"] == "file" && incoming["type"] == "text" && incoming["text"] == historicalAttachmentPlaceholder
}

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
