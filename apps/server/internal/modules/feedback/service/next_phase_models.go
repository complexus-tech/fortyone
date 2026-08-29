package feedback

import feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"

const (
	ParticipationModeVerifiedGuest        = feedbackdomain.ParticipationModeVerifiedGuest
	ParticipationIntentVerifiedGuest      = feedbackdomain.ParticipationIntentVerifiedGuest
	ParticipationIntentExternal           = feedbackdomain.ParticipationIntentExternal
	ContributorKindVerifiedGuest          = feedbackdomain.ContributorKindVerifiedGuest
	ContributorKindExternal               = feedbackdomain.ContributorKindExternal
	GuestIdentityPolicyShowIdentity       = feedbackdomain.GuestIdentityPolicyShowIdentity
	GuestIdentityPolicyAllowPublicMasking = feedbackdomain.GuestIdentityPolicyAllowPublicMasking
	GuestIdentityPolicyAlwaysMaskGuests   = feedbackdomain.GuestIdentityPolicyAlwaysMaskGuests
	ContributorSessionSourcePortal        = feedbackdomain.ContributorSessionSourcePortal
	ContributorSessionSourceWidget        = feedbackdomain.ContributorSessionSourceWidget
	ContributorSessionSourcePreferences   = feedbackdomain.ContributorSessionSourcePreferences
	FeedbackUpdateStatusDraft             = feedbackdomain.FeedbackUpdateStatusDraft
	FeedbackUpdateStatusPublished         = feedbackdomain.FeedbackUpdateStatusPublished
)

type CoreParticipant = feedbackdomain.CoreParticipant
type CoreParticipantSession = feedbackdomain.CoreParticipantSession
type CoreVerificationRequest = feedbackdomain.CoreVerificationRequest
type CoreVerificationChallenge = feedbackdomain.CoreVerificationChallenge
type CoreVerificationConfirmation = feedbackdomain.CoreVerificationConfirmation
type CoreContributorSessionResult = feedbackdomain.CoreContributorSessionResult
type CoreUnsubscribeTokenInput = feedbackdomain.CoreUnsubscribeTokenInput
type CoreResolvedParticipant = feedbackdomain.CoreResolvedParticipant
type CoreContributorItemInput = feedbackdomain.CoreContributorItemInput
type CoreContributorCommentInput = feedbackdomain.CoreContributorCommentInput
type CoreContributorVoteInput = feedbackdomain.CoreContributorVoteInput
type CoreFollowState = feedbackdomain.CoreFollowState
type CoreContributorPreferences = feedbackdomain.CoreContributorPreferences
type CoreDeliveryRecipient = feedbackdomain.CoreDeliveryRecipient
type CoreAccountUpdateRecipient = feedbackdomain.CoreAccountUpdateRecipient
type CoreMergeItemInput = feedbackdomain.CoreMergeItemInput
type CoreMergeOutboxEvent = feedbackdomain.CoreMergeOutboxEvent
type CoreMergeRecipient = feedbackdomain.CoreMergeRecipient
type CoreMergeCandidate = feedbackdomain.CoreMergeCandidate
type CoreMergeCandidatesPage = feedbackdomain.CoreMergeCandidatesPage
type CoreCreateDeliveryInput = feedbackdomain.CoreCreateDeliveryInput
type CoreDelivery = feedbackdomain.CoreDelivery
type CoreFeedbackUpdate = feedbackdomain.CoreFeedbackUpdate
type CoreUpdateItem = feedbackdomain.CoreUpdateItem
type CoreUpdateInput = feedbackdomain.CoreUpdateInput
type CoreUpdatesPage = feedbackdomain.CoreUpdatesPage
type CoreWidgetSettings = feedbackdomain.CoreWidgetSettings
type CoreWidgetSettingsInput = feedbackdomain.CoreWidgetSettingsInput
type CoreWidgetSecretResult = feedbackdomain.CoreWidgetSecretResult
type CoreWidgetIdentityAssertion = feedbackdomain.CoreWidgetIdentityAssertion
type CoreWidgetSessionInput = feedbackdomain.CoreWidgetSessionInput
