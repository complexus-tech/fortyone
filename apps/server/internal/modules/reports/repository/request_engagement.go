package reportsrepository

import (
	"context"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetRequestSourceAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreRequestSourceAnalytics, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CoreRequestSourceAnalytics{}, err
	}
	rows, err := r.queries.ListRequestSourcePerformance(ctx, reportssql.ListRequestSourcePerformanceParams{
		WorkspaceID: query.workspaceID, TeamIds: query.teamIDs, AssigneeIds: query.assigneeIDs,
		SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs,
		StartDate: query.startDate, EndDate: query.endDate,
	})
	if err != nil {
		return reportdomain.CoreRequestSourceAnalytics{}, fmt.Errorf("selecting request source performance: %w", err)
	}

	result := reportdomain.CoreRequestSourceAnalytics{
		Providers: make([]reportdomain.CoreRequestProviderPerformance, len(rows)),
	}
	for i, row := range rows {
		provider := reportdomain.CoreRequestProviderPerformance{
			Provider: row.Provider, TotalRequests: int(row.TotalRequests), PendingRequests: int(row.PendingRequests),
			AcceptedRequests: int(row.AcceptedRequests), DeclinedRequests: int(row.DeclinedRequests),
			UrgentRequests: int(row.UrgentRequests), HighRequests: int(row.HighRequests),
			StaleRequests: int(row.StaleRequests), AcceptanceRate: row.AcceptanceRate,
		}
		result.Providers[i] = provider
		result.TotalRequests += provider.TotalRequests
		result.PendingRequests += provider.PendingRequests
		result.AcceptedRequests += provider.AcceptedRequests
		result.DeclinedRequests += provider.DeclinedRequests
	}

	return result, nil
}

func (r *repo) GetWorkspaceEngagementAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreWorkspaceEngagementAnalytics, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CoreWorkspaceEngagementAnalytics{}, err
	}
	totals, err := r.queries.GetWorkspaceEngagementTotals(ctx, reportssql.GetWorkspaceEngagementTotalsParams{
		WorkspaceID: query.workspaceID, TeamIds: query.teamIDs, AssigneeIds: query.assigneeIDs,
		SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs, StartDate: query.startDate, EndDate: query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkspaceEngagementAnalytics{}, fmt.Errorf("selecting engagement totals: %w", err)
	}
	nameRows, err := r.queries.ListWorkspaceEngagementByName(ctx, reportssql.ListWorkspaceEngagementByNameParams{
		WorkspaceID: query.workspaceID, TeamIds: query.teamIDs, AssigneeIds: query.assigneeIDs,
		SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs, StartDate: query.startDate, EndDate: query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkspaceEngagementAnalytics{}, fmt.Errorf("selecting engagement event names: %w", err)
	}
	surfaceRows, err := r.queries.ListWorkspaceEngagementBySurface(ctx, reportssql.ListWorkspaceEngagementBySurfaceParams{
		WorkspaceID: query.workspaceID, TeamIds: query.teamIDs, AssigneeIds: query.assigneeIDs,
		SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs, StartDate: query.startDate, EndDate: query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkspaceEngagementAnalytics{}, fmt.Errorf("selecting engagement surfaces: %w", err)
	}
	userRows, err := r.queries.ListWorkspaceEngagementTopUsers(ctx, reportssql.ListWorkspaceEngagementTopUsersParams{
		WorkspaceID: query.workspaceID, TeamIds: query.teamIDs, AssigneeIds: query.assigneeIDs,
		SprintIds: query.sprintIDs, ObjectiveIds: query.objectiveIDs, StartDate: query.startDate, EndDate: query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkspaceEngagementAnalytics{}, fmt.Errorf("selecting engagement users: %w", err)
	}

	result := reportdomain.CoreWorkspaceEngagementAnalytics{
		TotalEvents: int(totals.TotalEvents), UniqueUsers: int(totals.UniqueUsers),
		EventsByName:    make([]reportdomain.CoreWorkspaceEngagementCount, len(nameRows)),
		EventsBySurface: make([]reportdomain.CoreWorkspaceEngagementCount, len(surfaceRows)),
		TopUsers:        make([]reportdomain.CoreWorkspaceEngagementUser, len(userRows)),
	}
	for i, row := range nameRows {
		result.EventsByName[i] = reportdomain.CoreWorkspaceEngagementCount{Name: row.Name, Count: int(row.Count)}
	}
	for i, row := range surfaceRows {
		result.EventsBySurface[i] = reportdomain.CoreWorkspaceEngagementCount{Name: row.Name, Count: int(row.Count)}
	}
	for i, row := range userRows {
		if row.UserID == uuid.Nil {
			return reportdomain.CoreWorkspaceEngagementAnalytics{}, fmt.Errorf("mapping engagement user: %w", ErrInvalidProjection)
		}
		result.TopUsers[i] = reportdomain.CoreWorkspaceEngagementUser{
			UserID: row.UserID, FullName: row.FullName, Username: row.Username, AvatarURL: row.AvatarURL, Events: int(row.Events),
		}
	}

	return result, nil
}
