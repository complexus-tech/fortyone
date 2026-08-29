package stories

import (
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

var (
	ErrNotFound               = storydomain.ErrNotFound
	ErrInvalidStoryReadScope  = storydomain.ErrInvalidReadScope
	ErrInvalidStoryReadQuery  = storydomain.ErrInvalidReadQuery
	ErrInvalidStoryMutation   = storydomain.ErrInvalidMutation
	ErrStoryMutationForbidden = storydomain.ErrMutationForbidden
)

type StoryPatch = storydomain.StoryPatch
type Field[T any] = storydomain.Field[T]

func SetField[T any](value T) storydomain.Field[T] {
	return storydomain.SetField(value)
}

func ClearField[T any]() storydomain.Field[T] {
	return storydomain.ClearField[T]()
}

type StoryReadScope = storydomain.ReadScope

// BulkDeleteAuthorization carries the actor information required to enforce
// the same delete policy for every story in a bulk mutation.
type BulkDeleteAuthorization struct {
	ActorID uuid.UUID
	IsAdmin bool
}

// HardBulkDeleteResult separates the deleted-story receipt from retired media.
// AttachmentObjectDeletionDeferred is true when durable outbox delivery owns
// object removal; false preserves the legacy post-transaction cleanup path.
type HardBulkDeleteResult struct {
	StoryIDs                         []uuid.UUID
	OrphanedAttachmentIDs            []uuid.UUID
	AttachmentObjectDeletionDeferred bool
}

// BulkUpdateItemResult records the outcome for one requested story while
// preserving input order in a bulk-update receipt.
type BulkUpdateItemResult struct {
	StoryID uuid.UUID
	Success bool
	Error   string
}

// BulkUpdateResult describes a completed bulk-update attempt. Item failures
// are data, not a top-level error, because other stories may already be
// durably updated.
type BulkUpdateResult struct {
	TotalCount     int
	SucceededCount int
	FailedCount    int
	Partial        bool
	Items          []BulkUpdateItemResult
}

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

type CoreTeamSummary = storydomain.TeamSummary
type CoreObjectiveSummary = storydomain.ObjectiveSummary
type CoreSprintSummary = storydomain.SprintSummary
type CoreStoryList = storydomain.StoryList
type CoreSingleStory = storydomain.Story
type CoreStoryFilters = storydomain.StoryFilters
type CoreStoryQuery = storydomain.StoryQuery
type CoreStoryGroup = storydomain.StoryGroup
type StoryGroupBy = storydomain.StoryGroupBy
type StoryOrderBy = storydomain.StoryOrderBy
type SortDirection = storydomain.SortDirection

const (
	StoryGroupNone      = storydomain.StoryGroupNone
	StoryGroupStatus    = storydomain.StoryGroupStatus
	StoryGroupAssignee  = storydomain.StoryGroupAssignee
	StoryGroupPriority  = storydomain.StoryGroupPriority
	StoryGroupTeam      = storydomain.StoryGroupTeam
	StoryGroupSprint    = storydomain.StoryGroupSprint
	StoryOrderCreated   = storydomain.StoryOrderCreated
	StoryOrderUpdated   = storydomain.StoryOrderUpdated
	StoryOrderPriority  = storydomain.StoryOrderPriority
	StoryOrderDeadline  = storydomain.StoryOrderDeadline
	StoryOrderCompleted = storydomain.StoryOrderCompleted
	SortAscending       = storydomain.SortAscending
	SortDescending      = storydomain.SortDescending
)

type CoreNewStory = storydomain.NewStory

type CoreUpdateStory struct {
	Title                    *string
	EstimateValue            *int16
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	AutoSchedulingEnabled    *bool
	AutoSchedulingLocked     *bool
	Description              *string
	DescriptionHTML          *string
	Parent                   *uuid.UUID
	Objective                *uuid.UUID
	Status                   *uuid.UUID
	Assignee                 *uuid.UUID
	Priority                 *string
	Sprint                   *uuid.UUID
	KeyResult                *uuid.UUID
	StartDate                *time.Time
	EndDate                  *time.Time
	CompletedAt              *time.Time
}

func toCoreSingleStory(ns CoreNewStory, workspaceId uuid.UUID) CoreSingleStory {
	now := time.Now()
	if ns.Priority == "" {
		ns.Priority = "No Priority"
	}
	return CoreSingleStory{
		Workspace:                workspaceId,
		Title:                    ns.Title,
		EstimateValue:            ns.EstimateValue,
		EstimatedDurationMinutes: ns.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: ns.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    ns.AutoSchedulingEnabled,
		AutoSchedulingLocked:     ns.AutoSchedulingLocked,
		AutoSchedulingStatus:     AutoSchedulingStatusOff,
		Description:              ns.Description,
		DescriptionHTML:          ns.DescriptionHTML,
		Parent:                   ns.Parent,
		Objective:                ns.Objective,
		Status:                   ns.Status,
		Assignee:                 ns.Assignee,
		BlockedBy:                ns.BlockedBy,
		Blocking:                 ns.Blocking,
		Related:                  ns.Related,
		Reporter:                 ns.Reporter,
		Priority:                 ns.Priority,
		Sprint:                   ns.Sprint,
		KeyResult:                ns.KeyResult,
		Labels:                   append([]uuid.UUID(nil), ns.LabelIDs...),
		StartDate:                ns.StartDate,
		EndDate:                  ns.EndDate,
		Team:                     ns.Team,
		CreationKey:              ns.CreationKey,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
}

// CoreStoryAssociation represents a relationship between two stories.
type CoreStoryAssociation = storydomain.StoryAssociation

// CoreKeyResultReference contains the strategy details needed when linking a story.
type CoreKeyResultReference struct {
	ObjectiveID uuid.UUID
	Name        string
}

type CoreActivity = storydomain.Activity

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
	Reason       *string     `json:"reason,omitempty"`
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
	IsSystem  bool      `json:"isSystem"`
}

type CoreNewComment = storydomain.NewComment

type CoreComment = storydomain.Comment

type CreateCommentCommand = storydomain.CreateCommentCommand
