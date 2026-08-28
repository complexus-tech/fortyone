package reportsrepository

import (
	"context"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetTimelineTrends(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreTimelineTrends, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CoreTimelineTrends{}, err
	}
	startDate, endDate, err := query.requiredDates()
	if err != nil {
		return reportdomain.CoreTimelineTrends{}, err
	}

	storyRows, err := r.queries.ListStoryCompletionTimeline(ctx, reportssql.ListStoryCompletionTimelineParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reportdomain.CoreTimelineTrends{}, fmt.Errorf("selecting story completion timeline: %w", err)
	}
	objectiveRows, err := r.queries.ListObjectiveProgressTimeline(ctx, reportssql.ListObjectiveProgressTimelineParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reportdomain.CoreTimelineTrends{}, fmt.Errorf("selecting objective progress timeline: %w", err)
	}
	velocityRows, err := r.queries.ListTeamVelocityTimeline(ctx, reportssql.ListTeamVelocityTimelineParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reportdomain.CoreTimelineTrends{}, fmt.Errorf("selecting team velocity timeline: %w", err)
	}
	metricRows, err := r.queries.ListKeyMetricsTimeline(ctx, reportssql.ListKeyMetricsTimelineParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reportdomain.CoreTimelineTrends{}, fmt.Errorf("selecting key metrics timeline: %w", err)
	}

	result := reportdomain.CoreTimelineTrends{
		StoryCompletion:   make([]reportdomain.CoreStoryCompletionPoint, len(storyRows)),
		ObjectiveProgress: make([]reportdomain.CoreObjectiveProgressPoint, len(objectiveRows)),
		TeamVelocity:      make([]reportdomain.CoreTeamVelocityPoint, len(velocityRows)),
		KeyMetricsTrend:   make([]reportdomain.CoreKeyMetricsTrendPoint, len(metricRows)),
	}
	for i, row := range storyRows {
		result.StoryCompletion[i] = reportdomain.CoreStoryCompletionPoint{Date: row.Date, Created: int(row.Created), Completed: int(row.Completed)}
	}
	for i, row := range objectiveRows {
		result.ObjectiveProgress[i] = reportdomain.CoreObjectiveProgressPoint{Date: row.Date, TotalObjectives: int(row.TotalObjectives), CompletedObjectives: int(row.CompletedObjectives)}
	}
	for i, row := range velocityRows {
		if row.TeamID == uuid.Nil {
			return reportdomain.CoreTimelineTrends{}, fmt.Errorf("mapping team velocity timeline: %w", ErrInvalidProjection)
		}
		result.TeamVelocity[i] = reportdomain.CoreTeamVelocityPoint{Date: row.Date, TeamID: row.TeamID, Velocity: int(row.Velocity)}
	}
	for i, row := range metricRows {
		result.KeyMetricsTrend[i] = reportdomain.CoreKeyMetricsTrendPoint{Date: row.Date, ActiveUsers: int(row.ActiveUsers), StoriesPerDay: row.StoriesPerDay, AvgCycleTime: row.AvgCycleTime}
	}

	return result, nil
}
