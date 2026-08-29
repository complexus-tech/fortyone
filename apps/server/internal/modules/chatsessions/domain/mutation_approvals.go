// Package chatsessionsdomain owns transport-neutral chat-session invariants and
// persistence contracts shared by the service and repository adapters.
package chatsessionsdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound                 = errors.New("chat session not found")
	ErrMutationApprovalConflict = errors.New("mutation approval does not match the prepared tool call")
	ErrMutationApprovalLease    = errors.New("mutation approval execution lease is invalid or expired")
)

type MutationApprovalExecutionState string

const (
	MutationApprovalExecutionClaimed    MutationApprovalExecutionState = "claimed"
	MutationApprovalExecutionStarted    MutationApprovalExecutionState = "started"
	MutationApprovalExecutionReady      MutationApprovalExecutionState = "ready"
	MutationApprovalExecutionExecuting  MutationApprovalExecutionState = "executing"
	MutationApprovalExecutionInProgress MutationApprovalExecutionState = "in_progress"
	MutationApprovalExecutionCompleted  MutationApprovalExecutionState = "completed"
	MutationApprovalExecutionFailed     MutationApprovalExecutionState = "failed_uncertain"
)

type MutationApprovalReconciliationResolution string

const (
	MutationApprovalReconciliationVerifiedCompleted  MutationApprovalReconciliationResolution = "verified_completed"
	MutationApprovalReconciliationVerifiedNotApplied MutationApprovalReconciliationResolution = "verified_not_applied"
)

type MutationApprovalReconciliationEvidenceKind string

const (
	MutationApprovalReconciliationEvidenceIdempotencyLookup MutationApprovalReconciliationEvidenceKind = "idempotency_lookup"
	MutationApprovalReconciliationEvidenceWorkspaceProbe    MutationApprovalReconciliationEvidenceKind = "workspace_state_probe"
)

type CoreMutationApprovalExecution struct {
	State          MutationApprovalExecutionState `json:"state"`
	Output         json.RawMessage                `json:"output,omitempty"`
	LeaseToken     *uuid.UUID                     `json:"leaseToken,omitempty"`
	LeaseExpiresAt *time.Time                     `json:"leaseExpiresAt,omitempty"`
	FailureCode    string                         `json:"failureCode,omitempty"`
}

type MutationApprovalExecutionParams struct {
	SessionID   string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	ToolCallID  string
	Fingerprint string
	LeaseToken  uuid.UUID
}

type MutationApprovalReconciliationEvidence struct {
	Kind      MutationApprovalReconciliationEvidenceKind `json:"kind"`
	Reference string                                     `json:"reference"`
	Summary   string                                     `json:"summary"`
}

type MutationApprovalReconciliation struct {
	Resolution MutationApprovalReconciliationResolution
	Evidence   MutationApprovalReconciliationEvidence
	Output     json.RawMessage
}
