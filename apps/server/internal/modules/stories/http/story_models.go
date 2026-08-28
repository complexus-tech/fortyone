package storieshttp

import (
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

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
	ID                       uuid.UUID             `json:"id"`
	SequenceID               int                   `json:"sequenceId"`
	Title                    string                `json:"title"`
	EstimateLabel            *string               `json:"estimateLabel"`
	EstimateValue            *int16                `json:"estimateValue"`
	EstimateScheme           string                `json:"estimateScheme"`
	EstimatedDurationMinutes *int                  `json:"estimatedDurationMinutes"`
	MinimumFocusBlockMinutes *int                  `json:"minimumFocusBlockMinutes"`
	AutoSchedulingEnabled    bool                  `json:"autoSchedulingEnabled"`
	AutoSchedulingLocked     bool                  `json:"autoSchedulingLocked"`
	AutoSchedulingStatus     string                `json:"autoSchedulingStatus"`
	AutoSchedulingReason     *string               `json:"autoSchedulingReason"`
	AutoSchedulingUpdatedAt  *time.Time            `json:"autoSchedulingUpdatedAt"`
	TeamCode                 string                `json:"teamCode"`
	Description              *string               `json:"description"`
	DescriptionHTML          *string               `json:"descriptionHTML"`
	Parent                   *uuid.UUID            `json:"parentId"`
	Status                   *uuid.UUID            `json:"statusId"`
	AssigneeID               *uuid.UUID            `json:"assigneeId"`
	Assignee                 *AppUserSummary       `json:"assignee"`
	CollaboratorIDs          []uuid.UUID           `json:"collaboratorIds"`
	Collaborators            []AppUserSummary      `json:"collaborators"`
	CollaboratorCount        int                   `json:"collaboratorCount"`
	WatcherCount             int                   `json:"watcherCount"`
	Watchers                 []AppUserSummary      `json:"watchers"`
	IsWatching               bool                  `json:"isWatching"`
	WatchingReason           *string               `json:"watchingReason"`
	BlockedBy                *uuid.UUID            `json:"blockedById"`
	Blocking                 *uuid.UUID            `json:"blockingId"`
	Related                  *uuid.UUID            `json:"relatedId"`
	ReporterID               *uuid.UUID            `json:"reporterId"`
	Reporter                 *AppUserSummary       `json:"reporter"`
	Priority                 string                `json:"priority"`
	Sprint                   *uuid.UUID            `json:"sprintId"`
	Epic                     *uuid.UUID            `json:"epicId"`
	Objective                *uuid.UUID            `json:"objectiveId"`
	KeyResult                *uuid.UUID            `json:"keyResultId"`
	Team                     uuid.UUID             `json:"teamId"`
	Workspace                uuid.UUID             `json:"workspaceId"`
	StartDate                *time.Time            `json:"startDate"`
	EndDate                  *time.Time            `json:"endDate"`
	CreatedAt                time.Time             `json:"createdAt"`
	UpdatedAt                time.Time             `json:"updatedAt"`
	DeletedAt                *time.Time            `json:"deletedAt"`
	ArchivedAt               *time.Time            `json:"archivedAt"`
	CompletedAt              *time.Time            `json:"completedAt"`
	SubStories               []AppStoryList        `json:"subStories"`
	Labels                   []uuid.UUID           `json:"labels"`
	Associations             []AppStoryAssociation `json:"associations"`
}

type AppStoryAssociation struct {
	ID          uuid.UUID    `json:"id"`
	FromStoryID uuid.UUID    `json:"fromStoryId"`
	ToStoryID   uuid.UUID    `json:"toStoryId"`
	Type        string       `json:"type"` // "blocking", "related", "duplicate"
	Story       AppStoryList `json:"story"`
}

// AppStoryList represents a single story in the list of stories in the application.
type AppStoryList struct {
	ID                       uuid.UUID            `json:"id"`
	SequenceID               int                  `json:"sequenceId"`
	Title                    string               `json:"title"`
	EstimateLabel            *string              `json:"estimateLabel"`
	EstimateValue            *int16               `json:"estimateValue"`
	EstimateScheme           string               `json:"estimateScheme"`
	EstimatedDurationMinutes *int                 `json:"estimatedDurationMinutes"`
	MinimumFocusBlockMinutes *int                 `json:"minimumFocusBlockMinutes"`
	AutoSchedulingEnabled    bool                 `json:"autoSchedulingEnabled"`
	AutoSchedulingLocked     bool                 `json:"autoSchedulingLocked"`
	AutoSchedulingStatus     string               `json:"autoSchedulingStatus"`
	AutoSchedulingReason     *string              `json:"autoSchedulingReason"`
	AutoSchedulingUpdatedAt  *time.Time           `json:"autoSchedulingUpdatedAt"`
	Objective                *uuid.UUID           `json:"objectiveId"`
	ObjectiveSummary         *AppObjectiveSummary `json:"objective"`
	Status                   *uuid.UUID           `json:"statusId"`
	AssigneeID               *uuid.UUID           `json:"assigneeId"`
	Assignee                 *AppUserSummary      `json:"assignee"`
	CollaboratorCount        int                  `json:"collaboratorCount"`
	ReporterID               *uuid.UUID           `json:"reporterId"`
	Reporter                 *AppUserSummary      `json:"reporter"`
	Priority                 string               `json:"priority"`
	Sprint                   *uuid.UUID           `json:"sprintId"`
	SprintSummary            *AppSprintSummary    `json:"sprint"`
	KeyResult                *uuid.UUID           `json:"keyResultId"`
	Workspace                uuid.UUID            `json:"workspaceId"`
	Team                     uuid.UUID            `json:"teamId"`
	TeamSummary              *AppTeamSummary      `json:"team"`
	StartDate                *time.Time           `json:"startDate"`
	EndDate                  *time.Time           `json:"endDate"`
	CreatedAt                time.Time            `json:"createdAt"`
	UpdatedAt                time.Time            `json:"updatedAt"`
	CompletedAt              *time.Time           `json:"completedAt"`
	DeletedAt                *time.Time           `json:"deletedAt"`
	ArchivedAt               *time.Time           `json:"archivedAt"`
	Labels                   []uuid.UUID          `json:"labels"`
	SubStories               []AppStoryList       `json:"subStories"`
}

func toAppUserSummary(user users.CoreUser) AppUserSummary {
	return AppUserSummary{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		IsSystem:  user.IsSystem,
	}
}

func toAppStory(i stories.CoreSingleStory, usersByID map[uuid.UUID]AppUserSummary) AppSingleStory {
	return AppSingleStory{
		ID:                       i.ID,
		SequenceID:               i.SequenceID,
		EstimateLabel:            i.EstimateLabel,
		EstimateValue:            i.EstimateValue,
		EstimateScheme:           i.EstimateScheme,
		EstimatedDurationMinutes: i.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: i.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    i.AutoSchedulingEnabled,
		AutoSchedulingLocked:     i.AutoSchedulingLocked,
		AutoSchedulingStatus:     i.AutoSchedulingStatus,
		AutoSchedulingReason:     i.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  i.AutoSchedulingUpdatedAt,
		TeamCode:                 i.TeamCode,
		Description:              i.Description,
		DescriptionHTML:          i.DescriptionHTML,
		Parent:                   i.Parent,
		Title:                    i.Title,
		Objective:                i.Objective,
		Status:                   i.Status,
		AssigneeID:               i.Assignee,
		Assignee:                 findAppUserSummary(usersByID, i.Assignee),
		CollaboratorIDs:          i.Collaborators,
		Collaborators:            findAppUserSummaries(usersByID, i.Collaborators),
		CollaboratorCount:        len(i.Collaborators),
		WatcherCount:             i.WatcherCount,
		Watchers:                 findAppUserSummaries(usersByID, i.WatcherIDs),
		IsWatching:               i.IsWatching,
		WatchingReason:           i.WatchingReason,
		Priority:                 i.Priority,
		Sprint:                   i.Sprint,
		Epic:                     i.Epic,
		KeyResult:                i.KeyResult,
		Team:                     i.Team,
		Workspace:                i.Workspace,
		StartDate:                i.StartDate,
		EndDate:                  i.EndDate,
		CreatedAt:                i.CreatedAt,
		UpdatedAt:                i.UpdatedAt,
		DeletedAt:                i.DeletedAt,
		ArchivedAt:               i.ArchivedAt,
		CompletedAt:              i.CompletedAt,
		BlockedBy:                i.BlockedBy,
		Blocking:                 i.Blocking,
		Related:                  i.Related,
		ReporterID:               i.Reporter,
		Reporter:                 findAppUserSummary(usersByID, i.Reporter),
		SubStories:               toAppStories(i.SubStories, usersByID),
		Labels:                   i.Labels,
		Associations:             toAppStoryAssociations(i.Associations, usersByID),
	}
}

func toAppStoryAssociations(associations []stories.CoreStoryAssociation, usersByID map[uuid.UUID]AppUserSummary) []AppStoryAssociation {
	appAssociations := make([]AppStoryAssociation, len(associations))
	for i, association := range associations {
		appAssociations[i] = AppStoryAssociation{
			ID:          association.ID,
			FromStoryID: association.FromStoryID,
			ToStoryID:   association.ToStoryID,
			Type:        association.Type,
			Story:       toAppStoryListItem(association.Story, usersByID),
		}
	}
	return appAssociations
}

func toAppStories(stories []stories.CoreStoryList, usersByID map[uuid.UUID]AppUserSummary) []AppStoryList {
	appStories := make([]AppStoryList, len(stories))

	for i, story := range stories {
		appStories[i] = toAppStoryListItem(story, usersByID)
	}
	return appStories
}

func toAppStoryListItem(story stories.CoreStoryList, usersByID map[uuid.UUID]AppUserSummary) AppStoryList {
	return AppStoryList{
		ID:                       story.ID,
		SequenceID:               story.SequenceID,
		Title:                    story.Title,
		EstimateLabel:            story.EstimateLabel,
		EstimateValue:            story.EstimateValue,
		EstimateScheme:           story.EstimateScheme,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
		AutoSchedulingLocked:     story.AutoSchedulingLocked,
		AutoSchedulingStatus:     story.AutoSchedulingStatus,
		AutoSchedulingReason:     story.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
		Objective:                story.Objective,
		ObjectiveSummary:         toAppObjectiveSummary(story.ObjectiveSummary),
		Team:                     story.Team,
		TeamSummary:              toAppTeamSummary(story.TeamSummary),
		Workspace:                story.Workspace,
		Status:                   story.Status,
		AssigneeID:               story.Assignee,
		Assignee:                 findAppUserSummary(usersByID, story.Assignee),
		CollaboratorCount:        story.CollaboratorCount,
		ReporterID:               story.Reporter,
		Reporter:                 findAppUserSummary(usersByID, story.Reporter),
		Priority:                 story.Priority,
		Sprint:                   story.Sprint,
		SprintSummary:            toAppSprintSummary(story.SprintSummary),
		KeyResult:                story.KeyResult,
		StartDate:                story.StartDate,
		EndDate:                  story.EndDate,
		CreatedAt:                story.CreatedAt,
		UpdatedAt:                story.UpdatedAt,
		CompletedAt:              story.CompletedAt,
		DeletedAt:                story.DeletedAt,
		ArchivedAt:               story.ArchivedAt,
		Labels:                   story.Labels,
		SubStories:               toAppStories(story.SubStories, usersByID),
	}
}

func toAppTeamSummary(team *stories.CoreTeamSummary) *AppTeamSummary {
	if team == nil {
		return nil
	}

	return &AppTeamSummary{
		ID:   team.ID,
		Name: team.Name,
		Code: team.Code,
	}
}

func toAppObjectiveSummary(objective *stories.CoreObjectiveSummary) *AppObjectiveSummary {
	if objective == nil {
		return nil
	}

	return &AppObjectiveSummary{
		ID:          objective.ID,
		Name:        objective.Name,
		Description: objective.Description,
	}
}

func toAppSprintSummary(sprint *stories.CoreSprintSummary) *AppSprintSummary {
	if sprint == nil {
		return nil
	}

	return &AppSprintSummary{
		ID:        sprint.ID,
		Name:      sprint.Name,
		Goal:      sprint.Goal,
		StartDate: sprint.StartDate,
		EndDate:   sprint.EndDate,
	}
}

func findAppUserSummary(usersByID map[uuid.UUID]AppUserSummary, userID *uuid.UUID) *AppUserSummary {
	if userID == nil {
		return nil
	}

	user, ok := usersByID[*userID]
	if !ok {
		return nil
	}

	userCopy := user
	return &userCopy
}

func findAppUserSummaries(usersByID map[uuid.UUID]AppUserSummary, userIDs []uuid.UUID) []AppUserSummary {
	summaries := make([]AppUserSummary, 0, len(userIDs))
	for _, userID := range userIDs {
		if user, ok := usersByID[userID]; ok {
			summaries = append(summaries, user)
		}
	}
	return summaries
}
