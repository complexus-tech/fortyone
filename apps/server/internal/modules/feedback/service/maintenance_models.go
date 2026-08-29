package feedback

import feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"

const (
	DigestDeliverySent    = feedbackdomain.DigestDeliverySent
	DigestDeliverySkipped = feedbackdomain.DigestDeliverySkipped
)

type MaintenanceStore = feedbackdomain.MaintenanceStore
type CoreContributorArtifactCutoffs = feedbackdomain.CoreContributorArtifactCutoffs
type CoreContributorArtifactPurgeResult = feedbackdomain.CoreContributorArtifactPurgeResult
type CoreDeletedFeedbackPurgeResult = feedbackdomain.CoreDeletedFeedbackPurgeResult
type DigestStore = feedbackdomain.DigestStore
type CoreDigestRecipientCursor = feedbackdomain.CoreDigestRecipientCursor
type CoreDigestRecipient = feedbackdomain.CoreDigestRecipient
type CoreDigestSubscription = feedbackdomain.CoreDigestSubscription
type CoreDigestDeliveryClaim = feedbackdomain.CoreDigestDeliveryClaim
type CoreDigestItemsQuery = feedbackdomain.CoreDigestItemsQuery
type CoreDigestItem = feedbackdomain.CoreDigestItem
type CoreDigestDeliveryStatus = feedbackdomain.CoreDigestDeliveryStatus
type CoreDigestDeliveryCompletion = feedbackdomain.CoreDigestDeliveryCompletion
