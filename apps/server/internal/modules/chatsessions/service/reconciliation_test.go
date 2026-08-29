package chatsessions

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateMutationApprovalReconciliation(t *testing.T) {
	t.Parallel()

	evidence := MutationApprovalReconciliationEvidence{
		Kind:      MutationApprovalReconciliationEvidenceIdempotencyLookup,
		Reference: "story-create:chat-1:call-1",
		Summary:   "The idempotency lookup found the completed story receipt.",
	}
	for _, reconciliation := range []MutationApprovalReconciliation{
		{
			Resolution: MutationApprovalReconciliationVerifiedCompleted,
			Evidence:   evidence,
			Output:     json.RawMessage(`{"success":true}`),
		},
		{
			Resolution: MutationApprovalReconciliationVerifiedNotApplied,
			Evidence: MutationApprovalReconciliationEvidence{
				Kind:      MutationApprovalReconciliationEvidenceWorkspaceProbe,
				Reference: "story:missing:launch-checklist",
				Summary:   "The exact target was absent from the owner-scoped workspace query.",
			},
		},
	} {
		if err := validateMutationApprovalReconciliation(reconciliation); err != nil {
			t.Fatalf("valid reconciliation rejected: %v", err)
		}
	}
}

func TestValidateMutationApprovalReconciliationRejectsUnprovenReset(t *testing.T) {
	t.Parallel()

	base := MutationApprovalReconciliation{
		Resolution: MutationApprovalReconciliationVerifiedNotApplied,
		Evidence: MutationApprovalReconciliationEvidence{
			Kind:      MutationApprovalReconciliationEvidenceWorkspaceProbe,
			Reference: "story:missing:launch-checklist",
			Summary:   "The exact target was absent.",
		},
	}

	tests := []MutationApprovalReconciliation{
		{
			Resolution: MutationApprovalReconciliationVerifiedNotApplied,
			Evidence:   MutationApprovalReconciliationEvidence{},
		},
		{
			Resolution: MutationApprovalReconciliationVerifiedNotApplied,
			Evidence:   base.Evidence,
			Output:     json.RawMessage(`null`),
		},
		{
			Resolution: MutationApprovalReconciliationVerifiedCompleted,
			Evidence:   base.Evidence,
		},
		{
			Resolution: MutationApprovalReconciliationVerifiedNotApplied,
			Evidence: MutationApprovalReconciliationEvidence{
				Kind:      MutationApprovalReconciliationEvidenceWorkspaceProbe,
				Reference: strings.Repeat("x", 256),
				Summary:   "Absent.",
			},
		},
	}

	for _, reconciliation := range tests {
		if err := validateMutationApprovalReconciliation(reconciliation); err == nil {
			t.Fatalf("invalid reconciliation accepted: %#v", reconciliation)
		}
	}
}
