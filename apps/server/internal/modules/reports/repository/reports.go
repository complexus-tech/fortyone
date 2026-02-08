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

// Workspace Reports Repository Methods
