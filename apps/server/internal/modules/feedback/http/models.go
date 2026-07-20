package feedbackhttp

import (
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
)

type AppPortal struct {
	ID           uuid.UUID  `json:"id"`
	WorkspaceID  uuid.UUID  `json:"workspaceId"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	IsPublic     bool       `json:"isPublic"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	Boards       []AppBoard `json:"boards,omitempty"`
	Items        []AppItem  `json:"items,omitempty"`
	ItemsHasMore bool       `json:"itemsHasMore"`
}

type AppBoard struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	PortalID    uuid.UUID `json:"portalId"`
	TeamID      uuid.UUID `json:"teamId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Color       string    `json:"color"`
	OrderIndex  int       `json:"orderIndex"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AppItem struct {
	ID             uuid.UUID      `json:"id"`
	WorkspaceID    uuid.UUID      `json:"workspaceId"`
	PortalID       uuid.UUID      `json:"portalId"`
	BoardID        uuid.UUID      `json:"boardId"`
	AuthorID       uuid.UUID      `json:"authorId"`
	AuthorName     string         `json:"authorName"`
	AuthorAvatar   *string        `json:"authorAvatar"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Slug           string         `json:"slug"`
	Status         string         `json:"status"`
	VoteCount      int            `json:"voteCount"`
	CommentCount   int            `json:"commentCount"`
	RoadmapSummary *string        `json:"roadmapSummary,omitempty"`
	ReadAt         *time.Time     `json:"readAt"`
	Board          *AppBoard      `json:"board,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Comments       []AppComment   `json:"comments"`
	StoryLinks     []AppStoryLink `json:"storyLinks"`
}

type AppComment struct {
	ID           uuid.UUID `json:"id"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
	ItemID       uuid.UUID `json:"itemId"`
	AuthorID     uuid.UUID `json:"authorId"`
	AuthorName   string    `json:"authorName"`
	AuthorAvatar *string   `json:"authorAvatar"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AppContributorStats struct {
	FeedbackCount int `json:"feedbackCount"`
	CommentCount  int `json:"commentCount"`
	VoteScore     int `json:"voteScore"`
}

type AppContributor struct {
	ID        uuid.UUID           `json:"id"`
	Name      string              `json:"name"`
	AvatarURL *string             `json:"avatarUrl"`
	JoinedAt  time.Time           `json:"joinedAt"`
	Stats     AppContributorStats `json:"stats"`
}

type AppContributorFeedback struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Slug  string    `json:"slug"`
}

type AppContributorComment struct {
	ID        uuid.UUID              `json:"id"`
	Body      string                 `json:"body"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Feedback  AppContributorFeedback `json:"feedback"`
}

type AppStoryLink struct {
	ID              uuid.UUID `json:"id"`
	WorkspaceID     uuid.UUID `json:"workspaceId"`
	ItemID          uuid.UUID `json:"itemId"`
	StoryID         uuid.UUID `json:"storyId"`
	StoryTitle      string    `json:"storyTitle,omitempty"`
	Relationship    string    `json:"relationship"`
	IsPrimary       bool      `json:"isPrimary"`
	CreatedByUserID uuid.UUID `json:"createdByUserId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AppStoryFeedbackLink struct {
	ID            uuid.UUID `json:"id"`
	WorkspaceID   uuid.UUID `json:"workspaceId"`
	ItemID        uuid.UUID `json:"itemId"`
	StoryID       uuid.UUID `json:"storyId"`
	TeamID        uuid.UUID `json:"teamId"`
	FeedbackTitle string    `json:"feedbackTitle"`
	Relationship  string    `json:"relationship"`
	IsPrimary     bool      `json:"isPrimary"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AppTeamFeedbackSummary struct {
	TeamID      uuid.UUID `json:"teamId"`
	Enabled     bool      `json:"enabled"`
	TotalCount  int       `json:"totalCount"`
	UnreadCount int       `json:"unreadCount"`
}

type AppFeedbackReadState struct {
	ReadAt *time.Time `json:"readAt"`
}

type AppVoteResult struct {
	Vote      int  `json:"vote"`
	Voted     bool `json:"voted"`
	VoteCount int  `json:"voteCount"`
}

type AppVoteInput struct {
	Vote int `json:"vote"`
}

type AppUpdatePortal struct {
	IsPublic bool `json:"isPublic"`
}

type AppCreateBoard struct {
	PortalID   uuid.UUID `json:"portalId"`
	TeamID     uuid.UUID `json:"teamId"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	Color      string    `json:"color"`
	OrderIndex int       `json:"orderIndex"`
}

type AppBoardReviewer struct {
	UserID         uuid.UUID `json:"userId"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	EmailFrequency string    `json:"emailFrequency"`
}

type AppSetBoardReviewer struct {
	EmailFrequency string `json:"emailFrequency"`
}

type AppCreateItem struct {
	PortalID    uuid.UUID `json:"portalId"`
	BoardID     uuid.UUID `json:"boardId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type AppCreatePublicItem struct {
	BoardID     uuid.UUID `json:"boardId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type AppUpdateItemStatus struct {
	Status         string  `json:"status"`
	RoadmapSummary *string `json:"roadmapSummary"`
}

type AppCreateComment struct {
	Body string `json:"body"`
}

type AppCreateStoryFromItem struct {
	TeamID   uuid.UUID  `json:"teamId"`
	StoryID  *uuid.UUID `json:"storyId"`
	StatusID *uuid.UUID `json:"statusId"`
}

type AppCreateStoryResult struct {
	ItemID  uuid.UUID `json:"itemId"`
	StoryID uuid.UUID `json:"storyId"`
	LinkID  uuid.UUID `json:"linkId"`
	Created bool      `json:"created"`
}

type AppItemsPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

type AppTeamFeedbackResponse struct {
	Feedback   []AppItem          `json:"feedback"`
	Pagination AppItemsPagination `json:"pagination"`
}

type AppContributorCommentsResponse struct {
	Comments   []AppContributorComment `json:"comments"`
	Pagination AppItemsPagination      `json:"pagination"`
}

func toAppPortal(core feedback.CorePortal) AppPortal {
	return AppPortal{
		ID:          core.ID,
		WorkspaceID: core.WorkspaceID,
		Name:        core.Name,
		Slug:        core.Slug,
		IsPublic:    core.IsPublic,
		CreatedAt:   core.CreatedAt,
		UpdatedAt:   core.UpdatedAt,
	}
}

func toAppBoard(core feedback.CoreBoard) AppBoard {
	return AppBoard{
		ID:          core.ID,
		WorkspaceID: core.WorkspaceID,
		PortalID:    core.PortalID,
		TeamID:      core.TeamID,
		Name:        core.Name,
		Slug:        core.Slug,
		Color:       core.Color,
		OrderIndex:  core.OrderIndex,
		CreatedAt:   core.CreatedAt,
		UpdatedAt:   core.UpdatedAt,
	}
}

func toAppBoardReviewer(core feedback.CoreBoardReviewer) AppBoardReviewer {
	return AppBoardReviewer{
		UserID:         core.UserID,
		Name:           core.Name,
		Email:          core.Email,
		Role:           core.Role,
		EmailFrequency: core.EmailFrequency,
	}
}

func toAppItem(core feedback.CoreItem, comments []AppComment, links []AppStoryLink) AppItem {
	item := AppItem{
		ID:             core.ID,
		WorkspaceID:    core.WorkspaceID,
		PortalID:       core.PortalID,
		BoardID:        core.BoardID,
		AuthorID:       core.AuthorID,
		AuthorName:     core.AuthorName,
		AuthorAvatar:   core.AuthorAvatar,
		Title:          core.Title,
		Description:    core.Description,
		Slug:           core.Slug,
		Status:         core.Status,
		VoteCount:      core.VoteCount,
		CommentCount:   core.CommentCount,
		RoadmapSummary: core.RoadmapSummary,
		ReadAt:         core.ReadAt,
		CreatedAt:      core.CreatedAt,
		UpdatedAt:      core.UpdatedAt,
		Comments:       comments,
		StoryLinks:     links,
	}
	if core.Board.ID != uuid.Nil {
		board := toAppBoard(core.Board)
		item.Board = &board
	}
	return item
}

func toAppComment(core feedback.CoreComment) AppComment {
	return AppComment{
		ID:           core.ID,
		WorkspaceID:  core.WorkspaceID,
		ItemID:       core.ItemID,
		AuthorID:     core.AuthorID,
		AuthorName:   core.AuthorName,
		AuthorAvatar: core.AuthorAvatar,
		Body:         core.Body,
		CreatedAt:    core.CreatedAt,
		UpdatedAt:    core.UpdatedAt,
	}
}

func toAppContributor(core feedback.CoreContributor) AppContributor {
	return AppContributor{
		ID:        core.ID,
		Name:      core.Name,
		AvatarURL: core.AvatarURL,
		JoinedAt:  core.JoinedAt,
		Stats: AppContributorStats{
			FeedbackCount: core.Stats.FeedbackCount,
			CommentCount:  core.Stats.CommentCount,
			VoteScore:     core.Stats.VoteScore,
		},
	}
}

func toAppContributorComment(core feedback.CoreContributorComment) AppContributorComment {
	return AppContributorComment{
		ID:        core.ID,
		Body:      core.Body,
		CreatedAt: core.CreatedAt,
		UpdatedAt: core.UpdatedAt,
		Feedback: AppContributorFeedback{
			ID:    core.ItemID,
			Title: core.FeedbackTitle,
			Slug:  core.FeedbackSlug,
		},
	}
}

func toAppStoryLink(core feedback.CoreStoryLink) AppStoryLink {
	return AppStoryLink{
		ID:              core.ID,
		WorkspaceID:     core.WorkspaceID,
		ItemID:          core.ItemID,
		StoryID:         core.StoryID,
		StoryTitle:      core.StoryTitle,
		Relationship:    core.Relationship,
		IsPrimary:       core.IsPrimary,
		CreatedByUserID: core.CreatedByUserID,
		CreatedAt:       core.CreatedAt,
	}
}

func toAppStoryFeedbackLink(core feedback.CoreStoryFeedbackLink) AppStoryFeedbackLink {
	return AppStoryFeedbackLink{
		ID:            core.ID,
		WorkspaceID:   core.WorkspaceID,
		ItemID:        core.ItemID,
		StoryID:       core.StoryID,
		TeamID:        core.TeamID,
		FeedbackTitle: core.FeedbackTitle,
		Relationship:  core.Relationship,
		IsPrimary:     core.IsPrimary,
		CreatedAt:     core.CreatedAt,
	}
}

func toAppPortalSnapshot(core feedback.CorePortalSnapshot) AppPortal {
	commentsByItem := map[uuid.UUID][]AppComment{}
	for _, comment := range core.Comments {
		commentsByItem[comment.ItemID] = append(commentsByItem[comment.ItemID], toAppComment(comment))
	}
	linksByItem := map[uuid.UUID][]AppStoryLink{}
	for _, link := range core.Links {
		linksByItem[link.ItemID] = append(linksByItem[link.ItemID], toAppStoryLink(link))
	}
	portal := toAppPortal(core.Portal)
	portal.Boards = make([]AppBoard, 0, len(core.Boards))
	for _, board := range core.Boards {
		portal.Boards = append(portal.Boards, toAppBoard(board))
	}
	portal.Items = make([]AppItem, 0, len(core.Items))
	for _, item := range core.Items {
		portal.Items = append(portal.Items, toAppItem(item, commentsByItem[item.ID], linksByItem[item.ID]))
	}
	portal.ItemsHasMore = core.ItemsHasMore
	return portal
}
