package storiesgrp

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AppActivity represents an activity in the application layer
type AppActivity struct {
	ID           uuid.UUID `json:"id"`
	StoryID      uuid.UUID `json:"storyId"`
	UserID       uuid.UUID `json:"userId"`
	Type         string    `json:"type"`
	Field        string    `json:"field"`
	CurrentValue string    `json:"currentValue"`
	CreatedAt    time.Time `json:"createdAt"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
}

// AppNewLabels represents a new label in the application.
type AppNewLabels struct {
	Labels []uuid.UUID `json:"labels"`
}

func toAppActivity(i stories.CoreActivity) AppActivity {
	return AppActivity{
		ID:           i.ID,
		StoryID:      i.StoryID,
		UserID:       i.UserID,
		Type:         i.Type,
		Field:        i.Field,
		CurrentValue: i.CurrentValue,
		CreatedAt:    i.CreatedAt,
		WorkspaceID:  i.WorkspaceID,
	}
}

// toAppActivities converts a list of core activities to a list of application activities
func toAppActivities(activities []stories.CoreActivity) []AppActivity {
	appActivities := make([]AppActivity, len(activities))
	for i, activity := range activities {
		appActivities[i] = toAppActivity(activity)
	}
	return appActivities
}

// ActivitiesPagination represents pagination information for activities
type ActivitiesPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

// ActivitiesResponse represents the response for paginated activities
type ActivitiesResponse struct {
	Activities []AppActivity        `json:"activities"`
	Pagination ActivitiesPagination `json:"pagination"`
}

// CommentsPagination represents pagination information for comments
type CommentsPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

// CommentsResponse represents the response for paginated comments
type CommentsResponse struct {
	Comments   []AppComment       `json:"comments"`
	Pagination CommentsPagination `json:"pagination"`
}

// AppBulkDeleteRequest represents a request to delete multiple stories.
type AppBulkDeleteRequest struct {
	StoryIDs   []uuid.UUID `json:"storyIds"`
	HardDelete *bool       `json:"hardDelete,omitempty"`
}

// AppBulkRestoreRequest represents a request to restore multiple stories.
type AppBulkRestoreRequest struct {
	StoryIDs []uuid.UUID `json:"storyIds"`
}

// AppBulkUnarchiveRequest represents a request to unarchive multiple stories.
type AppBulkUnarchiveRequest struct {
	StoryIDs []uuid.UUID `json:"storyIds"`
}

// AppBulkArchiveRequest represents a request to archive multiple stories.
type AppBulkArchiveRequest struct {
	StoryIDs []uuid.UUID `json:"storyIds"`
}

// AppBulkUpdateRequest represents a request to update multiple stories.
type AppBulkUpdateRequest struct {
	StoryIDs []uuid.UUID    `json:"storyIds" validate:"required,min=1"`
	Updates  AppUpdateStory `json:"updates" validate:"required"`
}

type AppLabel struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	ProjectID   uuid.UUID  `json:"projectId"`
	TeamID      *uuid.UUID `json:"teamId"`
	WorkspaceID *uuid.UUID `json:"workspaceId"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// AppSingleStory represents a single story in the application.
type AppSingleStory struct {
	ID              uuid.UUID      `json:"id"`
	SequenceID      int            `json:"sequenceId"`
	Title           string         `json:"title"`
	Description     *string        `json:"description"`
	DescriptionHTML *string        `json:"descriptionHTML"`
	Parent          *uuid.UUID     `json:"parentId"`
	Status          *uuid.UUID     `json:"statusId"`
	Assignee        *uuid.UUID     `json:"assigneeId"`
	BlockedBy       *uuid.UUID     `json:"blockedById"`
	Blocking        *uuid.UUID     `json:"blockingId"`
	Related         *uuid.UUID     `json:"relatedId"`
	Reporter        *uuid.UUID     `json:"reporterId"`
	Priority        string         `json:"priority"`
	Sprint          *uuid.UUID     `json:"sprintId"`
	Epic            *uuid.UUID     `json:"epicId"`
	Objective       *uuid.UUID     `json:"objectiveId"`
	KeyResult       *uuid.UUID     `json:"keyResultId"`
	Team            uuid.UUID      `json:"teamId"`
	Workspace       uuid.UUID      `json:"workspaceId"`
	StartDate       *time.Time     `json:"startDate"`
	EndDate         *time.Time     `json:"endDate"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       *time.Time     `json:"deletedAt"`
	CompletedAt     *time.Time     `json:"completedAt"`
	SubStories      []AppStoryList `json:"subStories"`
	Labels          []uuid.UUID    `json:"labels"`
}

// AppStoryList represents a single story in the list of stories in the application.
type AppStoryList struct {
	ID          uuid.UUID      `json:"id"`
	SequenceID  int            `json:"sequenceId"`
	Title       string         `json:"title"`
	Objective   *uuid.UUID     `json:"objectiveId"`
	Status      *uuid.UUID     `json:"statusId"`
	Assignee    *uuid.UUID     `json:"assigneeId"`
	Reporter    *uuid.UUID     `json:"reporterId"`
	Priority    string         `json:"priority"`
	Sprint      *uuid.UUID     `json:"sprintId"`
	KeyResult   *uuid.UUID     `json:"keyResultId"`
	Workspace   uuid.UUID      `json:"workspaceId"`
	Team        uuid.UUID      `json:"teamId"`
	StartDate   *time.Time     `json:"startDate"`
	EndDate     *time.Time     `json:"endDate"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	CompletedAt *time.Time     `json:"completedAt"`
	Labels      []uuid.UUID    `json:"labels"`
	SubStories  []AppStoryList `json:"subStories"`
}

func toAppStory(i stories.CoreSingleStory) AppSingleStory {
	return AppSingleStory{
		ID:              i.ID,
		SequenceID:      i.SequenceID,
		Description:     i.Description,
		DescriptionHTML: i.DescriptionHTML,
		Parent:          i.Parent,
		Title:           i.Title,
		Objective:       i.Objective,
		Status:          i.Status,
		Assignee:        i.Assignee,
		Priority:        i.Priority,
		Sprint:          i.Sprint,
		Epic:            i.Epic,
		KeyResult:       i.KeyResult,
		Team:            i.Team,
		Workspace:       i.Workspace,
		StartDate:       i.StartDate,
		EndDate:         i.EndDate,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		DeletedAt:       i.DeletedAt,
		CompletedAt:     i.CompletedAt,
		BlockedBy:       i.BlockedBy,
		Blocking:        i.Blocking,
		Related:         i.Related,
		Reporter:        i.Reporter,
		SubStories:      toAppStories(i.SubStories),
		Labels:          i.Labels,
	}
}

func toAppStories(stories []stories.CoreStoryList) []AppStoryList {
	appStories := make([]AppStoryList, len(stories))
	for i, story := range stories {
		appStories[i] = AppStoryList{
			ID:          story.ID,
			SequenceID:  story.SequenceID,
			Title:       story.Title,
			Objective:   story.Objective,
			Team:        story.Team,
			Workspace:   story.Workspace,
			Status:      story.Status,
			Assignee:    story.Assignee,
			Reporter:    story.Reporter,
			Priority:    story.Priority,
			Sprint:      story.Sprint,
			KeyResult:   story.KeyResult,
			StartDate:   story.StartDate,
			EndDate:     story.EndDate,
			CreatedAt:   story.CreatedAt,
			UpdatedAt:   story.UpdatedAt,
			CompletedAt: story.CompletedAt,
			Labels:      story.Labels,
			SubStories:  toAppStories(story.SubStories),
		}
	}
	return appStories
}

// AppNewStory represents a new story in the application. Make all fields are optional and have both json and db tags.
type AppUpdateStory struct {
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description" db:"description"`
	DescriptionHTML string     `json:"descriptionHTML" db:"description_html"`
	Parent          uuid.UUID  `json:"parentId" db:"parent_id"`
	Objective       uuid.UUID  `json:"objectiveId" db:"objective_id"`
	Status          uuid.UUID  `json:"statusId" db:"status_id"`
	Assignee        *uuid.UUID `json:"assigneeId" db:"assignee_id"`
	Priority        string     `json:"priority" db:"priority" validate:"omitempty,oneof='No Priority' Low Medium High Urgent"`
	Sprint          uuid.UUID  `json:"sprintId" db:"sprint_id"`
	KeyResult       uuid.UUID  `json:"keyResultId" db:"key_result_id"`
	StartDate       *time.Time `json:"startDate" db:"start_date"`
	EndDate         *time.Time `json:"endDate" db:"end_date"`
	CompletedAt     *time.Time `json:"completedAt" db:"completed_at"`
}

type AppNewStory struct {
	Title           string     `json:"title" validate:"required"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Objective       *uuid.UUID `json:"objectiveId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	Priority        string     `json:"priority" validate:"oneof='No Priority' Low Medium High Urgent"`
	Sprint          *uuid.UUID `json:"sprintId"`
	KeyResult       *uuid.UUID `json:"keyResultId"`
	Team            uuid.UUID  `json:"teamId" validate:"required"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
}

type AppNewComment struct {
	Comment  string      `json:"comment" validate:"required"`
	Parent   *uuid.UUID  `json:"parentId"`
	Mentions []uuid.UUID `json:"mentions"`
}

type AppComment struct {
	ID          uuid.UUID    `json:"id"`
	StoryID     uuid.UUID    `json:"storyId"`
	Parent      *uuid.UUID   `json:"parentId"`
	UserID      uuid.UUID    `json:"userId"`
	Comment     string       `json:"comment"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	SubComments []AppComment `json:"subComments"`
}

// StoryFilters represents filtering options for stories at the handler level
type StoryFilters struct {
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
	// Date range filters
	CreatedAfter    *time.Time `json:"createdAfter"`
	CreatedBefore   *time.Time `json:"createdBefore"`
	UpdatedAfter    *time.Time `json:"updatedAfter"`
	UpdatedBefore   *time.Time `json:"updatedBefore"`
	DeadlineAfter   *time.Time `json:"deadlineAfter"`
	DeadlineBefore  *time.Time `json:"deadlineBefore"`
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

// AppFilters represents the filters for stories.
type AppFilters struct {
	Parent    *uuid.UUID `json:"parentId" db:"parent_id"`
	Objective *uuid.UUID `json:"objectiveId" db:"objective_id"`
	Status    *uuid.UUID `json:"statusId" db:"status_id"`
	Assignee  *uuid.UUID `json:"assigneeId" db:"assignee_id"`
	Priority  *string    `json:"priority" db:"priority" validate:"omitempty,oneof='No Priority' Low Medium High Urgent"`
	Sprint    *uuid.UUID `json:"sprintId" db:"sprint_id"`
	Team      *uuid.UUID `json:"teamId" db:"team_id"`
	Epic      *uuid.UUID `json:"epicId" db:"epic_id"`
	Reporter  *uuid.UUID `json:"reporterId" db:"reporter_id"`
	KeyResult *uuid.UUID `json:"keyResultId" db:"key_result_id"`
}

func toAppComment(i comments.CoreComment) AppComment {
	return AppComment{
		ID:          i.ID,
		StoryID:     i.StoryID,
		Parent:      i.Parent,
		UserID:      i.UserID,
		Comment:     i.Comment,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		SubComments: toAppComments(i.SubComments),
	}
}

func toAppComments(i []comments.CoreComment) []AppComment {
	appComments := make([]AppComment, len(i))
	for i, comment := range i {
		appComments[i] = toAppComment(comment)
	}
	return appComments
}

func toCoreNewStory(a AppNewStory, userID uuid.UUID) stories.CoreNewStory {
	return stories.CoreNewStory{
		Title:           a.Title,
		Description:     a.Description,
		DescriptionHTML: a.DescriptionHTML,
		Parent:          a.Parent,
		Objective:       a.Objective,
		Status:          a.Status,
		Assignee:        a.Assignee,
		Reporter:        &userID,
		Priority:        a.Priority,
		Sprint:          a.Sprint,
		KeyResult:       a.KeyResult,
		StartDate:       a.StartDate,
		EndDate:         a.EndDate,
		Team:            a.Team,
	}
}

func (a AppNewStory) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(a)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				fieldName := getJSONTagName(reflect.TypeOf(a), e.Field())
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldName))
				case "oneof":
					options := strings.Split(e.Param(), " ")
					formattedOptions := formatOptions(options)
					errorMessages = append(errorMessages, fmt.Sprintf("%s should be one of: %s", fieldName, formattedOptions))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation: %s", fieldName, e.Tag()))
				}
			}
			return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
		}
	}
	return err
}

func formatOptions(options []string) string {
	for i, option := range options {
		options[i] = "'" + option + "'"
	}
	return strings.Join(options, ", ")
}

func getJSONTagName(t reflect.Type, fieldName string) string {
	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return fieldName // Return original field name if no JSON tag
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return fieldName // Return original field name if JSON tag is "-"
	}

	return parts[0] // Return the JSON tag name
}
