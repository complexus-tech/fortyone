package chatsessions

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MutationApprovalExecutionState string

type MessageWriteOperation string

const (
	MutationApprovalExecutionClaimed    MutationApprovalExecutionState = "claimed"
	MutationApprovalExecutionStarted    MutationApprovalExecutionState = "started"
	MutationApprovalExecutionReady      MutationApprovalExecutionState = "ready"
	MutationApprovalExecutionExecuting  MutationApprovalExecutionState = "executing"
	MutationApprovalExecutionInProgress MutationApprovalExecutionState = "in_progress"
	MutationApprovalExecutionCompleted  MutationApprovalExecutionState = "completed"
	MutationApprovalExecutionFailed     MutationApprovalExecutionState = "failed_uncertain"
)

const (
	MessageWriteAppend     MessageWriteOperation = "append"
	MessageWriteRegenerate MessageWriteOperation = "regenerate"
	MessageWriteApproval   MessageWriteOperation = "approval"
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

type CoreChatSession struct {
	ID          string     `json:"id"`
	UserID      uuid.UUID  `json:"userId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type CoreNewChatSession struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Title       string    `json:"title"`
	Messages    []any     `json:"messages"`
}

type BeginMessageWriteParams struct {
	Session         CoreChatSession
	Messages        []any
	Operation       MessageWriteOperation
	TargetMessageID string
}

type CoreMessageWriteReservation struct {
	Generation int64     `json:"generation"`
	Token      uuid.UUID `json:"token"`
	// Messages is populated only when the server repaired request history.
	// It is a client-safe projection and never rehydrates omitted attachments.
	Messages []any `json:"messages,omitempty"`
}

type FinalizeMessageWriteParams struct {
	SessionID   string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Messages    []any
	Generation  int64
	Token       uuid.UUID
}

type CoreMessageWriteResult struct {
	Applied bool `json:"applied"`
}

type RecoverMutationApprovalOutputParams struct {
	SessionID   string
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	ToolCallID  string
	Fingerprint string
}

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
