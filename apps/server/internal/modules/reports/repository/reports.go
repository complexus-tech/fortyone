package reportsrepository

import (
	"strings"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// Helper function to build filter conditions and named parameters
func buildFilters(filters reports.ReportFilters, namedParams map[string]any) (teamFilter, sprintFilter, objectiveFilter string) {
	var tb, sb, ob strings.Builder

	if len(filters.TeamIDs) > 0 {
		tb.WriteString("AND team_id = ANY(:team_ids)")
		namedParams["team_ids"] = filters.TeamIDs
	}

	if len(filters.SprintIDs) > 0 {
		sb.WriteString("AND sprint_id = ANY(:sprint_ids)")
		namedParams["sprint_ids"] = filters.SprintIDs
	}

	if len(filters.ObjectiveIDs) > 0 {
		ob.WriteString("AND objective_id = ANY(:objective_ids)")
		namedParams["objective_ids"] = filters.ObjectiveIDs
	}

	return tb.String(), sb.String(), ob.String()
}

func buildWorkloadStoryFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	if len(filters.TeamIDs) > 0 {
		where.WriteString("AND s.team_id = ANY(:team_ids)")
		namedParams["team_ids"] = filters.TeamIDs
	}

	if len(filters.AssigneeIDs) > 0 {
		where.WriteString(" AND s.assignee_id = ANY(:assignee_ids)")
		namedParams["assignee_ids"] = filters.AssigneeIDs
	}

	if len(filters.SprintIDs) > 0 {
		where.WriteString(" AND s.sprint_id = ANY(:sprint_ids)")
		namedParams["sprint_ids"] = filters.SprintIDs
	}

	if len(filters.ObjectiveIDs) > 0 {
		where.WriteString(" AND s.objective_id = ANY(:objective_ids)")
		namedParams["objective_ids"] = filters.ObjectiveIDs
	}

	if filters.StartDate != nil {
		where.WriteString(" AND s.created_at >= :start_date")
		namedParams["start_date"] = *filters.StartDate
	}

	if filters.EndDate != nil {
		where.WriteString(" AND s.created_at <= :end_date")
		namedParams["end_date"] = *filters.EndDate
	}

	return where.String()
}

// Workspace Reports Repository Methods
