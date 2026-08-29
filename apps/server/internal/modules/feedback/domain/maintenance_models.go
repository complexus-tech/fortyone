package feedbackdomain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MaintenanceStore owns retention work for feedback data. The scheduler only
// decides when to invoke these operations; cutoff rules and atomic deletion of
// related feedback records stay inside the feedback persistence adapter.
type MaintenanceStore interface {
	PurgeExpiredContributorArtifacts(context.Context, CoreContributorArtifactCutoffs) (CoreContributorArtifactPurgeResult, error)
	PurgeDeletedFeedback(context.Context, time.Time) (CoreDeletedFeedbackPurgeResult, error)
}

type CoreContributorArtifactCutoffs struct {
	RetainedBefore time.Time
	ExpiredBefore  time.Time
}

type CoreContributorArtifactPurgeResult struct {
	VerificationsDeleted   int64
	SessionsDeleted        int64
	UnsubscribeTokens      int64
	WidgetNoncesDeleted    int64
	SecretRotationsDeleted int64
}

func (result CoreContributorArtifactPurgeResult) TotalDeleted() int64 {
	return result.VerificationsDeleted +
		result.SessionsDeleted +
		result.UnsubscribeTokens +
		result.WidgetNoncesDeleted +
		result.SecretRotationsDeleted
}

type CoreDeletedFeedbackPurgeResult struct {
	ItemsDeleted        int64
	ContributorsDeleted int64
}

// DigestStore is the durable state-machine boundary for reviewer digests.
// Implementations must fence recipients through current workspace and team
// membership on every read and must complete cursor advancement and delivery
// state changes in one transaction.
type DigestStore interface {
	ListDigestRecipients(context.Context, CoreDigestRecipientCursor) ([]CoreDigestRecipient, error)
	ListDigestSubscriptions(context.Context, uuid.UUID, uuid.UUID) ([]CoreDigestSubscription, error)
	ClaimDigestDelivery(context.Context, CoreDigestDeliveryClaim) (uuid.UUID, bool, error)
	ListDigestItems(context.Context, CoreDigestItemsQuery) ([]CoreDigestItem, error)
	CompleteDigestDelivery(context.Context, CoreDigestDeliveryCompletion) error
	FailDigestDelivery(context.Context, uuid.UUID, string) error
}

type CoreDigestRecipientCursor struct {
	Limit            int32
	HasCursor        bool
	AfterWorkspaceID uuid.UUID
	AfterUserID      uuid.UUID
}

type CoreDigestRecipient struct {
	UserID        uuid.UUID
	UserEmail     string
	UserName      string
	Timezone      string
	WorkspaceID   uuid.UUID
	WorkspaceName string
	WorkspaceSlug string
}

type CoreDigestSubscription struct {
	BoardID            uuid.UUID
	TeamID             uuid.UUID
	EmailFrequency     string
	CreatedAt          time.Time
	LastDigestSentAt   *time.Time
	LastDigestCursorAt *time.Time
}

type CoreDigestDeliveryClaim struct {
	WorkspaceID uuid.UUID
	RecipientID uuid.UUID
	LocalDate   time.Time
	WindowStart time.Time
	WindowEnd   time.Time
	StaleBefore time.Time
}

type CoreDigestItemsQuery struct {
	RecipientID  uuid.UUID
	WorkspaceID  uuid.UUID
	BoardIDs     []uuid.UUID
	WindowStarts []time.Time
	WindowEnd    time.Time
	Limit        int32
}

type CoreDigestItem struct {
	ID                 uuid.UUID
	TeamID             uuid.UUID
	Title              string
	Description        string
	AuthorName         string
	TeamName           string
	Status             string
	CreatedAt          time.Time
	TotalCount         int32
	PendingReviewCount int32
}

type CoreDigestDeliveryStatus string

const (
	DigestDeliverySent    CoreDigestDeliveryStatus = "sent"
	DigestDeliverySkipped CoreDigestDeliveryStatus = "skipped"
)

type CoreDigestDeliveryCompletion struct {
	DeliveryID  uuid.UUID
	RecipientID uuid.UUID
	WorkspaceID uuid.UUID
	BoardIDs    []uuid.UUID
	DeliveredAt time.Time
	WindowEnd   time.Time
	Status      CoreDigestDeliveryStatus
	ItemCount   int32
}
