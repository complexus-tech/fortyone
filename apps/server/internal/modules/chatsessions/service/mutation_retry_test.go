package chatsessions

import (
	"errors"
	"reflect"
	"testing"
)

func TestApprovedMutationFingerprintMatchesTypeScriptFixedVectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		toolName    string
		input       any
		fingerprint string
	}{
		{
			name:     "create applies schema defaults and strict-null normalization",
			toolName: "createStory",
			input: map[string]any{
				"autoSchedulingEnabled":    nil,
				"description":              "Line 1\nLine 2",
				"estimatedDurationMinutes": 90,
				"ignoredByTheSchema":       "never fingerprinted",
				"labelIds":                 []any{"β", "alpha"},
				"minimumFocusBlockMinutes": 30,
				"teamId":                   "team-1",
				"title":                    "Launch <Maya> — 東京\u2028done",
			},
			fingerprint: "ac769a74f5f27cb788cee0d73e9a597016083624026e73f8489069af3779e0d5",
		},
		{
			name:     "bulk create preserves nested arrays nulls numbers and unicode",
			toolName: "bulkCreateStories",
			input: map[string]any{
				"sharedValues": map[string]any{
					"autoSchedulingEnabled": nil,
					"estimateValue":         5,
					"ignoredByTheSchema":    true,
					"priority":              nil,
					"statusId":              nil,
					"teamId":                "team-1",
				},
				"storiesData": []any{
					map[string]any{
						"autoSchedulingEnabled":    nil,
						"description":              nil,
						"estimatedDurationMinutes": 30,
						"ignoredByTheSchema":       "never fingerprinted",
						"labelIds":                 []any{"β", "alpha"},
						"priority":                 nil,
						"teamId":                   nil,
						"title":                    "One",
					},
					map[string]any{
						"descriptionHTML":          "<p>Crème</p>",
						"minimumFocusBlockMinutes": 15,
						"startDate":                "2026-09-01",
						"title":                    "Two",
					},
				},
			},
			fingerprint: "c7515047c7421d07df71bfb1f23407baa0526a8eb422a3a3783f7b675e3a4dd2",
		},
		{
			name:     "delete trims the title and forces confirmation",
			toolName: "deleteStory",
			input: map[string]any{
				"confirmed":  false,
				"storyId":    "11111111-1111-4111-8111-111111111111",
				"storyTitle": "\uFEFF  Maya\u2028launch  ",
			},
			fingerprint: "1310c2d915346d60cf4c8a92f2293244889bd90e70875e62462412f596bf30d9",
		},
		{
			name:     "bulk delete trims every title and forces confirmation",
			toolName: "bulkDeleteStories",
			input: map[string]any{
				"confirmed": false,
				"storyIds": []any{
					"11111111-1111-4111-8111-111111111111",
					"22222222-2222-4222-8222-222222222222",
				},
				"storyTitles": []any{" One ", " Deux — 東京\uFEFF"},
			},
			fingerprint: "62548ee08353e2e712b9f3a170a6613849cef7b3031339d8485320cd32cfe2c1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fingerprint, err := approvedMutationFingerprint(test.toolName, test.input)
			if err != nil {
				t.Fatalf("approved mutation fingerprint: %v", err)
			}
			if fingerprint != test.fingerprint {
				t.Fatalf("fingerprint = %q, want %q", fingerprint, test.fingerprint)
			}
		})
	}
}

func TestPrepareMutationApprovalRetriesReopensOnlyVerifiedSafeCalls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		toolName string
		input    map[string]any
	}{
		{name: "create story", toolName: "createStory", input: map[string]any{"teamId": "team-1", "title": "One"}},
		{name: "bulk create stories", toolName: "bulkCreateStories", input: map[string]any{"storiesData": []any{map[string]any{"title": "One"}}}},
		{name: "delete story", toolName: "deleteStory", input: map[string]any{"storyId": "11111111-1111-4111-8111-111111111111", "storyTitle": "One"}},
		{name: "bulk delete stories", toolName: "bulkDeleteStories", input: map[string]any{"storyIds": []any{"11111111-1111-4111-8111-111111111111"}, "storyTitles": []any{"One"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			current, incoming := retryApprovalTranscripts(test.toolName, test.input)
			original := mustCloneRetryMessages(t, current)
			calls := 0
			reopened, err := PrepareMutationApprovalRetries(current, incoming, func(intent MutationApprovalRetryIntent) (bool, error) {
				calls++
				if intent.ToolCallID != "call-1" || intent.ToolName != test.toolName || len(intent.Fingerprint) != 64 {
					t.Fatalf("unexpected retry intent: %#v", intent)
				}
				return true, nil
			})
			if err != nil {
				t.Fatalf("prepare retry: %v", err)
			}
			if calls != 1 {
				t.Fatalf("prepare calls = %d, want 1", calls)
			}
			part := retryPartAt(t, reopened, 0)
			if part["state"] != "approval-responded" {
				t.Fatalf("state = %#v, want approval-responded", part["state"])
			}
			if _, exists := part["output"]; exists {
				t.Fatal("reopened retry must not retain uncertainty output")
			}
			if !reflect.DeepEqual(current, original) {
				t.Fatal("retry preparation mutated the caller transcript")
			}
		})
	}
}

func TestPrepareMutationApprovalRetriesRejectsUntrustedOrUnsafeEvidence(t *testing.T) {
	t.Parallel()

	baseInput := map[string]any{"teamId": "team-1", "title": "One"}
	tests := []struct {
		name   string
		mutate func(current, incoming []any)
	}{
		{
			name: "unsupported tool",
			mutate: func(current, incoming []any) {
				retryPartAt(t, current, 0)["type"] = "tool-updateStory"
				retryPartAt(t, incoming, 0)["type"] = "tool-updateStory"
			},
		},
		{
			name: "hard delete",
			mutate: func(current, incoming []any) {
				for _, transcript := range [][]any{current, incoming} {
					part := retryPartAt(t, transcript, 0)
					part["type"] = "tool-bulkDeleteStories"
					part["input"] = map[string]any{
						"hardDelete":  true,
						"storyIds":    []any{"11111111-1111-4111-8111-111111111111"},
						"storyTitles": []any{"One"},
					}
				}
			},
		},
		{
			name: "caller changes input",
			mutate: func(_ []any, incoming []any) {
				retryPartAt(t, incoming, 0)["input"] = map[string]any{"teamId": "team-1", "title": "Two"}
			},
		},
		{
			name: "caller changes tool call",
			mutate: func(_ []any, incoming []any) {
				retryPartAt(t, incoming, 0)["toolCallId"] = "call-2"
			},
		},
		{
			name: "caller changes approval id",
			mutate: func(_ []any, incoming []any) {
				retryPartAt(t, incoming, 0)["approval"] = map[string]any{"approved": true, "id": "approval-2"}
			},
		},
		{
			name: "caller denies approval",
			mutate: func(_ []any, incoming []any) {
				retryPartAt(t, incoming, 0)["approval"] = map[string]any{"approved": false, "id": "approval-1"}
			},
		},
		{
			name: "persisted approval was not granted",
			mutate: func(current, _ []any) {
				retryPartAt(t, current, 0)["approval"] = map[string]any{"approved": false, "id": "approval-1"}
			},
		},
		{
			name: "noncanonical uncertainty output",
			mutate: func(current, _ []any) {
				retryPartAt(t, current, 0)["output"] = map[string]any{"error": "caller supplied", "success": false}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			current, incoming := retryApprovalTranscripts("createStory", baseInput)
			test.mutate(current, incoming)
			original := mustCloneRetryMessages(t, current)
			calls := 0
			reopened, err := PrepareMutationApprovalRetries(current, incoming, func(MutationApprovalRetryIntent) (bool, error) {
				calls++
				return true, nil
			})
			if err != nil {
				t.Fatalf("prepare retry: %v", err)
			}
			if calls != 0 {
				t.Fatalf("prepare calls = %d, want 0", calls)
			}
			if !reflect.DeepEqual(reopened, original) {
				t.Fatalf("unsafe evidence changed transcript: %#v", reopened)
			}
		})
	}
}

func TestPrepareMutationApprovalRetriesLeavesSkippedFollowingApprovalsTerminal(t *testing.T) {
	t.Parallel()

	current, incoming := retryApprovalTranscripts("createStory", map[string]any{"teamId": "team-1", "title": "One"})
	currentParts := current[0].(map[string]any)["parts"].([]any)
	incomingParts := incoming[0].(map[string]any)["parts"].([]any)
	skipped := retryApprovalPart("createStory", "call-2", map[string]any{"teamId": "team-1", "title": "Two"}, "output-available")
	skipped["approval"] = map[string]any{"approved": true, "id": "approval-2"}
	skipped["output"] = map[string]any{"error": MutationApprovalSkippedOutputMessage, "success": false}
	current[0].(map[string]any)["parts"] = append(currentParts, skipped)
	incomingSecond := retryApprovalPart("createStory", "call-2", map[string]any{"teamId": "team-1", "title": "Two"}, "approval-responded")
	incomingSecond["approval"] = map[string]any{"approved": true, "id": "approval-2"}
	incoming[0].(map[string]any)["parts"] = append(incomingParts, incomingSecond)

	calls := 0
	reopened, err := PrepareMutationApprovalRetries(current, incoming, func(MutationApprovalRetryIntent) (bool, error) {
		calls++
		return true, nil
	})
	if err != nil {
		t.Fatalf("prepare retry: %v", err)
	}
	if calls != 1 {
		t.Fatalf("prepare calls = %d, want only the ledger-backed uncertainty", calls)
	}
	parts := reopened[0].(map[string]any)["parts"].([]any)
	if retryPartAt(t, reopened, 0)["state"] != "approval-responded" {
		t.Fatal("the exact uncertain approval was not reopened")
	}
	second := parts[1].(map[string]any)
	if second["state"] != "output-available" || !reflect.DeepEqual(second["output"], skipped["output"]) {
		t.Fatal("a skipped approval without a ledger row must require a fresh proposal")
	}
}

func TestPrepareMutationApprovalRetriesPropagatesLedgerFailure(t *testing.T) {
	t.Parallel()

	current, incoming := retryApprovalTranscripts("createStory", map[string]any{"teamId": "team-1", "title": "One"})
	wantErr := errors.New("database unavailable")
	_, err := PrepareMutationApprovalRetries(current, incoming, func(MutationApprovalRetryIntent) (bool, error) {
		return false, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func retryApprovalTranscripts(toolName string, input map[string]any) ([]any, []any) {
	currentPart := retryApprovalPart(toolName, "call-1", input, "output-available")
	currentPart["output"] = map[string]any{
		"error":   MutationApprovalUncertainOutputMessage,
		"success": false,
	}
	incomingPart := retryApprovalPart(toolName, "call-1", input, "approval-responded")
	return []any{map[string]any{"id": "assistant-1", "parts": []any{currentPart}, "role": "assistant"}},
		[]any{map[string]any{"id": "assistant-1", "parts": []any{incomingPart}, "role": "assistant"}}
}

func retryApprovalPart(toolName, toolCallID string, input map[string]any, state string) map[string]any {
	return map[string]any{
		"approval":   map[string]any{"approved": true, "id": "approval-1"},
		"input":      input,
		"state":      state,
		"toolCallId": toolCallID,
		"type":       "tool-" + toolName,
	}
}

func retryPartAt(t *testing.T, transcript []any, index int) map[string]any {
	t.Helper()
	return transcript[0].(map[string]any)["parts"].([]any)[index].(map[string]any)
}

func mustCloneRetryMessages(t *testing.T, messages []any) []any {
	t.Helper()
	cloned, err := cloneJSONValue(messages)
	if err != nil {
		t.Fatalf("clone messages: %v", err)
	}
	return cloned.([]any)
}
