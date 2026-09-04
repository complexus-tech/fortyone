package feedbackdomain

import (
	"context"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

// RoadmapSummaryText returns the public explanation for a status transition,
// falling back to a non-sensitive status description when no explanation was
// supplied.
func (item CoreItem) RoadmapSummaryText() string {
	if item.RoadmapSummary != nil && strings.TrimSpace(*item.RoadmapSummary) != "" {
		return strings.TrimSpace(*item.RoadmapSummary)
	}
	return "The status of this feedback changed to " + strings.ReplaceAll(item.Status, "_", " ") + "."
}

const (
	StatusPending     = "pending"
	StatusReviewing   = "reviewing"
	StatusPlanned     = "planned"
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusClosed      = "closed"
	ListStatusTrashed = "trashed"

	RelationshipCreatedFrom = "created_from"
	RelationshipLinked      = "linked"
	RelationshipSolves      = "solves"

	SubmissionSourceInternal    = "internal"
	SubmissionSourcePortal      = "portal"
	SubmissionSourceWidget      = "widget"
	SubmissionSourceIntegration = "integration"

	EmailFrequencyOff    = "off"
	EmailFrequencyDaily  = "daily"
	EmailFrequencyWeekly = "weekly"

	ParticipationModeAccountRequired  = "account_required"
	ParticipationModeAnonymousAllowed = "anonymous_allowed"
	ParticipationIntentAccount        = "account"
	ParticipationIntentAnonymous      = "anonymous"

	ContributorKindAccount   = "account"
	ContributorKindAnonymous = "anonymous"
)

type CorePortal struct {
	ID                  uuid.UUID
	WorkspaceID         uuid.UUID
	Name                string
	Slug                string
	IsPublic            bool
	ParticipationMode   string
	GuestIdentityPolicy string
	HasPublishedUpdates bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CoreContributorActivity struct {
	ID            uuid.UUID
	Type          string
	FeedbackID    uuid.UUID
	FeedbackTitle string
	FeedbackSlug  string
	Body          string
	BoardName     string
	Status        string
	VoteCount     int
	CommentCount  int
	PortalSlug    string
	WorkspaceName string
	WorkspaceSlug string
	CreatedAt     time.Time
}

type CoreContributorActivityPage struct {
	Activities    []CoreContributorActivity
	Page          int
	PageSize      int
	HasMore       bool
	FeedbackCount int
	CommentCount  int
	VoteScore     int
	PortalCount   int
}

type CoreListContributorActivityInput struct {
	UserID       uuid.UUID
	ActivityType string
	Page         int
	PageSize     int
}

type CoreBoard struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	PortalID    uuid.UUID
	TeamID      uuid.UUID
	Name        string
	Slug        string
	Color       string
	OrderIndex  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CoreItem struct {
	ID               uuid.UUID
	WorkspaceID      uuid.UUID
	PortalID         uuid.UUID
	BoardID          uuid.UUID
	ContributorID    uuid.UUID
	AuthorID         uuid.UUID
	AuthorName       string
	AuthorEmail      string
	AuthorAvatar     *string
	ParticipantKind  string
	AuthorMasked     bool
	MergedIntoItemID *uuid.UUID
	MergedAt         *time.Time
	MergedByUserID   *uuid.UUID
	Following        bool
	Title            string
	Description      string
	DescriptionHTML  string
	Slug             string
	Status           string
	VoteCount        int
	UpvoteCount      int
	DownvoteCount    int
	CommentCount     int
	RoadmapSummary   *string
	Board            CoreBoard
	StoryLinks       []CoreStoryLink
	ReadAt           *time.Time
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CoreItemAttachment struct {
	ID          uuid.UUID
	ItemID      uuid.UUID
	WorkspaceID uuid.UUID
	Filename    string
	Size        int64
	MimeType    string
	CreatedAt   time.Time
}

// CorePrivateAuthor is kept separate from CoreItem so contact data cannot
// accidentally flow through a public item serializer. Only the
// administrator-only private-author endpoint returns this model.
type CorePrivateAuthor struct {
	ContributorID uuid.UUID
	UserID        *uuid.UUID
	Kind          string
	DisplayName   string
	Email         *string
	AvatarURL     *string
	PublicMasked  bool
}

type CoreMergeItemResult struct {
	SourceItemID         uuid.UUID
	TargetItemID         uuid.UUID
	PortalID             uuid.UUID
	MergedAt             time.Time
	MergedByUserID       uuid.UUID
	MovedFollowerCount   int
	MovedUpdateLinkCount int
	MovedStoryLinkCount  int
	Target               CoreItem
}

type CoreCanonicalItem struct {
	ItemID   uuid.UUID
	ItemSlug string
	Merged   bool
}

type CoreSimilarItem struct {
	ID           uuid.UUID
	Slug         string
	Title        string
	AuthorID     *uuid.UUID
	AuthorName   string
	AuthorAvatar *string
	Status       string
	VoteCount    int
	CommentCount int
	Confidence   float64
	IsDuplicate  bool
}

type CoreComment struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	ItemID          uuid.UUID
	AuthorID        uuid.UUID
	ContributorID   uuid.UUID
	ParentID        *uuid.UUID
	AuthorName      string
	AuthorAvatar    *string
	ParticipantKind string
	AuthorMasked    bool
	Body            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CoreContributorStats struct {
	FeedbackCount int
	CommentCount  int
	VoteScore     int
}

type CoreContributor struct {
	ID        uuid.UUID
	PortalID  uuid.UUID
	UserID    uuid.UUID
	Kind      string
	Name      string
	AvatarURL *string
	JoinedAt  time.Time
	Stats     CoreContributorStats
}

type CoreContributorComment struct {
	ID            uuid.UUID
	ItemID        uuid.UUID
	FeedbackTitle string
	FeedbackSlug  string
	Body          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CoreStoryLink struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	ItemID          uuid.UUID
	StoryID         uuid.UUID
	StoryTitle      string
	Relationship    string
	IsPrimary       bool
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
}

type CoreStoryFeedbackLink struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	ItemID        uuid.UUID
	StoryID       uuid.UUID
	TeamID        uuid.UUID
	FeedbackTitle string
	Relationship  string
	IsPrimary     bool
	CreatedAt     time.Time
}

type CoreTeamSummary struct {
	TeamID      uuid.UUID
	Enabled     bool
	TotalCount  int
	UnreadCount int
}

type CoreBoardReviewer struct {
	UserID         uuid.UUID
	Name           string
	Email          string
	AvatarURL      *string
	Role           string
	EmailFrequency string
}

type CorePortalSnapshot struct {
	Portal       CorePortal
	Boards       []CoreBoard
	Items        []CoreItem
	ItemsHasMore bool
	Comments     []CoreComment
	Links        []CoreStoryLink
	Attachments  []CoreItemAttachment
}

type CorePortalInput struct {
	WorkspaceID         uuid.UUID
	IsPublic            *bool
	ParticipationMode   *string
	GuestIdentityPolicy *string
}

type CoreWorkspacePortalInput struct {
	WorkspaceID   uuid.UUID
	WorkspaceName string
	WorkspaceSlug string
}

type CoreBoardInput struct {
	WorkspaceID uuid.UUID
	PortalID    uuid.UUID
	TeamID      uuid.UUID
	CreatorID   uuid.UUID
	Name        string
	Slug        string
	Color       string
	OrderIndex  int
}

type CoreItemInput struct {
	WorkspaceID     uuid.UUID
	PortalID        uuid.UUID
	BoardID         uuid.UUID
	ContributorID   uuid.UUID
	AuthorID        uuid.UUID
	Title           string
	Description     string
	DescriptionHTML string
	Slug            string
	Source          string
}

type CoreBoardReviewerInput struct {
	WorkspaceID    uuid.UUID
	BoardID        uuid.UUID
	UserID         uuid.UUID
	EmailFrequency string
}

type CorePublicItemInput struct {
	PortalSlug          string
	BoardID             uuid.UUID
	AuthorID            uuid.UUID
	Title               string
	Description         string
	DescriptionHTML     string
	Source              string
	ParticipationIntent string
	Participant         *CoreParticipant
}

type CorePublicItemResult struct {
	Item            CoreItem
	Anonymous       bool
	ParticipantKind string
	Following       bool
}

type CorePublicCommentInput struct {
	PortalSlug  string
	ItemID      uuid.UUID
	AuthorID    uuid.UUID
	Participant *CoreParticipant
	ParentID    *uuid.UUID
	Body        string
}

type CorePublicVoteInput struct {
	PortalSlug  string
	ItemID      uuid.UUID
	UserID      uuid.UUID
	Participant *CoreParticipant
	Vote        int
}

type CoreUpdateItemStatusInput struct {
	Status         string
	RoadmapSummary *string
	ActorID        uuid.UUID
	AllowLinked    bool
}

type CoreCommentInput struct {
	WorkspaceID uuid.UUID
	ItemID      uuid.UUID
	AuthorID    uuid.UUID
	ParentID    *uuid.UUID
	Body        string
}

type CoreVoteResult struct {
	Vote      int
	VoteCount int
}

type CoreStoryLinkInput struct {
	WorkspaceID     uuid.UUID
	ItemID          uuid.UUID
	StoryID         uuid.UUID
	Relationship    string
	IsPrimary       bool
	CreatedByUserID uuid.UUID
}

type CoreCreateStoryInput struct {
	TeamID   uuid.UUID
	StoryID  *uuid.UUID
	StatusID *uuid.UUID
}

type CoreCreateStoryResult struct {
	ItemID  uuid.UUID
	StoryID uuid.UUID
	LinkID  uuid.UUID
	Created bool
}

type CoreItemDetails struct {
	Item       CoreItem
	Comments   []CoreComment
	StoryLinks []CoreStoryLink
}

type CoreListItemsInput struct {
	WorkspaceID uuid.UUID
	PortalID    uuid.UUID
	TeamID      *uuid.UUID
	ViewerID    uuid.UUID
	AuthorID    uuid.UUID
	ItemID      uuid.UUID
	Status      string
	BoardID     *uuid.UUID
	Search      string
	Sort        string
	Page        int
	PageSize    int
	DeletedOnly bool
}

type CorePortalSnapshotInput struct {
	AuthorID    uuid.UUID
	ItemID      uuid.UUID
	Status      string
	BoardID     *uuid.UUID
	Search      string
	Sort        string
	Page        int
	PageSize    int
	SummaryOnly bool
}

type CoreItemsPage struct {
	Items   []CoreItem
	HasMore bool
}

type CoreListContributorCommentsInput struct {
	PortalID uuid.UUID
	AuthorID uuid.UUID
	Page     int
	PageSize int
}

type CoreContributorCommentsPage struct {
	Comments []CoreContributorComment
	Page     int
	PageSize int
	HasMore  bool
}

// StoryPlan is the small story projection feedback needs when deciding whether
// an item can be linked or how its public status should be projected. Keeping
// this type feedback-owned prevents the stories persistence/service model from
// becoming part of feedback's domain contract.
type StoryPlan struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
	StatusID    *uuid.UUID
	DeletedAt   *time.Time
}

// StoryDraft describes the only story shape feedback may create. Fields that
// belong exclusively to stories remain behind the adapter rather than leaking
// into this module.
type StoryDraft struct {
	Title       string
	Description string
	StatusID    *uuid.UUID
	ReporterID  uuid.UUID
	TeamID      uuid.UUID
}

// StoryPlanner is a caller-owned capability port. Implementations must enforce
// the actor, workspace, and team policies of the stories module; feedback does
// not receive a broad stories service.
type StoryPlanner interface {
	CreateFromFeedback(ctx context.Context, workspaceID, actorID uuid.UUID, draft StoryDraft) (StoryPlan, error)
	GetForFeedback(ctx context.Context, workspaceID, storyID uuid.UUID) (StoryPlan, error)
	DeleteCreatedFromFeedback(ctx context.Context, workspaceID, storyID, actorID uuid.UUID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}
