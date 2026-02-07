package stories

import (
	"time"

	"github.com/google/uuid"
)

type CoreLabel struct {
	LabelID     uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	ProjectID   uuid.UUID  `json:"projectId"`
	TeamID      *uuid.UUID `json:"teamId"`
	WorkspaceID *uuid.UUID `json:"workspaceId"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// CoreStoryList represents a list of stories.
type CoreStoryList struct {
	ID          uuid.UUID       `json:"id"`
	SequenceID  int             `json:"sequence_id"`
	Title       string          `json:"title"`
	Parent      *uuid.UUID      `json:"parent_id"`
	Objective   *uuid.UUID      `json:"objective_id"`
	Epic        *uuid.UUID      `json:"epic_id"`
	Status      *uuid.UUID      `json:"status_id"`
	Assignee    *uuid.UUID      `json:"assignee_id"`
	Reporter    *uuid.UUID      `json:"reporter_id"`
	Priority    string          `json:"priority"`
	Sprint      *uuid.UUID      `json:"sprint_id"`
	KeyResult   *uuid.UUID      `json:"key_result_id"`
	Team        uuid.UUID       `json:"team_id"`
	Workspace   uuid.UUID       `json:"workspace_id"`
	StartDate   *time.Time      `json:"start_date"`
	EndDate     *time.Time      `json:"end_date"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	ArchivedAt  *time.Time      `json:"archived_at"`
	Labels      []uuid.UUID     `json:"labels"`
	SubStories  []CoreStoryList `json:"subStories"`
}

// CoreSingleStory represents a single story.
type CoreSingleStory struct {
	ID              uuid.UUID
	SequenceID      int
	Title           string
	TeamCode        string
	Description     *string
	DescriptionHTML *string
	Parent          *uuid.UUID
	Objective       *uuid.UUID
	Status          *uuid.UUID
	Assignee        *uuid.UUID
	BlockedBy       *uuid.UUID
	Blocking        *uuid.UUID
	Related         *uuid.UUID
	Reporter        *uuid.UUID
	Priority        string
	Sprint          *uuid.UUID
	Epic            *uuid.UUID
	KeyResult       *uuid.UUID
	Team            uuid.UUID
	Workspace       uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	ArchivedAt      *time.Time
	CompletedAt     *time.Time
	SubStories      []CoreStoryList
	Labels          []uuid.UUID
	Associations    []CoreStoryAssociation
}

type CoreNewStory struct {
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Objective       *uuid.UUID `json:"objectiveId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	BlockedBy       *uuid.UUID `json:"blockedById"`
	Blocking        *uuid.UUID `json:"blockingId"`
	Related         *uuid.UUID `json:"relatedId"`
	Reporter        *uuid.UUID `json:"reporterId"`
	Priority        string     `json:"priority"`
	Sprint          *uuid.UUID `json:"sprintId"`
	KeyResult       *uuid.UUID `json:"keyResultId"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
	Team            uuid.UUID  `json:"teamId"`
}

type CoreUpdateStory struct {
	Title           *string
	Description     *string
	DescriptionHTML *string
	Parent          *uuid.UUID
	Objective       *uuid.UUID
	Status          *uuid.UUID
	Assignee        *uuid.UUID
	Priority        *string
	Sprint          *uuid.UUID
	KeyResult       *uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CompletedAt     *time.Time
}

func toCoreSingleStory(ns CoreNewStory, workspaceId uuid.UUID) CoreSingleStory {
	now := time.Now()
	if ns.Priority == "" {
		ns.Priority = "No Priority"
	}

	return CoreSingleStory{
		Workspace:       workspaceId,
		Title:           ns.Title,
		Description:     ns.Description,
		DescriptionHTML: ns.DescriptionHTML,
		Parent:          ns.Parent,
		Objective:       ns.Objective,
		Status:          ns.Status,
		Assignee:        ns.Assignee,
		BlockedBy:       ns.BlockedBy,
		Blocking:        ns.Blocking,
		Related:         ns.Related,
		Reporter:        ns.Reporter,
		Priority:        ns.Priority,
		Sprint:          ns.Sprint,
		KeyResult:       ns.KeyResult,
		StartDate:       ns.StartDate,
		EndDate:         ns.EndDate,
		Team:            ns.Team,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// CoreStoryAssociation represents a relationship between two stories.
type CoreStoryAssociation struct {
	ID          uuid.UUID     `json:"id"`
	FromStoryID uuid.UUID     `json:"fromStoryId"`
	ToStoryID   uuid.UUID     `json:"toStoryId"`
	Type        string        `json:"type"` // "blocking", "related", "duplicate"
	Story       CoreStoryList `json:"story"`
}

// CoreActivity represents the core model for an activity.
type CoreActivity struct {
	ID           uuid.UUID `json:"id"`
	StoryID      uuid.UUID `json:"storyId"`
	UserID       uuid.UUID `json:"userId"`
	Type         string    `json:"type"`
	Field        string    `json:"field"`
	CurrentValue string    `json:"currentValue"`
	OldValue     any       `json:"oldValue"`
	NewValue     any       `json:"newValue"`
	CreatedAt    time.Time `json:"createdAt"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
}

// CoreActivityWithUser represents an activity with embedded user details
type CoreActivityWithUser struct {
	ID           uuid.UUID   `json:"id"`
	StoryID      uuid.UUID   `json:"storyId"`
	UserID       uuid.UUID   `json:"userId"`
	Type         string      `json:"type"`
	Field        string      `json:"field"`
	CurrentValue string      `json:"currentValue"`
	OldValue     any         `json:"oldValue"`
	NewValue     any         `json:"newValue"`
	CreatedAt    time.Time   `json:"createdAt"`
	WorkspaceID  uuid.UUID   `json:"workspaceId"`
	User         UserDetails `json:"user"`
}

// UserDetails represents basic user information for activities
type UserDetails struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
	IsActive  bool      `json:"isActive"`
}

type CoreNewComment struct {
	StoryID  uuid.UUID
	Parent   *uuid.UUID
	UserID   uuid.UUID
	Comment  string
	Mentions []uuid.UUID
}

// CoreStoryFilters represents filtering options for stories
type CoreStoryFilters struct {
	StatusIDs     []uuid.UUID `json:"statusIds"`
	AssigneeIDs   []uuid.UUID `json:"assigneeIds"`
	ReporterIDs   []uuid.UUID `json:"reporterIds"`
	Priorities    []string    `json:"priorities"`
	Categories    []string    `json:"categories"`
	TeamIDs       []uuid.UUID `json:"teamIds"`
	SprintIDs     []uuid.UUID `json:"sprintIds"`
	LabelIDs      []uuid.UUID `json:"labelIds"`
	Parent        *uuid.UUID  `json:"parentId"`
	Objective     *uuid.UUID  `json:"objectiveId"`
	Epic          *uuid.UUID  `json:"epicId"`
	KeyResult     *uuid.UUID  `json:"keyResultId"`
	HasNoAssignee *bool       `json:"hasNoAssignee"`
	AssignedToMe  *bool       `json:"assignedToMe"`
	CreatedByMe   *bool       `json:"createdByMe"`
	CurrentUserID uuid.UUID   `json:"currentUserId"`
	WorkspaceID   uuid.UUID   `json:"workspaceId"`
	// Date range filters
	CreatedAfter    *time.Time `json:"createdAfter"`
	CreatedBefore   *time.Time `json:"createdBefore"`
	UpdatedAfter    *time.Time `json:"updatedAfter"`
	UpdatedBefore   *time.Time `json:"updatedBefore"`
	DeadlineAfter   *time.Time `json:"deadlineAfter"`
	DeadlineBefore  *time.Time `json:"deadlineBefore"`
	CompletedAfter  *time.Time `json:"completedAfter"`
	CompletedBefore *time.Time `json:"completedBefore"`
	IsCompleted     *bool      `json:"isCompleted"`
	IsNotCompleted  *bool      `json:"isNotCompleted"`
	IncludeArchived *bool      `json:"includeArchived"`
	IncludeDeleted  *bool      `json:"includeDeleted"`
}

// CoreStoryQuery represents query parameters for grouped stories
type CoreStoryQuery struct {
	Filters         CoreStoryFilters `json:"filters"`
	GroupBy         string           `json:"groupBy"`
	OrderBy         string           `json:"orderBy"`
	OrderDirection  string           `json:"orderDirection"`
	StoriesPerGroup int              `json:"storiesPerGroup"`
	GroupKey        string           `json:"groupKey"`
	Page            int              `json:"page"`
	PageSize        int              `json:"pageSize"`
}

// CoreStoryGroup represents a group of stories
type CoreStoryGroup struct {
	Key         string          `json:"key"`
	LoadedCount int             `json:"loadedCount"`
	TotalCount  int             `json:"totalCount"`
	HasMore     bool            `json:"hasMore"`
	Stories     []CoreStoryList `json:"stories"`
	NextPage    int             `json:"nextPage"`
}
