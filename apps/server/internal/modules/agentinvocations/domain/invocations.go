// Package agentinvocationsdomain defines durable, surface-neutral assistant runs.
package agentinvocationsdomain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued           Status = "queued"
	StatusRunning          Status = "running"
	StatusNeedsInput       Status = "needs_input"
	StatusAwaitingApproval Status = "awaiting_approval"
	StatusSucceeded        Status = "succeeded"
	StatusFailed           Status = "failed"
	StatusCancelled        Status = "cancelled"
	MaximumPayloadBytes           = 64 << 10
	MaximumAttempts               = 3
)

var (
	ErrInvalidInput     = errors.New("invalid agent invocation")
	ErrNotFound         = errors.New("agent invocation not found")
	ErrConflict         = errors.New("agent invocation changed or cannot make this transition")
	ErrLeaseLost        = errors.New("agent invocation lease lost")
	ErrSourceChanged    = errors.New("agent invocation source changed")
	ErrPermissionDenied = errors.New("agent invocation permission denied")
	ErrUnavailable      = errors.New("agent invocation runtime unavailable")
)

type Input struct {
	WorkspaceID    uuid.UUID
	RequesterID    uuid.UUID
	TeamID         uuid.UUID
	Surface        string
	SourceID       uuid.UUID
	ConversationID uuid.UUID
	EntityID       uuid.UUID
	SourceRevision string
	SourceHash     string
	IdempotencyKey string
	ContinuationID *uuid.UUID
	// Payload is a bounded context snapshot, not an authority or action source.
	// Replays preserve it when display metadata changes; authoritative source
	// content must be bound by SourceRevision and SourceHash.
	Payload json.RawMessage
}

// Run exposes no model arguments, credentials, receipts, or fencing material in JSON.
// Transports should still map an explicit safe DTO and authorize its destination.
type Run struct {
	ID                  uuid.UUID       `json:"id"`
	WorkspaceID         uuid.UUID       `json:"-"`
	RequesterID         uuid.UUID       `json:"requesterId"`
	TeamID              uuid.UUID       `json:"-"`
	Surface             string          `json:"-"`
	SourceID            uuid.UUID       `json:"sourceId"`
	ConversationID      uuid.UUID       `json:"-"`
	EntityID            uuid.UUID       `json:"-"`
	ParentRunID         *uuid.UUID      `json:"-"`
	SourceRevision      string          `json:"-"`
	SourceHash          string          `json:"-"`
	Payload             json.RawMessage `json:"-"`
	Status              Status          `json:"status"`
	Summary             string          `json:"summary"`
	Version             int64           `json:"version"`
	Proposal            *Proposal       `json:"-"`
	Outcome             *Outcome        `json:"-"`
	ReplyCommentID      *uuid.UUID      `json:"replyCommentId"`
	ProposalReplyID     *uuid.UUID      `json:"-"`
	ProposalDeliveredAt *time.Time      `json:"-"`
	FailureReplyID      *uuid.UUID      `json:"-"`
	FailurePending      bool            `json:"-"`
	LeaseToken          *uuid.UUID      `json:"-"`
	LeaseGeneration     int64           `json:"-"`
	LeaseExpiresAt      *time.Time      `json:"-"`
	AttemptCount        int32           `json:"-"`
	RetryAttemptCount   int32           `json:"-"`
	ActionStartedAt     *time.Time      `json:"-"`
	ApprovedAt          *time.Time      `json:"-"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type Proposal struct {
	Summary          string          `json:"summary"`
	Review           string          `json:"review,omitempty"`
	Reply            string          `json:"reply"`
	Action           json.RawMessage `json:"action,omitempty"`
	RequiresApproval bool            `json:"requiresApproval"`
	NeedsInput       bool            `json:"needsInput"`
}

type Outcome struct {
	Reply   string          `json:"reply"`
	Receipt json.RawMessage `json:"receipt,omitempty"`
}

// HasDurableResponse separates bounded preparation retries from completion of
// already committed work. Transient publication/delivery outages must not strand
// an action receipt or an approval question after the model retry limit.
func (run Run) HasDurableResponse() bool {
	return run.Outcome != nil || run.ProposalReplyID != nil || run.ActionStartedAt != nil
}

type Control struct {
	WorkspaceID     uuid.UUID
	RequesterID     uuid.UUID
	RunID           uuid.UUID
	ExpectedVersion int64
}

// ContinuationQuery is derived by the surface adapter after source creation.
// SourceID identifies that new source, allowing replay to retain its original
// linkage even after the parent clarification has been completed.
type ContinuationQuery struct {
	WorkspaceID    uuid.UUID
	RequesterID    uuid.UUID
	Surface        string
	EntityID       uuid.UUID
	ConversationID uuid.UUID
	ReplyToID      uuid.UUID
	SourceID       uuid.UUID
}

// Failure supplies deliberately safe copy; arbitrary provider errors are never
// copied into the user-visible summary. Cause remains available to worker logs.
type Failure struct {
	Code    string
	Summary string
	Cause   error
}

func (e *Failure) Error() string { return fmt.Sprintf("agent invocation %s: %v", e.Code, e.Cause) }
func (e *Failure) Unwrap() error { return e.Cause }

func (input Input) Validate() error {
	if input.WorkspaceID == uuid.Nil || input.RequesterID == uuid.Nil || input.SourceID == uuid.Nil || input.ConversationID == uuid.Nil || input.EntityID == uuid.Nil {
		return ErrInvalidInput
	}
	if (input.Surface == "document_comment") != (input.TeamID == uuid.Nil) {
		return ErrInvalidInput
	}
	if input.ContinuationID != nil && *input.ContinuationID == uuid.Nil {
		return ErrInvalidInput
	}
	for _, field := range []struct {
		value string
		limit int
	}{{input.Surface, 64}, {input.IdempotencyKey, 256}, {input.SourceRevision, 256}, {input.SourceHash, 128}} {
		if strings.TrimSpace(field.value) == "" || len(field.value) > field.limit {
			return ErrInvalidInput
		}
	}
	if len(input.Payload) > MaximumPayloadBytes || (len(input.Payload) > 0 && !json.Valid(input.Payload)) {
		return ErrInvalidInput
	}
	return nil
}

func (proposal Proposal) HasAction() bool {
	return len(proposal.Action) > 0 && strings.TrimSpace(string(proposal.Action)) != "null"
}

func (proposal Proposal) Validate() error {
	if len(proposal.Summary) > 2000 || len(proposal.Review) > MaximumPayloadBytes || len(proposal.Reply) > MaximumPayloadBytes || len(proposal.Action) > MaximumPayloadBytes || (proposal.HasAction() && !json.Valid(proposal.Action)) {
		return ErrInvalidInput
	}
	if proposal.NeedsInput && (proposal.HasAction() || proposal.RequiresApproval) {
		return ErrInvalidInput
	}
	if proposal.RequiresApproval && (!proposal.HasAction() || strings.TrimSpace(proposal.Review) == "" || strings.TrimSpace(proposal.Reply) == "") {
		return ErrInvalidInput
	}
	if !proposal.HasAction() && strings.TrimSpace(proposal.Reply) == "" {
		return ErrInvalidInput
	}
	return nil
}

func (outcome Outcome) Validate() error {
	if strings.TrimSpace(outcome.Reply) == "" || len(outcome.Reply) > MaximumPayloadBytes || len(outcome.Receipt) > MaximumPayloadBytes || (len(outcome.Receipt) > 0 && !json.Valid(outcome.Receipt)) {
		return ErrInvalidInput
	}
	return nil
}
