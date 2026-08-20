package objectives

import (
	"time"

	"github.com/google/uuid"
)

// ObjectiveHealth represents the possible health states of an objective
type ObjectiveHealth string

const (
	HealthAtRisk   ObjectiveHealth = "At Risk"
	HealthOnTrack  ObjectiveHealth = "On Track"
	HealthOffTrack ObjectiveHealth = "Off Track"
)

const DefaultObjectiveColor = "#4A90E2"

// ObjectiveScheduleStatus describes the objective's live delivery forecast.
// It is deliberately separate from the manually managed objective health.
type ObjectiveScheduleStatus string

const (
	ScheduleStatusOnTrack    ObjectiveScheduleStatus = "on_track"
	ScheduleStatusAtRisk     ObjectiveScheduleStatus = "at_risk"
	ScheduleStatusNoTarget   ObjectiveScheduleStatus = "no_target"
	ScheduleStatusNoSchedule ObjectiveScheduleStatus = "no_schedule"
)

type CoreForecastCauseStory struct {
	ID         uuid.UUID
	SequenceID int
	Title      string
	Source     string
}

type CoreObjective struct {
	ID                 uuid.UUID
	SequenceID         int
	Name               string
	Description        *string
	ShortSummary       *string
	LeadUser           *uuid.UUID
	Team               uuid.UUID
	Workspace          uuid.UUID
	StartDate          *time.Time
	EndDate            *time.Time
	IsPrivate          bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Status             uuid.UUID
	CreatedBy          uuid.UUID
	Priority           *string
	Health             *ObjectiveHealth
	Color              string
	ForecastStartDate  *time.Time
	ForecastEndDate    *time.Time
	ScheduleStatus     ObjectiveScheduleStatus
	ForecastDaysDelta  int
	ForecastCauseStory *CoreForecastCauseStory
	KeyResultCount     int
	TotalStories       int
	CancelledStories   int
	CompletedStories   int
	StartedStories     int
	UnstartedStories   int
	BacklogStories     int
}

type CoreNewObjective struct {
	Name         string
	Description  *string
	ShortSummary *string
	LeadUser     *uuid.UUID
	Team         uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	IsPrivate    bool
	Status       uuid.UUID
	Priority     *string
	Color        string
	CreatedBy    uuid.UUID
}

type CoreUpdateObjective struct {
	Name         *string
	Description  *string
	ShortSummary *string
	LeadUser     *uuid.UUID
	Team         *uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	IsPrivate    *bool
	Visibility   *string
	Status       *uuid.UUID
	Priority     *string
	Health       *ObjectiveHealth
	Color        *string
}

// ApplyScheduleForecast derives delivery risk from the committed objective
// target and the latest active linked work. It never mutates target dates.
func (objective *CoreObjective) ApplyScheduleForecast() {
	objective.ForecastDaysDelta = 0

	if objective.EndDate == nil {
		objective.ScheduleStatus = ScheduleStatusNoTarget
		return
	}
	if objective.ForecastEndDate == nil {
		objective.ScheduleStatus = ScheduleStatusNoSchedule
		return
	}

	objective.ForecastDaysDelta = calendarDayDelta(*objective.EndDate, *objective.ForecastEndDate)
	if objective.ForecastDaysDelta > 0 {
		objective.ScheduleStatus = ScheduleStatusAtRisk
		return
	}

	objective.ScheduleStatus = ScheduleStatusOnTrack
}

func calendarDayDelta(from, to time.Time) int {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(toDate.Sub(fromDate).Hours() / 24)
}

type CoreStrategyMap struct {
	UltimateGoal string
	Description  *string
	Pillars      []CoreStrategicPillar
}

type CoreStrategicPillar struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	OrderIndex   int
	ObjectiveIDs []uuid.UUID
}

type CoreStrategyUpdate struct {
	UltimateGoal string
	Description  *string
}

type CoreNewStrategicPillar struct {
	Name        string
	Description *string
	OrderIndex  int
}

type CoreUpdateStrategicPillar struct {
	Name        *string
	Description *string
	OrderIndex  *int
}

// Objective Analytics Models

type CoreObjectiveAnalytics struct {
	ObjectiveID       uuid.UUID
	PriorityBreakdown []CorePriorityBreakdown
	ProgressBreakdown CoreProgressBreakdown
	TeamAllocation    []CoreTeamMemberAllocation
	ProgressChart     []CoreObjectiveProgressDataPoint
}

type CorePriorityBreakdown struct {
	Priority string `db:"priority"`
	Count    int    `db:"count"`
}

type CoreProgressBreakdown struct {
	Total      int `db:"total"`
	Completed  int `db:"completed"`
	InProgress int `db:"in_progress"`
	Todo       int `db:"todo"`
	Blocked    int `db:"blocked"`
	Cancelled  int `db:"cancelled"`
}

type CoreTeamMemberAllocation struct {
	MemberID  uuid.UUID `db:"user_id"`
	Username  string    `db:"username"`
	AvatarURL *string   `db:"avatar_url"`
	Assigned  int       `db:"assigned"`
	Completed int       `db:"completed"`
}

type CoreObjectiveProgressDataPoint struct {
	Date       time.Time `json:"date" db:"completion_date"`
	Completed  int       `json:"completed" db:"stories_completed"`
	InProgress int       `json:"inProgress" db:"stories_in_progress"`
	Total      int       `json:"total" db:"total_stories"`
}
