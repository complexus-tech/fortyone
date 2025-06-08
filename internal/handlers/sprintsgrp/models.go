package sprintsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/google/uuid"
)

// AppSprint represents a single sprint without stats in the application layer.
type AppSprint struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Goal      *string    `json:"goal"`
	Objective *uuid.UUID `json:"objectiveId"`
	Team      uuid.UUID  `json:"teamId"`
	Workspace uuid.UUID  `json:"workspaceId"`
	StartDate time.Time  `json:"startDate"`
	EndDate   time.Time  `json:"endDate"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// AppSprintList represents a sprint in the application layer.
type AppSprintsList struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Goal      *string     `json:"goal"`
	Objective *uuid.UUID  `json:"objectiveId"`
	Team      uuid.UUID   `json:"teamId"`
	Workspace uuid.UUID   `json:"workspaceId"`
	StartDate time.Time   `json:"startDate"`
	EndDate   time.Time   `json:"endDate"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Stats     SprintStats `json:"stats"`
}

type SprintStats struct {
	Total     int `json:"total"`
	Cancelled int `json:"cancelled"`
	Completed int `json:"completed"`
	Started   int `json:"started"`
	Unstarted int `json:"unstarted"`
	Backlog   int `json:"backlog"`
}

type AppFilters struct {
	Objective *uuid.UUID `json:"objectiveId" db:"objective_id"`
	Team      *uuid.UUID `json:"teamId" db:"team_id"`
}

type AppNewSprint struct {
	Name      string     `json:"name" validate:"required"`
	Goal      *string    `json:"goal"`
	Objective *uuid.UUID `json:"objectiveId"`
	Team      uuid.UUID  `json:"teamId" validate:"required"`
	StartDate time.Time  `json:"startDate" validate:"required"`
	EndDate   time.Time  `json:"endDate" validate:"required"`
}

type AppUpdateSprint struct {
	Name      *string    `json:"name,omitempty"`
	Goal      *string    `json:"goal,omitempty"`
	Objective *uuid.UUID `json:"objectiveId,omitempty"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

// toAppSprints converts a list of core sprints to a list of application sprints.
func toAppSprints(sprints []sprints.CoreSprint) []AppSprintsList {
	appSprints := make([]AppSprintsList, len(sprints))
	for i, sprint := range sprints {
		appSprints[i] = AppSprintsList{
			ID:        sprint.ID,
			Name:      sprint.Name,
			Goal:      sprint.Goal,
			Objective: sprint.Objective,
			Team:      sprint.Team,
			Workspace: sprint.Workspace,
			StartDate: sprint.StartDate,
			EndDate:   sprint.EndDate,
			CreatedAt: sprint.CreatedAt,
			UpdatedAt: sprint.UpdatedAt,
			Stats: SprintStats{
				Total:     sprint.TotalStories,
				Cancelled: sprint.CancelledStories,
				Completed: sprint.CompletedStories,
				Started:   sprint.StartedStories,
				Unstarted: sprint.UnstartedStories,
				Backlog:   sprint.BacklogStories,
			},
		}
	}
	return appSprints
}

// toAppSprint converts a core sprint to a simple application sprint (without stats).
func toAppSprint(sprint sprints.CoreSprint) AppSprint {
	return AppSprint{
		ID:        sprint.ID,
		Name:      sprint.Name,
		Goal:      sprint.Goal,
		Objective: sprint.Objective,
		Team:      sprint.Team,
		Workspace: sprint.Workspace,
		StartDate: sprint.StartDate,
		EndDate:   sprint.EndDate,
		CreatedAt: sprint.CreatedAt,
		UpdatedAt: sprint.UpdatedAt,
	}
}

// Sprint Analytics Models

type AppSprintAnalytics struct {
	SprintID         uuid.UUID              `json:"sprintId"`
	Overview         SprintOverview         `json:"overview"`
	StoryBreakdown   StoryBreakdown         `json:"storyBreakdown"`
	Burndown         []BurndownDataPoint    `json:"burndown"`
	TeamAllocation   []TeamMemberAllocation `json:"teamAllocation"`
	HealthIndicators SprintHealthIndicators `json:"healthIndicators"`
}

type SprintOverview struct {
	CompletionPercentage int    `json:"completionPercentage"`
	DaysElapsed          int    `json:"daysElapsed"`
	DaysRemaining        int    `json:"daysRemaining"`
	Status               string `json:"status"` // "on_track", "at_risk", "behind"
}

type StoryBreakdown struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"inProgress"`
	Todo       int `json:"todo"`
	Blocked    int `json:"blocked"`
	Cancelled  int `json:"cancelled"`
}

type BurndownDataPoint struct {
	Date      time.Time `json:"date"`
	Remaining int       `json:"remaining"`
}

type TeamMemberAllocation struct {
	MemberID  uuid.UUID `json:"memberId"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatarUrl"`
	Assigned  int       `json:"assigned"`
	Completed int       `json:"completed"`
}

type SprintHealthIndicators struct {
	BlockedCount   int `json:"blockedCount"`
	OverdueCount   int `json:"overdueCount"`
	AddedMidSprint int `json:"addedMidSprint"`
}

// toAppSprintAnalytics converts core sprint analytics to app sprint analytics.
func toAppSprintAnalytics(analytics sprints.CoreSprintAnalytics) AppSprintAnalytics {
	return AppSprintAnalytics{
		SprintID: analytics.SprintID,
		Overview: SprintOverview{
			CompletionPercentage: analytics.Overview.CompletionPercentage,
			DaysElapsed:          analytics.Overview.DaysElapsed,
			DaysRemaining:        analytics.Overview.DaysRemaining,
			Status:               analytics.Overview.Status,
		},
		StoryBreakdown: StoryBreakdown{
			Total:      analytics.StoryBreakdown.Total,
			Completed:  analytics.StoryBreakdown.Completed,
			InProgress: analytics.StoryBreakdown.InProgress,
			Todo:       analytics.StoryBreakdown.Todo,
			Blocked:    analytics.StoryBreakdown.Blocked,
			Cancelled:  analytics.StoryBreakdown.Cancelled,
		},
		Burndown:       toAppBurndownData(analytics.Burndown),
		TeamAllocation: toAppTeamAllocation(analytics.TeamAllocation),
		HealthIndicators: SprintHealthIndicators{
			BlockedCount:   analytics.HealthIndicators.BlockedCount,
			OverdueCount:   analytics.HealthIndicators.OverdueCount,
			AddedMidSprint: analytics.HealthIndicators.AddedMidSprint,
		},
	}
}

func toAppBurndownData(burndown []sprints.CoreBurndownDataPoint) []BurndownDataPoint {
	result := make([]BurndownDataPoint, len(burndown))
	for i, point := range burndown {
		result[i] = BurndownDataPoint{
			Date:      point.Date,
			Remaining: point.Remaining,
		}
	}
	return result
}

func toAppTeamAllocation(allocation []sprints.CoreTeamMemberAllocation) []TeamMemberAllocation {
	result := make([]TeamMemberAllocation, len(allocation))
	for i, member := range allocation {
		result[i] = TeamMemberAllocation{
			MemberID:  member.MemberID,
			Username:  member.Username,
			AvatarURL: member.AvatarURL,
			Assigned:  member.Assigned,
			Completed: member.Completed,
		}
	}
	return result
}
