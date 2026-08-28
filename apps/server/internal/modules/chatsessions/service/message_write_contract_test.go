package chatsessions

import (
	"errors"
	"reflect"
	"testing"
)

func TestReserveMessageWritePreservesHistoricalAttachments(t *testing.T) {
	t.Parallel()

	filePart := map[string]any{
		"filename":  "brief.pdf",
		"mediaType": "application/pdf",
		"type":      "file",
		"url":       "data:application/pdf;base64,private",
	}
	current := []any{
		message("legacy-user", "user", filePart),
		message("assistant-1", "assistant", textPart("I read the brief.")),
	}
	incoming := []any{
		message("legacy-user", "user", map[string]any{
			"text": historicalAttachmentPlaceholder,
			"type": "text",
		}),
		message("assistant-1", "assistant", textPart("I read the brief.")),
		message("user-2", "user", textPart("Create the stories.")),
	}

	reserved, err := ReserveMessageWrite(current, incoming, MessageWriteAppend)
	if err != nil {
		t.Fatalf("reserve append: %v", err)
	}
	first := reserved[0].(map[string]any)
	parts := first["parts"].([]any)
	if !reflect.DeepEqual(parts[0], filePart) {
		t.Fatalf("historical file was not preserved: %#v", parts[0])
	}
}

func TestReserveMessageWriteAllowsLegacyMissingAndDuplicatePrefixIDs(t *testing.T) {
	t.Parallel()

	first := message("", "user", textPart("Legacy one"))
	delete(first, "id")
	current := []any{
		first,
		message("duplicate", "assistant", textPart("First answer")),
		message("duplicate", "user", textPart("Legacy follow-up")),
	}
	incoming := append([]any{}, current...)
	incoming = append(incoming, message("fresh-user", "user", textPart("Continue")))

	reserved, err := ReserveMessageWrite(current, incoming, MessageWriteAppend)
	if err != nil {
		t.Fatalf("legacy prefix should remain writable: %v", err)
	}
	if len(reserved) != 4 {
		t.Fatalf("reserved length = %d, want 4", len(reserved))
	}
}

func TestReserveMessageWriteRejectsDuplicateNewMessageID(t *testing.T) {
	t.Parallel()

	current := []any{message("user-1", "user", textPart("Existing"))}
	incoming := append([]any{}, current...)
	incoming = append(incoming, message("user-1", "user", textPart("Duplicate")))

	_, err := ReserveMessageWrite(current, incoming, MessageWriteAppend)
	if !errors.Is(err, ErrMessageWriteInvalid) {
		t.Fatalf("error = %v, want invalid write", err)
	}
}

func TestReserveMessageWriteRejectsRegenerationAcrossUnresolvedApproval(t *testing.T) {
	t.Parallel()

	current := []any{
		message("user-1", "user", textPart("Delete it")),
		message("assistant-1", "assistant", approvalPart("approval-requested", map[string]any{"id": "approval-call-1"})),
	}
	incoming := current[:1]

	_, err := ReserveMessageWrite(current, incoming, MessageWriteRegenerate)
	if !errors.Is(err, ErrMessageWriteApprovalOpen) {
		t.Fatalf("error = %v, want open approval", err)
	}
}

func TestHistoricalAbandonedApprovalDoesNotBrickNewTurns(t *testing.T) {
	t.Parallel()

	current := []any{
		message("user-1", "user", textPart("Old request")),
		message("assistant-1", "assistant", approvalPart("approval-requested", map[string]any{"id": "approval-call-1"})),
		message("user-2", "user", textPart("Ignore that")),
		message("assistant-2", "assistant", textPart("Understood.")),
	}
	incoming := append([]any{}, current...)
	incoming = append(incoming, message("user-3", "user", textPart("New request")))

	if _, err := ReserveMessageWrite(current, incoming, MessageWriteAppend); err != nil {
		t.Fatalf("historical non-actionable approval should not block append: %v", err)
	}
}

func TestApprovalWriteOnlyMovesRequestedToResponded(t *testing.T) {
	t.Parallel()

	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-requested", map[string]any{"id": "approval-call-1"})),
	}
	incoming := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-call-1",
		})),
	}

	reserved, err := ReserveMessageWrite(current, incoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve approval: %v", err)
	}
	part := lastPart(reserved)
	if part["state"] != "approval-responded" {
		t.Fatalf("state = %v, want approval-responded", part["state"])
	}
}

func TestApprovalWriteBindsPersistedApprovalID(t *testing.T) {
	t.Parallel()

	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-requested", map[string]any{"id": "approval-original"})),
	}
	incoming := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-forged",
		})),
	}

	if _, err := ReserveMessageWrite(current, incoming, MessageWriteApproval); !errors.Is(err, ErrMessageWriteConflict) {
		t.Fatalf("error = %v, want approval identity conflict", err)
	}
}

func TestDeniedApprovalIsTerminalInTheReservationTransaction(t *testing.T) {
	t.Parallel()

	requestedApproval := map[string]any{"id": "approval-call-1"}
	deniedApproval := map[string]any{
		"approved": false,
		"id":       "approval-call-1",
	}
	current := []any{
		message("user-1", "user", textPart("Delete it")),
		message("assistant-1", "assistant", approvalPart("approval-requested", requestedApproval)),
	}
	incoming := []any{
		message("user-1", "user", textPart("Delete it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", deniedApproval)),
	}

	reserved, err := ReserveMessageWrite(current, incoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve denial: %v", err)
	}
	if part := lastPart(reserved); part["state"] != "output-denied" {
		t.Fatalf("denial state = %v, want output-denied", part["state"])
	}

	// A response-loss retry is idempotent even though denial has no ledger row.
	replayed, err := ReserveMessageWrite(reserved, incoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("replay denial: %v", err)
	}
	if part := lastPart(replayed); part["state"] != "output-denied" {
		t.Fatalf("replayed denial state = %v", part["state"])
	}
}

func TestDeniedApprovalWriteLossDoesNotBlockNextTurn(t *testing.T) {
	t.Parallel()

	deniedApproval := map[string]any{
		"approved": false,
		"id":       "approval-call-1",
	}
	// This is the pre-fix state left behind if the UI output chunk arrived but
	// transcript finalization failed.
	current := []any{
		message("user-1", "user", textPart("Delete it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", deniedApproval)),
	}
	repaired, err := RecoverDurableApprovalOutputs(
		current,
		func(string) (any, bool, error) { return nil, false, nil },
	)
	if err != nil {
		t.Fatalf("repair denial: %v", err)
	}
	if part := lastPart(repaired); part["state"] != "output-denied" {
		t.Fatalf("repaired denial state = %v", part["state"])
	}

	staleAppend := append([]any{}, current...)
	staleAppend = append(staleAppend, message("user-2", "user", textPart("Keep going")))
	if _, err := ReserveMessageWrite(repaired, staleAppend, MessageWriteAppend); err != nil {
		t.Fatalf("append after denial repair: %v", err)
	}
	if _, err := ReserveMessageWriteForTarget(repaired, repaired[:1], MessageWriteRegenerate, "assistant-1"); err != nil {
		t.Fatalf("regenerate after denial repair: %v", err)
	}
}

func TestMixedMultiApprovalDenialAndApprovalRemainOrdered(t *testing.T) {
	t.Parallel()

	requested := func(toolCallID string) map[string]any {
		part := approvalPart("approval-requested", map[string]any{"id": "approval-" + toolCallID})
		part["toolCallId"] = toolCallID
		return part
	}
	responded := func(toolCallID string, approved bool) map[string]any {
		part := approvalPart("approval-responded", map[string]any{
			"approved": approved,
			"id":       "approval-" + toolCallID,
		})
		part["toolCallId"] = toolCallID
		return part
	}
	current := []any{
		message("user-1", "user", textPart("Handle both")),
		message("assistant-1", "assistant", requested("call-a"), requested("call-b")),
	}
	incoming := []any{
		message("user-1", "user", textPart("Handle both")),
		message("assistant-1", "assistant", responded("call-a", false), responded("call-b", true)),
	}
	reserved, err := ReserveMessageWrite(current, incoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve mixed approvals: %v", err)
	}
	parts := reserved[len(reserved)-1].(map[string]any)["parts"].([]any)
	if parts[0].(map[string]any)["state"] != "output-denied" || parts[1].(map[string]any)["state"] != "approval-responded" {
		t.Fatalf("mixed approval states were reordered or lost: %#v", parts)
	}
}

func TestStaleSameLengthApprovalCannotRollBackOutput(t *testing.T) {
	t.Parallel()

	output := map[string]any{"storyId": "story-1", "success": true}
	terminalPart := approvalPart("output-available", map[string]any{
		"approved": true,
		"id":       "approval-call-1",
	})
	terminalPart["output"] = output
	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", terminalPart),
	}
	stale := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-call-1",
		})),
	}

	if _, err := ReserveMessageWrite(current, stale, MessageWriteApproval); !errors.Is(err, ErrMessageWriteConflict) {
		t.Fatalf("reserve error = %v, want conflict", err)
	}
	if _, err := FinalizeMessageWriteTransition(current, stale, MessageWriteApproval); !errors.Is(err, ErrMessageWriteConflict) {
		t.Fatalf("finalize error = %v, want conflict", err)
	}
}

func TestApprovalFinalizationProducesTerminalOutput(t *testing.T) {
	t.Parallel()

	approval := map[string]any{"approved": true, "id": "approval-call-1"}
	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", approval)),
	}
	finishedPart := approvalPart("output-available", approval)
	finishedPart["output"] = map[string]any{"storyId": "story-1", "success": true}
	incoming := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", finishedPart),
	}

	finished, err := FinalizeMessageWriteTransition(current, incoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("finalize approval: %v", err)
	}
	part := lastPart(finished)
	if part["state"] != "output-available" || !reflect.DeepEqual(part["output"], finishedPart["output"]) {
		t.Fatalf("unexpected terminal part: %#v", part)
	}
}

func TestPartialMultiApprovalRetryRecoversCompletedOutputBeforeReservation(t *testing.T) {
	t.Parallel()

	responded := func(toolCallID string) map[string]any {
		part := approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-" + toolCallID,
		})
		part["toolCallId"] = toolCallID
		return part
	}
	current := []any{
		message("user-1", "user", textPart("Create both")),
		message("assistant-1", "assistant", responded("call-a"), responded("call-b")),
	}
	durableOutput := map[string]any{"storyId": "story-a", "success": true}
	completedA := responded("call-a")
	completedA["state"] = "output-available"
	completedA["output"] = durableOutput
	incoming := []any{
		message("user-1", "user", textPart("Create both")),
		message("assistant-1", "assistant", completedA, responded("call-b")),
	}

	recovered, err := RecoverCompletedApprovalOutputsForReservation(
		current,
		incoming,
		func(toolCallID string) (any, bool, error) {
			if toolCallID == "call-a" {
				return durableOutput, true, nil
			}
			return nil, false, nil
		},
	)
	if err != nil {
		t.Fatalf("recover partial output: %v", err)
	}
	reserved, err := ReserveMessageWrite(recovered, incoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve remaining approval: %v", err)
	}
	last := reserved[len(reserved)-1].(map[string]any)
	parts := last["parts"].([]any)
	partA := parts[0].(map[string]any)
	partB := parts[1].(map[string]any)
	if partA["state"] != "output-available" || !reflect.DeepEqual(partA["output"], durableOutput) {
		t.Fatalf("completed approval A was not preserved: %#v", partA)
	}
	if partB["state"] != "approval-responded" {
		t.Fatalf("approval B state = %v, want approval-responded", partB["state"])
	}
}

func TestServerAheadApprovalRetryKeepsExactLedgerOutput(t *testing.T) {
	t.Parallel()

	approval := map[string]any{"approved": true, "id": "approval-call-1"}
	durableOutput := map[string]any{"storyId": "story-1", "success": true}
	terminal := approvalPart("output-available", approval)
	terminal["output"] = durableOutput
	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", terminal),
	}
	staleIncoming := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", approvalPart("approval-responded", approval)),
	}

	reconciledCurrent, reconciledIncoming, err := ReconcileCompletedApprovalReservation(
		current,
		staleIncoming,
		func(toolCallID string) (any, bool, error) {
			return durableOutput, toolCallID == "call-1", nil
		},
	)
	if err != nil {
		t.Fatalf("reconcile server-ahead retry: %v", err)
	}
	reserved, err := ReserveMessageWrite(reconciledCurrent, reconciledIncoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve server-ahead retry: %v", err)
	}
	if part := lastPart(reserved); part["state"] != "output-available" || !reflect.DeepEqual(part["output"], durableOutput) {
		t.Fatalf("durable output was rolled back: %#v", part)
	}
}

func TestAllCompletedMultiApprovalRetryIsIdempotent(t *testing.T) {
	t.Parallel()

	approval := func(toolCallID string) map[string]any {
		part := approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-" + toolCallID,
		})
		part["toolCallId"] = toolCallID
		return part
	}
	outputA := map[string]any{"storyId": "story-a", "success": true}
	outputB := map[string]any{"storyId": "story-b", "success": true}
	terminalA := approval("call-a")
	terminalA["state"] = "output-available"
	terminalA["output"] = outputA
	terminalB := approval("call-b")
	terminalB["state"] = "output-available"
	terminalB["output"] = outputB
	current := []any{
		message("user-1", "user", textPart("Create both")),
		message("assistant-1", "assistant", terminalA, terminalB),
	}
	staleIncoming := []any{
		message("user-1", "user", textPart("Create both")),
		message("assistant-1", "assistant", approval("call-a"), approval("call-b")),
	}
	outputs := map[string]any{"call-a": outputA, "call-b": outputB}
	reconciledCurrent, reconciledIncoming, err := ReconcileCompletedApprovalReservation(
		current,
		staleIncoming,
		func(toolCallID string) (any, bool, error) {
			output, found := outputs[toolCallID]
			return output, found, nil
		},
	)
	if err != nil {
		t.Fatalf("reconcile completed approvals: %v", err)
	}
	reserved, err := ReserveMessageWrite(reconciledCurrent, reconciledIncoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve completed replay: %v", err)
	}
	parts := reserved[len(reserved)-1].(map[string]any)["parts"].([]any)
	if parts[0].(map[string]any)["state"] != "output-available" || parts[1].(map[string]any)["state"] != "output-available" {
		t.Fatalf("terminal states were not preserved: %#v", parts)
	}
}
