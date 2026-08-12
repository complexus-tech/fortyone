package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	EmailMessageDirectionInbound  = "inbound"
	EmailMessageDirectionOutbound = "outbound"

	EmailMessageRoleUser      = "user"
	EmailMessageRoleAssistant = "assistant"
	EmailMessageRoleSystem    = "system"

	EmailMessageKindGuidance     = "guidance"
	EmailMessageKindReply        = "reply"
	EmailMessageKindAnswer       = "answer"
	EmailMessageKindProposal     = "proposal"
	EmailMessageKindConfirmation = "confirmation"
	EmailMessageKindReceipt      = "receipt"
	EmailMessageKindError        = "error"
)

type EmailActionProposalStatus string

const (
	EmailActionProposalPending    EmailActionProposalStatus = "pending"
	EmailActionProposalConfirmed  EmailActionProposalStatus = "confirmed"
	EmailActionProposalApplying   EmailActionProposalStatus = "applying"
	EmailActionProposalApplied    EmailActionProposalStatus = "applied"
	EmailActionProposalFailed     EmailActionProposalStatus = "failed"
	EmailActionProposalCancelled  EmailActionProposalStatus = "cancelled"
	EmailActionProposalExpired    EmailActionProposalStatus = "expired"
	EmailActionProposalSuperseded EmailActionProposalStatus = "superseded"
)

var (
	ErrInvalidEmailConversation = errors.New("invalid email conversation")
	ErrInvalidEmailReplyToken   = errors.New("invalid email reply token")
	ErrEmailSummaryConflict     = errors.New("email conversation summary changed")
	ErrInvalidEmailProposal     = errors.New("invalid email action proposal")
	ErrEmailProposalConflict    = errors.New("email action proposal transition conflict")
	ErrEmailProposalExpired     = errors.New("email action proposal expired")
	ErrEmailProposalApplyBusy   = errors.New("email action proposal apply is already in progress")
)

// EmailConversationStore persists complete email history and arbitrates the
// lifecycle of reply-triggered action proposals. Raw reply tokens are never
// passed to this interface; callers hash them before persistence or lookup.
type EmailConversationStore interface {
	AcquireEmailThreadProcessingLease(context.Context, EmailThreadProcessingLease) (func() error, error)
	CreateEmailThread(context.Context, EmailThreadInput) (EmailThreadRecord, bool, error)
	CreateEmailReplyTokenAlias(context.Context, EmailReplyTokenInput) (EmailReplyTokenRecord, bool, error)
	FindEmailThreadByReplyToken(context.Context, EmailReplyTokenLookup) (EmailThreadLookup, error)
	GetEmailThread(context.Context, EmailThreadKey) (EmailThreadRecord, error)
	AppendEmailMessage(context.Context, EmailMessageInput) (EmailMessageRecord, bool, error)
	ListEmailMessages(context.Context, EmailMessagePageInput) (EmailMessagePage, error)
	UpdateEmailThreadSummary(context.Context, EmailThreadSummaryUpdate) (EmailThreadRecord, error)
	RegisterEmailActionProposal(context.Context, EmailActionProposalInput) (EmailActionProposalRecord, bool, error)
	ListPendingEmailActionProposals(context.Context, EmailActionProposalListInput) ([]EmailActionProposalRecord, error)
	FindLatestEmailActionProposalForControl(context.Context, EmailActionProposalControlLookup) (EmailActionProposalRecord, bool, error)
	GetEmailActionProposal(context.Context, EmailActionProposalKey) (EmailActionProposalRecord, error)
	DecideEmailActionProposal(context.Context, EmailActionProposalDecision) (EmailActionProposalRecord, bool, error)
	ClaimEmailActionProposalApply(context.Context, EmailActionProposalApplyClaim) (EmailActionProposalRecord, bool, error)
	CompleteEmailActionProposalApply(context.Context, EmailActionProposalApplyCompletion) (EmailActionProposalRecord, bool, error)
}

type EmailThreadProcessingLease struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type EmailThreadInput struct {
	Provider              string
	WorkspaceID           uuid.UUID
	UserID                uuid.UUID
	RecipientEmail        string
	ExternalThreadID      string
	RootInternetMessageID string
	Context               json.RawMessage
	ReplyTokenHash        []byte
	ReplyTokenExpiresAt   time.Time
}

type EmailThreadRecord struct {
	ID                      uuid.UUID       `db:"id"`
	WorkspaceID             uuid.UUID       `db:"workspace_id"`
	UserID                  uuid.UUID       `db:"user_id"`
	Provider                string          `db:"provider"`
	RecipientEmail          string          `db:"recipient_email"`
	ExternalThreadID        string          `db:"external_thread_id"`
	RootInternetMessageID   string          `db:"root_internet_message_id"`
	LatestInternetMessageID string          `db:"latest_internet_message_id"`
	Context                 json.RawMessage `db:"context"`
	Summary                 string          `db:"summary"`
	SummaryThroughSequence  int64           `db:"summary_through_sequence"`
	NextMessageSequence     int64           `db:"next_message_sequence"`
	CreatedAt               time.Time       `db:"created_at"`
	UpdatedAt               time.Time       `db:"updated_at"`
}

type EmailThreadKey struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type EmailReplyTokenInput struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	TokenHash   []byte
	ExpiresAt   time.Time
}

type EmailReplyTokenRecord struct {
	ID        uuid.UUID  `db:"id"`
	ThreadID  uuid.UUID  `db:"thread_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type EmailReplyTokenLookup struct {
	Provider  string
	TokenHash []byte
	Now       time.Time
}

type EmailThreadLookup struct {
	Thread              EmailThreadRecord
	ReplyTokenID        uuid.UUID
	ReplyTokenExpiresAt time.Time
}

type EmailMessageInput struct {
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

type EmailMessageRecord struct {
	ID                 uuid.UUID       `db:"id"`
	ThreadID           uuid.UUID       `db:"thread_id"`
	Sequence           int64           `db:"sequence"`
	InboundEventID     *uuid.UUID      `db:"inbound_event_id"`
	IdempotencyKey     string          `db:"idempotency_key"`
	Direction          string          `db:"direction"`
	Role               string          `db:"role"`
	Kind               string          `db:"kind"`
	ProviderMessageID  *string         `db:"provider_message_id"`
	InternetMessageID  *string         `db:"internet_message_id"`
	InReplyToMessageID *string         `db:"in_reply_to_message_id"`
	Subject            string          `db:"subject"`
	Content            string          `db:"content"`
	Context            json.RawMessage `db:"context"`
	ProviderMetadata   json.RawMessage `db:"provider_metadata"`
	CreatedAt          time.Time       `db:"created_at"`
}

type EmailMessagePageInput struct {
	ThreadID      uuid.UUID
	WorkspaceID   uuid.UUID
	UserID        uuid.UUID
	AfterSequence int64
	Limit         int
}

type EmailMessagePage struct {
	Messages     []EmailMessageRecord
	NextSequence int64
	HasMore      bool
}

type EmailThreadSummaryUpdate struct {
	ThreadID                       uuid.UUID
	WorkspaceID                    uuid.UUID
	UserID                         uuid.UUID
	ExpectedSummaryThroughSequence int64
	Summary                        string
	ThroughSequence                int64
}

type EmailActionProposalInput struct {
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

type EmailActionProposalRecord struct {
	ID                    uuid.UUID                 `db:"id"`
	ThreadID              uuid.UUID                 `db:"thread_id"`
	WorkspaceID           uuid.UUID                 `db:"workspace_id"`
	UserID                uuid.UUID                 `db:"user_id"`
	SourceMessageID       uuid.UUID                 `db:"source_message_id"`
	IdempotencyKey        string                    `db:"idempotency_key"`
	ActionKind            string                    `db:"action_kind"`
	EntityType            string                    `db:"entity_type"`
	EntityID              uuid.UUID                 `db:"entity_id"`
	ExpectedEntityVersion string                    `db:"expected_entity_version"`
	ProposedDiff          json.RawMessage           `db:"proposed_diff"`
	Status                EmailActionProposalStatus `db:"status"`
	ApplyAttempt          int                       `db:"apply_attempt"`
	Result                json.RawMessage           `db:"result"`
	LastError             *string                   `db:"last_error"`
	ExpiresAt             time.Time                 `db:"expires_at"`
	ConfirmedAt           *time.Time                `db:"confirmed_at"`
	ApplyingAt            *time.Time                `db:"applying_at"`
	AppliedAt             *time.Time                `db:"applied_at"`
	FailedAt              *time.Time                `db:"failed_at"`
	CancelledAt           *time.Time                `db:"cancelled_at"`
	ExpiredAt             *time.Time                `db:"expired_at"`
	SupersededAt          *time.Time                `db:"superseded_at"`
	CreatedAt             time.Time                 `db:"created_at"`
	UpdatedAt             time.Time                 `db:"updated_at"`
}

type EmailActionProposalDecision struct {
	ProposalID     uuid.UUID
	ThreadID       uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	ReplyTokenHash []byte
	Decision       EmailActionProposalStatus
	Now            time.Time
}

type EmailActionProposalListInput struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Now         time.Time
}

// EmailActionProposalControlLookup recovers the durable proposal selected by
// an exact CONFIRM or CANCEL after that proposal has already left pending
// state. This is strictly a retry primitive; model turns still receive only
// pending proposals.
type EmailActionProposalControlLookup struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Control     EmailActionProposalStatus
}

type EmailActionProposalKey struct {
	ProposalID  uuid.UUID
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type EmailActionProposalApplyClaim struct {
	ProposalID  uuid.UUID
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Now         time.Time
	RetryAfter  time.Duration
}

type EmailActionProposalApplyCompletion struct {
	ProposalID   uuid.UUID
	ThreadID     uuid.UUID
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	ApplyAttempt int
	Status       EmailActionProposalStatus
	Result       json.RawMessage
	ErrorMessage string
	Now          time.Time
}
