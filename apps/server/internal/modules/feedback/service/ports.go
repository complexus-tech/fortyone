package feedback

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository composes small capability ports. Services can use the complete
// adapter while focused workers and tests depend only on the capability they
// actually need.
type Repository interface {
	PortalStore
	BoardStore
	ItemStore
	CommentStore
	StoryLinkStore
	StatusStore
}

type PortalStore interface {
	GetPortalBySlug(context.Context, string) (CorePortal, error)
	GetPortalByWorkspaceSlugAndSlug(context.Context, string, string) (CorePortal, error)
	GetPortal(context.Context, uuid.UUID, uuid.UUID) (CorePortal, error)
	ListPortals(context.Context, uuid.UUID) ([]CorePortal, error)
	CreatePortal(context.Context, CorePortalInput) (CorePortal, error)
	UpdatePortal(context.Context, uuid.UUID, uuid.UUID, CorePortalInput) (CorePortal, error)
}

type BoardStore interface {
	ListBoards(context.Context, uuid.UUID) ([]CoreBoard, error)
	GetBoard(context.Context, uuid.UUID, uuid.UUID) (CoreBoard, error)
	CreateBoard(context.Context, CoreBoardInput) (CoreBoard, error)
	DeleteBoard(context.Context, uuid.UUID, uuid.UUID) error
	ListBoardReviewers(context.Context, uuid.UUID, uuid.UUID) ([]CoreBoardReviewer, error)
	SetBoardReviewer(context.Context, CoreBoardReviewerInput) (CoreBoardReviewer, error)
}

type ItemStore interface {
	ListContributorActivity(context.Context, CoreListContributorActivityInput) (CoreContributorActivityPage, error)
	ListItems(context.Context, CoreListItemsInput) (CoreItemsPage, error)
	GetContributor(context.Context, uuid.UUID, uuid.UUID) (CoreContributor, error)
	ContributorExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	ListContributorComments(context.Context, CoreListContributorCommentsInput) (CoreContributorCommentsPage, error)
	GetItem(context.Context, uuid.UUID, uuid.UUID) (CoreItem, error)
	GetPrivateAuthor(context.Context, uuid.UUID, uuid.UUID) (CorePrivateAuthor, error)
	GetItemReadAt(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (*time.Time, error)
	ListTeamSummaries(context.Context, uuid.UUID, uuid.UUID) ([]CoreTeamSummary, error)
	MarkItemRead(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (time.Time, error)
	MarkItemUnread(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	GetItemByPortal(context.Context, uuid.UUID, uuid.UUID) (CoreItem, error)
	ResolveCanonicalItem(context.Context, uuid.UUID, string) (CoreCanonicalItem, error)
	ListSimilarItems(context.Context, uuid.UUID, string, string, int) ([]CoreSimilarItem, error)
	CreateItem(context.Context, CoreItemInput) (CoreItem, error)
	GetOrCreateAccountContributor(context.Context, uuid.UUID, uuid.UUID) (CoreContributor, error)
	CreateAnonymousItem(context.Context, CoreItemInput) (CoreItem, error)
	LinkItemAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreItemAttachment, error)
	GetItemAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreItemAttachment, error)
	ListItemAttachments(context.Context, uuid.UUID, []uuid.UUID) ([]CoreItemAttachment, error)
	UpdateItemStatus(context.Context, uuid.UUID, uuid.UUID, CoreUpdateItemStatusInput) (CoreItem, bool, error)
	UpdateItemStatusIfUnchanged(context.Context, uuid.UUID, uuid.UUID, time.Time, CoreUpdateItemStatusInput) (CoreItem, bool, bool, error)
	TrashItem(context.Context, uuid.UUID, uuid.UUID) error
	RestoreItem(context.Context, uuid.UUID, uuid.UUID) error
	ToggleVote(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int) (CoreVoteResult, error)
}

type CommentStore interface {
	ListComments(context.Context, uuid.UUID, []uuid.UUID) ([]CoreComment, error)
	ListItemComments(context.Context, uuid.UUID, uuid.UUID) ([]CoreComment, error)
	GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreComment, error)
	CreateComment(context.Context, CoreCommentInput) (CoreComment, error)
}

type StoryLinkStore interface {
	ListStoryLinks(context.Context, uuid.UUID, []uuid.UUID) ([]CoreStoryLink, error)
	ListItemStoryLinks(context.Context, uuid.UUID, uuid.UUID) ([]CoreStoryLink, error)
	ListStoryFeedbackLinks(context.Context, uuid.UUID, uuid.UUID) ([]CoreStoryFeedbackLink, error)
	LinkStory(context.Context, CoreStoryLinkInput) (CoreStoryLink, error)
}

type StatusStore interface {
	FindFirstStatusByCategory(context.Context, uuid.UUID, string) (*uuid.UUID, error)
	GetStatusCategory(context.Context, uuid.UUID, uuid.UUID) (string, error)
}

type NextPhaseRepository interface {
	ContributorIdentityStore
	ContributorEngagementStore
	MergeStore
	DeliveryIntentStore
	UpdateStore
	WidgetStore
}

type ContributorIdentityStore interface {
	CreateContributorVerification(context.Context, CoreVerificationRequest) (CoreVerificationChallenge, error)
	ConfirmContributorVerification(context.Context, CoreVerificationConfirmation) (CoreParticipant, CoreParticipantSession, error)
	GetContributorSession(context.Context, uuid.UUID, []byte, string) (CoreParticipant, CoreParticipantSession, error)
	RevokeContributorSession(context.Context, uuid.UUID, []byte) error
	GetParticipantByUser(context.Context, uuid.UUID, uuid.UUID) (CoreParticipant, error)
	ConsumeUnsubscribeToken(context.Context, uuid.UUID, []byte, []byte, time.Time) (CoreParticipant, CoreParticipantSession, error)
	CreateExternalContributorSession(context.Context, uuid.UUID, string, string, string, *string, []byte, time.Time) (CoreParticipant, CoreParticipantSession, error)
}

type ContributorEngagementStore interface {
	CreateContributorItemAndFollow(context.Context, CoreContributorItemInput) (CoreItem, error)
	CreateContributorComment(context.Context, CoreContributorCommentInput) (CoreComment, error)
	ToggleContributorVote(context.Context, CoreContributorVoteInput) (CoreVoteResult, error)
	GetItemFollow(context.Context, uuid.UUID, uuid.UUID) (CoreFollowState, error)
	SetItemFollow(context.Context, uuid.UUID, uuid.UUID, bool) (CoreFollowState, error)
	GetContributorPreferences(context.Context, uuid.UUID, uuid.UUID) (CoreContributorPreferences, error)
	SetPortalEmailPreference(context.Context, uuid.UUID, uuid.UUID, bool) (CoreContributorPreferences, error)
	GetUnreadUpdateCount(context.Context, uuid.UUID, uuid.UUID) (int, error)
	MarkUpdatesSeen(context.Context, uuid.UUID, uuid.UUID) (time.Time, error)
	ListDeliveryRecipients(context.Context, uuid.UUID, *uuid.UUID, *uuid.UUID, uuid.UUID) ([]CoreDeliveryRecipient, error)
	ListAccountItemFollowers(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
	ListAccountUpdateRecipients(context.Context, uuid.UUID, uuid.UUID) ([]CoreAccountUpdateRecipient, error)
	ListPrimaryStoryItems(context.Context, uuid.UUID, uuid.UUID) ([]CoreItem, error)
}

type MergeStore interface {
	ListItemCandidates(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, int) (CoreMergeCandidatesPage, error)
	MergeItems(context.Context, CoreMergeItemInput) (CoreMergeItemResult, error)
	ClaimMergeOutboxEvents(context.Context, int, time.Duration) ([]CoreMergeOutboxEvent, error)
	ListMergeRecipients(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) ([]CoreMergeRecipient, error)
	CompleteMergeOutboxEvent(context.Context, uuid.UUID, uuid.UUID) error
	RetryMergeOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string, time.Time, bool) error
}

type DeliveryIntentStore interface {
	CreateContributorDelivery(context.Context, CoreCreateDeliveryInput) (CoreDelivery, bool, error)
}

type UpdateStore interface {
	ListWorkspaceUpdates(context.Context, uuid.UUID, int, int) (CoreUpdatesPage, error)
	GetWorkspaceUpdate(context.Context, uuid.UUID, uuid.UUID) (CoreFeedbackUpdate, error)
	CreateUpdate(context.Context, CoreUpdateInput) (CoreFeedbackUpdate, error)
	UpdateUpdate(context.Context, CoreUpdateInput) (CoreFeedbackUpdate, error)
	DeleteUpdate(context.Context, uuid.UUID, uuid.UUID) error
	PublishUpdate(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreFeedbackUpdate, bool, error)
	UnpublishUpdate(context.Context, uuid.UUID, uuid.UUID) (CoreFeedbackUpdate, error)
	ListPublicUpdates(context.Context, uuid.UUID, int, int) (CoreUpdatesPage, error)
	GetPublicUpdate(context.Context, uuid.UUID, string) (CoreFeedbackUpdate, error)
}

type WidgetStore interface {
	GetWidgetSettings(context.Context, uuid.UUID, uuid.UUID) (CoreWidgetSettings, error)
	GetPublicWidgetSettings(context.Context, uuid.UUID) (CoreWidgetSettings, error)
	UpsertWidgetSettings(context.Context, CoreWidgetSettingsInput) (CoreWidgetSettings, error)
	SetInitialWidgetSecret(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, int) (CoreWidgetSettings, error)
	RotateWidgetSecret(context.Context, uuid.UUID, uuid.UUID, string, int, time.Time) (CoreWidgetSettings, error)
	GetWidgetSigningSecret(context.Context, uuid.UUID, uuid.UUID, int) (string, error)
	ConsumeWidgetAssertionNonce(context.Context, uuid.UUID, uuid.UUID, int, string, string, time.Time) error
}

// ScopedNextPhaseStore is the authenticated persistence contract used by the
// production adapter. Keeping it separate from the legacy-shaped test port
// makes the actor and credential-team boundary explicit without letting HTTP
// or repository packages reach into authentication context internals.
type ScopedNextPhaseStore interface {
	ListWorkspaceUpdatesScoped(context.Context, CoreAccessScope, int, int) (CoreUpdatesPage, error)
	GetWorkspaceUpdateScoped(context.Context, CoreAccessScope, uuid.UUID) (CoreFeedbackUpdate, error)
	DeleteUpdateScoped(context.Context, CoreAccessScope, uuid.UUID) error
	PublishUpdateScoped(context.Context, CoreAccessScope, uuid.UUID) (CoreFeedbackUpdate, bool, error)
	UnpublishUpdateScoped(context.Context, CoreAccessScope, uuid.UUID) (CoreFeedbackUpdate, error)
	ListItemCandidatesScoped(context.Context, CoreAccessScope, uuid.UUID, uuid.UUID, string, int) (CoreMergeCandidatesPage, error)
	MergeItemsScoped(context.Context, CoreAccessScope, CoreMergeItemInput) (CoreMergeItemResult, error)
	GetWidgetSettingsScoped(context.Context, CoreAccessScope, uuid.UUID) (CoreWidgetSettings, error)
	UpsertWidgetSettingsScoped(context.Context, CoreAccessScope, CoreWidgetSettingsInput) (CoreWidgetSettings, error)
	SetInitialWidgetSecretScoped(context.Context, CoreAccessScope, uuid.UUID, uuid.UUID, string, int) (CoreWidgetSettings, error)
	RotateWidgetSecretScoped(context.Context, CoreAccessScope, uuid.UUID, string, int, time.Time) (CoreWidgetSettings, error)
}

// ScopedCoreStore is the authenticated item/community persistence boundary.
// Public portal operations intentionally remain on their contributor-aware
// ports; internal operations always receive the actor and credential team
// fence explicitly from the service.
type ScopedCoreStore interface {
	ListItemsScoped(context.Context, CoreAccessScope, CoreListItemsInput) (CoreItemsPage, error)
	GetItemScoped(context.Context, CoreAccessScope, uuid.UUID) (CoreItem, error)
	GetPrivateAuthorScoped(context.Context, CoreAccessScope, uuid.UUID) (CorePrivateAuthor, error)
	GetItemReadAtScoped(context.Context, CoreAccessScope, uuid.UUID) (*time.Time, error)
	ListTeamSummariesScoped(context.Context, CoreAccessScope) ([]CoreTeamSummary, error)
	MarkItemReadScoped(context.Context, CoreAccessScope, uuid.UUID) (time.Time, error)
	MarkItemUnreadScoped(context.Context, CoreAccessScope, uuid.UUID) error
	ListItemCommentsScoped(context.Context, CoreAccessScope, uuid.UUID) ([]CoreComment, error)
	GetCommentScoped(context.Context, CoreAccessScope, uuid.UUID, uuid.UUID) (CoreComment, error)
	CreateCommentScoped(context.Context, CoreAccessScope, CoreCommentInput) (CoreComment, error)
	ListItemStoryLinksScoped(context.Context, CoreAccessScope, uuid.UUID) ([]CoreStoryLink, error)
	ListStoryFeedbackLinksScoped(context.Context, CoreAccessScope, uuid.UUID) ([]CoreStoryFeedbackLink, error)
	LinkStoryScoped(context.Context, CoreAccessScope, CoreStoryLinkInput) (CoreStoryLink, error)
	TrashItemScoped(context.Context, CoreAccessScope, uuid.UUID) error
	RestoreItemScoped(context.Context, CoreAccessScope, uuid.UUID) error
	ToggleVoteScoped(context.Context, CoreAccessScope, uuid.UUID, int) (CoreVoteResult, error)
}
