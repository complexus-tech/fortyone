package mayaadapter

import (
	"context"
	"errors"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type SystemWorkloadReader interface {
	GetSystemTeamWorkload(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (reports.CoreWorkloadAnalysis, error)
}

type reportService struct {
	interactive maya.ReportsService
	system      SystemWorkloadReader
	actorID     uuid.UUID
}

func NewReports(interactive maya.ReportsService, system SystemWorkloadReader, actorID uuid.UUID) maya.ReportsService {
	return reportService{interactive: interactive, system: system, actorID: actorID}
}

func (s reportService) GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkloadAnalysis, error) {
	if caller, err := auth.GetActor(ctx); err == nil {
		if caller.WorkspaceID != uuid.Nil && caller.WorkspaceID != workspaceID {
			return reports.CoreWorkloadAnalysis{}, reports.ErrReportsAccessDenied
		}
		if caller.Kind != auth.PrincipalSystem {
			return s.interactive.GetWorkloadAnalysis(ctx, workspaceID, filters)
		}
		if caller.PrincipalID != s.actorID || !caller.TeamAccess.IsUnrestricted() {
			return reports.CoreWorkloadAnalysis{}, reports.ErrReportsAccessDenied
		}
	} else if !errors.Is(err, auth.ErrActorNotFound) {
		return reports.CoreWorkloadAnalysis{}, err
	}
	if s.system == nil || s.actorID == uuid.Nil || len(filters.TeamIDs) != 1 ||
		len(filters.AssigneeIDs) != 0 || len(filters.SprintIDs) != 0 || len(filters.ObjectiveIDs) != 0 ||
		filters.StartDate != nil || filters.EndDate != nil {
		return reports.CoreWorkloadAnalysis{}, reports.ErrInvalidReportFilters
	}
	return s.system.GetSystemTeamWorkload(ctx, s.actorID, workspaceID, filters.TeamIDs[0])
}
