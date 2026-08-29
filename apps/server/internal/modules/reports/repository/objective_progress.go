package reportsrepository

import (
	"context"
	"fmt"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetObjectiveProgress(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreObjectiveProgress, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reports.CoreObjectiveProgress{}, err
	}
	startDate, endDate, err := query.requiredDates()
	if err != nil {
		return reports.CoreObjectiveProgress{}, err
	}

	healthRows, err := r.queries.ListObjectiveHealthDistribution(ctx, reportssql.ListObjectiveHealthDistributionParams{
		WorkspaceID:  query.workspaceID,
		StartDate:    startDate,
		EndDate:      endDate,
		TeamIds:      query.teamIDs,
		ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reports.CoreObjectiveProgress{}, fmt.Errorf("selecting objective health distribution: %w", err)
	}

	statusRows, err := r.queries.ListObjectiveStatusBreakdown(ctx, reportssql.ListObjectiveStatusBreakdownParams{
		WorkspaceID:  query.workspaceID,
		StartDate:    startDate,
		EndDate:      endDate,
		TeamIds:      query.teamIDs,
		ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reports.CoreObjectiveProgress{}, fmt.Errorf("selecting objective status breakdown: %w", err)
	}

	keyResultRows, err := r.queries.ListKeyResultProgress(ctx, reportssql.ListKeyResultProgressParams{
		WorkspaceID:  query.workspaceID,
		StartDate:    startDate,
		EndDate:      endDate,
		TeamIds:      query.teamIDs,
		ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reports.CoreObjectiveProgress{}, fmt.Errorf("selecting key result progress: %w", err)
	}

	teamRows, err := r.queries.ListObjectiveProgressByTeam(ctx, reportssql.ListObjectiveProgressByTeamParams{
		WorkspaceID:  query.workspaceID,
		StartDate:    startDate,
		EndDate:      endDate,
		TeamIds:      query.teamIDs,
		ObjectiveIds: query.objectiveIDs,
	})
	if err != nil {
		return reports.CoreObjectiveProgress{}, fmt.Errorf("selecting objective progress by team: %w", err)
	}

	result := reports.CoreObjectiveProgress{
		HealthDistribution: make([]reports.CoreHealthDistributionItem, len(healthRows)),
		StatusBreakdown:    make([]reports.CoreObjectiveStatusItem, len(statusRows)),
		KeyResultsProgress: make([]reports.CoreKeyResultProgressItem, len(keyResultRows)),
		ProgressByTeam:     make([]reports.CoreObjectiveTeamProgressItem, len(teamRows)),
	}
	for i, row := range healthRows {
		result.HealthDistribution[i] = reports.CoreHealthDistributionItem{
			Status: row.Status,
			Count:  int(row.Count),
		}
	}
	for i, row := range statusRows {
		result.StatusBreakdown[i] = reports.CoreObjectiveStatusItem{
			StatusName: row.StatusName,
			Count:      int(row.Count),
		}
	}
	for i, row := range keyResultRows {
		result.KeyResultsProgress[i] = reports.CoreKeyResultProgressItem{
			ObjectiveID:   row.ObjectiveID,
			ObjectiveName: row.ObjectiveName,
			Total:         int(row.Total),
			Completed:     int(row.Completed),
			AvgProgress:   row.AvgProgress,
		}
	}
	for i, row := range teamRows {
		result.ProgressByTeam[i] = reports.CoreObjectiveTeamProgressItem{
			TeamID:     row.TeamID,
			TeamName:   row.TeamName,
			Objectives: int(row.Objectives),
			Completed:  int(row.Completed),
		}
	}

	return result, nil
}
