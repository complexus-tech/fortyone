package domain

import (
	"time"

	"github.com/google/uuid"
)

type ObjectiveHealth string

const (
	HealthAtRisk   ObjectiveHealth = "At Risk"
	HealthOnTrack  ObjectiveHealth = "On Track"
	HealthOffTrack ObjectiveHealth = "Off Track"
)

const DefaultObjectiveColor = "#4A90E2"

type ObjectiveScheduleStatus string

const (
	ScheduleStatusOnTrack    ObjectiveScheduleStatus = "on_track"
	ScheduleStatusAtRisk     ObjectiveScheduleStatus = "at_risk"
	ScheduleStatusNoTarget   ObjectiveScheduleStatus = "no_target"
	ScheduleStatusNoSchedule ObjectiveScheduleStatus = "no_schedule"
)

type ForecastCauseStory struct {
	ID         uuid.UUID
	SequenceID int
	Title      string
	Source     string
}

type Objective struct {
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
	ForecastCauseStory *ForecastCauseStory
	KeyResultCount     int
	TotalStories       int
	CancelledStories   int
	CompletedStories   int
	StartedStories     int
	UnstartedStories   int
	BacklogStories     int
}

func (objective *Objective) ApplyScheduleForecast() {
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

type ObjectiveAnalytics struct {
	ObjectiveID       uuid.UUID
	PriorityBreakdown []PriorityBreakdown
	ProgressBreakdown ProgressBreakdown
	TeamAllocation    []TeamMemberAllocation
	ProgressChart     []ObjectiveProgressDataPoint
}

type PriorityBreakdown struct {
	Priority string
	Count    int
}

type ProgressBreakdown struct {
	Total      int
	Completed  int
	InProgress int
	Todo       int
	Blocked    int
	Cancelled  int
}

type TeamMemberAllocation struct {
	MemberID  uuid.UUID
	Username  string
	AvatarURL *string
	Assigned  int
	Completed int
}

type ObjectiveProgressDataPoint struct {
	Date       time.Time `json:"date"`
	Completed  int       `json:"completed"`
	InProgress int       `json:"inProgress"`
	Total      int       `json:"total"`
}
