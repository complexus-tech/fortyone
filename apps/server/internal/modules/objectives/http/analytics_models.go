package objectiveshttp

import (
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
)

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

func toAppObjectiveAnalytics(value objectives.CoreObjectiveAnalytics) AppObjectiveAnalytics {
	priority := make([]AppPriorityBreakdown, len(value.PriorityBreakdown))
	for index, item := range value.PriorityBreakdown {
		priority[index] = AppPriorityBreakdown{Priority: item.Priority, Count: item.Count}
	}
	allocation := make([]AppTeamMemberAllocation, len(value.TeamAllocation))
	for index, item := range value.TeamAllocation {
		allocation[index] = AppTeamMemberAllocation{
			MemberID: item.MemberID, Username: item.Username, AvatarURL: item.AvatarURL,
			Assigned: item.Assigned, Completed: item.Completed,
		}
	}
	chart := make([]AppObjectiveProgressDataPoint, len(value.ProgressChart))
	for index, item := range value.ProgressChart {
		chart[index] = AppObjectiveProgressDataPoint{
			Date: item.Date, Completed: item.Completed, InProgress: item.InProgress, Total: item.Total,
		}
	}
	return AppObjectiveAnalytics{
		ObjectiveID: value.ObjectiveID, PriorityBreakdown: priority,
		ProgressBreakdown: AppProgressBreakdown{
			Total: value.ProgressBreakdown.Total, Completed: value.ProgressBreakdown.Completed,
			InProgress: value.ProgressBreakdown.InProgress, Todo: value.ProgressBreakdown.Todo,
			Blocked: value.ProgressBreakdown.Blocked, Cancelled: value.ProgressBreakdown.Cancelled,
		},
		TeamAllocation: allocation, ProgressChart: chart,
	}
}
