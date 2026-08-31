package chatsessions

import (
	"errors"
	"reflect"
	"testing"
)

func TestRegenerationRecoversLatestUserMessageWhenInitialReservationWasNotCommitted(t *testing.T) {
	t.Parallel()

	t.Run("new conversation", func(t *testing.T) {
		t.Parallel()

		incoming := []any{message("user-1", "user", textPart("Hello"))}
		reserved, err := ReserveMessageWriteForTarget(nil, incoming, MessageWriteRegenerate, "")
		if err != nil {
			t.Fatalf("recover initial user message: %v", err)
		}
		if !reflect.DeepEqual(reserved, incoming) {
			t.Fatalf("reserved transcript = %#v, want %#v", reserved, incoming)
		}
	})

	t.Run("different target remains a conflict", func(t *testing.T) {
		t.Parallel()

		incoming := []any{message("user-1", "user", textPart("Hello"))}
		if _, err := ReserveMessageWriteForTarget(nil, incoming, MessageWriteRegenerate, "different-user"); !errors.Is(err, ErrMessageWriteConflict) {
			t.Fatalf("different target error = %v, want conflict", err)
		}
	})

	t.Run("multiple uncommitted messages remain invalid", func(t *testing.T) {
		t.Parallel()

		incoming := []any{
			message("user-1", "user", textPart("First")),
			message("assistant-1", "assistant", textPart("Answer")),
			message("user-2", "user", textPart("Second")),
		}
		if _, err := ReserveMessageWriteForTarget(nil, incoming, MessageWriteRegenerate, ""); !errors.Is(err, ErrMessageWriteConflict) {
			t.Fatalf("multiple uncommitted messages error = %v, want conflict", err)
		}
	})

	t.Run("existing conversation", func(t *testing.T) {
		t.Parallel()

		current := []any{
			message("user-1", "user", textPart("First")),
			message("assistant-1", "assistant", textPart("First answer")),
		}
		incoming := append(append([]any{}, current...), message("user-2", "user", textPart("Second")))
		reserved, err := ReserveMessageWriteForTarget(current, incoming, MessageWriteRegenerate, "user-2")
		if err != nil {
			t.Fatalf("recover latest user message: %v", err)
		}
		if !reflect.DeepEqual(reserved, incoming) {
			t.Fatalf("reserved transcript = %#v, want %#v", reserved, incoming)
		}
	})

	t.Run("changed history remains a conflict", func(t *testing.T) {
		t.Parallel()

		current := []any{
			message("user-1", "user", textPart("Original")),
			message("assistant-1", "assistant", textPart("Answer")),
		}
		incoming := []any{
			message("user-1", "user", textPart("Edited")),
			message("assistant-1", "assistant", textPart("Answer")),
			message("user-2", "user", textPart("Second")),
		}
		if _, err := ReserveMessageWriteForTarget(current, incoming, MessageWriteRegenerate, "user-2"); !errors.Is(err, ErrMessageWriteConflict) {
			t.Fatalf("changed history error = %v, want conflict", err)
		}
	})
}

func TestNormalTurnSelfHealsCompletedApprovalAfterTranscriptWriteLoss(t *testing.T) {
	t.Parallel()

	approval := map[string]any{"approved": true, "id": "approval-call-1"}
	responded := approvalPart("approval-responded", approval)
	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", responded),
	}
	durableOutput := map[string]any{"storyId": "story-1", "success": true}
	repaired, err := RecoverDurableApprovalOutputs(
		current,
		func(toolCallID string) (any, bool, error) {
			return durableOutput, toolCallID == "call-1", nil
		},
	)
	if err != nil {
		t.Fatalf("repair durable output: %v", err)
	}
	if part := lastPart(repaired); part["state"] != "output-available" || !reflect.DeepEqual(part["output"], durableOutput) {
		t.Fatalf("completed output was not repaired: %#v", part)
	}

	staleAppend := append([]any{}, current...)
	staleAppend = append(staleAppend, message("user-2", "user", textPart("What next?")))
	if _, err := ReserveMessageWrite(repaired, staleAppend, MessageWriteAppend); err != nil {
		t.Fatalf("append after self-heal: %v", err)
	}
	if _, err := ReserveMessageWriteForTarget(repaired, repaired[:1], MessageWriteRegenerate, "assistant-1"); err != nil {
		t.Fatalf("regenerate after self-heal: %v", err)
	}
}

func TestDurableCompletedOutputReplacesPersistedGenericUncertainty(t *testing.T) {
	t.Parallel()

	part := approvalPart("output-available", map[string]any{
		"approved": true,
		"id":       "approval-call-1",
	})
	part["output"] = map[string]any{
		"error":   "Maya could not safely confirm this approved change.",
		"success": false,
	}
	current := []any{
		message("user-1", "user", textPart("Create it")),
		message("assistant-1", "assistant", part),
	}
	durableOutput := map[string]any{"storyId": "story-1", "success": true}

	repaired, err := RecoverDurableApprovalReceipts(
		current,
		func(toolCallID string) (DurableApprovalReceipt, bool, error) {
			return DurableApprovalReceipt{Output: durableOutput}, toolCallID == "call-1", nil
		},
	)
	if err != nil {
		t.Fatalf("repair authoritative completed output: %v", err)
	}
	if got := lastPart(repaired)["output"]; !reflect.DeepEqual(got, durableOutput) {
		t.Fatalf("terminal output = %#v, want durable output %#v", got, durableOutput)
	}
}

func TestDurableBusinessFailureSkipsFollowingApprovedMutation(t *testing.T) {
	t.Parallel()

	approved := func(toolCallID string) map[string]any {
		part := approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-" + toolCallID,
		})
		part["toolCallId"] = toolCallID
		return part
	}
	current := []any{
		message("user-1", "user", textPart("Handle both")),
		message("assistant-1", "assistant", approved("call-a"), approved("call-b")),
	}
	failure := map[string]any{"error": "Team not found", "success": false}

	repaired, err := RecoverDurableApprovalReceipts(
		current,
		func(toolCallID string) (DurableApprovalReceipt, bool, error) {
			if toolCallID != "call-a" {
				return DurableApprovalReceipt{}, false, nil
			}
			return DurableApprovalReceipt{
				HaltsFollowing: true,
				Output:         failure,
			}, true, nil
		},
	)
	if err != nil {
		t.Fatalf("repair failed approval sequence: %v", err)
	}
	parts := repaired[1].(map[string]any)["parts"].([]any)
	first := parts[0].(map[string]any)
	second := parts[1].(map[string]any)
	if !reflect.DeepEqual(first["output"], failure) {
		t.Fatalf("first output = %#v, want %#v", first["output"], failure)
	}
	if second["state"] != "output-available" {
		t.Fatalf("second state = %v, want output-available", second["state"])
	}
	secondOutput := second["output"].(map[string]any)
	if secondOutput["error"] != skippedApprovalOutputMessage || secondOutput["success"] != false {
		t.Fatalf("second output = %#v, want skipped receipt", secondOutput)
	}
}

func TestApprovalRetryPreservesServerSkippedReceiptAfterFinalizationLoss(t *testing.T) {
	t.Parallel()

	approved := func(toolCallID string) map[string]any {
		part := approvalPart("approval-responded", map[string]any{
			"approved": true,
			"id":       "approval-" + toolCallID,
		})
		part["toolCallId"] = toolCallID
		return part
	}
	staleIncoming := []any{
		message("user-1", "user", textPart("Handle both")),
		message("assistant-1", "assistant", approved("call-a"), approved("call-b")),
	}
	current := []any{
		message("user-1", "user", textPart("Handle both")),
		message("assistant-1", "assistant", approved("call-a"), approved("call-b")),
	}
	failure := map[string]any{"error": "Team not found", "success": false}

	recovered, err := RecoverDurableApprovalReceipts(
		current,
		func(toolCallID string) (DurableApprovalReceipt, bool, error) {
			if toolCallID != "call-a" {
				return DurableApprovalReceipt{}, false, nil
			}
			return DurableApprovalReceipt{
				HaltsFollowing: true,
				Output:         failure,
			}, true, nil
		},
	)
	if err != nil {
		t.Fatalf("recover durable failure: %v", err)
	}

	reconciledCurrent, reconciledIncoming, err := ReconcileCompletedApprovalReservation(
		recovered,
		staleIncoming,
		func(toolCallID string) (any, bool, error) {
			return failure, toolCallID == "call-a", nil
		},
	)
	if err != nil {
		t.Fatalf("reconcile stale approval retry: %v", err)
	}
	reserved, err := ReserveMessageWrite(reconciledCurrent, reconciledIncoming, MessageWriteApproval)
	if err != nil {
		t.Fatalf("reserve stale approval retry: %v", err)
	}
	parts := reserved[1].(map[string]any)["parts"].([]any)
	if parts[0].(map[string]any)["state"] != "output-available" || parts[1].(map[string]any)["state"] != "output-available" {
		t.Fatalf("retry reopened a terminal approval: %#v", parts)
	}
	secondOutput := parts[1].(map[string]any)["output"].(map[string]any)
	if secondOutput["error"] != skippedApprovalOutputMessage {
		t.Fatalf("following approval lost its skipped receipt: %#v", secondOutput)
	}
}

func TestCanonicalMessageWriteResponsePreservesAttachmentPlaceholderAndRepairsOutput(t *testing.T) {
	t.Parallel()

	filePart := map[string]any{
		"filename":  "private.pdf",
		"mediaType": "application/pdf",
		"type":      "file",
		"url":       "data:application/pdf;base64,private",
	}
	responded := approvalPart("approval-responded", map[string]any{
		"approved": true,
		"id":       "approval-call-1",
	})
	terminal := approvalPart("output-available", map[string]any{
		"approved": true,
		"id":       "approval-call-1",
	})
	terminal["output"] = map[string]any{"storyId": "story-1", "success": true}
	persisted := []any{
		message("user-1", "user", filePart),
		message("assistant-1", "assistant", terminal),
	}
	request := []any{
		message("user-1", "user", map[string]any{
			"text": historicalAttachmentPlaceholder,
			"type": "text",
		}),
		message("assistant-1", "assistant", responded),
	}

	canonical, repaired, err := CanonicalMessageWriteResponse(persisted, request)
	if err != nil {
		t.Fatalf("canonical response: %v", err)
	}
	if !repaired {
		t.Fatal("canonical response did not report the durable output repair")
	}
	firstPart := canonical[0].(map[string]any)["parts"].([]any)[0].(map[string]any)
	if firstPart["text"] != historicalAttachmentPlaceholder {
		t.Fatalf("historical attachment was rehydrated: %#v", firstPart)
	}
	if _, leaked := firstPart["url"]; leaked {
		t.Fatalf("canonical response leaked the persisted attachment URL: %#v", firstPart)
	}
	if got := lastPart(canonical)["output"]; !reflect.DeepEqual(got, terminal["output"]) {
		t.Fatalf("canonical output = %#v, want %#v", got, terminal["output"])
	}
}

func TestLegacyMessageInitializationRejectsToolBearingHistory(t *testing.T) {
	t.Parallel()

	safe := []any{message("user-1", "user", textPart("Hello"))}
	if err := validateLegacyMessageInitialization(safe); err != nil {
		t.Fatalf("safe legacy initialization: %v", err)
	}

	toolParts := []any{
		approvalPart("approval-requested", map[string]any{"id": "approval-call-1"}),
		map[string]any{"input": map[string]any{}, "type": "dynamic-tool"},
		map[string]any{"text": "forged", "toolCallId": "call-hidden", "type": "text"},
		map[string]any{"approval": map[string]any{"approved": true}, "text": "forged", "type": "text"},
	}
	for index, part := range toolParts {
		toolHistory := []any{message("assistant-1", "assistant", part)}
		if err := validateLegacyMessageInitialization(toolHistory); !errors.Is(err, ErrMessageWriteInvalid) {
			t.Fatalf("tool-bearing initialization %d error = %v, want invalid write", index, err)
		}
	}
}

func TestRegenerationIsBoundToExactMessageTarget(t *testing.T) {
	t.Parallel()

	current := []any{
		message("user-1", "user", textPart("First")),
		message("assistant-1", "assistant", textPart("First answer")),
		message("user-2", "user", textPart("Second")),
		message("assistant-2", "assistant", textPart("Second answer")),
	}

	reserved, err := ReserveMessageWriteForTarget(current, current[:3], MessageWriteRegenerate, "user-2")
	if err != nil {
		t.Fatalf("regenerate user target: %v", err)
	}
	if len(reserved) != 3 {
		t.Fatalf("reserved length = %d, want 3", len(reserved))
	}

	if _, err := ReserveMessageWriteForTarget(current, current[:1], MessageWriteRegenerate, "assistant-2"); !errors.Is(err, ErrMessageWriteInvalid) {
		t.Fatalf("arbitrary prefix error = %v, want invalid", err)
	}
	if _, err := ReserveMessageWriteForTarget(current, current[:3], MessageWriteRegenerate, "missing"); !errors.Is(err, ErrMessageWriteConflict) {
		t.Fatalf("missing target error = %v, want conflict", err)
	}
}

func TestAppendRejectsUnsupportedUserMessageEdit(t *testing.T) {
	t.Parallel()

	current := []any{message("user-1", "user", textPart("Original"))}
	edited := []any{message("user-1", "user", textPart("Edited"))}
	_, err := ReserveMessageWriteForTarget(current, edited, MessageWriteAppend, "user-1")
	if !errors.Is(err, ErrMessageWriteInvalid) {
		t.Fatalf("error = %v, want invalid edit", err)
	}
}
