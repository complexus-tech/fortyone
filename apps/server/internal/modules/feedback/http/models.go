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
	Description  string     `json:"description"`
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

type AppStoryLink struct {
	ID              uuid.UUID `json:"id"`
	WorkspaceID     uuid.UUID `json:"workspaceId"`
	ItemID          uuid.UUID `json:"itemId"`
	StoryID         uuid.UUID `json:"storyId"`
	Relationship    string    `json:"relationship"`
	CreatedByUserID uuid.UUID `json:"createdByUserId"`
	CreatedAt       time.Time `json:"createdAt"`
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
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
}

type AppCreateBoard struct {
	PortalID   uuid.UUID `json:"portalId"`
	TeamID     uuid.UUID `json:"teamId"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	Color      string    `json:"color"`
	OrderIndex int       `json:"orderIndex"`
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
	StatusID *uuid.UUID `json:"statusId"`
}

type AppCreateStoryResult struct {
	ItemID  uuid.UUID `json:"itemId"`
	StoryID uuid.UUID `json:"storyId"`
	LinkID  uuid.UUID `json:"linkId"`
}

func toAppPortal(core feedback.CorePortal) AppPortal {
	return AppPortal{
		ID:          core.ID,
		WorkspaceID: core.WorkspaceID,
		Name:        core.Name,
		Slug:        core.Slug,
		Description: core.Description,
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

func toAppItem(core feedback.CoreItem, comments []AppComment, links []AppStoryLink) AppItem {
	return AppItem{
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
		CreatedAt:      core.CreatedAt,
		UpdatedAt:      core.UpdatedAt,
		Comments:       comments,
		StoryLinks:     links,
	}
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

func toAppStoryLink(core feedback.CoreStoryLink) AppStoryLink {
	return AppStoryLink{
		ID:              core.ID,
		WorkspaceID:     core.WorkspaceID,
		ItemID:          core.ItemID,
		StoryID:         core.StoryID,
		Relationship:    core.Relationship,
		CreatedByUserID: core.CreatedByUserID,
		CreatedAt:       core.CreatedAt,
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
