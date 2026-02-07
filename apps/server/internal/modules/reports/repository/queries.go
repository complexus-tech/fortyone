package reportsrepository

import (
	"context"
	"fmt"
	"strings"
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// GetStoryStats gets story statistics for a workspace.
func (r *repo) GetStoryStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StoryStatsFilters) (reports.CoreStoryStats, error) {
	const query = `
		WITH story_stats AS (
			SELECT 
				COUNT(CASE WHEN s.category = 'completed' AND st.assignee_id = :user_id THEN 1 END) as closed,
				COUNT(CASE 
					WHEN st.end_date < CURRENT_DATE 
					AND s.category NOT IN ('completed', 'cancelled')
					AND st.assignee_id = :user_id
					THEN 1 
				END) as overdue,
				COUNT(CASE 
					WHEN s.category = 'started' 
					AND st.assignee_id = :user_id 
					THEN 1 
				END) as in_progress,
				COUNT(CASE WHEN st.reporter_id = :user_id THEN 1 END) as created,
				COUNT(CASE WHEN st.assignee_id = :user_id THEN 1 END) as assigned
			FROM stories st
			JOIN statuses s ON st.status_id = s.status_id
			WHERE st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
		)
		SELECT * FROM story_stats`

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		r.log.Error(ctx, "failed to get user id", "error", err)
		return reports.CoreStoryStats{}, fmt.Errorf("getting user id: %w", err)
	}

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"start_date":   filters.StartDate,
		"end_date":     filters.EndDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return reports.CoreStoryStats{}, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var stats dbStoryStats
	if err := stmt.GetContext(ctx, &stats, params); err != nil {
		r.log.Error(ctx, "failed to get story stats", "error", err)
		return reports.CoreStoryStats{}, fmt.Errorf("selecting story stats: %w", err)
	}

	return toCoreStoryStats(stats), nil
}

// GetContributionStats gets contribution statistics for a user.
func (r *repo) GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, startDate time.Time, endDate time.Time) ([]reports.CoreContributionStats, error) {
	const query = `
		WITH dates AS (
			SELECT CAST(generate_series(CAST(:start_date AS date), CAST(:end_date AS date), INTERVAL '1 day') AS date) as date
		),
		activity_counts AS (
			SELECT 
				DATE(created_at) as date,
				COUNT(*) as contributions
			FROM story_activities
			WHERE user_id = :user_id
			AND workspace_id = :workspace_id
			AND created_at >= :start_date
			AND created_at <= :end_date
			GROUP BY DATE(created_at)
		)
		SELECT 
			d.date,
			COALESCE(ac.contributions, 0) as contributions
		FROM dates d
		LEFT JOIN activity_counts ac ON d.date = ac.date
		ORDER BY d.date`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
		"start_date":   startDate,
		"end_date":     endDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return nil, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var stats []dbContributionStats
	if err := stmt.SelectContext(ctx, &stats, params); err != nil {
		r.log.Error(ctx, "failed to get contribution stats", "error", err)
		return nil, fmt.Errorf("selecting contribution stats: %w", err)
	}

	return toCoreContributionsStats(stats), nil
}

// GetUserStats gets user-specific statistics.
func (r *repo) GetUserStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (reports.CoreUserStats, error) {
	const query = `
		WITH user_stats AS (
			SELECT 
				COUNT(CASE WHEN assignee_id = :user_id THEN 1 END) as assigned_to_me,
				COUNT(CASE WHEN reporter_id = :user_id THEN 1 END) as created_by_me
			FROM stories
			WHERE workspace_id = :workspace_id
		)
		SELECT * FROM user_stats`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return reports.CoreUserStats{}, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var stats dbUserStats
	if err := stmt.GetContext(ctx, &stats, params); err != nil {
		r.log.Error(ctx, "failed to get user stats", "error", err)
		return reports.CoreUserStats{}, fmt.Errorf("selecting user stats: %w", err)
	}

	return toCoreUserStats(stats), nil
}

func (r *repo) GetStatusStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StatsFilters) ([]reports.CoreStatusStats, error) {
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		r.log.Error(ctx, "failed to get user id", "error", err)
		return nil, fmt.Errorf("getting user id: %w", err)
	}

	query := `
		WITH user_teams AS (
			SELECT team_id 
			FROM team_members 
			WHERE user_id = :user_id
		),
		story_stats AS (
			SELECT 
				s.status_id,
				s.name,
				COUNT(st.id) as count
			FROM statuses s
			INNER JOIN stories st ON st.status_id = s.status_id 
				AND st.deleted_at IS NULL
				AND st.is_draft = false
				AND st.team_id IN (SELECT team_id FROM user_teams)
			WHERE s.workspace_id = :workspace_id
				AND st.created_at >= :start_date
				AND st.created_at <= :end_date
			GROUP BY s.status_id, s.name, s.order_index
			ORDER BY s.order_index
		)
		SELECT 
			name,
			CAST(count AS integer)
		FROM story_stats`

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"start_date":   filters.StartDate,
		"end_date":     filters.EndDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return nil, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var stats []dbStatusStats
	if err := stmt.SelectContext(ctx, &stats, params); err != nil {
		r.log.Error(ctx, "failed to get status stats", "error", err)
		return nil, fmt.Errorf("selecting status stats: %w", err)
	}

	return toCoreStatusStats(stats), nil
}

func (r *repo) GetPriorityStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StatsFilters) ([]reports.CorePriorityStats, error) {
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		r.log.Error(ctx, "failed to get user id", "error", err)
		return nil, fmt.Errorf("getting user id: %w", err)
	}

	query := `
		WITH user_teams AS (
			SELECT team_id 
			FROM team_members 
			WHERE user_id = :user_id
		),
		priority_stats AS (
			SELECT 
				st.priority,
				COUNT(st.id) as count
			FROM stories st
			WHERE st.workspace_id = :workspace_id
				AND st.deleted_at IS NULL
				AND st.is_draft = false
				AND st.team_id IN (SELECT team_id FROM user_teams)
				AND st.created_at >= :start_date
				AND st.created_at <= :end_date
			GROUP BY st.priority
			ORDER BY 
				CASE st.priority
					WHEN 'Urgent' THEN 1
					WHEN 'High' THEN 2
					WHEN 'Medium' THEN 3
					WHEN 'Low' THEN 4
					WHEN 'No Priority' THEN 5
					ELSE 6
				END
		)
		SELECT 
			priority,
			CAST(count AS integer)
		FROM priority_stats`

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"start_date":   filters.StartDate,
		"end_date":     filters.EndDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return nil, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var stats []dbPriorityStats
	if err := stmt.SelectContext(ctx, &stats, params); err != nil {
		r.log.Error(ctx, "failed to get priority stats", "error", err)
		return nil, fmt.Errorf("selecting priority stats: %w", err)
	}

	return toCorePriorityStats(stats), nil
}

func (r *repo) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkspaceOverview, error) {
	r.log.Info(ctx, "reportsrepository.GetWorkspaceOverview")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetWorkspaceOverview")
	defer span.End()

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"start_date":   *filters.StartDate,
		"end_date":     *filters.EndDate,
	}

	// Build filter conditions
	teamFilter, _, _ := buildFilters(filters, namedParams)

	// Get workspace metrics
	metricsQuery := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT st.id) as total_stories,
			COUNT(DISTINCT CASE WHEN stat.category = 'completed' THEN st.id END) as completed_stories,
			COUNT(DISTINCT o.objective_id) as active_objectives,
			COUNT(DISTINCT s.sprint_id) as active_sprints,
			COUNT(DISTINCT tm.user_id) as total_team_members
		FROM stories st
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		LEFT JOIN objectives o ON st.objective_id = o.objective_id
		LEFT JOIN sprints s ON st.sprint_id = s.sprint_id
		LEFT JOIN team_members tm ON tm.team_id = st.team_id
		WHERE st.workspace_id = :workspace_id 
			AND st.deleted_at IS NULL 
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
			%s
	`, teamFilter)

	var metrics reports.CoreWorkspaceMetrics
	metricsStmt, err := r.db.PrepareNamedContext(ctx, metricsQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare metrics query", "error", err)
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("preparing metrics query: %w", err)
	}
	defer metricsStmt.Close()

	err = metricsStmt.GetContext(ctx, &metrics, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute metrics query", "error", err)
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("executing metrics query: %w", err)
	}

	// Get completion trend (weekly data points)
	trendQuery := fmt.Sprintf(`
		WITH weekly_data AS (
			SELECT 
				DATE_TRUNC('week', st.created_at) as week_start,
				COUNT(DISTINCT st.id) as total,
				COUNT(DISTINCT CASE WHEN stat.category = 'completed' THEN st.id END) as completed
			FROM stories st
			LEFT JOIN statuses stat ON st.status_id = stat.status_id
			WHERE st.workspace_id = :workspace_id 
				AND st.deleted_at IS NULL 
				AND st.is_draft = false
				AND st.created_at >= :start_date
				AND st.created_at <= :end_date
				%s
			GROUP BY DATE_TRUNC('week', st.created_at)
			ORDER BY week_start
		)
		SELECT week_start, completed, total FROM weekly_data
	`, teamFilter)

	type dbCompletionTrend struct {
		Date      time.Time `db:"week_start"`
		Completed int       `db:"completed"`
		Total     int       `db:"total"`
	}

	trendStmt, err := r.db.PrepareNamedContext(ctx, trendQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare trend query", "error", err)
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("preparing trend query: %w", err)
	}
	defer trendStmt.Close()

	var dbCompletionTrends []dbCompletionTrend
	err = trendStmt.SelectContext(ctx, &dbCompletionTrends, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute trend query", "error", err)
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("executing trend query: %w", err)
	}

	// Convert to core model
	completionTrend := make([]reports.CoreCompletionTrendPoint, len(dbCompletionTrends))
	for i, trend := range dbCompletionTrends {
		completionTrend[i] = reports.CoreCompletionTrendPoint{
			Date:      trend.Date,
			Completed: trend.Completed,
			Total:     trend.Total,
		}
	}

	// Get velocity trend (weekly story completion)
	velocityQuery := fmt.Sprintf(`
		WITH weekly_velocity AS (
			SELECT 
				TO_CHAR(DATE_TRUNC('week', st.updated_at), 'Mon DD') as period,
				COUNT(DISTINCT st.id) as velocity
			FROM stories st
			LEFT JOIN statuses stat ON st.status_id = stat.status_id
			WHERE st.workspace_id = :workspace_id 
				AND st.deleted_at IS NULL 
				AND st.is_draft = false
				AND stat.category = 'completed'
				AND st.updated_at >= :start_date
				AND st.updated_at <= :end_date
				%s
			GROUP BY DATE_TRUNC('week', st.updated_at)
			ORDER BY DATE_TRUNC('week', st.updated_at)
		)
		SELECT period, velocity FROM weekly_velocity
	`, teamFilter)

	type dbVelocityTrend struct {
		Period   string `db:"period"`
		Velocity int    `db:"velocity"`
	}

	velocityStmt, err := r.db.PrepareNamedContext(ctx, velocityQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare velocity query", "error", err)
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("preparing velocity query: %w", err)
	}
	defer velocityStmt.Close()

	var dbVelocityTrends []dbVelocityTrend
	err = velocityStmt.SelectContext(ctx, &dbVelocityTrends, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute velocity query", "error", err)
		return reports.CoreWorkspaceOverview{}, fmt.Errorf("executing velocity query: %w", err)
	}

	// Convert to core model
	velocityTrend := make([]reports.CoreVelocityTrendPoint, len(dbVelocityTrends))
	for i, velocity := range dbVelocityTrends {
		velocityTrend[i] = reports.CoreVelocityTrendPoint{
			Period:   velocity.Period,
			Velocity: velocity.Velocity,
		}
	}

	return reports.CoreWorkspaceOverview{
		WorkspaceID:     workspaceID,
		ReportDate:      time.Now(),
		Filters:         filters,
		Metrics:         metrics,
		CompletionTrend: completionTrend,
		VelocityTrend:   velocityTrend,
	}, nil
}

func (r *repo) GetStoryAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreStoryAnalytics, error) {
	r.log.Info(ctx, "reportsrepository.GetStoryAnalytics")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetStoryAnalytics")
	defer span.End()

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"start_date":   *filters.StartDate,
		"end_date":     *filters.EndDate,
	}

	// Build filter conditions
	var teamFilter strings.Builder
	var sprintFilter strings.Builder

	if len(filters.TeamIDs) > 0 {
		teamFilter.WriteString("AND st.team_id = ANY(:team_ids)")
		namedParams["team_ids"] = filters.TeamIDs
	}

	if len(filters.SprintIDs) > 0 {
		sprintFilter.WriteString("AND st.sprint_id = ANY(:sprint_ids)")
		namedParams["sprint_ids"] = filters.SprintIDs
	}

	// Get status breakdown
	statusQuery := fmt.Sprintf(`
		SELECT 
			s.name as status_name,
			COUNT(st.id) as count,
			st.team_id as team_id
		FROM stories st
		INNER JOIN statuses s ON s.status_id = st.status_id
		WHERE st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
			%s
			%s
		GROUP BY s.status_id, s.name, s.order_index, st.team_id
		ORDER BY s.order_index
	`, teamFilter.String(), sprintFilter.String())

	type dbStatusBreakdown struct {
		StatusName string     `db:"status_name"`
		Count      int        `db:"count"`
		TeamID     *uuid.UUID `db:"team_id"`
	}

	statusStmt, err := r.db.PrepareNamedContext(ctx, statusQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare status query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("preparing status query: %w", err)
	}
	defer statusStmt.Close()

	var dbStatusBreakdowns []dbStatusBreakdown
	err = statusStmt.SelectContext(ctx, &dbStatusBreakdowns, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute status query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("executing status query: %w", err)
	}

	// Convert to core model
	statusBreakdown := make([]reports.CoreStatusBreakdownItem, len(dbStatusBreakdowns))
	for i, status := range dbStatusBreakdowns {
		statusBreakdown[i] = reports.CoreStatusBreakdownItem{
			StatusName: status.StatusName,
			Count:      status.Count,
			TeamID:     status.TeamID,
		}
	}

	// Get priority distribution
	priorityQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(st.priority, 'No Priority') as priority,
			COUNT(st.id) as count
		FROM stories st
		WHERE st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
			%s
			%s
		GROUP BY st.priority
		ORDER BY 
			CASE COALESCE(st.priority, 'No Priority')
				WHEN 'Urgent' THEN 1
				WHEN 'High' THEN 2
				WHEN 'Medium' THEN 3
				WHEN 'Low' THEN 4
				WHEN 'No Priority' THEN 5
				ELSE 6
			END
	`, teamFilter.String(), sprintFilter.String())

	type dbPriorityDistribution struct {
		Priority string `db:"priority"`
		Count    int    `db:"count"`
	}

	priorityStmt, err := r.db.PrepareNamedContext(ctx, priorityQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare priority query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("preparing priority query: %w", err)
	}
	defer priorityStmt.Close()

	var dbPriorityDistributions []dbPriorityDistribution
	err = priorityStmt.SelectContext(ctx, &dbPriorityDistributions, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute priority query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("executing priority query: %w", err)
	}

	// Convert to core model
	priorityDistribution := make([]reports.CorePriorityDistributionItem, len(dbPriorityDistributions))
	for i, priority := range dbPriorityDistributions {
		priorityDistribution[i] = reports.CorePriorityDistributionItem{
			Priority: priority.Priority,
			Count:    priority.Count,
		}
	}

	// Get completion by team
	teamQuery := fmt.Sprintf(`
		SELECT 
			t.team_id as team_id,
			t.name as team_name,
			COUNT(st.id) as total,
			COUNT(CASE WHEN stat.category = 'completed' THEN st.id END) as completed
		FROM teams t
		LEFT JOIN stories st ON st.team_id = t.team_id 
			AND st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
			%s
			%s
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		WHERE t.workspace_id = :workspace_id
		GROUP BY t.team_id, t.name
		ORDER BY t.name
	`, teamFilter.String(), sprintFilter.String())

	type dbTeamCompletion struct {
		TeamID    uuid.UUID `db:"team_id"`
		TeamName  string    `db:"team_name"`
		Total     int       `db:"total"`
		Completed int       `db:"completed"`
	}

	teamStmt, err := r.db.PrepareNamedContext(ctx, teamQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare team query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("preparing team query: %w", err)
	}
	defer teamStmt.Close()

	var dbTeamCompletions []dbTeamCompletion
	err = teamStmt.SelectContext(ctx, &dbTeamCompletions, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute team query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("executing team query: %w", err)
	}

	// Convert to core model
	completionByTeam := make([]reports.CoreTeamCompletionItem, len(dbTeamCompletions))
	for i, team := range dbTeamCompletions {
		completionByTeam[i] = reports.CoreTeamCompletionItem{
			TeamID:    team.TeamID,
			TeamName:  team.TeamName,
			Total:     team.Total,
			Completed: team.Completed,
		}
	}

	// Get burndown data (daily completion counts)
	burndownQuery := fmt.Sprintf(`
		WITH daily_completion AS (
			SELECT 
				DATE(st.updated_at) as completion_date,
				COUNT(st.id) as remaining
			FROM stories st
			LEFT JOIN statuses stat ON st.status_id = stat.status_id
			WHERE st.workspace_id = :workspace_id
				AND st.deleted_at IS NULL
				AND st.is_draft = false
				AND stat.category = 'completed'
				AND st.updated_at >= :start_date
				AND st.updated_at <= :end_date
				%s
				%s
			GROUP BY DATE(st.updated_at)
			ORDER BY completion_date
		)
		SELECT completion_date, remaining FROM daily_completion
	`, teamFilter.String(), sprintFilter.String())

	type dbBurndown struct {
		Date      time.Time `db:"completion_date"`
		Remaining int       `db:"remaining"`
	}

	burndownStmt, err := r.db.PrepareNamedContext(ctx, burndownQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare burndown query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("preparing burndown query: %w", err)
	}
	defer burndownStmt.Close()

	var dbBurndowns []dbBurndown
	err = burndownStmt.SelectContext(ctx, &dbBurndowns, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute burndown query", "error", err)
		return reports.CoreStoryAnalytics{}, fmt.Errorf("executing burndown query: %w", err)
	}

	// Convert to core model
	burndown := make([]reports.CoreBurndownPoint, len(dbBurndowns))
	for i, point := range dbBurndowns {
		burndown[i] = reports.CoreBurndownPoint{
			Date:      point.Date,
			Remaining: point.Remaining,
		}
	}

	return reports.CoreStoryAnalytics{
		StatusBreakdown:      statusBreakdown,
		PriorityDistribution: priorityDistribution,
		CompletionByTeam:     completionByTeam,
		Burndown:             burndown,
	}, nil
}

func (r *repo) GetObjectiveProgress(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreObjectiveProgress, error) {
	r.log.Info(ctx, "reportsrepository.GetObjectiveProgress")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetObjectiveProgress")
	defer span.End()

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"start_date":   *filters.StartDate,
		"end_date":     *filters.EndDate,
	}

	// Build filter conditions
	var teamFilter strings.Builder
	var objectiveFilter strings.Builder

	if len(filters.TeamIDs) > 0 {
		teamFilter.WriteString("AND o.team_id = ANY(:team_ids)")
		namedParams["team_ids"] = filters.TeamIDs
	}

	if len(filters.ObjectiveIDs) > 0 {
		objectiveFilter.WriteString("AND o.objective_id = ANY(:objective_ids)")
		namedParams["objective_ids"] = filters.ObjectiveIDs
	}

	// Get health distribution
	healthQuery := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN o.health IS NULL THEN 'Not Set'
				ELSE CAST(o.health AS text)
			END as status,
			COUNT(o.objective_id) as count
		FROM objectives o
		WHERE o.workspace_id = :workspace_id
			AND o.created_at >= :start_date
			AND o.created_at <= :end_date
			%s
			%s
		GROUP BY 
			CASE 
				WHEN o.health IS NULL THEN 'Not Set'
				ELSE CAST(o.health AS text)
			END
		ORDER BY 
			CASE 
				WHEN CASE 
					WHEN o.health IS NULL THEN 'Not Set'
					ELSE CAST(o.health AS text)
				END = 'On Track' THEN 1
				WHEN CASE 
					WHEN o.health IS NULL THEN 'Not Set'
					ELSE CAST(o.health AS text)
				END = 'At Risk' THEN 2
				WHEN CASE 
					WHEN o.health IS NULL THEN 'Not Set'
					ELSE CAST(o.health AS text)
				END = 'Off Track' THEN 3
				WHEN CASE 
					WHEN o.health IS NULL THEN 'Not Set'
					ELSE CAST(o.health AS text)
				END = 'Not Set' THEN 4
				ELSE 5
			END
	`, teamFilter.String(), objectiveFilter.String())

	type dbHealthDistribution struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}

	healthStmt, err := r.db.PrepareNamedContext(ctx, healthQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare health query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("preparing health query: %w", err)
	}
	defer healthStmt.Close()

	var dbHealthDistributions []dbHealthDistribution
	err = healthStmt.SelectContext(ctx, &dbHealthDistributions, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute health query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("executing health query: %w", err)
	}

	// Convert to core model
	healthDistribution := make([]reports.CoreHealthDistributionItem, len(dbHealthDistributions))
	for i, health := range dbHealthDistributions {
		healthDistribution[i] = reports.CoreHealthDistributionItem{
			Status: health.Status,
			Count:  health.Count,
		}
	}

	// Get status breakdown
	statusQuery := fmt.Sprintf(`
		SELECT 
			s.name as status_name,
			COUNT(o.objective_id) as count
		FROM objectives o
		INNER JOIN objective_statuses s ON s.status_id = o.status_id
		WHERE o.workspace_id = :workspace_id
			AND o.created_at >= :start_date
			AND o.created_at <= :end_date
			%s
			%s
		GROUP BY s.status_id, s.name, s.order_index
		ORDER BY s.order_index
	`, teamFilter.String(), objectiveFilter.String())

	type dbObjectiveStatus struct {
		StatusName string `db:"status_name"`
		Count      int    `db:"count"`
	}

	statusStmt, err := r.db.PrepareNamedContext(ctx, statusQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare status query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("preparing status query: %w", err)
	}
	defer statusStmt.Close()

	var dbObjectiveStatuses []dbObjectiveStatus
	err = statusStmt.SelectContext(ctx, &dbObjectiveStatuses, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute status query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("executing status query: %w", err)
	}

	// Convert to core model
	statusBreakdown := make([]reports.CoreObjectiveStatusItem, len(dbObjectiveStatuses))
	for i, status := range dbObjectiveStatuses {
		statusBreakdown[i] = reports.CoreObjectiveStatusItem{
			StatusName: status.StatusName,
			Count:      status.Count,
		}
	}

	// Get key results progress
	keyResultsQuery := fmt.Sprintf(`
		SELECT 
			o.objective_id,
			o.name as objective_name,
			COUNT(kr.id) as total,
			COUNT(CASE 
				WHEN kr.measurement_type = 'percentage' AND kr.current_value >= kr.target_value THEN kr.id
				WHEN kr.measurement_type = 'number' AND kr.current_value >= kr.target_value THEN kr.id
				WHEN kr.measurement_type = 'boolean' AND kr.current_value = kr.target_value THEN kr.id
			END) as completed,
			COALESCE(AVG(
				CASE 
					WHEN kr.measurement_type = 'percentage' THEN LEAST(kr.current_value, 100)
					WHEN kr.measurement_type = 'number' AND kr.target_value != kr.start_value THEN 
						LEAST(((kr.current_value - kr.start_value) / NULLIF(kr.target_value - kr.start_value, 0)) * 100, 100)
					WHEN kr.measurement_type = 'boolean' THEN 
						CASE WHEN kr.current_value = kr.target_value THEN 100 ELSE 0 END
					ELSE 0
				END
			), 0) as avg_progress
		FROM objectives o
		LEFT JOIN key_results kr ON kr.objective_id = o.objective_id
		WHERE o.workspace_id = :workspace_id
			AND o.created_at >= :start_date
			AND o.created_at <= :end_date
			%s
			%s
		GROUP BY o.objective_id, o.name
		ORDER BY o.name
	`, teamFilter.String(), objectiveFilter.String())

	type dbKeyResultProgress struct {
		ObjectiveID   uuid.UUID `db:"objective_id"`
		ObjectiveName string    `db:"objective_name"`
		Total         int       `db:"total"`
		Completed     int       `db:"completed"`
		AvgProgress   float64   `db:"avg_progress"`
	}

	keyResultsStmt, err := r.db.PrepareNamedContext(ctx, keyResultsQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare key results query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("preparing key results query: %w", err)
	}
	defer keyResultsStmt.Close()

	var dbKeyResultProgresses []dbKeyResultProgress
	err = keyResultsStmt.SelectContext(ctx, &dbKeyResultProgresses, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute key results query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("executing key results query: %w", err)
	}

	// Convert to core model
	keyResultsProgress := make([]reports.CoreKeyResultProgressItem, len(dbKeyResultProgresses))
	for i, kr := range dbKeyResultProgresses {
		keyResultsProgress[i] = reports.CoreKeyResultProgressItem{
			ObjectiveID:   kr.ObjectiveID,
			ObjectiveName: kr.ObjectiveName,
			Total:         kr.Total,
			Completed:     kr.Completed,
			AvgProgress:   kr.AvgProgress,
		}
	}

	// Get progress by team
	teamProgressQuery := fmt.Sprintf(`
		SELECT 
			t.team_id,
			t.name as team_name,
			COUNT(o.objective_id) as objectives,
			COUNT(CASE WHEN stat.category = 'completed' THEN o.objective_id END) as completed
		FROM teams t
		LEFT JOIN objectives o ON o.team_id = t.team_id 
			AND o.workspace_id = :workspace_id
			AND o.created_at >= :start_date
			AND o.created_at <= :end_date
			%s
		LEFT JOIN objective_statuses stat ON o.status_id = stat.status_id
		WHERE t.workspace_id = :workspace_id
		GROUP BY t.team_id, t.name
		ORDER BY t.name
	`, objectiveFilter.String())

	type dbObjectiveTeamProgress struct {
		TeamID     uuid.UUID `db:"team_id"`
		TeamName   string    `db:"team_name"`
		Objectives int       `db:"objectives"`
		Completed  int       `db:"completed"`
	}

	teamProgressStmt, err := r.db.PrepareNamedContext(ctx, teamProgressQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare team progress query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("preparing team progress query: %w", err)
	}
	defer teamProgressStmt.Close()

	var dbObjectiveTeamProgresses []dbObjectiveTeamProgress
	err = teamProgressStmt.SelectContext(ctx, &dbObjectiveTeamProgresses, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute team progress query", "error", err)
		return reports.CoreObjectiveProgress{}, fmt.Errorf("executing team progress query: %w", err)
	}

	// Convert to core model
	progressByTeam := make([]reports.CoreObjectiveTeamProgressItem, len(dbObjectiveTeamProgresses))
	for i, team := range dbObjectiveTeamProgresses {
		progressByTeam[i] = reports.CoreObjectiveTeamProgressItem{
			TeamID:     team.TeamID,
			TeamName:   team.TeamName,
			Objectives: team.Objectives,
			Completed:  team.Completed,
		}
	}

	return reports.CoreObjectiveProgress{
		HealthDistribution: healthDistribution,
		StatusBreakdown:    statusBreakdown,
		KeyResultsProgress: keyResultsProgress,
		ProgressByTeam:     progressByTeam,
	}, nil
}

func (r *repo) GetTeamPerformance(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreTeamPerformance, error) {
	r.log.Info(ctx, "reportsrepository.GetTeamPerformance")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetTeamPerformance")
	defer span.End()

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"start_date":   *filters.StartDate,
		"end_date":     *filters.EndDate,
	}

	// Build filter conditions
	teamFilter, _, _ := buildFilters(filters, namedParams)

	// Get team workload
	workloadQuery := fmt.Sprintf(`
		SELECT 
			t.team_id,
			t.name as team_name,
			COUNT(st.id) as assigned,
			COUNT(CASE WHEN stat.category = 'completed' THEN st.id END) as completed,
			COUNT(DISTINCT tm.user_id) * 40 as capacity
		FROM teams t
		LEFT JOIN stories st ON st.team_id = t.team_id 
			AND st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		LEFT JOIN team_members tm ON tm.team_id = t.team_id
		WHERE t.workspace_id = :workspace_id
			%s
		GROUP BY t.team_id, t.name
		ORDER BY t.name
	`, teamFilter)

	type dbTeamWorkload struct {
		TeamID    uuid.UUID `db:"team_id"`
		TeamName  string    `db:"team_name"`
		Assigned  int       `db:"assigned"`
		Completed int       `db:"completed"`
		Capacity  int       `db:"capacity"`
	}

	workloadStmt, err := r.db.PrepareNamedContext(ctx, workloadQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare workload query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("preparing workload query: %w", err)
	}
	defer workloadStmt.Close()

	var dbTeamWorkloads []dbTeamWorkload
	err = workloadStmt.SelectContext(ctx, &dbTeamWorkloads, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute workload query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("executing workload query: %w", err)
	}

	// Convert to core model
	teamWorkload := make([]reports.CoreTeamWorkloadItem, len(dbTeamWorkloads))
	for i, workload := range dbTeamWorkloads {
		teamWorkload[i] = reports.CoreTeamWorkloadItem{
			TeamID:    workload.TeamID,
			TeamName:  workload.TeamName,
			Assigned:  workload.Assigned,
			Completed: workload.Completed,
			Capacity:  workload.Capacity,
		}
	}

	// Get member contributions
	contributionsQuery := fmt.Sprintf(`
		SELECT 
			u.user_id,
			u.username,
			u.avatar_url,
			tm.team_id,
			COUNT(st.id) as assigned,
			COUNT(CASE WHEN stat.category = 'completed' THEN st.id END) as completed
		FROM users u
		INNER JOIN team_members tm ON tm.user_id = u.user_id
		INNER JOIN teams t ON t.team_id = tm.team_id
		LEFT JOIN stories st ON st.assignee_id = u.user_id 
			AND st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		WHERE t.workspace_id = :workspace_id
			%s
		GROUP BY u.user_id, u.username, u.avatar_url, tm.team_id
		ORDER BY u.username
	`, teamFilter)

	type dbMemberContribution struct {
		UserID    uuid.UUID `db:"user_id"`
		Username  string    `db:"username"`
		AvatarURL string    `db:"avatar_url"`
		TeamID    uuid.UUID `db:"team_id"`
		Assigned  int       `db:"assigned"`
		Completed int       `db:"completed"`
	}

	contributionsStmt, err := r.db.PrepareNamedContext(ctx, contributionsQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare contributions query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("preparing contributions query: %w", err)
	}
	defer contributionsStmt.Close()

	var dbMemberContributions []dbMemberContribution
	err = contributionsStmt.SelectContext(ctx, &dbMemberContributions, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute contributions query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("executing contributions query: %w", err)
	}

	// Convert to core model
	memberContributions := make([]reports.CoreMemberContributionItem, len(dbMemberContributions))
	for i, contrib := range dbMemberContributions {
		memberContributions[i] = reports.CoreMemberContributionItem{
			UserID:    contrib.UserID,
			Username:  contrib.Username,
			AvatarURL: contrib.AvatarURL,
			TeamID:    contrib.TeamID,
			Assigned:  contrib.Assigned,
			Completed: contrib.Completed,
		}
	}

	// Get velocity by team (last 3 weeks)
	velocityQuery := fmt.Sprintf(`
		SELECT 
			t.team_id,
			t.name as team_name,
			0 as week1,  -- Simplified - these could be calculated separately if needed
			0 as week2,
			0 as week3,
			ROUND(COUNT(st.id) / 3.0, 2) as average
		FROM teams t
		LEFT JOIN stories st ON st.team_id = t.team_id 
			AND st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND DATE(st.updated_at) >= DATE(:start_date) - INTERVAL '3 weeks'
			AND st.updated_at <= :end_date
		LEFT JOIN statuses stat ON st.status_id = stat.status_id 
			AND stat.category = 'completed'
		WHERE t.workspace_id = :workspace_id
			%s
		GROUP BY t.team_id, t.name
		ORDER BY t.name
	`, teamFilter)

	type dbTeamVelocity struct {
		TeamID   uuid.UUID `db:"team_id"`
		TeamName string    `db:"team_name"`
		Week1    int       `db:"week1"`
		Week2    int       `db:"week2"`
		Week3    int       `db:"week3"`
		Average  float64   `db:"average"`
	}

	velocityStmt, err := r.db.PrepareNamedContext(ctx, velocityQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare velocity query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("preparing velocity query: %w", err)
	}
	defer velocityStmt.Close()

	var dbTeamVelocities []dbTeamVelocity
	err = velocityStmt.SelectContext(ctx, &dbTeamVelocities, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute velocity query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("executing velocity query: %w", err)
	}

	// Convert to core model
	velocityByTeam := make([]reports.CoreTeamVelocityItem, len(dbTeamVelocities))
	for i, velocity := range dbTeamVelocities {
		velocityByTeam[i] = reports.CoreTeamVelocityItem{
			TeamID:   velocity.TeamID,
			TeamName: velocity.TeamName,
			Week1:    velocity.Week1,
			Week2:    velocity.Week2,
			Week3:    velocity.Week3,
			Average:  velocity.Average,
		}
	}

	// Get workload trend (daily)
	trendQuery := fmt.Sprintf(`
		SELECT 
			DATE(st.created_at) as date,
			COUNT(st.id) as assigned,
			COUNT(CASE WHEN stat.category = 'completed' THEN st.id END) as completed
		FROM stories st
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		LEFT JOIN teams t ON t.team_id = st.team_id
		WHERE st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
			%s
		GROUP BY DATE(st.created_at)
		ORDER BY date
	`, teamFilter)

	type dbWorkloadTrend struct {
		Date      time.Time `db:"date"`
		Assigned  int       `db:"assigned"`
		Completed int       `db:"completed"`
	}

	trendStmt, err := r.db.PrepareNamedContext(ctx, trendQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare trend query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("preparing trend query: %w", err)
	}
	defer trendStmt.Close()

	var dbWorkloadTrends []dbWorkloadTrend
	err = trendStmt.SelectContext(ctx, &dbWorkloadTrends, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute trend query", "error", err)
		return reports.CoreTeamPerformance{}, fmt.Errorf("executing trend query: %w", err)
	}

	// Convert to core model
	workloadTrend := make([]reports.CoreWorkloadTrendPoint, len(dbWorkloadTrends))
	for i, trend := range dbWorkloadTrends {
		workloadTrend[i] = reports.CoreWorkloadTrendPoint{
			Date:      trend.Date,
			Assigned:  trend.Assigned,
			Completed: trend.Completed,
		}
	}

	return reports.CoreTeamPerformance{
		TeamWorkload:        teamWorkload,
		MemberContributions: memberContributions,
		VelocityByTeam:      velocityByTeam,
		WorkloadTrend:       workloadTrend,
	}, nil
}

func (r *repo) GetSprintAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreSprintAnalyticsWorkspace, error) {
	r.log.Info(ctx, "reportsrepository.GetSprintAnalytics")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetSprintAnalytics")
	defer span.End()

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"start_date":   *filters.StartDate,
		"end_date":     *filters.EndDate,
	}

	// Build filter conditions
	var teamFilter strings.Builder
	var sprintFilter strings.Builder

	if len(filters.TeamIDs) > 0 {
		teamFilter.WriteString("AND s.team_id = ANY(:team_ids)")
		namedParams["team_ids"] = filters.TeamIDs
	}

	if len(filters.SprintIDs) > 0 {
		sprintFilter.WriteString("AND s.sprint_id = ANY(:sprint_ids)")
		namedParams["sprint_ids"] = filters.SprintIDs
	}

	// Get sprint progress
	progressQuery := fmt.Sprintf(`
		SELECT 
			s.sprint_id,
			s.name as sprint_name,
			s.team_id,
			COUNT(st.id) as total,
			COUNT(CASE WHEN stat.category = 'completed' THEN st.id END) as completed,
			CASE 
				WHEN s.end_date < CURRENT_DATE THEN 'Completed'
				WHEN s.start_date > CURRENT_DATE THEN 'Not Started'
				WHEN s.start_date <= CURRENT_DATE AND s.end_date >= CURRENT_DATE THEN 'Active'
				ELSE 'Unknown'
			END as status
		FROM sprints s
		LEFT JOIN stories st ON st.sprint_id = s.sprint_id 
			AND st.deleted_at IS NULL
			AND st.is_draft = false
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		WHERE s.workspace_id = :workspace_id
			AND s.created_at >= :start_date
			AND s.created_at <= :end_date
			%s
			%s
		GROUP BY s.sprint_id, s.name, s.team_id, s.start_date, s.end_date
		ORDER BY s.name
	`, teamFilter.String(), sprintFilter.String())

	type dbSprintProgress struct {
		SprintID   uuid.UUID `db:"sprint_id"`
		SprintName string    `db:"sprint_name"`
		TeamID     uuid.UUID `db:"team_id"`
		Total      int       `db:"total"`
		Completed  int       `db:"completed"`
		Status     string    `db:"status"`
	}

	progressStmt, err := r.db.PrepareNamedContext(ctx, progressQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare progress query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("preparing progress query: %w", err)
	}
	defer progressStmt.Close()

	var dbSprintProgresses []dbSprintProgress
	err = progressStmt.SelectContext(ctx, &dbSprintProgresses, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute progress query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("executing progress query: %w", err)
	}

	// Convert to core model
	sprintProgress := make([]reports.CoreSprintProgressItem, len(dbSprintProgresses))
	for i, progress := range dbSprintProgresses {
		sprintProgress[i] = reports.CoreSprintProgressItem{
			SprintID:   progress.SprintID,
			SprintName: progress.SprintName,
			TeamID:     progress.TeamID,
			Total:      progress.Total,
			Completed:  progress.Completed,
			Status:     progress.Status,
		}
	}

	// Get combined burndown (daily across all sprints)
	burndownQuery := fmt.Sprintf(`
		WITH daily_burndown AS (
			SELECT 
				DATE(st.updated_at) as date,
				COUNT(st.id) as actual,
				COUNT(st.id) FILTER (WHERE st.created_at <= DATE(st.updated_at)) as planned
			FROM stories st
			INNER JOIN sprints s ON s.sprint_id = st.sprint_id
			LEFT JOIN statuses stat ON st.status_id = stat.status_id
			WHERE s.workspace_id = :workspace_id
				AND st.deleted_at IS NULL
				AND st.is_draft = false
				AND stat.category = 'completed'
				AND st.updated_at >= :start_date
				AND st.updated_at <= :end_date
				%s
				%s
			GROUP BY DATE(st.updated_at)
			ORDER BY date
		)
		SELECT date, planned, actual FROM daily_burndown
	`, teamFilter.String(), sprintFilter.String())

	type dbCombinedBurndown struct {
		Date    time.Time `db:"date"`
		Planned int       `db:"planned"`
		Actual  int       `db:"actual"`
	}

	burndownStmt, err := r.db.PrepareNamedContext(ctx, burndownQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare burndown query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("preparing burndown query: %w", err)
	}
	defer burndownStmt.Close()

	var dbCombinedBurndowns []dbCombinedBurndown
	err = burndownStmt.SelectContext(ctx, &dbCombinedBurndowns, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute burndown query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("executing burndown query: %w", err)
	}

	// Convert to core model
	combinedBurndown := make([]reports.CoreCombinedBurndownPoint, len(dbCombinedBurndowns))
	for i, burndown := range dbCombinedBurndowns {
		combinedBurndown[i] = reports.CoreCombinedBurndownPoint{
			Date:    burndown.Date,
			Planned: burndown.Planned,
			Actual:  burndown.Actual,
		}
	}

	// Get team allocation
	allocationQuery := fmt.Sprintf(`
		SELECT 
			t.team_id,
			t.name as team_name,
			COUNT(DISTINCT s.sprint_id) as active_sprints,
			COUNT(st.id) as total_stories,
			COUNT(CASE WHEN stat.category = 'completed' THEN st.id END) as completed_stories
		FROM teams t
		LEFT JOIN sprints s ON s.team_id = t.team_id 
			AND s.workspace_id = :workspace_id
			AND s.created_at >= :start_date
			AND s.created_at <= :end_date
			%s
		LEFT JOIN stories st ON st.sprint_id = s.sprint_id 
			AND st.deleted_at IS NULL
			AND st.is_draft = false
		LEFT JOIN statuses stat ON st.status_id = stat.status_id
		WHERE t.workspace_id = :workspace_id
		GROUP BY t.team_id, t.name
		ORDER BY t.name
	`, sprintFilter.String())

	type dbSprintTeamAllocation struct {
		TeamID           uuid.UUID `db:"team_id"`
		TeamName         string    `db:"team_name"`
		ActiveSprints    int       `db:"active_sprints"`
		TotalStories     int       `db:"total_stories"`
		CompletedStories int       `db:"completed_stories"`
	}

	allocationStmt, err := r.db.PrepareNamedContext(ctx, allocationQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare allocation query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("preparing allocation query: %w", err)
	}
	defer allocationStmt.Close()

	var dbSprintTeamAllocations []dbSprintTeamAllocation
	err = allocationStmt.SelectContext(ctx, &dbSprintTeamAllocations, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute allocation query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("executing allocation query: %w", err)
	}

	// Convert to core model
	teamAllocation := make([]reports.CoreSprintTeamAllocation, len(dbSprintTeamAllocations))
	for i, allocation := range dbSprintTeamAllocations {
		teamAllocation[i] = reports.CoreSprintTeamAllocation{
			TeamID:           allocation.TeamID,
			TeamName:         allocation.TeamName,
			ActiveSprints:    allocation.ActiveSprints,
			TotalStories:     allocation.TotalStories,
			CompletedStories: allocation.CompletedStories,
		}
	}

	// Get sprint health
	healthQuery := fmt.Sprintf(`
		WITH sprint_status AS (
			SELECT 
				CASE 
					WHEN s.end_date < CURRENT_DATE THEN 'Completed'
					WHEN s.start_date > CURRENT_DATE THEN 'Not Started'
					WHEN s.end_date <= CURRENT_DATE + INTERVAL '3 days' THEN 'At Risk'
					WHEN s.start_date <= CURRENT_DATE AND s.end_date >= CURRENT_DATE THEN 'On Track'
					ELSE 'Unknown'
				END as status
			FROM sprints s
			WHERE s.workspace_id = :workspace_id
				AND s.created_at >= :start_date
				AND s.created_at <= :end_date
				%s
				%s
		)
		SELECT 
			status,
			COUNT(*) as count
		FROM sprint_status
		GROUP BY status
		ORDER BY 
			CASE status
				WHEN 'On Track' THEN 1
				WHEN 'At Risk' THEN 2
				WHEN 'Not Started' THEN 3
				WHEN 'Completed' THEN 4
				ELSE 5
			END
	`, teamFilter.String(), sprintFilter.String())

	type dbSprintHealth struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}

	healthStmt, err := r.db.PrepareNamedContext(ctx, healthQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare health query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("preparing health query: %w", err)
	}
	defer healthStmt.Close()

	var dbSprintHealths []dbSprintHealth
	err = healthStmt.SelectContext(ctx, &dbSprintHealths, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute health query", "error", err)
		return reports.CoreSprintAnalyticsWorkspace{}, fmt.Errorf("executing health query: %w", err)
	}

	// Convert to core model
	sprintHealth := make([]reports.CoreSprintHealthItem, len(dbSprintHealths))
	for i, health := range dbSprintHealths {
		sprintHealth[i] = reports.CoreSprintHealthItem{
			Status: health.Status,
			Count:  health.Count,
		}
	}

	return reports.CoreSprintAnalyticsWorkspace{
		SprintProgress:   sprintProgress,
		CombinedBurndown: combinedBurndown,
		TeamAllocation:   teamAllocation,
		SprintHealth:     sprintHealth,
	}, nil
}

func (r *repo) GetTimelineTrends(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreTimelineTrends, error) {
	r.log.Info(ctx, "reportsrepository.GetTimelineTrends")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetTimelineTrends")
	defer span.End()

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"start_date":   *filters.StartDate,
		"end_date":     *filters.EndDate,
	}

	// Build filter conditions
	teamFilter, sprintFilter, objectiveFilter := buildFilters(filters, namedParams)

	// Get story completion timeline
	storyCompletionQuery := fmt.Sprintf(`
		WITH daily_stats AS (
			SELECT 
				DATE(st.created_at) as date,
				COUNT(st.id) as created,
				COUNT(CASE 
					WHEN stat.category = 'completed' AND DATE(st.updated_at) = DATE(st.created_at) 
					THEN st.id 
				END) as completed
			FROM stories st
			LEFT JOIN statuses stat ON st.status_id = stat.status_id
			WHERE st.workspace_id = :workspace_id
				AND st.deleted_at IS NULL
				AND st.is_draft = false
				AND st.created_at >= :start_date
				AND st.created_at <= :end_date
				%s
				%s
				%s
			GROUP BY DATE(st.created_at)
		)
		SELECT date, created, COALESCE(completed, 0) as completed
		FROM daily_stats
		ORDER BY date
	`, teamFilter, sprintFilter, objectiveFilter)

	type dbStoryCompletion struct {
		Date      time.Time `db:"date"`
		Created   int       `db:"created"`
		Completed int       `db:"completed"`
	}

	storyCompletionStmt, err := r.db.PrepareNamedContext(ctx, storyCompletionQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare story completion query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("preparing story completion query: %w", err)
	}
	defer storyCompletionStmt.Close()

	var dbStoryCompletions []dbStoryCompletion
	err = storyCompletionStmt.SelectContext(ctx, &dbStoryCompletions, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute story completion query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("executing story completion query: %w", err)
	}

	// Convert to core model
	storyCompletion := make([]reports.CoreStoryCompletionPoint, len(dbStoryCompletions))
	for i, completion := range dbStoryCompletions {
		storyCompletion[i] = reports.CoreStoryCompletionPoint{
			Date:      completion.Date,
			Created:   completion.Created,
			Completed: completion.Completed,
		}
	}

	// Get objective progress timeline
	objectiveProgressQuery := fmt.Sprintf(`
		SELECT 
			DATE(o.created_at) as date,
			COUNT(o.objective_id) as total_objectives,
			COUNT(CASE WHEN stat.category = 'completed' THEN o.objective_id END) as completed_objectives
		FROM objectives o
		LEFT JOIN objective_statuses stat ON o.status_id = stat.status_id
		WHERE o.workspace_id = :workspace_id
			AND o.created_at >= :start_date
			AND o.created_at <= :end_date
			%s
			%s
		GROUP BY DATE(o.created_at)
		ORDER BY date
	`, teamFilter, objectiveFilter)

	type dbObjectiveProgress struct {
		Date                time.Time `db:"date"`
		TotalObjectives     int       `db:"total_objectives"`
		CompletedObjectives int       `db:"completed_objectives"`
	}

	objectiveProgressStmt, err := r.db.PrepareNamedContext(ctx, objectiveProgressQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare objective progress query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("preparing objective progress query: %w", err)
	}
	defer objectiveProgressStmt.Close()

	var dbObjectiveProgresses []dbObjectiveProgress
	err = objectiveProgressStmt.SelectContext(ctx, &dbObjectiveProgresses, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute objective progress query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("executing objective progress query: %w", err)
	}

	// Convert to core model
	objectiveProgress := make([]reports.CoreObjectiveProgressPoint, len(dbObjectiveProgresses))
	for i, progress := range dbObjectiveProgresses {
		objectiveProgress[i] = reports.CoreObjectiveProgressPoint{
			Date:                progress.Date,
			TotalObjectives:     progress.TotalObjectives,
			CompletedObjectives: progress.CompletedObjectives,
		}
	}

	// Get team velocity timeline
	teamVelocityQuery := fmt.Sprintf(`
		SELECT 
			DATE(st.updated_at) as date,
			st.team_id,
			COUNT(st.id) as velocity
		FROM stories st
		INNER JOIN statuses stat ON st.status_id = stat.status_id
		WHERE st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND stat.category = 'completed'
			AND st.updated_at >= :start_date
			AND st.updated_at <= :end_date
			%s
			%s
			%s
		GROUP BY DATE(st.updated_at), st.team_id
		ORDER BY date, team_id
	`, teamFilter, sprintFilter, objectiveFilter)

	type dbTeamVelocityPoint struct {
		Date     time.Time `db:"date"`
		TeamID   uuid.UUID `db:"team_id"`
		Velocity int       `db:"velocity"`
	}

	teamVelocityStmt, err := r.db.PrepareNamedContext(ctx, teamVelocityQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare team velocity query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("preparing team velocity query: %w", err)
	}
	defer teamVelocityStmt.Close()

	var dbTeamVelocityPoints []dbTeamVelocityPoint
	err = teamVelocityStmt.SelectContext(ctx, &dbTeamVelocityPoints, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute team velocity query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("executing team velocity query: %w", err)
	}

	// Convert to core model
	teamVelocity := make([]reports.CoreTeamVelocityPoint, len(dbTeamVelocityPoints))
	for i, velocity := range dbTeamVelocityPoints {
		teamVelocity[i] = reports.CoreTeamVelocityPoint{
			Date:     velocity.Date,
			TeamID:   velocity.TeamID,
			Velocity: velocity.Velocity,
		}
	}

	// Get key metrics trend
	keyMetricsQuery := fmt.Sprintf(`
		SELECT 
			DATE(st.created_at) as date,
			COUNT(DISTINCT st.assignee_id) as active_users,
			COUNT(st.id) as stories_per_day,
			ROUND(COALESCE(AVG(EXTRACT(EPOCH FROM (st.updated_at - st.created_at)) / 86400), 0), 2) as avg_cycle_time
		FROM stories st
		WHERE st.workspace_id = :workspace_id
			AND st.deleted_at IS NULL
			AND st.is_draft = false
			AND st.created_at >= :start_date
			AND st.created_at <= :end_date
			%s
			%s
			%s
		GROUP BY DATE(st.created_at)
		ORDER BY date
	`, teamFilter, sprintFilter, objectiveFilter)

	type dbKeyMetricsTrend struct {
		Date          time.Time `db:"date"`
		ActiveUsers   int       `db:"active_users"`
		StoriesPerDay float64   `db:"stories_per_day"`
		AvgCycleTime  float64   `db:"avg_cycle_time"`
	}

	keyMetricsStmt, err := r.db.PrepareNamedContext(ctx, keyMetricsQuery)
	if err != nil {
		r.log.Error(ctx, "failed to prepare key metrics query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("preparing key metrics query: %w", err)
	}
	defer keyMetricsStmt.Close()

	var dbKeyMetricsTrends []dbKeyMetricsTrend
	err = keyMetricsStmt.SelectContext(ctx, &dbKeyMetricsTrends, namedParams)
	if err != nil {
		r.log.Error(ctx, "failed to execute key metrics query", "error", err)
		return reports.CoreTimelineTrends{}, fmt.Errorf("executing key metrics query: %w", err)
	}

	// Convert to core model
	keyMetricsTrend := make([]reports.CoreKeyMetricsTrendPoint, len(dbKeyMetricsTrends))
	for i, metrics := range dbKeyMetricsTrends {
		keyMetricsTrend[i] = reports.CoreKeyMetricsTrendPoint{
			Date:          metrics.Date,
			ActiveUsers:   metrics.ActiveUsers,
			StoriesPerDay: metrics.StoriesPerDay,
			AvgCycleTime:  metrics.AvgCycleTime,
		}
	}

	return reports.CoreTimelineTrends{
		StoryCompletion:   storyCompletion,
		ObjectiveProgress: objectiveProgress,
		TeamVelocity:      teamVelocity,
		KeyMetricsTrend:   keyMetricsTrend,
	}, nil
}
