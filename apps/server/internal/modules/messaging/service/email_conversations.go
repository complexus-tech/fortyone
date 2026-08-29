package messaging

import (
	"context"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
)

const (
	EmailMessageDirectionInbound  = messagingdomain.EmailMessageDirectionInbound
	EmailMessageDirectionOutbound = messagingdomain.EmailMessageDirectionOutbound

	EmailMessageRoleUser      = messagingdomain.EmailMessageRoleUser
	EmailMessageRoleAssistant = messagingdomain.EmailMessageRoleAssistant
	EmailMessageRoleSystem    = messagingdomain.EmailMessageRoleSystem

	EmailMessageKindGuidance     = messagingdomain.EmailMessageKindGuidance
	EmailMessageKindReply        = messagingdomain.EmailMessageKindReply
	EmailMessageKindAnswer       = messagingdomain.EmailMessageKindAnswer
	EmailMessageKindProposal     = messagingdomain.EmailMessageKindProposal
	EmailMessageKindConfirmation = messagingdomain.EmailMessageKindConfirmation
	EmailMessageKindReceipt      = messagingdomain.EmailMessageKindReceipt
	EmailMessageKindError        = messagingdomain.EmailMessageKindError
)

type EmailActionProposalStatus = messagingdomain.EmailActionProposalStatus

const (
	EmailActionProposalPending    = messagingdomain.EmailActionProposalPending
	EmailActionProposalConfirmed  = messagingdomain.EmailActionProposalConfirmed
	EmailActionProposalApplying   = messagingdomain.EmailActionProposalApplying
	EmailActionProposalApplied    = messagingdomain.EmailActionProposalApplied
	EmailActionProposalFailed     = messagingdomain.EmailActionProposalFailed
	EmailActionProposalCancelled  = messagingdomain.EmailActionProposalCancelled
	EmailActionProposalExpired    = messagingdomain.EmailActionProposalExpired
	EmailActionProposalSuperseded = messagingdomain.EmailActionProposalSuperseded
)

var (
	ErrInvalidEmailConversation = messagingdomain.ErrInvalidEmailConversation
	ErrInvalidEmailReplyToken   = messagingdomain.ErrInvalidEmailReplyToken
	ErrEmailSummaryConflict     = messagingdomain.ErrEmailSummaryConflict
	ErrInvalidEmailProposal     = messagingdomain.ErrInvalidEmailProposal
	ErrEmailProposalConflict    = messagingdomain.ErrEmailProposalConflict
	ErrEmailProposalExpired     = messagingdomain.ErrEmailProposalExpired
	ErrEmailProposalApplyBusy   = messagingdomain.ErrEmailProposalApplyBusy
)

type EmailThreadProcessingLease = messagingdomain.EmailThreadProcessingLease
type EmailThreadInput = messagingdomain.EmailThreadInput
type EmailThreadRecord = messagingdomain.EmailThreadRecord
type EmailThreadKey = messagingdomain.EmailThreadKey
type EmailReplyTokenInput = messagingdomain.EmailReplyTokenInput
type EmailReplyTokenRecord = messagingdomain.EmailReplyTokenRecord
type EmailReplyTokenLookup = messagingdomain.EmailReplyTokenLookup
type EmailThreadLookup = messagingdomain.EmailThreadLookup
type EmailMessageInput = messagingdomain.EmailMessageInput
type EmailMessageRecord = messagingdomain.EmailMessageRecord
type EmailMessagePageInput = messagingdomain.EmailMessagePageInput
type EmailMessagePage = messagingdomain.EmailMessagePage
type EmailThreadSummaryUpdate = messagingdomain.EmailThreadSummaryUpdate
type EmailActionProposalInput = messagingdomain.EmailActionProposalInput
type EmailActionProposalRecord = messagingdomain.EmailActionProposalRecord
type EmailActionProposalDecision = messagingdomain.EmailActionProposalDecision
type EmailActionProposalListInput = messagingdomain.EmailActionProposalListInput
type EmailActionProposalControlLookup = messagingdomain.EmailActionProposalControlLookup
type EmailActionProposalKey = messagingdomain.EmailActionProposalKey
type EmailActionProposalApplyClaim = messagingdomain.EmailActionProposalApplyClaim
type EmailActionProposalApplyCompletion = messagingdomain.EmailActionProposalApplyCompletion

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
