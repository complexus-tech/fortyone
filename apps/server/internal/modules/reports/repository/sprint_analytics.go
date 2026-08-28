package reportsrepository

import (
	"context"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetSprintAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreSprintAnalyticsWorkspace, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CoreSprintAnalyticsWorkspace{}, err
	}
	startDate, endDate, err := query.requiredDates()
	if err != nil {
		return reportdomain.CoreSprintAnalyticsWorkspace{}, err
	}

	progressRows, err := r.queries.ListSprintProgress(ctx, reportssql.ListSprintProgressParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs,
	})
	if err != nil {
		return reportdomain.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("selecting sprint progress: %w", err)
	}
	burndownRows, err := r.queries.ListCombinedSprintBurndown(ctx, reportssql.ListCombinedSprintBurndownParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs,
	})
	if err != nil {
		return reportdomain.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("selecting combined sprint burndown: %w", err)
	}
	allocationRows, err := r.queries.ListSprintTeamAllocation(ctx, reportssql.ListSprintTeamAllocationParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs,
	})
	if err != nil {
		return reportdomain.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("selecting sprint team allocation: %w", err)
	}
	healthRows, err := r.queries.ListSprintHealth(ctx, reportssql.ListSprintHealthParams{
		WorkspaceID: query.workspaceID, StartDate: startDate, EndDate: endDate,
		TeamIds: query.teamIDs, SprintIds: query.sprintIDs,
	})
	if err != nil {
		return reportdomain.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("selecting sprint health: %w", err)
	}

	result := reportdomain.CoreSprintAnalyticsWorkspace{
		SprintProgress:   make([]reportdomain.CoreSprintProgressItem, len(progressRows)),
		CombinedBurndown: make([]reportdomain.CoreCombinedBurndownPoint, len(burndownRows)),
		TeamAllocation:   make([]reportdomain.CoreSprintTeamAllocation, len(allocationRows)),
		SprintHealth:     make([]reportdomain.CoreSprintHealthItem, len(healthRows)),
	}
	for i, row := range progressRows {
		if row.SprintID == uuid.Nil || row.TeamID == uuid.Nil {
			return reportdomain.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("mapping sprint progress: %w", ErrInvalidProjection)
		}
		result.SprintProgress[i] = reportdomain.CoreSprintProgressItem{
			SprintID: row.SprintID, SprintName: row.SprintName, TeamID: row.TeamID,
			Total: int(row.Total), Completed: int(row.Completed), Status: row.Status,
		}
	}
	for i, row := range burndownRows {
		result.CombinedBurndown[i] = reportdomain.CoreCombinedBurndownPoint{
			Date: row.Date, Planned: int(row.Planned), Actual: int(row.Actual),
		}
	}
	for i, row := range allocationRows {
		if row.TeamID == uuid.Nil {
			return reportdomain.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("mapping sprint allocation: %w", ErrInvalidProjection)
		}
		result.TeamAllocation[i] = reportdomain.CoreSprintTeamAllocation{
			TeamID: row.TeamID, TeamName: row.TeamName, ActiveSprints: int(row.ActiveSprints),
			TotalStories: int(row.TotalStories), CompletedStories: int(row.CompletedStories),
		}
	}
	for i, row := range healthRows {
		result.SprintHealth[i] = reportdomain.CoreSprintHealthItem{Status: row.Status, Count: int(row.Count)}
	}

	return result, nil
}
