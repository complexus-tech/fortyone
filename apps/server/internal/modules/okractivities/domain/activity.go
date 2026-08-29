package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultPageSize  = 20
	MaximumPageSize  = 100
	MaximumBatchSize = 100
)

var (
	ErrInvalid       = errors.New("invalid OKR activity")
	ErrForbidden     = errors.New("OKR activity access is forbidden")
	ErrScopeMismatch = errors.New("OKR activity resource scope does not match")
)

type UpdateType string

const (
	UpdateTypeObjective UpdateType = "objective"
	UpdateTypeKeyResult UpdateType = "key_result"
)

func (updateType UpdateType) valid() bool {
	return updateType == UpdateTypeObjective || updateType == UpdateTypeKeyResult
}

type ActivityType string

const (
	ActivityTypeCreate ActivityType = "create"
	ActivityTypeUpdate ActivityType = "update"
	ActivityTypeDelete ActivityType = "delete"
)

func (activityType ActivityType) valid() bool {
	return activityType == ActivityTypeCreate || activityType == ActivityTypeUpdate || activityType == ActivityTypeDelete
}

type Activity struct {
	ID           uuid.UUID
	ObjectiveID  uuid.UUID
	KeyResultID  *uuid.UUID
	UserID       uuid.UUID
	Type         ActivityType
	UpdateType   UpdateType
	Field        string
	CurrentValue string
	Comment      string
	CreatedAt    time.Time
	WorkspaceID  uuid.UUID
	User         UserDetails
}

type UserDetails struct {
	ID        uuid.UUID
	Username  string
	FullName  string
	AvatarURL string
	IsActive  bool
}

type NewActivity struct {
	ObjectiveID  uuid.UUID
	KeyResultID  *uuid.UUID
	UserID       uuid.UUID
	Type         ActivityType
	UpdateType   UpdateType
	Field        string
	CurrentValue string
	Comment      string
	WorkspaceID  uuid.UUID
}

func (activity NewActivity) Normalize() (NewActivity, error) {
	if activity.ObjectiveID == uuid.Nil || activity.UserID == uuid.Nil || activity.WorkspaceID == uuid.Nil {
		return NewActivity{}, fmt.Errorf("%w: objective, actor, and workspace are required", ErrInvalid)
	}
	if activity.KeyResultID != nil && *activity.KeyResultID == uuid.Nil {
		return NewActivity{}, fmt.Errorf("%w: key result cannot be a zero id", ErrInvalid)
	}
	if !activity.Type.valid() || !activity.UpdateType.valid() {
		return NewActivity{}, fmt.Errorf("%w: unsupported activity type", ErrInvalid)
	}
	if activity.UpdateType == UpdateTypeKeyResult && activity.Type != ActivityTypeDelete && activity.KeyResultID == nil {
		return NewActivity{}, fmt.Errorf("%w: key result activity requires a key result", ErrInvalid)
	}
	activity.Field = strings.TrimSpace(activity.Field)
	activity.Comment = strings.TrimSpace(activity.Comment)
	if len([]rune(activity.Field)) > 100 || len([]rune(activity.Comment)) > 10_000 || len([]rune(activity.CurrentValue)) > 100_000 {
		return NewActivity{}, fmt.Errorf("%w: activity text exceeds its limit", ErrInvalid)
	}
	return activity, nil
}

type CreateBatchCommand struct {
	Activities []NewActivity
}

func (command CreateBatchCommand) Normalize() (CreateBatchCommand, error) {
	if len(command.Activities) == 0 {
		return command, nil
	}
	if len(command.Activities) > MaximumBatchSize {
		return CreateBatchCommand{}, fmt.Errorf("%w: activity batch cannot exceed %d", ErrInvalid, MaximumBatchSize)
	}
	for index := range command.Activities {
		normalized, err := command.Activities[index].Normalize()
		if err != nil {
			return CreateBatchCommand{}, fmt.Errorf("activity %d: %w", index, err)
		}
		command.Activities[index] = normalized
	}
	return command, nil
}

type ListQuery struct {
	ObjectiveID uuid.UUID
	KeyResultID *uuid.UUID
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Page        int
	PageSize    int
}

func (query ListQuery) Normalize() (ListQuery, error) {
	if query.WorkspaceID == uuid.Nil || query.ActorID == uuid.Nil {
		return ListQuery{}, fmt.Errorf("%w: workspace and actor are required", ErrInvalid)
	}
	if query.KeyResultID != nil && *query.KeyResultID == uuid.Nil {
		return ListQuery{}, fmt.Errorf("%w: key result cannot be a zero id", ErrInvalid)
	}
	if query.KeyResultID == nil && query.ObjectiveID == uuid.Nil {
		return ListQuery{}, fmt.Errorf("%w: objective is required", ErrInvalid)
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = DefaultPageSize
	}
	if query.PageSize > MaximumPageSize {
		query.PageSize = MaximumPageSize
	}
	return query, nil
}
