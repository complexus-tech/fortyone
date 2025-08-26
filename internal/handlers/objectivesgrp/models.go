package objectivesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/google/uuid"
)

// AppObjectiveList represents a list of objectives in the application.
type AppObjectiveList struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	LeadUser    *uuid.UUID     `json:"leadUser"`
	Team        uuid.UUID      `json:"teamId"`
	Workspace   uuid.UUID      `json:"workspaceId"`
	StartDate   *time.Time     `json:"startDate"`
	EndDate     *time.Time     `json:"endDate"`
	IsPrivate   bool           `json:"isPrivate"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Status      uuid.UUID      `json:"statusId"`
	Priority    *string        `json:"priority"`
	Health      *string        `json:"health"`
	CreatedBy   uuid.UUID      `json:"createdBy"`
	Stats       ObjectiveStats `json:"stats"`
}

type ObjectiveStats struct {
	Total     int `json:"total"`
	Cancelled int `json:"cancelled"`
	Completed int `json:"completed"`
	Started   int `json:"started"`
	Unstarted int `json:"unstarted"`
	Backlog   int `json:"backlog"`
}

type AppFilters struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// toAppObjectives converts a list of core objectives to a list of application objectives.
func toAppObjectives(objectives []objectives.CoreObjective) []AppObjectiveList {
	appObjectives := make([]AppObjectiveList, len(objectives))
	for i, objective := range objectives {
		var healthStr *string
		if objective.Health != nil {
			h := string(*objective.Health)
			healthStr = &h
		}

		appObjectives[i] = AppObjectiveList{
			ID:          objective.ID,
			Name:        objective.Name,
			Description: objective.Description,
			LeadUser:    objective.LeadUser,
			Team:        objective.Team,
			Workspace:   objective.Workspace,
			StartDate:   objective.StartDate,
			EndDate:     objective.EndDate,
			IsPrivate:   objective.IsPrivate,
			CreatedAt:   objective.CreatedAt,
			UpdatedAt:   objective.UpdatedAt,
			Status:      objective.Status,
			Priority:    objective.Priority,
			CreatedBy:   objective.CreatedBy,
			Health:      healthStr,
			Stats: ObjectiveStats{
				Total:     objective.TotalStories,
				Cancelled: objective.CancelledStories,
				Completed: objective.CompletedStories,
				Started:   objective.StartedStories,
				Unstarted: objective.UnstartedStories,
				Backlog:   objective.BacklogStories,
			},
		}
	}
	return appObjectives
}

// AppNewObjective represents the data needed to create a new objective
type AppNewObjective struct {
	Name        string            `json:"name" validate:"required"`
	Description *string           `json:"description"`
	LeadUser    *uuid.UUID        `json:"leadUser"`
	Team        uuid.UUID         `json:"teamId" validate:"required"`
	StartDate   *time.Time        `json:"startDate"`
	EndDate     *time.Time        `json:"endDate"`
	IsPrivate   bool              `json:"isPrivate"`
	Status      uuid.UUID         `json:"statusId"`
	Priority    *string           `json:"priority"`
	KeyResults  []AppNewKeyResult `json:"keyResults,omitempty"`
}

// AppNewKeyResult represents a new key result to be created
type AppNewKeyResult struct {
	Name            string      `json:"name" validate:"required"`
	MeasurementType string      `json:"measurementType" validate:"required,oneof=percentage number boolean"`
	StartValue      float64     `json:"startValue"`
	CurrentValue    float64     `json:"currentValue"`
	TargetValue     float64     `json:"targetValue"`
	Lead            *uuid.UUID  `json:"lead"`
	Contributors    []uuid.UUID `json:"contributors"`
	StartDate       *time.Time  `json:"startDate"`
	EndDate         *time.Time  `json:"endDate"`
	CreatedBy       uuid.UUID   `json:"createdBy"`
}

// toCoreNewObjective converts an AppNewObjective to a CoreNewObjective
func toCoreNewObjective(ano AppNewObjective, createdBy uuid.UUID) objectives.CoreNewObjective {
	return objectives.CoreNewObjective{
		Name:        ano.Name,
		Description: ano.Description,
		LeadUser:    ano.LeadUser,
		Team:        ano.Team,
		StartDate:   ano.StartDate,
		EndDate:     ano.EndDate,
		IsPrivate:   ano.IsPrivate,
		Status:      ano.Status,
		Priority:    ano.Priority,
		CreatedBy:   createdBy,
	}
}

// toAppObjective converts a single core objective to an application objective.
func toAppObjective(objective objectives.CoreObjective) AppObjectiveList {
	var healthStr *string
	if objective.Health != nil {
		h := string(*objective.Health)
		healthStr = &h
	}

	return AppObjectiveList{
		ID:          objective.ID,
		Name:        objective.Name,
		Description: objective.Description,
		LeadUser:    objective.LeadUser,
		Team:        objective.Team,
		Workspace:   objective.Workspace,
		StartDate:   objective.StartDate,
		EndDate:     objective.EndDate,
		IsPrivate:   objective.IsPrivate,
		CreatedAt:   objective.CreatedAt,
		UpdatedAt:   objective.UpdatedAt,
		Status:      objective.Status,
		Priority:    objective.Priority,
		CreatedBy:   objective.CreatedBy,
		Health:      healthStr,
		Stats: ObjectiveStats{
			Total:     objective.TotalStories,
			Cancelled: objective.CancelledStories,
			Completed: objective.CompletedStories,
			Started:   objective.StartedStories,
			Unstarted: objective.UnstartedStories,
			Backlog:   objective.BacklogStories,
		},
	}
}

// AppUpdateObjective represents the data needed to update an objective
type AppUpdateObjective struct {
	Name        *string    `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	LeadUser    *uuid.UUID `json:"leadUser" db:"lead_user_id"`
	StartDate   *time.Time `json:"startDate" db:"start_date"`
	EndDate     *time.Time `json:"endDate" db:"end_date"`
	IsPrivate   *bool      `json:"isPrivate" db:"is_private"`
	Status      *uuid.UUID `json:"statusId" db:"status_id"`
	Priority    *string    `json:"priority" db:"priority"`
	Health      *string    `json:"health" db:"health"`
}

// AppKeyResult represents a key result in the application
type AppKeyResult struct {
	ID              uuid.UUID   `json:"id"`
	ObjectiveID     uuid.UUID   `json:"objectiveId"`
	Name            string      `json:"name"`
	MeasurementType string      `json:"measurementType"`
	StartValue      float64     `json:"startValue"`
	CurrentValue    float64     `json:"currentValue"`
	TargetValue     float64     `json:"targetValue"`
	Lead            *uuid.UUID  `json:"lead"`
	Contributors    []uuid.UUID `json:"contributors"`
	StartDate       *time.Time  `json:"startDate"`
	EndDate         *time.Time  `json:"endDate"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
	CreatedBy       uuid.UUID   `json:"createdBy"`
}

func toAppKeyResult(kr keyresults.CoreKeyResult) AppKeyResult {
	return AppKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		CurrentValue:    kr.CurrentValue,
		TargetValue:     kr.TargetValue,
		Lead:            kr.Lead,
		Contributors:    kr.Contributors,
		StartDate:       kr.StartDate,
		EndDate:         kr.EndDate,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
		CreatedBy:       kr.CreatedBy,
	}
}

func toAppKeyResults(krs []keyresults.CoreKeyResult) []AppKeyResult {
	result := make([]AppKeyResult, len(krs))
	for i, kr := range krs {
		result[i] = toAppKeyResult(kr)
	}
	return result
}

// Objective Analytics Models

type AppObjectiveAnalytics struct {
	ObjectiveID       uuid.UUID                       `json:"objectiveId"`
	PriorityBreakdown []AppPriorityBreakdown          `json:"priorityBreakdown"`
	ProgressBreakdown AppProgressBreakdown            `json:"progressBreakdown"`
	TeamAllocation    []AppTeamMemberAllocation       `json:"teamAllocation"`
	ProgressChart     []AppObjectiveProgressDataPoint `json:"progressChart"`
}

type AppPriorityBreakdown struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

type AppProgressBreakdown struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"inProgress"`
	Todo       int `json:"todo"`
	Blocked    int `json:"blocked"`
	Cancelled  int `json:"cancelled"`
}

type AppTeamMemberAllocation struct {
	MemberID  uuid.UUID `json:"memberId"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatarUrl"`
	Assigned  int       `json:"assigned"`
	Completed int       `json:"completed"`
}

type AppObjectiveProgressDataPoint struct {
	Date       time.Time `json:"date"`
	Completed  int       `json:"completed"`
	InProgress int       `json:"inProgress"`
	Total      int       `json:"total"`
}

// Conversion functions

func toAppObjectiveAnalytics(analytics objectives.CoreObjectiveAnalytics) AppObjectiveAnalytics {
	return AppObjectiveAnalytics{
		ObjectiveID:       analytics.ObjectiveID,
		PriorityBreakdown: toAppPriorityBreakdown(analytics.PriorityBreakdown),
		ProgressBreakdown: toAppProgressBreakdown(analytics.ProgressBreakdown),
		TeamAllocation:    toAppTeamAllocation(analytics.TeamAllocation),
		ProgressChart:     toAppProgressChart(analytics.ProgressChart),
	}
}

func toAppPriorityBreakdown(breakdown []objectives.CorePriorityBreakdown) []AppPriorityBreakdown {
	result := make([]AppPriorityBreakdown, len(breakdown))
	for i, b := range breakdown {
		result[i] = AppPriorityBreakdown{
			Priority: b.Priority,
			Count:    b.Count,
		}
	}
	return result
}

func toAppProgressBreakdown(breakdown objectives.CoreProgressBreakdown) AppProgressBreakdown {
	return AppProgressBreakdown{
		Total:      breakdown.Total,
		Completed:  breakdown.Completed,
		InProgress: breakdown.InProgress,
		Todo:       breakdown.Todo,
		Blocked:    breakdown.Blocked,
		Cancelled:  breakdown.Cancelled,
	}
}

func toAppTeamAllocation(allocation []objectives.CoreTeamMemberAllocation) []AppTeamMemberAllocation {
	result := make([]AppTeamMemberAllocation, len(allocation))
	for i, a := range allocation {
		result[i] = AppTeamMemberAllocation{
			MemberID:  a.MemberID,
			Username:  a.Username,
			AvatarURL: a.AvatarURL,
			Assigned:  a.Assigned,
			Completed: a.Completed,
		}
	}
	return result
}

func toAppProgressChart(progressChart []objectives.CoreObjectiveProgressDataPoint) []AppObjectiveProgressDataPoint {
	result := make([]AppObjectiveProgressDataPoint, len(progressChart))
	for i, p := range progressChart {
		result[i] = AppObjectiveProgressDataPoint{
			Date:       p.Date,
			Completed:  p.Completed,
			InProgress: p.InProgress,
			Total:      p.Total,
		}
	}
	return result
}

// AppObjectiveActivity represents an objective activity in the application
type AppObjectiveActivity struct {
	ID           uuid.UUID  `json:"id"`
	ObjectiveID  uuid.UUID  `json:"objectiveId"`
	KeyResultID  *uuid.UUID `json:"keyResultId"`
	UserID       uuid.UUID  `json:"userId"`
	Type         string     `json:"type"`
	UpdateType   string     `json:"updateType"`
	Field        string     `json:"field"`
	CurrentValue string     `json:"currentValue"`
	Comment      string     `json:"comment"`
	CreatedAt    time.Time  `json:"createdAt"`
	WorkspaceID  uuid.UUID  `json:"workspaceId"`
}

// toAppObjectiveActivity converts a CoreActivity to an AppObjectiveActivity
func toAppObjectiveActivity(a okractivities.CoreActivity) AppObjectiveActivity {
	return AppObjectiveActivity{
		ID:           a.ID,
		ObjectiveID:  a.ObjectiveID,
		KeyResultID:  a.KeyResultID,
		UserID:       a.UserID,
		Type:         string(a.Type),
		UpdateType:   string(a.UpdateType),
		Field:        a.Field,
		CurrentValue: a.CurrentValue,
		Comment:      a.Comment,
		CreatedAt:    a.CreatedAt,
		WorkspaceID:  a.WorkspaceID,
	}
}

// toAppObjectiveActivities converts a slice of CoreActivity to a slice of AppObjectiveActivity
func toAppObjectiveActivities(acts []okractivities.CoreActivity) []AppObjectiveActivity {
	result := make([]AppObjectiveActivity, len(acts))
	for i, a := range acts {
		result[i] = toAppObjectiveActivity(a)
	}
	return result
}
