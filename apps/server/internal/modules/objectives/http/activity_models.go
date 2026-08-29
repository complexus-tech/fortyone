package objectiveshttp

import (
	"time"

	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/google/uuid"
)

type AppObjectiveActivity struct {
	ID           uuid.UUID      `json:"id"`
	ObjectiveID  uuid.UUID      `json:"objectiveId"`
	KeyResultID  *uuid.UUID     `json:"keyResultId"`
	UserID       uuid.UUID      `json:"userId"`
	Type         string         `json:"type"`
	UpdateType   string         `json:"updateType"`
	Field        string         `json:"field"`
	CurrentValue string         `json:"currentValue"`
	Comment      string         `json:"comment"`
	CreatedAt    time.Time      `json:"createdAt"`
	WorkspaceID  uuid.UUID      `json:"workspaceId"`
	User         AppUserDetails `json:"user"`
}

type AppUserDetails struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
	IsActive  bool      `json:"isActive"`
}

type AppActivityPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
}

type AppObjectiveActivitiesResponse struct {
	Activities []AppObjectiveActivity `json:"activities"`
	Pagination AppActivityPagination  `json:"pagination"`
}

func toAppObjectiveActivity(value okractivities.CoreActivity) AppObjectiveActivity {
	return AppObjectiveActivity{
		ID: value.ID, ObjectiveID: value.ObjectiveID, KeyResultID: value.KeyResultID,
		UserID: value.UserID, Type: string(value.Type), UpdateType: string(value.UpdateType),
		Field: value.Field, CurrentValue: value.CurrentValue, Comment: value.Comment,
		CreatedAt: value.CreatedAt, WorkspaceID: value.WorkspaceID,
		User: AppUserDetails{
			ID: value.User.ID, Username: value.User.Username, FullName: value.User.FullName,
			AvatarURL: value.User.AvatarURL, IsActive: value.User.IsActive,
		},
	}
}

func toAppObjectiveActivities(values []okractivities.CoreActivity) []AppObjectiveActivity {
	result := make([]AppObjectiveActivity, len(values))
	for index, value := range values {
		result[index] = toAppObjectiveActivity(value)
	}
	return result
}
