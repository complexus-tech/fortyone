package reportsrepository

import (
	"strings"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
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

func buildUUIDArrayFilter(column string, paramName string, values []uuid.UUID, namedParams map[string]any) string {
	if len(values) == 0 {
		return ""
	}

	namedParams[paramName] = values

	var filter strings.Builder
	filter.WriteString(" AND ")
	filter.WriteString(column)
	filter.WriteString(" = ANY(CAST(:")
	filter.WriteString(paramName)
	filter.WriteString(" AS uuid[]))")

	return filter.String()
}

func buildFiltersForColumns(filters reports.ReportFilters, namedParams map[string]any, teamColumn string, sprintColumn string, objectiveColumn string) (teamFilter, sprintFilter, objectiveFilter string) {
	if teamColumn != "" {
		teamFilter = buildUUIDArrayFilter(teamColumn, "team_ids", filters.TeamIDs, namedParams)
	}
	if sprintColumn != "" {
		sprintFilter = buildUUIDArrayFilter(sprintColumn, "sprint_ids", filters.SprintIDs, namedParams)
	}
	if objectiveColumn != "" {
		objectiveFilter = buildUUIDArrayFilter(objectiveColumn, "objective_ids", filters.ObjectiveIDs, namedParams)
	}

	return teamFilter, sprintFilter, objectiveFilter
}

func buildWorkloadStoryFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	where.WriteString(buildUUIDArrayFilter("s.team_id", "team_ids", filters.TeamIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("s.assignee_id", "assignee_ids", filters.AssigneeIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("s.sprint_id", "sprint_ids", filters.SprintIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("s.objective_id", "objective_ids", filters.ObjectiveIDs, namedParams))

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

func buildPulseSprintFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	where.WriteString(buildUUIDArrayFilter("sp.team_id", "team_ids", filters.TeamIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("sp.sprint_id", "sprint_ids", filters.SprintIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("sp.objective_id", "objective_ids", filters.ObjectiveIDs, namedParams))

	if filters.StartDate != nil {
		where.WriteString(" AND sp.created_at >= :start_date")
		namedParams["start_date"] = *filters.StartDate
	}

	if filters.EndDate != nil {
		where.WriteString(" AND sp.created_at <= :end_date")
		namedParams["end_date"] = *filters.EndDate
	}

	return where.String()
}

func buildPulseObjectiveFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	where.WriteString(buildUUIDArrayFilter("o.team_id", "team_ids", filters.TeamIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("o.lead_user_id", "assignee_ids", filters.AssigneeIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("o.objective_id", "objective_ids", filters.ObjectiveIDs, namedParams))

	if filters.StartDate != nil {
		where.WriteString(" AND o.created_at >= :start_date")
		namedParams["start_date"] = *filters.StartDate
	}

	if filters.EndDate != nil {
		where.WriteString(" AND o.created_at <= :end_date")
		namedParams["end_date"] = *filters.EndDate
	}

	return where.String()
}

func buildPulseRequestFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	where.WriteString(buildUUIDArrayFilter("ir.team_id", "team_ids", filters.TeamIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("ir.assignee_id", "assignee_ids", filters.AssigneeIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("ir.sprint_id", "sprint_ids", filters.SprintIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("ir.objective_id", "objective_ids", filters.ObjectiveIDs, namedParams))

	if filters.StartDate != nil {
		where.WriteString(" AND ir.created_at >= :start_date")
		namedParams["start_date"] = *filters.StartDate
	}

	if filters.EndDate != nil {
		where.WriteString(" AND ir.created_at <= :end_date")
		namedParams["end_date"] = *filters.EndDate
	}

	return where.String()
}

func buildRequestSourceFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	where.WriteString(buildUUIDArrayFilter("ir.team_id", "team_ids", filters.TeamIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("ir.assignee_id", "assignee_ids", filters.AssigneeIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("ir.sprint_id", "sprint_ids", filters.SprintIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("ir.objective_id", "objective_ids", filters.ObjectiveIDs, namedParams))

	if filters.StartDate != nil {
		where.WriteString(" AND ir.created_at >= :start_date")
		namedParams["start_date"] = *filters.StartDate
	}

	if filters.EndDate != nil {
		where.WriteString(" AND ir.created_at <= :end_date")
		namedParams["end_date"] = *filters.EndDate
	}

	return where.String()
}

func buildWorkspaceEngagementFilter(filters reports.ReportFilters, namedParams map[string]any) string {
	var where strings.Builder

	where.WriteString(buildUUIDArrayFilter("wae.team_id", "team_ids", filters.TeamIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("wae.user_id", "assignee_ids", filters.AssigneeIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("wae.sprint_id", "sprint_ids", filters.SprintIDs, namedParams))
	where.WriteString(buildUUIDArrayFilter("wae.objective_id", "objective_ids", filters.ObjectiveIDs, namedParams))

	if filters.StartDate != nil {
		where.WriteString(" AND wae.occurred_at >= :start_date")
		namedParams["start_date"] = *filters.StartDate
	}

	if filters.EndDate != nil {
		where.WriteString(" AND wae.occurred_at <= :end_date")
		namedParams["end_date"] = *filters.EndDate
	}

	return where.String()
}

// Workspace Reports Repository Methods
