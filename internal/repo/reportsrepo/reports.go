package reportsrepo

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/reports"
	"github.com/complexus-tech/projects-api/internal/web/mid"
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

// GetStoryStats gets story statistics for a workspace.
func (r *repo) GetStoryStats(ctx context.Context, workspaceID uuid.UUID) (reports.CoreStoryStats, error) {
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
		)
		SELECT * FROM story_stats`

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		r.log.Error(ctx, "failed to get user id", "error", err)
		return reports.CoreStoryStats{}, fmt.Errorf("getting user id: %w", err)
	}

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
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
func (r *repo) GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, days int) ([]reports.CoreContributionStats, error) {
	const query = `
		WITH RECURSIVE dates AS (
			SELECT CURRENT_DATE - make_interval(days => :days) as date
			UNION ALL
			SELECT date + INTERVAL '1 day'
			FROM dates
			WHERE date < CURRENT_DATE
		),
		activity_counts AS (
			SELECT 
				DATE(created_at) as date,
				COUNT(*) as contributions
			FROM story_activities
			WHERE user_id = :user_id
			AND workspace_id = :workspace_id
			AND created_at >= CURRENT_DATE - make_interval(days => :days)
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
		"days":         days,
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
				COUNT(CASE WHEN creator_id = :user_id THEN 1 END) as created_by_me
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
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		r.log.Error(ctx, "failed to get user id", "error", err)
		return nil, fmt.Errorf("getting user id: %w", err)
	}

	query := `
		WITH user_teams AS (
			SELECT team_id 
			FROM team_members 
			WHERE user_id = $2
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
			WHERE s.workspace_id = $1
			GROUP BY s.status_id, s.name, s.order_index
			ORDER BY s.order_index
		)
		SELECT 
			name,
			count::integer
		FROM story_stats`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, userID)
	if err != nil {
		r.log.Error(ctx, "failed to execute query", "error", err)
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer rows.Close()

	var stats []dbStatusStats
	for rows.Next() {
		var stat dbStatusStats
		if err := rows.Scan(&stat.Name, &stat.Count); err != nil {
			r.log.Error(ctx, "failed to scan row", "error", err)
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		r.log.Error(ctx, "error iterating rows", "error", err)
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return toCoreStatusStats(stats), nil
}

func (r *repo) GetPriorityStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StatsFilters) ([]reports.CorePriorityStats, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		r.log.Error(ctx, "failed to get user id", "error", err)
		return nil, fmt.Errorf("getting user id: %w", err)
	}

	query := `
		WITH priority_stats AS (
			SELECT 
				COALESCE(priority, 'No Priority') as priority,
				COUNT(*) as count
			FROM stories
			WHERE workspace_id = $1
				AND deleted_at IS NULL
				AND is_draft = false
				AND team_id IN (
					SELECT team_id 
					FROM team_members 
					WHERE user_id = $2
				)
			GROUP BY priority
		)
		SELECT 
			priority,
			count::integer
		FROM priority_stats
		ORDER BY 
			CASE priority
				WHEN 'Urgent' THEN 1
				WHEN 'High' THEN 2
				WHEN 'Medium' THEN 3
				WHEN 'Low' THEN 4
				WHEN 'No Priority' THEN 5
			END`

	rows, err := r.db.QueryContext(ctx, query, workspaceID, userID)
	if err != nil {
		r.log.Error(ctx, "failed to execute query", "error", err)
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer rows.Close()

	var stats []dbPriorityStats
	for rows.Next() {
		var stat dbPriorityStats
		if err := rows.Scan(&stat.Priority, &stat.Count); err != nil {
			r.log.Error(ctx, "failed to scan row", "error", err)
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		r.log.Error(ctx, "error iterating rows", "error", err)
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return toCorePriorityStats(stats), nil
}
