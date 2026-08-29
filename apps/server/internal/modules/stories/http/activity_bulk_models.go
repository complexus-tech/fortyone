package storieshttp

import (
	"fmt"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type AppUserSummary struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
	IsActive  bool      `json:"isActive"`
	IsSystem  bool      `json:"isSystem"`
}

type AppTeamSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type AppStoryMedia struct {
	ID         uuid.UUID `json:"id"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mimeType"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"createdAt"`
	UploadedBy uuid.UUID `json:"uploadedBy"`
}

type AppObjectiveSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
}

type AppSprintSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Goal      *string   `json:"goal"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// AppActivityWithUser represents an activity with embedded user details
type AppActivityWithUser struct {
	ID           uuid.UUID `json:"id"`
	StoryID      uuid.UUID `json:"storyId"`
	UserID       uuid.UUID `json:"userId"`
	Type         string    `json:"type"`
	Field        string    `json:"field"`
	CurrentValue string    `json:"currentValue"`
	OldValue     any       `json:"oldValue"`
	NewValue     any       `json:"newValue"`
	Reason       *string   `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`

	// Embedded user details
	User AppUserDetails `json:"user"`
}

// AppUserDetails represents basic user information for activities
type AppUserDetails struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
	IsActive  bool      `json:"isActive"`
	IsSystem  bool      `json:"isSystem"`
}

// AppNewLabels represents a new label in the application.
type AppNewLabels struct {
	Labels []uuid.UUID `json:"labels"`
}

type AppUpdateCollaborators struct {
	CollaboratorIDs []uuid.UUID `json:"collaboratorIds"`
}

type AppSetStoryWatching struct {
	Watching bool `json:"watching"`
}

func toAppActivityWithUser(i stories.CoreActivityWithUser) AppActivityWithUser {
	return AppActivityWithUser{
		ID:           i.ID,
		StoryID:      i.StoryID,
		UserID:       i.UserID,
		Type:         i.Type,
		Field:        i.Field,
		CurrentValue: i.CurrentValue,
		OldValue:     i.OldValue,
		NewValue:     i.NewValue,
		Reason:       i.Reason,
		CreatedAt:    i.CreatedAt,
		WorkspaceID:  i.WorkspaceID,
		User: AppUserDetails{
			ID:        i.User.ID,
			Username:  i.User.Username,
			FullName:  i.User.FullName,
			AvatarURL: i.User.AvatarURL,
			IsActive:  i.User.IsActive,
			IsSystem:  i.User.IsSystem,
		},
	}
}

// toAppActivitiesWithUser converts a list of core activities with user details to a list of application activities
func toAppActivitiesWithUser(activities []stories.CoreActivityWithUser) []AppActivityWithUser {
	appActivities := make([]AppActivityWithUser, len(activities))
	for i, activity := range activities {
		appActivities[i] = toAppActivityWithUser(activity)
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

// ActivitiesResponseWithUser represents the response for paginated activities with user details
type ActivitiesResponseWithUser struct {
	Activities []AppActivityWithUser `json:"activities"`
	Pagination ActivitiesPagination  `json:"pagination"`
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

func (a AppBulkDeleteRequest) Validate() error {
	if len(a.StoryIDs) == 0 {
		return fmt.Errorf("storyIds must contain at least one story")
	}
	if len(a.StoryIDs) > 50 {
		return fmt.Errorf("storyIds cannot contain more than 50 stories")
	}

	seen := make(map[uuid.UUID]struct{}, len(a.StoryIDs))
	for index, storyID := range a.StoryIDs {
		if storyID == uuid.Nil {
			return fmt.Errorf("storyIds[%d] must be a valid story ID", index)
		}
		if _, exists := seen[storyID]; exists {
			return fmt.Errorf("storyIds cannot contain duplicate story IDs")
		}
		seen[storyID] = struct{}{}
	}

	return nil
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

type AppBulkUpdateItemResult struct {
	StoryID uuid.UUID `json:"storyId"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
}

type AppBulkUpdateResult struct {
	TotalCount     int                       `json:"totalCount"`
	SucceededCount int                       `json:"succeededCount"`
	FailedCount    int                       `json:"failedCount"`
	Partial        bool                      `json:"partial"`
	Items          []AppBulkUpdateItemResult `json:"items"`
}

func toAppBulkUpdateResult(core stories.BulkUpdateResult) AppBulkUpdateResult {
	items := make([]AppBulkUpdateItemResult, 0, len(core.Items))
	for _, item := range core.Items {
		items = append(items, AppBulkUpdateItemResult{
			StoryID: item.StoryID,
			Success: item.Success,
			Error:   item.Error,
		})
	}

	return AppBulkUpdateResult{
		TotalCount:     core.TotalCount,
		SucceededCount: core.SucceededCount,
		FailedCount:    core.FailedCount,
		Partial:        core.Partial,
		Items:          items,
	}
}
