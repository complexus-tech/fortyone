package reportsrepository

import (
	"context"
	"fmt"
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkspaceOverview, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reports.CoreWorkspaceOverview{}, err
	}
	startDate, endDate, err := query.requiredDates()
	if err != nil {
		return reports.CoreWorkspaceOverview{}, err
	}

	metrics, err := r.queries.GetWorkspaceMetrics(ctx, reportssql.GetWorkspaceMetricsParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("selecting workspace metrics: %w", err)
	}

	completionRows, err := r.queries.ListWorkspaceCompletionTrend(ctx, reportssql.ListWorkspaceCompletionTrendParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("selecting workspace completion trend: %w", err)
	}

	velocityRows, err := r.queries.ListWorkspaceVelocityTrend(ctx, reportssql.ListWorkspaceVelocityTrendParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("selecting workspace velocity trend: %w", err)
	}

	completionTrend := make([]reports.CoreCompletionTrendPoint, len(completionRows))
	for i, row := range completionRows {
		completionTrend[i] = reports.CoreCompletionTrendPoint{
			Date:      row.WeekStart,
			Completed: int(row.Completed),
			Total:     int(row.Total),
		}
	}

	velocityTrend := make([]reports.CoreVelocityTrendPoint, len(velocityRows))
	for i, row := range velocityRows {
		velocityTrend[i] = reports.CoreVelocityTrendPoint{
			Period:   row.Period,
			Velocity: int(row.Velocity),
		}
	}

	return reports.CoreWorkspaceOverview{
		WorkspaceID: workspaceID,
		ReportDate:  time.Now().UTC(),
		Filters:     filters,
		Metrics: reports.CoreWorkspaceMetrics{
			TotalStories:     int(metrics.TotalStories),
			CompletedStories: int(metrics.CompletedStories),
			ActiveObjectives: int(metrics.ActiveObjectives),
			ActiveSprints:    int(metrics.ActiveSprints),
			TotalTeamMembers: int(metrics.TotalTeamMembers),
		},
		CompletionTrend: completionTrend,
		VelocityTrend:   velocityTrend,
	}, nil
}
