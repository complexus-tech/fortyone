package chatsessionsrepository

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMessageWriteRepositoryBlocksOnlyLiveApprovalExecutions(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("message_writes.go")
	if err != nil {
		t.Fatalf("read message write repository: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
	if !strings.Contains(source, "status in ('ready', 'executing')") || !strings.Contains(source, "status = 'retry_ready'") {
		t.Fatal("ordinary and regeneration writes must not cross a live mutation execution")
	}
	if strings.Contains(source, "status in ('ready', 'retry_ready', 'executing', 'failed_uncertain')") {
		t.Fatal("an uncertain execution must quarantine its fingerprint without permanently blocking unrelated chat turns")
	}
}

func TestApprovalReservationsPrepareSafeRetriesBeforeDurableRecovery(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("message_writes.go")
	if err != nil {
		t.Fatalf("read message write repository: %v", err)
	}
	source := strings.Join(strings.Fields(string(data)), " ")
	prepareIndex := strings.Index(source, "current, err = chatsessions.PrepareMutationApprovalRetries")
	recoverIndex := strings.Index(source, "current, err = chatsessions.RecoverDurableApprovalReceipts")
	approvalBranchIndex := strings.Index(source, "if params.Operation != chatsessions.MessageWriteApproval")
	reconcileIndex := strings.Index(source, "chatsessions.ReconcileCompletedApprovalReservation")
	if prepareIndex < 0 || recoverIndex < 0 || approvalBranchIndex < 0 || reconcileIndex < 0 {
		t.Fatal("message write reservation is missing safe retry preparation, durable recovery, or reconciliation")
	}
	if prepareIndex > recoverIndex || recoverIndex > approvalBranchIndex || recoverIndex > reconcileIndex {
		t.Fatal("safe retry intent must be prepared before durable receipts overwrite it and before approval reconciliation")
	}
}

func TestSafeRetryPreparationIsExactOriginAuditedAndBounded(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("message_writes.go")
	if err != nil {
		t.Fatalf("read message write repository: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")

	for _, contract := range []string{
		"execution.session_id = $1",
		"execution.user_id = $2",
		"execution.workspace_id = $3",
		"execution.tool_call_id = $4",
		"execution.fingerprint = $5",
		"execution.status = 'failed_uncertain'",
		"execution.reconciliation_count = 0",
		"set status = 'retry_ready'",
		"last_reconciliation_resolution = $6",
		"reconciliation_count = reconciliation_count + 1",
		"session.deleted_at is null",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("safe retry preparation is missing contract %q", contract)
		}
	}
}

func TestLegacyWholeArrayWritesCannotReplaceExistingTranscript(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read chat session commands: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
	if strings.Contains(source, "on conflict (session_id) do update") {
		t.Fatal("legacy chat writes must not replace an existing transcript")
	}
	if count := strings.Count(source, "on conflict (session_id) do nothing"); count < 2 {
		t.Fatalf("found %d create-only transcript writes, want at least 2", count)
	}
	if !strings.Contains(source, "return chatsessions.errmessagewriteconflict") {
		t.Fatal("legacy SaveMessages must report a conflict for an existing transcript")
	}
}

func TestDurableApprovalReceiptTerminalizesLeaseExpiryAndBusinessFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		stored     dbDurableApprovalReceipt
		wantFound  bool
		wantHalt   bool
		wantOutput any
	}{
		{
			name:      "live ready lease is not terminal",
			stored:    dbDurableApprovalReceipt{Status: "ready"},
			wantFound: false,
		},
		{
			name: "expired ready lease is terminal and definitely not run",
			stored: dbDurableApprovalReceipt{
				LeaseExpired: true,
				Status:       "ready",
			},
			wantFound: true,
			wantHalt:  true,
			wantOutput: map[string]any{
				"error":   expiredApprovalOutputMessage,
				"success": false,
			},
		},
		{
			name:      "uncertain execution is terminal",
			stored:    dbDurableApprovalReceipt{Status: "failed_uncertain"},
			wantFound: true,
			wantHalt:  true,
			wantOutput: map[string]any{
				"error":   uncertainApprovalOutputMessage,
				"success": false,
			},
		},
		{
			name:      "prepared safe retry is pending rather than terminal",
			stored:    dbDurableApprovalReceipt{Status: "retry_ready"},
			wantFound: false,
		},
		{
			name: "completed business failure halts ordered execution",
			stored: dbDurableApprovalReceipt{
				Output: json.RawMessage(`{"error":"team missing","success":false}`),
				Status: "completed",
			},
			wantFound: true,
			wantHalt:  true,
			wantOutput: map[string]any{
				"error":   "team missing",
				"success": false,
			},
		},
		{
			name: "completed success permits the next ordered mutation",
			stored: dbDurableApprovalReceipt{
				Output: json.RawMessage(`{"storyId":"story-1","success":true}`),
				Status: "completed",
			},
			wantFound: true,
			wantOutput: map[string]any{
				"storyId": "story-1",
				"success": true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receipt, found, err := toDurableApprovalReceipt(test.stored)
			if err != nil {
				t.Fatalf("to durable receipt: %v", err)
			}
			if found != test.wantFound || receipt.HaltsFollowing != test.wantHalt {
				t.Fatalf("found/halt = %v/%v, want %v/%v", found, receipt.HaltsFollowing, test.wantFound, test.wantHalt)
			}
			if !reflect.DeepEqual(receipt.Output, test.wantOutput) {
				t.Fatalf("output = %#v, want %#v", receipt.Output, test.wantOutput)
			}
		})
	}
}

func TestMergeCompletedToolOutputReplacesGenericTerminalReceipt(t *testing.T) {
	t.Parallel()

	messages := []any{map[string]any{
		"id":   "assistant-1",
		"role": "assistant",
		"parts": []any{map[string]any{
			"input":      map[string]any{"title": "Launch"},
			"output":     map[string]any{"error": "unconfirmed", "success": false},
			"state":      "output-available",
			"toolCallId": "call-1",
			"type":       "tool-createStory",
		}},
	}}
	durable := map[string]any{"storyId": "story-1", "success": true}

	applied, changed, err := mergeCompletedToolOutput(messages, "call-1", durable)
	if err != nil {
		t.Fatalf("merge durable output: %v", err)
	}
	if !applied || !changed {
		t.Fatalf("applied/changed = %v/%v, want true/true", applied, changed)
	}
	part := messages[0].(map[string]any)["parts"].([]any)[0].(map[string]any)
	if !reflect.DeepEqual(part["output"], durable) {
		t.Fatalf("output = %#v, want %#v", part["output"], durable)
	}
}
