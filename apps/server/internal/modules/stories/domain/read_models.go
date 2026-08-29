// Package domain contains transport- and persistence-neutral story contracts.
package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("story not found")
	ErrInvalidReadScope = errors.New("invalid story read scope")
	ErrInvalidReadQuery = errors.New("invalid story read query")
)

// ReadScope carries caller identity and credential restrictions. Persistence
// still verifies live user, workspace, and team membership on every query.
type ReadScope struct {
	ActorID                uuid.UUID
	WorkspaceID            uuid.UUID
	UnrestrictedTeamAccess bool
	AllowedTeamIDs         []uuid.UUID
}

func (s ReadScope) Validate() error {
	if s.ActorID == uuid.Nil || s.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: actor and workspace are required", ErrInvalidReadScope)
	}
	if s.UnrestrictedTeamAccess && len(s.AllowedTeamIDs) > 0 {
		return fmt.Errorf("%w: team access cannot be both unrestricted and restricted", ErrInvalidReadScope)
	}
	if !s.UnrestrictedTeamAccess && len(s.AllowedTeamIDs) == 0 {
		return fmt.Errorf("%w: restricted actors require at least one team", ErrInvalidReadScope)
	}
	for _, teamID := range s.AllowedTeamIDs {
		if teamID == uuid.Nil {
			return fmt.Errorf("%w: team restrictions cannot contain a zero id", ErrInvalidReadScope)
		}
	}
	return nil
}

type TeamSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type ObjectiveSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
}

type SprintSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Goal      *string   `json:"goal"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type StoryList struct {
	ID                       uuid.UUID         `json:"id"`
	SequenceID               int               `json:"sequence_id"`
	Title                    string            `json:"title"`
	EstimateLabel            *string           `json:"estimate_label"`
	EstimateValue            *int16            `json:"estimate_value"`
	EstimateScheme           string            `json:"estimate_scheme"`
	EstimatedDurationMinutes *int              `json:"estimated_duration_minutes"`
	MinimumFocusBlockMinutes *int              `json:"minimum_focus_block_minutes"`
	AutoSchedulingEnabled    bool              `json:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool              `json:"auto_scheduling_locked"`
	AutoSchedulingStatus     string            `json:"auto_scheduling_status"`
	AutoSchedulingReason     *string           `json:"auto_scheduling_reason"`
	AutoSchedulingUpdatedAt  *time.Time        `json:"auto_scheduling_updated_at"`
	Parent                   *uuid.UUID        `json:"parent_id"`
	Objective                *uuid.UUID        `json:"objective_id"`
	ObjectiveSummary         *ObjectiveSummary `json:"objective"`
	Epic                     *uuid.UUID        `json:"epic_id"`
	Status                   *uuid.UUID        `json:"status_id"`
	Assignee                 *uuid.UUID        `json:"assignee_id"`
	CollaboratorCount        int               `json:"collaborator_count"`
	Reporter                 *uuid.UUID        `json:"reporter_id"`
	Priority                 string            `json:"priority"`
	Sprint                   *uuid.UUID        `json:"sprint_id"`
	SprintSummary            *SprintSummary    `json:"sprint"`
	KeyResult                *uuid.UUID        `json:"key_result_id"`
	Team                     uuid.UUID         `json:"team_id"`
	TeamSummary              *TeamSummary      `json:"team"`
	Workspace                uuid.UUID         `json:"workspace_id"`
	StartDate                *time.Time        `json:"start_date"`
	EndDate                  *time.Time        `json:"end_date"`
	CreatedAt                time.Time         `json:"created_at"`
	UpdatedAt                time.Time         `json:"updated_at"`
	CompletedAt              *time.Time        `json:"completed_at"`
	DeletedAt                *time.Time        `json:"deleted_at"`
	ArchivedAt               *time.Time        `json:"archived_at"`
	Labels                   []uuid.UUID       `json:"labels"`
	SubStories               []StoryList       `json:"subStories"`
}

type Story struct {
	ID                       uuid.UUID
	SequenceID               int
	Title                    string
	EstimateLabel            *string
	EstimateValue            *int16
	EstimateScheme           string
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	AutoSchedulingEnabled    bool
	AutoSchedulingLocked     bool
	AutoSchedulingStatus     string
	AutoSchedulingReason     *string
	AutoSchedulingUpdatedAt  *time.Time
	TeamCode                 string
	Description              *string
	DescriptionHTML          *string
	Parent                   *uuid.UUID
	Objective                *uuid.UUID
	Status                   *uuid.UUID
	Assignee                 *uuid.UUID
	Collaborators            []uuid.UUID
	WatcherIDs               []uuid.UUID
	WatcherCount             int
	IsWatching               bool
	WatchingReason           *string
	BlockedBy                *uuid.UUID
	Blocking                 *uuid.UUID
	Related                  *uuid.UUID
	Reporter                 *uuid.UUID
	Priority                 string
	Sprint                   *uuid.UUID
	SprintSummary            *SprintSummary
	Epic                     *uuid.UUID
	KeyResult                *uuid.UUID
	Team                     uuid.UUID
	Workspace                uuid.UUID
	StartDate                *time.Time
	EndDate                  *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                *time.Time
	ArchivedAt               *time.Time
	CompletedAt              *time.Time
	SubStories               []StoryList
	Labels                   []uuid.UUID
	Associations             []StoryAssociation
	CreationKey              *string
	CreatedNow               bool
}

type StoryAssociation struct {
	ID             uuid.UUID `json:"id"`
	FromStoryID    uuid.UUID `json:"fromStoryId"`
	ToStoryID      uuid.UUID `json:"toStoryId"`
	Type           string    `json:"type"`
	PreviousType   *string   `json:"previousType,omitempty"`
	Story          StoryList `json:"story"`
	FromStoryTitle string    `json:"-"`
	ToStoryTitle   string    `json:"-"`
}

// StoryFilters is the typed, transport-neutral filter contract shared by
// ordinary and grouped story reads. Empty slices and nil pointers mean that a
// filter is not applied.
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
	CurrentUserID          uuid.UUID   `json:"currentUserId"`
	WorkspaceID            uuid.UUID   `json:"workspaceId"`
	CreatedAfter           *time.Time  `json:"createdAfter"`
	CreatedBefore          *time.Time  `json:"createdBefore"`
	UpdatedAfter           *time.Time  `json:"updatedAfter"`
	UpdatedBefore          *time.Time  `json:"updatedBefore"`
	StartDateAfter         *time.Time  `json:"startDateAfter"`
	StartDateBefore        *time.Time  `json:"startDateBefore"`
	StartDateNot           *time.Time  `json:"startDateNot"`
	DeadlineAfter          *time.Time  `json:"deadlineAfter"`
	DeadlineBefore         *time.Time  `json:"deadlineBefore"`
	DeadlineNot            *time.Time  `json:"deadlineNot"`
	CompletedAfter         *time.Time  `json:"completedAfter"`
	CompletedBefore        *time.Time  `json:"completedBefore"`
	IsCompleted            *bool       `json:"isCompleted"`
	IsNotCompleted         *bool       `json:"isNotCompleted"`
	IncludeArchived        *bool       `json:"includeArchived"`
	IncludeDeleted         *bool       `json:"includeDeleted"`
	Limit                  int         `json:"limit"`
	Offset                 int         `json:"offset"`
}

type StoryGroupBy string

const (
	StoryGroupNone     StoryGroupBy = "none"
	StoryGroupStatus   StoryGroupBy = "status"
	StoryGroupAssignee StoryGroupBy = "assignee"
	StoryGroupPriority StoryGroupBy = "priority"
	StoryGroupTeam     StoryGroupBy = "team"
	StoryGroupSprint   StoryGroupBy = "sprint"
)

type StoryOrderBy string

const (
	StoryOrderCreated   StoryOrderBy = "created"
	StoryOrderUpdated   StoryOrderBy = "updated"
	StoryOrderPriority  StoryOrderBy = "priority"
	StoryOrderDeadline  StoryOrderBy = "deadline"
	StoryOrderCompleted StoryOrderBy = "completed"
)

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type StoryQuery struct {
	Filters         StoryFilters  `json:"filters"`
	GroupBy         StoryGroupBy  `json:"groupBy"`
	OrderBy         StoryOrderBy  `json:"orderBy"`
	OrderDirection  SortDirection `json:"orderDirection"`
	StoriesPerGroup int           `json:"storiesPerGroup"`
	GroupKey        string        `json:"groupKey"`
	Page            int           `json:"page"`
	PageSize        int           `json:"pageSize"`
}

type StoryGroup struct {
	Key         string      `json:"key"`
	LoadedCount int         `json:"loadedCount"`
	TotalCount  int         `json:"totalCount"`
	HasMore     bool        `json:"hasMore"`
	Stories     []StoryList `json:"stories"`
	NextPage    int         `json:"nextPage"`
}
