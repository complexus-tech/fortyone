package feedbackhttp

import (
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
)

type AppPortal struct {
	ID                  uuid.UUID  `json:"id"`
	WorkspaceID         uuid.UUID  `json:"workspaceId"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug"`
	IsPublic            bool       `json:"isPublic"`
	ParticipationMode   string     `json:"participationMode"`
	GuestIdentityPolicy string     `json:"guestIdentityPolicy"`
	HasPublishedUpdates bool       `json:"hasPublishedUpdates"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	Boards              []AppBoard `json:"boards"`
	Items               []AppItem  `json:"items,omitempty"`
	ItemsHasMore        bool       `json:"itemsHasMore"`
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
	ID               uuid.UUID      `json:"id"`
	WorkspaceID      uuid.UUID      `json:"workspaceId"`
	PortalID         uuid.UUID      `json:"portalId"`
	BoardID          uuid.UUID      `json:"boardId"`
	AuthorID         *uuid.UUID     `json:"authorId"`
	AuthorName       string         `json:"authorName"`
	AuthorAvatar     *string        `json:"authorAvatar"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Slug             string         `json:"slug"`
	Status           string         `json:"status"`
	VoteCount        int            `json:"voteCount"`
	UpvoteCount      int            `json:"upvoteCount"`
	DownvoteCount    int            `json:"downvoteCount"`
	CommentCount     int            `json:"commentCount"`
	RoadmapSummary   *string        `json:"roadmapSummary,omitempty"`
	ReadAt           *time.Time     `json:"readAt"`
	DeletedAt        *time.Time     `json:"deletedAt,omitempty"`
	RestoreUntil     *time.Time     `json:"restoreUntil,omitempty"`
	Board            *AppBoard      `json:"board,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	Comments         []AppComment   `json:"comments"`
	StoryLinks       []AppStoryLink `json:"storyLinks"`
	Anonymous        bool           `json:"anonymous,omitempty"`
	ParticipantKind  string         `json:"participantKind,omitempty"`
	AuthorMasked     bool           `json:"authorMasked,omitempty"`
	MergedIntoItemID *uuid.UUID     `json:"mergedIntoItemId,omitempty"`
	MergedAt         *time.Time     `json:"mergedAt,omitempty"`
	MergedByUserID   *uuid.UUID     `json:"mergedByUserId,omitempty"`
	Following        bool           `json:"following,omitempty"`
}

type AppPrivateAuthor struct {
	ContributorID uuid.UUID  `json:"contributorId"`
	UserID        *uuid.UUID `json:"userId"`
	Kind          string     `json:"kind"`
	DisplayName   string     `json:"displayName"`
	Email         *string    `json:"email"`
	AvatarURL     *string    `json:"avatarUrl"`
	PublicMasked  bool       `json:"publicMasked"`
}

type AppCanonicalItem struct {
	ItemID   uuid.UUID `json:"itemId"`
	ItemSlug string    `json:"itemSlug"`
	Merged   bool      `json:"merged"`
}

type AppComment struct {
	ID              uuid.UUID  `json:"id"`
	WorkspaceID     uuid.UUID  `json:"workspaceId"`
	ItemID          uuid.UUID  `json:"itemId"`
	AuthorID        *uuid.UUID `json:"authorId"`
	ParentID        *uuid.UUID `json:"parentId"`
	AuthorName      string     `json:"authorName"`
	AuthorAvatar    *string    `json:"authorAvatar"`
	ParticipantKind string     `json:"participantKind,omitempty"`
	AuthorMasked    bool       `json:"authorMasked,omitempty"`
	Body            string     `json:"body"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
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

type AppContributorActivity struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	FeedbackID    uuid.UUID `json:"feedbackId"`
	FeedbackTitle string    `json:"feedbackTitle"`
	FeedbackSlug  string    `json:"feedbackSlug"`
	Body          string    `json:"body"`
	BoardName     string    `json:"boardName"`
	Status        string    `json:"status"`
	VoteCount     int       `json:"voteCount"`
	CommentCount  int       `json:"commentCount"`
	PortalSlug    string    `json:"portalSlug"`
	WorkspaceName string    `json:"workspaceName"`
	WorkspaceSlug string    `json:"workspaceSlug"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AppContributorActivityPage struct {
	Activities    []AppContributorActivity `json:"activities"`
	Page          int                      `json:"page"`
	PageSize      int                      `json:"pageSize"`
	HasMore       bool                     `json:"hasMore"`
	FeedbackCount int                      `json:"feedbackCount"`
	CommentCount  int                      `json:"commentCount"`
	VoteScore     int                      `json:"voteScore"`
	PortalCount   int                      `json:"portalCount"`
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
	Vote            int    `json:"vote"`
	Voted           bool   `json:"voted"`
	VoteCount       int    `json:"voteCount"`
	ParticipantKind string `json:"participantKind,omitempty"`
}

type AppVoteInput struct {
	Vote                int    `json:"vote"`
	ParticipationIntent string `json:"participationIntent"`
}

type AppUpdatePortal struct {
	IsPublic            *bool   `json:"isPublic"`
	ParticipationMode   *string `json:"participationMode"`
	GuestIdentityPolicy *string `json:"guestIdentityPolicy"`
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
	AvatarURL      *string   `json:"avatarUrl"`
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
	BoardID             uuid.UUID `json:"boardId"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	ParticipationIntent string    `json:"participationIntent"`
	Website             string    `json:"website"`
}

type AppSimilarItem struct {
	ID           uuid.UUID  `json:"id"`
	Slug         string     `json:"slug"`
	Title        string     `json:"title"`
	AuthorID     *uuid.UUID `json:"authorId"`
	AuthorName   string     `json:"authorName"`
	AuthorAvatar *string    `json:"authorAvatar"`
	Status       string     `json:"status"`
	VoteCount    int        `json:"voteCount"`
	CommentCount int        `json:"commentCount"`
	Confidence   float64    `json:"confidence"`
	IsDuplicate  bool       `json:"isDuplicate"`
}

type AppUpdateItemStatus struct {
	Status         string  `json:"status"`
	RoadmapSummary *string `json:"roadmapSummary"`
}

type AppMergeItemInput struct {
	TargetItemID uuid.UUID `json:"targetItemId"`
}

type AppMergeCandidate struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	VoteCount    int       `json:"voteCount"`
	CommentCount int       `json:"commentCount"`
}

type AppMergeCandidatesPage struct {
	Candidates []AppMergeCandidate `json:"candidates"`
	HasMore    bool                `json:"hasMore"`
}

type AppMergeItemResult struct {
	SourceItemID         uuid.UUID `json:"sourceItemId"`
	TargetItemID         uuid.UUID `json:"targetItemId"`
	PortalID             uuid.UUID `json:"portalId"`
	MergedAt             time.Time `json:"mergedAt"`
	MergedByUserID       uuid.UUID `json:"mergedByUserId"`
	MovedFollowerCount   int       `json:"movedFollowerCount"`
	MovedUpdateLinkCount int       `json:"movedUpdateLinkCount"`
	MovedStoryLinkCount  int       `json:"movedStoryLinkCount"`
	Target               AppItem   `json:"target"`
}

type AppCreateComment struct {
	Body                string     `json:"body"`
	ParentID            *uuid.UUID `json:"parentId"`
	ParticipationIntent string     `json:"participationIntent"`
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
		ID:                  core.ID,
		WorkspaceID:         core.WorkspaceID,
		Name:                core.Name,
		Slug:                core.Slug,
		IsPublic:            core.IsPublic,
		ParticipationMode:   core.ParticipationMode,
		GuestIdentityPolicy: core.GuestIdentityPolicy,
		HasPublishedUpdates: core.HasPublishedUpdates,
		CreatedAt:           core.CreatedAt,
		UpdatedAt:           core.UpdatedAt,
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
		AvatarURL:      core.AvatarURL,
		Role:           core.Role,
		EmailFrequency: core.EmailFrequency,
	}
}

func toAppItem(core feedback.CoreItem, comments []AppComment, links []AppStoryLink) AppItem {
	authorID := uuidPointer(core.AuthorID)
	authorAvatar := core.AuthorAvatar
	if core.AuthorMasked {
		authorID = nil
		authorAvatar = nil
	}
	item := AppItem{
		ID:               core.ID,
		WorkspaceID:      core.WorkspaceID,
		PortalID:         core.PortalID,
		BoardID:          core.BoardID,
		AuthorID:         authorID,
		AuthorName:       core.AuthorName,
		AuthorAvatar:     authorAvatar,
		ParticipantKind:  core.ParticipantKind,
		AuthorMasked:     core.AuthorMasked,
		MergedIntoItemID: core.MergedIntoItemID,
		MergedAt:         core.MergedAt,
		MergedByUserID:   core.MergedByUserID,
		Following:        core.Following,
		Title:            core.Title,
		Description:      core.Description,
		Slug:             core.Slug,
		Status:           core.Status,
		VoteCount:        core.VoteCount,
		UpvoteCount:      core.UpvoteCount,
		DownvoteCount:    core.DownvoteCount,
		CommentCount:     core.CommentCount,
		RoadmapSummary:   core.RoadmapSummary,
		ReadAt:           core.ReadAt,
		DeletedAt:        core.DeletedAt,
		CreatedAt:        core.CreatedAt,
		UpdatedAt:        core.UpdatedAt,
		Comments:         comments,
		StoryLinks:       links,
	}
	if core.DeletedAt != nil {
		restoreUntil := core.DeletedAt.Add(30 * 24 * time.Hour)
		item.RestoreUntil = &restoreUntil
	}
	if core.Board.ID != uuid.Nil {
		board := toAppBoard(core.Board)
		item.Board = &board
	}
	return item
}

func toAppSimilarItem(core feedback.CoreSimilarItem) AppSimilarItem {
	return AppSimilarItem{
		ID:           core.ID,
		Slug:         core.Slug,
		Title:        core.Title,
		AuthorID:     core.AuthorID,
		AuthorName:   core.AuthorName,
		AuthorAvatar: core.AuthorAvatar,
		Status:       core.Status,
		VoteCount:    core.VoteCount,
		CommentCount: core.CommentCount,
		Confidence:   core.Confidence,
		IsDuplicate:  core.IsDuplicate,
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	return &value
}

func toAppComment(core feedback.CoreComment) AppComment {
	authorID := uuidPointer(core.AuthorID)
	authorAvatar := core.AuthorAvatar
	if core.AuthorMasked {
		authorID = nil
		authorAvatar = nil
	}
	return AppComment{
		ID:              core.ID,
		WorkspaceID:     core.WorkspaceID,
		ItemID:          core.ItemID,
		AuthorID:        authorID,
		ParentID:        core.ParentID,
		AuthorName:      core.AuthorName,
		AuthorAvatar:    authorAvatar,
		ParticipantKind: core.ParticipantKind,
		AuthorMasked:    core.AuthorMasked,
		Body:            core.Body,
		CreatedAt:       core.CreatedAt,
		UpdatedAt:       core.UpdatedAt,
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

func toAppContributorActivityPage(core feedback.CoreContributorActivityPage) AppContributorActivityPage {
	activities := make([]AppContributorActivity, 0, len(core.Activities))
	for _, activity := range core.Activities {
		activities = append(activities, AppContributorActivity{
			ID:            activity.ID,
			Type:          activity.Type,
			FeedbackID:    activity.FeedbackID,
			FeedbackTitle: activity.FeedbackTitle,
			FeedbackSlug:  activity.FeedbackSlug,
			Body:          activity.Body,
			BoardName:     activity.BoardName,
			Status:        activity.Status,
			VoteCount:     activity.VoteCount,
			CommentCount:  activity.CommentCount,
			PortalSlug:    activity.PortalSlug,
			WorkspaceName: activity.WorkspaceName,
			WorkspaceSlug: activity.WorkspaceSlug,
			CreatedAt:     activity.CreatedAt,
		})
	}
	return AppContributorActivityPage{
		Activities:    activities,
		Page:          core.Page,
		PageSize:      core.PageSize,
		HasMore:       core.HasMore,
		FeedbackCount: core.FeedbackCount,
		CommentCount:  core.CommentCount,
		VoteScore:     core.VoteScore,
		PortalCount:   core.PortalCount,
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
