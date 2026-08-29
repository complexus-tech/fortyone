package reportsrepository

import (
	"context"
	"fmt"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetStoryAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreStoryAnalytics, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reports.CoreStoryAnalytics{}, err
	}
	startDate, endDate, err := query.requiredDates()
	if err != nil {
		return reports.CoreStoryAnalytics{}, err
	}

	statusRows, err := r.queries.ListStoryStatusBreakdown(ctx, reportssql.ListStoryStatusBreakdownParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
		SprintIds:   query.sprintIDs,
	})
	if err != nil {
		return reports.CoreStoryAnalytics{}, fmt.Errorf("selecting story status breakdown: %w", err)
	}

	priorityRows, err := r.queries.ListStoryPriorityDistribution(ctx, reportssql.ListStoryPriorityDistributionParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
		SprintIds:   query.sprintIDs,
	})
	if err != nil {
		return reports.CoreStoryAnalytics{}, fmt.Errorf("selecting story priority distribution: %w", err)
	}

	teamRows, err := r.queries.ListStoryCompletionByTeam(ctx, reportssql.ListStoryCompletionByTeamParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
		SprintIds:   query.sprintIDs,
	})
	if err != nil {
		return reports.CoreStoryAnalytics{}, fmt.Errorf("selecting story completion by team: %w", err)
	}

	burndownRows, err := r.queries.ListStoryBurndown(ctx, reportssql.ListStoryBurndownParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
		SprintIds:   query.sprintIDs,
	})
	if err != nil {
		return reports.CoreStoryAnalytics{}, fmt.Errorf("selecting story burndown: %w", err)
	}

	result := reports.CoreStoryAnalytics{
		StatusBreakdown:      make([]reports.CoreStatusBreakdownItem, len(statusRows)),
		PriorityDistribution: make([]reports.CorePriorityDistributionItem, len(priorityRows)),
		CompletionByTeam:     make([]reports.CoreTeamCompletionItem, len(teamRows)),
		Burndown:             make([]reports.CoreBurndownPoint, len(burndownRows)),
	}
	for i, row := range statusRows {
		teamID := row.TeamID
		result.StatusBreakdown[i] = reports.CoreStatusBreakdownItem{
			StatusName: row.StatusName,
			Count:      int(row.Count),
			TeamID:     &teamID,
		}
	}
	for i, row := range priorityRows {
		result.PriorityDistribution[i] = reports.CorePriorityDistributionItem{
			Priority: row.Priority,
			Count:    int(row.Count),
		}
	}
	for i, row := range teamRows {
		result.CompletionByTeam[i] = reports.CoreTeamCompletionItem{
			TeamID:    row.TeamID,
			TeamName:  row.TeamName,
			Total:     int(row.Total),
			Completed: int(row.Completed),
		}
	}
	for i, row := range burndownRows {
		result.Burndown[i] = reports.CoreBurndownPoint{
			Date:      row.CompletionDate,
			Remaining: int(row.Remaining),
		}
	}

	return result, nil
}
