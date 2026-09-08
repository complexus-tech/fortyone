package reportsrepository

import (
	"context"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

// GetSystemTeamWorkload is the narrow background planning port. It cannot
// request an unbounded workspace report or impersonate a workspace member.
func (r *repo) GetSystemTeamWorkload(ctx context.Context, actorID, workspaceID, teamID uuid.UUID) (reports.CoreWorkloadAnalysis, error) {
	if actorID == uuid.Nil || workspaceID == uuid.Nil || teamID == uuid.Nil {
		return reports.CoreWorkloadAnalysis{}, reports.ErrReportsAccessDenied
	}
	allowed, err := r.queries.SystemCanReadTeamWorkload(ctx, reportssql.SystemCanReadTeamWorkloadParams{
		ActorID: actorID, WorkspaceID: workspaceID, TeamID: teamID,
	})
	if err != nil {
		return reports.CoreWorkloadAnalysis{}, err
	}
	if !allowed {
		return reports.CoreWorkloadAnalysis{}, reports.ErrReportsAccessDenied
	}
	return r.workloadAnalysis(ctx, newQueryFilters(workspaceID, reports.ReportFilters{TeamIDs: []uuid.UUID{teamID}}))
}
