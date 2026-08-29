package messagingdomain

import (
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
	ID                      uuid.UUID
	WorkspaceID             uuid.UUID
	UserID                  uuid.UUID
	Provider                string
	RecipientEmail          string
	ExternalThreadID        string
	RootInternetMessageID   string
	LatestInternetMessageID string
	Context                 json.RawMessage
	Summary                 string
	SummaryThroughSequence  int64
	NextMessageSequence     int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
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
	ID        uuid.UUID
	ThreadID  uuid.UUID
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
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
	ID                 uuid.UUID
	ThreadID           uuid.UUID
	Sequence           int64
	InboundEventID     *uuid.UUID
	IdempotencyKey     string
	Direction          string
	Role               string
	Kind               string
	ProviderMessageID  *string
	InternetMessageID  *string
	InReplyToMessageID *string
	Subject            string
	Content            string
	Context            json.RawMessage
	ProviderMetadata   json.RawMessage
	CreatedAt          time.Time
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
	ID                    uuid.UUID
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
	Status                EmailActionProposalStatus
	ApplyAttempt          int
	Result                json.RawMessage
	LastError             *string
	ExpiresAt             time.Time
	ConfirmedAt           *time.Time
	ApplyingAt            *time.Time
	AppliedAt             *time.Time
	FailedAt              *time.Time
	CancelledAt           *time.Time
	ExpiredAt             *time.Time
	SupersededAt          *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
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
