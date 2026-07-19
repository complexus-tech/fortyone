package feedback

import (
	"context"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusReviewing  = "reviewing"
	StatusPlanned    = "planned"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusClosed     = "closed"

	RelationshipCreatedFrom = "created_from"
	RelationshipLinked      = "linked"
	RelationshipSolves      = "solves"
)

type CorePortal struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Slug        string
	Description string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	PortalID       uuid.UUID
	BoardID        uuid.UUID
	AuthorID       uuid.UUID
	AuthorName     string
	AuthorEmail    string
	AuthorAvatar   *string
	Title          string
	Description    string
	Slug           string
	Status         string
	VoteCount      int
	CommentCount   int
	RoadmapSummary *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CoreComment struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	ItemID       uuid.UUID
	AuthorID     uuid.UUID
	AuthorName   string
	AuthorAvatar *string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CoreStoryLink struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	ItemID          uuid.UUID
	StoryID         uuid.UUID
	Relationship    string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
}

type CorePortalSnapshot struct {
	Portal       CorePortal
	Boards       []CoreBoard
	Items        []CoreItem
	ItemsHasMore bool
	Comments     []CoreComment
	Links        []CoreStoryLink
}

type CorePortalInput struct {
	WorkspaceID uuid.UUID
	Description string
	IsPublic    bool
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
	Name        string
	Slug        string
	Color       string
	OrderIndex  int
}

type CoreItemInput struct {
	WorkspaceID uuid.UUID
	PortalID    uuid.UUID
	BoardID     uuid.UUID
	AuthorID    uuid.UUID
	Title       string
	Description string
	Slug        string
}

type CoreUpdateItemStatusInput struct {
	Status         string
	RoadmapSummary *string
}

type CoreCommentInput struct {
	WorkspaceID uuid.UUID
	ItemID      uuid.UUID
	AuthorID    uuid.UUID
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
	CreatedByUserID uuid.UUID
}

type CoreCreateStoryInput struct {
	TeamID   uuid.UUID
	StatusID *uuid.UUID
}

type CoreCreateStoryResult struct {
	ItemID  uuid.UUID
	StoryID uuid.UUID
	LinkID  uuid.UUID
}

type Repository interface {
	GetPortalBySlug(ctx context.Context, slug string) (CorePortal, error)
	GetPortalByWorkspaceSlugAndSlug(ctx context.Context, workspaceSlug, slug string) (CorePortal, error)
	GetPortal(ctx context.Context, workspaceID, portalID uuid.UUID) (CorePortal, error)
	ListPortals(ctx context.Context, workspaceID uuid.UUID) ([]CorePortal, error)
	CreatePortal(ctx context.Context, input CorePortalInput) (CorePortal, error)
	UpdatePortal(ctx context.Context, workspaceID, portalID uuid.UUID, input CorePortalInput) (CorePortal, error)
	ListBoards(ctx context.Context, portalID uuid.UUID) ([]CoreBoard, error)
	CreateBoard(ctx context.Context, input CoreBoardInput) (CoreBoard, error)
	ListItems(ctx context.Context, input CoreListItemsInput) (CoreItemsPage, error)
	ListComments(ctx context.Context, portalID uuid.UUID) ([]CoreComment, error)
	ListStoryLinks(ctx context.Context, portalID uuid.UUID) ([]CoreStoryLink, error)
	GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (CoreItem, error)
	CreateItem(ctx context.Context, input CoreItemInput) (CoreItem, error)
	UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input CoreUpdateItemStatusInput) (CoreItem, error)
	CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error)
	ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error)
	LinkStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error)
	FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error)
}

type CoreListItemsInput struct {
	PortalID uuid.UUID
	Status   string
	BoardID  *uuid.UUID
	Search   string
	Sort     string
	Page     int
	PageSize int
}

type CoreItemsPage struct {
	Items   []CoreItem
	HasMore bool
}

type StoryService interface {
	CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
}
