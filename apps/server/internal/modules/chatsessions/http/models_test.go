package chatsessionshttp

import (
	"encoding/json"
	"strings"
	"testing"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
)

func TestMutationApprovalRequestValidation(t *testing.T) {
	t.Parallel()

	fingerprint := strings.Repeat("a", 64)
	leaseToken := uuid.New()
	if err := (AppClaimMutationApprovalRequest{Fingerprint: fingerprint}).Validate(); err != nil {
		t.Fatalf("valid claim request: %v", err)
	}
	if err := (AppCompleteMutationApprovalRequest{
		Fingerprint: fingerprint,
		LeaseToken:  leaseToken,
		Output:      json.RawMessage(`null`),
	}).Validate(); err != nil {
		t.Fatalf("valid completion request: %v", err)
	}
	if err := (AppStartMutationApprovalRequest{
		Fingerprint: fingerprint,
		LeaseToken:  leaseToken,
	}).Validate(); err != nil {
		t.Fatalf("valid start request: %v", err)
	}
	if err := (AppFailMutationApprovalRequest{
		FailureCode: chatsessions.MutationApprovalFailureCompletionUncertain,
		Fingerprint: fingerprint,
		LeaseToken:  leaseToken,
	}).Validate(); err != nil {
		t.Fatalf("valid failure request: %v", err)
	}
	if err := (AppReconcileMutationApprovalRequest{
		Evidence: chatsessions.MutationApprovalReconciliationEvidence{
			Kind:      chatsessions.MutationApprovalReconciliationEvidenceWorkspaceProbe,
			Reference: "story:missing:launch-checklist",
			Summary:   "The exact target is absent from the workspace.",
		},
		Fingerprint: fingerprint,
		Resolution:  chatsessions.MutationApprovalReconciliationVerifiedNotApplied,
	}).Validate(); err != nil {
		t.Fatalf("valid reconciliation request: %v", err)
	}
	if err := (AppReconcileMutationApprovalRequest{
		Evidence: chatsessions.MutationApprovalReconciliationEvidence{
			Kind:      chatsessions.MutationApprovalReconciliationEvidenceIdempotencyLookup,
			Reference: "story-create:chat-1:call-1",
			Summary:   "The idempotency receipt confirms completion.",
		},
		Fingerprint: fingerprint,
		Output:      json.RawMessage(`{"success":true}`),
		Resolution:  chatsessions.MutationApprovalReconciliationVerifiedCompleted,
	}).Validate(); err != nil {
		t.Fatalf("valid completed reconciliation request: %v", err)
	}

	for _, invalid := range []AppClaimMutationApprovalRequest{
		{},
		{Fingerprint: strings.Repeat("a", 63)},
		{Fingerprint: strings.Repeat("A", 64)},
		{Fingerprint: strings.Repeat("z", 64)},
	} {
		if err := invalid.Validate(); err == nil {
			t.Fatalf("invalid fingerprint accepted: %q", invalid.Fingerprint)
		}
	}
	if err := (AppCompleteMutationApprovalRequest{
		Fingerprint: fingerprint,
		LeaseToken:  leaseToken,
		Output:      json.RawMessage(`not-json`),
	}).Validate(); err == nil {
		t.Fatal("invalid completion output must be rejected")
	}
	if err := (AppStartMutationApprovalRequest{
		Fingerprint: fingerprint,
	}).Validate(); err == nil {
		t.Fatal("zero start lease token must be rejected")
	}
	if err := (AppFailMutationApprovalRequest{
		FailureCode: "retryable",
		Fingerprint: fingerprint,
		LeaseToken:  leaseToken,
	}).Validate(); err == nil {
		t.Fatal("unsupported failure code must be rejected")
	}
	if err := (AppReconcileMutationApprovalRequest{
		Evidence: chatsessions.MutationApprovalReconciliationEvidence{
			Kind:      chatsessions.MutationApprovalReconciliationEvidenceWorkspaceProbe,
			Reference: "story:missing:launch-checklist",
			Summary:   "The exact target is absent from the workspace.",
		},
		Fingerprint: fingerprint,
		Output:      json.RawMessage(`null`),
		Resolution:  chatsessions.MutationApprovalReconciliationVerifiedNotApplied,
	}).Validate(); err == nil {
		t.Fatal("verified-not-applied reconciliation with output must be rejected")
	}
	if err := (AppReconcileMutationApprovalRequest{
		Evidence:    chatsessions.MutationApprovalReconciliationEvidence{},
		Fingerprint: fingerprint,
		Resolution:  chatsessions.MutationApprovalReconciliationVerifiedNotApplied,
	}).Validate(); err == nil {
		t.Fatal("reconciliation without evidence must be rejected")
	}
}
