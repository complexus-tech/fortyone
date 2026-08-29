package objectivesrepository

import (
	"context"
	"fmt"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
)

func (repository *Repository) GetAnalytics(
	ctx context.Context,
	query objectivesdomain.AnalyticsQuery,
	now time.Time,
) (objectivesdomain.ObjectiveAnalytics, error) {
	if err := query.Validate(); err != nil {
		return objectivesdomain.ObjectiveAnalytics{}, err
	}
	canRead, err := repository.queries.CanReadObjective(ctx, objectivessql.CanReadObjectiveParams{
		ActorID: query.ActorID, ObjectiveID: query.ObjectiveID, WorkspaceID: uuidPointer(query.WorkspaceID),
	})
	if err != nil {
		return objectivesdomain.ObjectiveAnalytics{}, fmt.Errorf("authorize objective analytics: %w", mapDatabaseError(err))
	}
	if !canRead {
		return objectivesdomain.ObjectiveAnalytics{}, objectivesdomain.ErrNotFound
	}

	priorityRows, err := repository.queries.GetObjectivePriorityBreakdown(ctx, objectivessql.GetObjectivePriorityBreakdownParams{
		ObjectiveID: uuidPointer(query.ObjectiveID), WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
	})
	if err != nil {
		return objectivesdomain.ObjectiveAnalytics{}, fmt.Errorf("get objective priority breakdown: %w", err)
	}
	progressRow, err := repository.queries.GetObjectiveProgressBreakdown(ctx, objectivessql.GetObjectiveProgressBreakdownParams{
		ObjectiveID: uuidPointer(query.ObjectiveID), WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
	})
	if err != nil {
		return objectivesdomain.ObjectiveAnalytics{}, fmt.Errorf("get objective progress breakdown: %w", err)
	}
	allocationRows, err := repository.queries.GetObjectiveTeamAllocation(ctx, objectivessql.GetObjectiveTeamAllocationParams{
		ObjectiveID: uuidPointer(query.ObjectiveID), WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
	})
	if err != nil {
		return objectivesdomain.ObjectiveAnalytics{}, fmt.Errorf("get objective team allocation: %w", err)
	}
	chartEnd := dateOnly(now.UTC())
	chartRows, err := repository.queries.GetObjectiveProgressChart(ctx, objectivessql.GetObjectiveProgressChartParams{
		ChartStart: chartEnd.AddDate(0, 0, -30), ChartEnd: chartEnd,
		ObjectiveID: query.ObjectiveID, WorkspaceID: uuidPointer(query.WorkspaceID), ActorID: query.ActorID,
	})
	if err != nil {
		return objectivesdomain.ObjectiveAnalytics{}, fmt.Errorf("get objective progress chart: %w", err)
	}

	analytics := objectivesdomain.ObjectiveAnalytics{
		ObjectiveID:       query.ObjectiveID,
		PriorityBreakdown: make([]objectivesdomain.PriorityBreakdown, 0, len(priorityRows)),
		ProgressBreakdown: objectivesdomain.ProgressBreakdown{
			Total: int(progressRow.Total), Completed: int(progressRow.Completed),
			InProgress: int(progressRow.InProgress), Todo: int(progressRow.Todo),
			Blocked: int(progressRow.Blocked), Cancelled: int(progressRow.Cancelled),
		},
		TeamAllocation: make([]objectivesdomain.TeamMemberAllocation, 0, len(allocationRows)),
		ProgressChart:  make([]objectivesdomain.ObjectiveProgressDataPoint, 0, len(chartRows)),
	}
	for _, row := range priorityRows {
		analytics.PriorityBreakdown = append(analytics.PriorityBreakdown, objectivesdomain.PriorityBreakdown{Priority: row.Priority, Count: int(row.Count)})
	}
	for _, row := range allocationRows {
		analytics.TeamAllocation = append(analytics.TeamAllocation, objectivesdomain.TeamMemberAllocation{
			MemberID: row.UserID, Username: row.Username, AvatarURL: row.AvatarURL,
			Assigned: int(row.Assigned), Completed: int(row.Completed),
		})
	}
	for _, row := range chartRows {
		analytics.ProgressChart = append(analytics.ProgressChart, objectivesdomain.ObjectiveProgressDataPoint{
			Date: row.CompletionDate.UTC(), Completed: int(row.StoriesCompleted),
			InProgress: int(row.StoriesInProgress), Total: int(row.TotalStories),
		})
	}
	return analytics, nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
