package agentinvocationsdomain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRunJSONDoesNotExposeExecutionMaterial(t *testing.T) {
	token := uuid.New()
	run := Run{ID: uuid.New(), SourceID: uuid.New(), RequesterID: uuid.New(), Status: StatusRunning,
		SourceHash: "private-source-hash", Payload: json.RawMessage(`{"secret":"private-payload"}`), LeaseToken: &token,
		Proposal: &Proposal{Action: json.RawMessage(`{"token":"signed-secret"}`)},
		Outcome:  &Outcome{Receipt: json.RawMessage(`{"private":"receipt"}`)},
	}
	encoded, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"sourceHash", "payload", "proposal", "outcome", "lease", token.String(), "signed-secret", "private-source-hash"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("run JSON exposes %q: %s", secret, encoded)
		}
	}
}

func TestProposalSeparatesClarificationFromAction(t *testing.T) {
	for _, test := range []struct {
		name     string
		proposal Proposal
		valid    bool
	}{
		{"question", Proposal{Reply: "The story is open."}, true},
		{"clarification", Proposal{Reply: "Which status?", NeedsInput: true}, true},
		{"null action", Proposal{Reply: "Ready", Action: json.RawMessage(" null ")}, true},
		{"approval", Proposal{Summary: "Close this story", Reply: "Should I close this story?", Review: "Status: Open → Closed", Action: json.RawMessage(`{"token":"opaque"}`), RequiresApproval: true}, true},
		{"approval without question", Proposal{Review: "Status: Open → Closed", Action: json.RawMessage(`{}`), RequiresApproval: true}, false},
		{"approval without review", Proposal{Summary: "Update this story", Action: json.RawMessage(`{}`), RequiresApproval: true}, false},
		{"oversized review", Proposal{Review: strings.Repeat("x", MaximumPayloadBytes+1), Reply: "Ready"}, false},
		{"empty", Proposal{}, false},
		{"clarification with mutation", Proposal{Reply: "Which status?", NeedsInput: true, Action: json.RawMessage(`{}`)}, false},
		{"approval without action", Proposal{Reply: "Ready", RequiresApproval: true}, false},
		{"invalid json", Proposal{Action: json.RawMessage(`{`)}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.proposal.Validate() == nil; got != test.valid {
				t.Fatalf("valid=%v, want %v", got, test.valid)
			}
		})
	}
}
