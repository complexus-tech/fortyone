package storieshttp

import (
	"time"

	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

type AppNewStory struct {
	Title                    string      `json:"title" validate:"required"`
	EstimateValue            *int16      `json:"estimateValue"`
	EstimatedDurationMinutes *int        `json:"estimatedDurationMinutes"`
	MinimumFocusBlockMinutes *int        `json:"minimumFocusBlockMinutes"`
	AutoSchedulingEnabled    bool        `json:"autoSchedulingEnabled"`
	Description              *string     `json:"description"`
	DescriptionHTML          *string     `json:"descriptionHTML"`
	Parent                   *uuid.UUID  `json:"parentId"`
	Objective                *uuid.UUID  `json:"objectiveId"`
	Status                   *uuid.UUID  `json:"statusId"`
	Assignee                 *uuid.UUID  `json:"assigneeId"`
	Priority                 string      `json:"priority" validate:"oneof='No Priority' Low Medium High Urgent"`
	Sprint                   *uuid.UUID  `json:"sprintId"`
	KeyResult                *uuid.UUID  `json:"keyResultId"`
	LabelIDs                 []uuid.UUID `json:"labelIds"`
	Team                     uuid.UUID   `json:"teamId" validate:"required"`
	StartDate                *date.Date  `json:"startDate"`
	EndDate                  *date.Date  `json:"endDate"`
	IdempotencyKey           *string     `json:"idempotencyKey" validate:"omitempty,max=128"`
}

type AppNewComment struct {
	Comment  string      `json:"comment" validate:"required"`
	Parent   *uuid.UUID  `json:"parentId"`
	Mentions []uuid.UUID `json:"mentions"`
}

type AppComment struct {
	ID          uuid.UUID      `json:"id"`
	StoryID     uuid.UUID      `json:"storyId"`
	Parent      *uuid.UUID     `json:"parentId"`
	UserID      uuid.UUID      `json:"userId"`
	User        AppUserSummary `json:"user"`
	Comment     string         `json:"comment"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	SubComments []AppComment   `json:"subComments"`
}

type AppNewAssociation struct {
	ToStoryID       uuid.UUID `json:"toStoryId" validate:"required"`
	AssociationType string    `json:"type" validate:"required"`
}

type AppUpdateAssociation struct {
	FromStoryID     uuid.UUID `json:"fromStoryId" validate:"required"`
	ToStoryID       uuid.UUID `json:"toStoryId" validate:"required"`
	AssociationType string    `json:"type" validate:"required"`
}

// StoryFilters represents filtering options for stories at the handler level
type StoryFilters struct {
	StatusIDs              []uuid.UUID `json:"statusIds"`
	ExcludedStatusIDs      []uuid.UUID `json:"excludedStatusIds"`
	AssigneeIDs            []uuid.UUID `json:"assigneeIds"`
	ExcludedAssigneeIDs    []uuid.UUID `json:"excludedAssigneeIds"`
	CollaboratorIDs        []uuid.UUID `json:"collaboratorIds"`
	ReporterIDs            []uuid.UUID `json:"reporterIds"`
	ExcludedReporterIDs    []uuid.UUID `json:"excludedReporterIds"`
	TitleContains          *string     `json:"titleContains"`
	TitleNotContains       *string     `json:"titleNotContains"`
	Priorities             []string    `json:"priorities"`
	ExcludedPriorities     []string    `json:"excludedPriorities"`
	Categories             []string    `json:"categories"`
	TeamIDs                []uuid.UUID `json:"teamIds"`
	ExcludedTeamIDs        []uuid.UUID `json:"excludedTeamIds"`
	SprintIDs              []uuid.UUID `json:"sprintIds"`
	ExcludedSprintIDs      []uuid.UUID `json:"excludedSprintIds"`
	LabelIDs               []uuid.UUID `json:"labelIds"`
	ExcludedLabelIDs       []uuid.UUID `json:"excludedLabelIds"`
	EstimateValues         []int16     `json:"estimateValues"`
	ExcludedEstimateValues []int16     `json:"excludedEstimateValues"`
	Parent                 *uuid.UUID  `json:"parentId"`
	Objective              *uuid.UUID  `json:"objectiveId"`
	ExcludedObjective      *uuid.UUID  `json:"excludedObjectiveId"`
	Epic                   *uuid.UUID  `json:"epicId"`
	KeyResult              *uuid.UUID  `json:"keyResultId"`
	HasNoAssignee          *bool       `json:"hasNoAssignee"`
	HasAssignee            *bool       `json:"hasAssignee"`
	HasBlockedBy           *bool       `json:"hasBlockedBy"`
	AssignedToMe           *bool       `json:"assignedToMe"`
	CollaboratingWithMe    *bool       `json:"collaboratingWithMe"`
	CreatedByMe            *bool       `json:"createdByMe"`
	ShowSubStories         *bool       `json:"showSubStories"`
	// Date range filters
	CreatedAfter    *time.Time `json:"createdAfter"`
	CreatedBefore   *time.Time `json:"createdBefore"`
	UpdatedAfter    *time.Time `json:"updatedAfter"`
	UpdatedBefore   *time.Time `json:"updatedBefore"`
	StartDateAfter  *time.Time `json:"startDateAfter"`
	StartDateBefore *time.Time `json:"startDateBefore"`
	StartDateNot    *time.Time `json:"startDateNot"`
	DeadlineAfter   *time.Time `json:"deadlineAfter"`
	DeadlineBefore  *time.Time `json:"deadlineBefore"`
	DeadlineNot     *time.Time `json:"deadlineNot"`
	CompletedAfter  *time.Time `json:"completedAfter"`
	CompletedBefore *time.Time `json:"completedBefore"`
	IncludeArchived *bool      `json:"includeArchived"`
	IncludeDeleted  *bool      `json:"includeDeleted"`
}

// StoryQuery represents query parameters for grouped stories at the handler level
type StoryQuery struct {
	Filters         StoryFilters `json:"filters"`
	GroupBy         string       `json:"groupBy"`
	OrderBy         string       `json:"orderBy"`
	OrderDirection  string       `json:"orderDirection"`
	StoriesPerGroup int          `json:"storiesPerGroup"`
	GroupKey        string       `json:"groupKey"`
	Page            int          `json:"page"`
	PageSize        int          `json:"pageSize"`
}

// StoryGroup represents a group of stories at the handler level
type StoryGroup struct {
	Key         string         `json:"key"`
	LoadedCount int            `json:"loadedCount"`
	TotalCount  int            `json:"totalCount"`
	HasMore     bool           `json:"hasMore"`
	Stories     []AppStoryList `json:"stories"`
	NextPage    int            `json:"nextPage"`
}

// GroupsMeta represents metadata for grouped stories response
type GroupsMeta struct {
	TotalGroups    int          `json:"totalGroups"`
	Filters        StoryFilters `json:"filters"`
	GroupBy        string       `json:"groupBy"`
	OrderBy        string       `json:"orderBy"`
	OrderDirection string       `json:"orderDirection"`
}

// StoriesResponse represents the response for stories (grouped or regular)
type StoriesResponse struct {
	Stories []AppStoryList `json:"stories,omitempty"`
	Groups  []StoryGroup   `json:"groups,omitempty"`
	Meta    GroupsMeta     `json:"meta"`
}

// GroupPagination represents pagination info for a specific group
type GroupPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

// GroupStoriesResponse represents the response for loading more stories in a group
type GroupStoriesResponse struct {
	GroupKey       string          `json:"groupKey"`
	Stories        []AppStoryList  `json:"stories"`
	Pagination     GroupPagination `json:"pagination"`
	Filters        StoryFilters    `json:"filters"`
	OrderBy        string          `json:"orderBy"`
	OrderDirection string          `json:"orderDirection"`
}

// CategoryPagination represents pagination info for category stories
type CategoryPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

// CategoryMeta represents metadata for category stories response
type CategoryMeta struct {
	Category    string    `json:"category"`
	TeamID      uuid.UUID `json:"teamId"`
	TotalLoaded int       `json:"totalLoaded"`
}

// CategoryStoriesResponse represents the response for stories filtered by category
type CategoryStoriesResponse struct {
	Stories    []AppStoryList     `json:"stories"`
	Pagination CategoryPagination `json:"pagination"`
	Meta       CategoryMeta       `json:"meta"`
}
