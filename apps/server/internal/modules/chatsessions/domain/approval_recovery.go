package chatsessionsdomain

import (
	"fmt"
	"reflect"
)

const HistoricalAttachmentPlaceholder = "[Historical attachment omitted from this request.]"

type CompletedApprovalOutputLookup func(toolCallID string) (output any, found bool, err error)

type DurableApprovalReceipt struct {
	Output         any
	HaltsFollowing bool
}

type DurableApprovalReceiptLookup func(toolCallID string) (receipt DurableApprovalReceipt, found bool, err error)

const SkippedApprovalOutputMessage = "Maya did not run this approved change because an earlier approved change was unresolved. Review the earlier result, then ask Maya to prepare this change again."

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
					"error":   SkippedApprovalOutputMessage,
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
