package feedback

import feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"

const (
	StatusPending                     = feedbackdomain.StatusPending
	StatusReviewing                   = feedbackdomain.StatusReviewing
	StatusPlanned                     = feedbackdomain.StatusPlanned
	StatusInProgress                  = feedbackdomain.StatusInProgress
	StatusCompleted                   = feedbackdomain.StatusCompleted
	StatusClosed                      = feedbackdomain.StatusClosed
	ListStatusTrashed                 = feedbackdomain.ListStatusTrashed
	RelationshipCreatedFrom           = feedbackdomain.RelationshipCreatedFrom
	RelationshipLinked                = feedbackdomain.RelationshipLinked
	RelationshipSolves                = feedbackdomain.RelationshipSolves
	SubmissionSourceInternal          = feedbackdomain.SubmissionSourceInternal
	SubmissionSourcePortal            = feedbackdomain.SubmissionSourcePortal
	SubmissionSourceWidget            = feedbackdomain.SubmissionSourceWidget
	SubmissionSourceIntegration       = feedbackdomain.SubmissionSourceIntegration
	EmailFrequencyOff                 = feedbackdomain.EmailFrequencyOff
	EmailFrequencyDaily               = feedbackdomain.EmailFrequencyDaily
	EmailFrequencyWeekly              = feedbackdomain.EmailFrequencyWeekly
	ParticipationModeAccountRequired  = feedbackdomain.ParticipationModeAccountRequired
	ParticipationModeAnonymousAllowed = feedbackdomain.ParticipationModeAnonymousAllowed
	ParticipationIntentAccount        = feedbackdomain.ParticipationIntentAccount
	ParticipationIntentAnonymous      = feedbackdomain.ParticipationIntentAnonymous
	ContributorKindAccount            = feedbackdomain.ContributorKindAccount
	ContributorKindAnonymous          = feedbackdomain.ContributorKindAnonymous
)

type CorePortal = feedbackdomain.CorePortal
type CoreContributorActivity = feedbackdomain.CoreContributorActivity
type CoreContributorActivityPage = feedbackdomain.CoreContributorActivityPage
type CoreListContributorActivityInput = feedbackdomain.CoreListContributorActivityInput
type CoreBoard = feedbackdomain.CoreBoard
type CoreItem = feedbackdomain.CoreItem
type CorePrivateAuthor = feedbackdomain.CorePrivateAuthor
type CoreMergeItemResult = feedbackdomain.CoreMergeItemResult
type CoreCanonicalItem = feedbackdomain.CoreCanonicalItem
type CoreSimilarItem = feedbackdomain.CoreSimilarItem
type CoreComment = feedbackdomain.CoreComment
type CoreContributorStats = feedbackdomain.CoreContributorStats
type CoreContributor = feedbackdomain.CoreContributor
type CoreContributorComment = feedbackdomain.CoreContributorComment
type CoreStoryLink = feedbackdomain.CoreStoryLink
type CoreStoryFeedbackLink = feedbackdomain.CoreStoryFeedbackLink
type CoreTeamSummary = feedbackdomain.CoreTeamSummary
type CoreBoardReviewer = feedbackdomain.CoreBoardReviewer
type CorePortalSnapshot = feedbackdomain.CorePortalSnapshot
type CorePortalInput = feedbackdomain.CorePortalInput
type CoreWorkspacePortalInput = feedbackdomain.CoreWorkspacePortalInput
type CoreBoardInput = feedbackdomain.CoreBoardInput
type CoreItemInput = feedbackdomain.CoreItemInput
type CoreBoardReviewerInput = feedbackdomain.CoreBoardReviewerInput
type CorePublicItemInput = feedbackdomain.CorePublicItemInput
type CorePublicItemResult = feedbackdomain.CorePublicItemResult
type CorePublicCommentInput = feedbackdomain.CorePublicCommentInput
type CorePublicVoteInput = feedbackdomain.CorePublicVoteInput
type CoreUpdateItemStatusInput = feedbackdomain.CoreUpdateItemStatusInput
type CoreCommentInput = feedbackdomain.CoreCommentInput
type CoreVoteResult = feedbackdomain.CoreVoteResult
type CoreStoryLinkInput = feedbackdomain.CoreStoryLinkInput
type CoreCreateStoryInput = feedbackdomain.CoreCreateStoryInput
type CoreCreateStoryResult = feedbackdomain.CoreCreateStoryResult
type CoreItemDetails = feedbackdomain.CoreItemDetails
type CoreListItemsInput = feedbackdomain.CoreListItemsInput
type CorePortalSnapshotInput = feedbackdomain.CorePortalSnapshotInput
type CoreItemsPage = feedbackdomain.CoreItemsPage
type CoreListContributorCommentsInput = feedbackdomain.CoreListContributorCommentsInput
type CoreContributorCommentsPage = feedbackdomain.CoreContributorCommentsPage
type StoryPlan = feedbackdomain.StoryPlan
type StoryDraft = feedbackdomain.StoryDraft
type StoryPlanner = feedbackdomain.StoryPlanner
type EventPublisher = feedbackdomain.EventPublisher
