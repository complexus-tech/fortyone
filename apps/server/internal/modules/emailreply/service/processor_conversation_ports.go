package emailreply

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	MessageDirectionInbound = "inbound"

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"

	MessageKindReply    = "reply"
	MessageKindAnswer   = "answer"
	MessageKindProposal = "proposal"
	MessageKindReceipt  = "receipt"
	MessageKindError    = "error"
)

type ProposalStatus string

const (
	ProposalPending    ProposalStatus = "pending"
	ProposalConfirmed  ProposalStatus = "confirmed"
	ProposalApplying   ProposalStatus = "applying"
	ProposalApplied    ProposalStatus = "applied"
	ProposalFailed     ProposalStatus = "failed"
	ProposalCancelled  ProposalStatus = "cancelled"
	ProposalSuperseded ProposalStatus = "superseded"
)

var (
	ErrInvalidConversation = errors.New("invalid email conversation")
	ErrProposalConflict    = errors.New("email action proposal transition conflict")
	ErrProposalExpired     = errors.New("email action proposal expired")
)

// Thread is the emailreply-owned conversation projection. It intentionally
// contains only the fields needed to authorize, order, summarize, and deliver
// an inbound reply.
type Thread struct {
	ID                      uuid.UUID
	WorkspaceID             uuid.UUID
	UserID                  uuid.UUID
	RecipientEmail          string
	ExternalThreadID        string
	RootInternetMessageID   string
	LatestInternetMessageID string
	Context                 json.RawMessage
	Summary                 string
	SummaryThroughSequence  int64
	NextMessageSequence     int64
}

type ThreadKey struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type ThreadLease struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type MessageInput struct {
	ThreadID           uuid.UUID
	WorkspaceID        uuid.UUID
	UserID             uuid.UUID
	InboundEventID     *uuid.UUID
	IdempotencyKey     string
	Direction          string
	Role               string
	Kind               string
	ProviderMessageID  string
	InternetMessageID  string
	InReplyToMessageID string
	Subject            string
	Content            string
	Context            json.RawMessage
	ProviderMetadata   json.RawMessage
}

type Message struct {
	ID                uuid.UUID
	Sequence          int64
	Role              string
	Content           string
	InternetMessageID *string
	CreatedAt         time.Time
}

type MessagePageInput struct {
	ThreadID      uuid.UUID
	WorkspaceID   uuid.UUID
	UserID        uuid.UUID
	AfterSequence int64
	Limit         int
}

type MessagePage struct {
	Messages     []Message
	NextSequence int64
	HasMore      bool
}

type ThreadSummaryUpdate struct {
	ThreadID                       uuid.UUID
	WorkspaceID                    uuid.UUID
	UserID                         uuid.UUID
	ExpectedSummaryThroughSequence int64
	Summary                        string
	ThroughSequence                int64
}

type ProposalInput struct {
	ThreadID              uuid.UUID
	WorkspaceID           uuid.UUID
	UserID                uuid.UUID
	SourceMessageID       uuid.UUID
	IdempotencyKey        string
	ActionKind            string
	EntityType            string
	EntityID              uuid.UUID
	ExpectedEntityVersion string
	ProposedDiff          json.RawMessage
	ExpiresAt             time.Time
	Now                   time.Time
}

type Proposal struct {
	ID              uuid.UUID
	ThreadID        uuid.UUID
	WorkspaceID     uuid.UUID
	UserID          uuid.UUID
	SourceMessageID uuid.UUID
	ActionKind      string
	ProposedDiff    json.RawMessage
	Status          ProposalStatus
	ApplyAttempt    int
}

type ProposalDecision struct {
	ProposalID     uuid.UUID
	ThreadID       uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	ReplyTokenHash []byte
	Decision       ProposalStatus
	Now            time.Time
}

type ProposalListInput struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Now         time.Time
}

type ProposalControlLookup struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Control     ProposalStatus
}

type ProposalApplyClaim struct {
	ProposalID  uuid.UUID
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Now         time.Time
	RetryAfter  time.Duration
}

type ProposalApplyCompletion struct {
	ProposalID   uuid.UUID
	ThreadID     uuid.UUID
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	ApplyAttempt int
	Status       ProposalStatus
	Result       json.RawMessage
	ErrorMessage string
	Now          time.Time
}

type InboundEvent struct {
	ID                  uuid.UUID
	WorkspaceID         *uuid.UUID
	ExternalWorkspaceID string
	ExternalEventID     string
	Status              string
	PayloadEncrypted    *string
}

type OutboundDeliveryInput struct {
	Provider                string
	WorkspaceID             uuid.UUID
	UserID                  *uuid.UUID
	ExternalWorkspaceID     string
	ExternalRecipientUserID string
	InboundEventID          *uuid.UUID
	IdempotencyKey          string
	ExternalChannelID       string
	ExternalThreadID        string
	Purpose                 string
	ExpiresAt               *time.Time
}

type OutboundDelivery struct {
	ID              uuid.UUID
	IdempotencyKey  string
	ProviderPayload []byte
	ExpiresAt       *time.Time
}

// ProcessorStore is a caller-owned port. The messaging module is adapted at
// bootstrap so its repository and service records never leak into this module.
type ProcessorStore interface {
	AcquireThreadLease(context.Context, ThreadLease) (func() error, error)
	GetThread(context.Context, ThreadKey) (Thread, error)
	AppendMessage(context.Context, MessageInput) (Message, bool, error)
	ListMessages(context.Context, MessagePageInput) (MessagePage, error)
	UpdateThreadSummary(context.Context, ThreadSummaryUpdate) (Thread, error)
	RegisterProposal(context.Context, ProposalInput) (Proposal, bool, error)
	ListPendingProposals(context.Context, ProposalListInput) ([]Proposal, error)
	FindLatestProposalForControl(context.Context, ProposalControlLookup) (Proposal, bool, error)
	DecideProposal(context.Context, ProposalDecision) (Proposal, bool, error)
	ClaimProposalApply(context.Context, ProposalApplyClaim) (Proposal, bool, error)
	CompleteProposalApply(context.Context, ProposalApplyCompletion) (Proposal, bool, error)
	HasEarlierInboundEvent(context.Context, string, string, uuid.UUID) (bool, error)
	GetInboundEvent(context.Context, string, string, string) (InboundEvent, error)
	StartInboundEvent(context.Context, string, string, string) (InboundEvent, bool, error)
	CompleteInboundEvent(context.Context, uuid.UUID, string, string) error
	StartOutboundDelivery(context.Context, OutboundDeliveryInput) (OutboundDelivery, bool, error)
	SetOutboundDeliveryContentAndProviderPayload(context.Context, uuid.UUID, string, []byte) error
	CompleteOutboundDelivery(context.Context, uuid.UUID, string) error
	FailOutboundDelivery(context.Context, uuid.UUID, string) error
}

type ReplyPreparation struct {
	Thread            Thread
	ReplyToken        string
	InternetMessageID string
	InReplyTo         string
	Subject           string
	Content           string
	Kind              string
	IdempotencyKey    string
	Context           json.RawMessage
}

// ReplyThreadPort rotates reply aliases and records the outbound conversation
// message. The concrete emailthread service is adapted in bootstrap.
type ReplyThreadPort interface {
	NewReplyToken(context.Context, Thread) (string, error)
	PrepareReply(context.Context, ReplyPreparation) (string, error)
}
