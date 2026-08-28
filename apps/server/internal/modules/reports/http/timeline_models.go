package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

// 6. Timeline Trends App Models
type AppTimelineTrends struct {
	StoryCompletion   []AppStoryCompletionPoint   `json:"storyCompletion"`
	ObjectiveProgress []AppObjectiveProgressPoint `json:"objectiveProgress"`
	TeamVelocity      []AppTeamVelocityPoint      `json:"teamVelocity"`
	KeyMetricsTrend   []AppKeyMetricsTrendPoint   `json:"keyMetricsTrend"`
}

type AppStoryCompletionPoint struct {
	Date      time.Time `json:"date"`
	Completed int       `json:"completed"`
	Created   int       `json:"created"`
}

type AppObjectiveProgressPoint struct {
	Date                time.Time `json:"date"`
	TotalObjectives     int       `json:"totalObjectives"`
	CompletedObjectives int       `json:"completedObjectives"`
}

type AppTeamVelocityPoint struct {
	Date     time.Time `json:"date"`
	TeamID   uuid.UUID `json:"teamId"`
	Velocity int       `json:"velocity"`
}

type AppKeyMetricsTrendPoint struct {
	Date          time.Time `json:"date"`
	ActiveUsers   int       `json:"activeUsers"`
	StoriesPerDay float64   `json:"storiesPerDay"`
	AvgCycleTime  float64   `json:"avgCycleTime"`
}

func toAppTimelineTrends(trends reports.CoreTimelineTrends) AppTimelineTrends {
	return AppTimelineTrends{
		StoryCompletion:   toAppStoryCompletionPoints(trends.StoryCompletion),
		ObjectiveProgress: toAppObjectiveProgressPoints(trends.ObjectiveProgress),
		TeamVelocity:      toAppTeamVelocityPoints(trends.TeamVelocity),
		KeyMetricsTrend:   toAppKeyMetricsTrendPoints(trends.KeyMetricsTrend),
	}
}

func toAppStoryCompletionPoints(points []reports.CoreStoryCompletionPoint) []AppStoryCompletionPoint {
	result := make([]AppStoryCompletionPoint, len(points))
	for i, point := range points {
		result[i] = AppStoryCompletionPoint{
			Date:      point.Date,
			Completed: point.Completed,
			Created:   point.Created,
		}
	}
	return result
}

func toAppObjectiveProgressPoints(points []reports.CoreObjectiveProgressPoint) []AppObjectiveProgressPoint {
	result := make([]AppObjectiveProgressPoint, len(points))
	for i, point := range points {
		result[i] = AppObjectiveProgressPoint{
			Date:                point.Date,
			TotalObjectives:     point.TotalObjectives,
			CompletedObjectives: point.CompletedObjectives,
		}
	}
	return result
}

func toAppTeamVelocityPoints(points []reports.CoreTeamVelocityPoint) []AppTeamVelocityPoint {
	result := make([]AppTeamVelocityPoint, len(points))
	for i, point := range points {
		result[i] = AppTeamVelocityPoint{
			Date:     point.Date,
			TeamID:   point.TeamID,
			Velocity: point.Velocity,
		}
	}
	return result
}

func toAppKeyMetricsTrendPoints(points []reports.CoreKeyMetricsTrendPoint) []AppKeyMetricsTrendPoint {
	result := make([]AppKeyMetricsTrendPoint, len(points))
	for i, point := range points {
		result[i] = AppKeyMetricsTrendPoint{
			Date:          point.Date,
			ActiveUsers:   point.ActiveUsers,
			StoriesPerDay: point.StoriesPerDay,
			AvgCycleTime:  point.AvgCycleTime,
		}
	}
	return result
}
