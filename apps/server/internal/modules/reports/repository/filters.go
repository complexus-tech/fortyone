package reportsrepository

import (
	"context"
	"fmt"
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

const maxRequestedReportTeamIDs = 100

type queryFilters struct {
	workspaceID  uuid.UUID
	teamIDs      []uuid.UUID
	assigneeIDs  []uuid.UUID
	sprintIDs    []uuid.UUID
	objectiveIDs []uuid.UUID
	startDate    *time.Time
	endDate      *time.Time
}

func newQueryFilters(workspaceID uuid.UUID, filters reports.ReportFilters) queryFilters {
	return queryFilters{
		workspaceID:  workspaceID,
		teamIDs:      cloneUUIDs(filters.TeamIDs),
		assigneeIDs:  cloneUUIDs(filters.AssigneeIDs),
		sprintIDs:    cloneUUIDs(filters.SprintIDs),
		objectiveIDs: cloneUUIDs(filters.ObjectiveIDs),
		startDate:    filters.StartDate,
		endDate:      filters.EndDate,
	}
}

func (r *repo) scopedQueryFilters(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (queryFilters, error) {
	teamIDs, err := r.resolveTeamScope(ctx, filters.ActorID, workspaceID, filters.TeamIDs)
	if err != nil {
		return queryFilters{}, err
	}
	filters.TeamIDs = teamIDs

	return newQueryFilters(workspaceID, filters), nil
}

func (r *repo) validateRequestedTeam(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID, teamID *uuid.UUID) error {
	if teamID == nil {
		return r.authorize(ctx, actorID, workspaceID)
	}
	_, err := r.resolveTeamScope(ctx, actorID, workspaceID, []uuid.UUID{*teamID})
	return err
}

func (r *repo) resolveTeamScope(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID, requestedTeamIDs []uuid.UUID) ([]uuid.UUID, error) {
	access, err := r.actorAccess(ctx, actorID, workspaceID)
	if err != nil {
		return nil, err
	}

	requestedTeamIDs, err = normalizeRequestedTeamIDs(requestedTeamIDs)
	if err != nil {
		return nil, err
	}
	if access.isAdmin && len(requestedTeamIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	visibleTeamIDs, err := r.queries.ListReportsVisibleTeamIDs(ctx, reportssql.ListReportsVisibleTeamIDsParams{
		ActorID:          actorID,
		WorkspaceID:      workspaceID,
		RequestedTeamIds: requestedTeamIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("resolving reports team scope: %w", err)
	}
	if len(requestedTeamIDs) > 0 && len(visibleTeamIDs) != len(requestedTeamIDs) {
		return nil, reports.ErrReportsAccessDenied
	}
	if len(visibleTeamIDs) == 0 {
		// An empty SQLC UUID array means "all teams" to report queries. A nil UUID
		// is therefore the intentional non-matching scope for a restricted actor
		// who cannot see any team.
		return []uuid.UUID{uuid.Nil}, nil
	}

	return visibleTeamIDs, nil
}

func normalizeRequestedTeamIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	if len(values) > maxRequestedReportTeamIDs {
		return nil, reports.ErrInvalidReportFilters
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return nil, reports.ErrInvalidReportFilters
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result, nil
}

func (f queryFilters) requiredDates() (time.Time, time.Time, error) {
	if f.startDate == nil || f.endDate == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("missing required report date range: %w", reports.ErrInvalidReportFilters)
	}

	return *f.startDate, *f.endDate, nil
}

func cloneUUIDs(values []uuid.UUID) []uuid.UUID {
	if len(values) == 0 {
		return []uuid.UUID{}
	}

	return append([]uuid.UUID(nil), values...)
}
