package reportsrepository

import (
	"context"
	"fmt"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type dbRequestProviderPerformance struct {
	Provider         string  `db:"provider"`
	TotalRequests    int     `db:"total_requests"`
	PendingRequests  int     `db:"pending_requests"`
	AcceptedRequests int     `db:"accepted_requests"`
	DeclinedRequests int     `db:"declined_requests"`
	UrgentRequests   int     `db:"urgent_requests"`
	HighRequests     int     `db:"high_requests"`
	StaleRequests    int     `db:"stale_requests"`
	AcceptanceRate   float64 `db:"acceptance_rate"`
}

type dbWorkspaceEngagementTotals struct {
	TotalEvents int `db:"total_events"`
	UniqueUsers int `db:"unique_users"`
}

type dbWorkspaceEngagementCount struct {
	Name  string `db:"name"`
	Count int    `db:"count"`
}

type dbWorkspaceEngagementUser struct {
	UserID    uuid.UUID `db:"user_id"`
	FullName  string    `db:"full_name"`
	Username  string    `db:"username"`
	AvatarURL string    `db:"avatar_url"`
	Events    int       `db:"events"`
}

func (r *repo) GetRequestSourceAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreRequestSourceAnalytics, error) {
	r.log.Info(ctx, "reportsrepository.GetRequestSourceAnalytics")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetRequestSourceAnalytics")
	defer span.End()

	namedParams := map[string]any{"workspace_id": workspaceID}
	where := buildRequestSourceFilter(filters, namedParams)
	query := fmt.Sprintf(`
		SELECT
			ir.provider,
			CAST(COUNT(*) AS int) AS total_requests,
			CAST(COUNT(*) FILTER (WHERE ir.status = 'pending') AS int) AS pending_requests,
			CAST(COUNT(*) FILTER (WHERE ir.status = 'accepted') AS int) AS accepted_requests,
			CAST(COUNT(*) FILTER (WHERE ir.status = 'declined') AS int) AS declined_requests,
			CAST(COUNT(*) FILTER (WHERE ir.priority = 'Urgent') AS int) AS urgent_requests,
			CAST(COUNT(*) FILTER (WHERE ir.priority = 'High') AS int) AS high_requests,
			CAST(COUNT(*) FILTER (
				WHERE ir.status = 'pending'
				  AND ir.created_at < NOW() - INTERVAL '7 days'
			) AS int) AS stale_requests,
			COALESCE(
				CAST(COUNT(*) FILTER (WHERE ir.status = 'accepted') AS float)
				/ NULLIF(COUNT(*), 0),
				0
			) AS acceptance_rate
		FROM integration_requests ir
		WHERE ir.workspace_id = :workspace_id
			%s
		GROUP BY ir.provider
		ORDER BY total_requests DESC, ir.provider ASC
	`, where)

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return reports.CoreRequestSourceAnalytics{}, fmt.Errorf("preparing request source query: %w", err)
	}
	defer stmt.Close()

	var rows []dbRequestProviderPerformance
	if err := stmt.SelectContext(ctx, &rows, namedParams); err != nil {
		return reports.CoreRequestSourceAnalytics{}, fmt.Errorf("selecting request source analytics: %w", err)
	}

	result := reports.CoreRequestSourceAnalytics{
		Providers: make([]reports.CoreRequestProviderPerformance, len(rows)),
	}
	for i, row := range rows {
		result.Providers[i] = reports.CoreRequestProviderPerformance{
			Provider:         row.Provider,
			TotalRequests:    row.TotalRequests,
			PendingRequests:  row.PendingRequests,
			AcceptedRequests: row.AcceptedRequests,
			DeclinedRequests: row.DeclinedRequests,
			UrgentRequests:   row.UrgentRequests,
			HighRequests:     row.HighRequests,
			StaleRequests:    row.StaleRequests,
			AcceptanceRate:   row.AcceptanceRate,
		}
		result.TotalRequests += row.TotalRequests
		result.PendingRequests += row.PendingRequests
		result.AcceptedRequests += row.AcceptedRequests
		result.DeclinedRequests += row.DeclinedRequests
	}

	return result, nil
}

func (r *repo) GetWorkspaceEngagementAnalytics(ctx context.Context, workspaceID uuid.UUID, filters reports.ReportFilters) (reports.CoreWorkspaceEngagementAnalytics, error) {
	r.log.Info(ctx, "reportsrepository.GetWorkspaceEngagementAnalytics")
	ctx, span := web.AddSpan(ctx, "reportsrepository.GetWorkspaceEngagementAnalytics")
	defer span.End()

	namedParams := map[string]any{"workspace_id": workspaceID}
	where := buildWorkspaceEngagementFilter(filters, namedParams)
	baseWhere := fmt.Sprintf("wae.workspace_id = :workspace_id %s", where)

	totals, err := r.getWorkspaceEngagementTotals(ctx, baseWhere, namedParams)
	if err != nil {
		return reports.CoreWorkspaceEngagementAnalytics{}, err
	}
	eventsByName, err := r.getWorkspaceEngagementCounts(ctx, baseWhere, namedParams, "wae.event_name")
	if err != nil {
		return reports.CoreWorkspaceEngagementAnalytics{}, err
	}
	eventsBySurface, err := r.getWorkspaceEngagementCounts(ctx, baseWhere, namedParams, "wae.surface")
	if err != nil {
		return reports.CoreWorkspaceEngagementAnalytics{}, err
	}
	topUsers, err := r.getWorkspaceEngagementTopUsers(ctx, baseWhere, namedParams)
	if err != nil {
		return reports.CoreWorkspaceEngagementAnalytics{}, err
	}

	return reports.CoreWorkspaceEngagementAnalytics{
		TotalEvents:     totals.TotalEvents,
		UniqueUsers:     totals.UniqueUsers,
		EventsByName:    eventsByName,
		EventsBySurface: eventsBySurface,
		TopUsers:        topUsers,
	}, nil
}

func (r *repo) getWorkspaceEngagementTotals(ctx context.Context, where string, namedParams map[string]any) (dbWorkspaceEngagementTotals, error) {
	query := fmt.Sprintf(`
		SELECT
			CAST(COUNT(*) AS int) AS total_events,
			CAST(COUNT(DISTINCT wae.user_id) AS int) AS unique_users
		FROM workspace_analytics_events wae
		WHERE %s
	`, where)

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return dbWorkspaceEngagementTotals{}, fmt.Errorf("preparing engagement totals query: %w", err)
	}
	defer stmt.Close()

	var totals dbWorkspaceEngagementTotals
	if err := stmt.GetContext(ctx, &totals, namedParams); err != nil {
		return dbWorkspaceEngagementTotals{}, fmt.Errorf("selecting engagement totals: %w", err)
	}
	return totals, nil
}

func (r *repo) getWorkspaceEngagementCounts(ctx context.Context, where string, namedParams map[string]any, dimension string) ([]reports.CoreWorkspaceEngagementCount, error) {
	query := fmt.Sprintf(`
		SELECT
			%s AS name,
			CAST(COUNT(*) AS int) AS count
		FROM workspace_analytics_events wae
		WHERE %s
		GROUP BY %s
		ORDER BY count DESC, name ASC
		LIMIT 20
	`, dimension, where, dimension)

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("preparing engagement count query: %w", err)
	}
	defer stmt.Close()

	var rows []dbWorkspaceEngagementCount
	if err := stmt.SelectContext(ctx, &rows, namedParams); err != nil {
		return nil, fmt.Errorf("selecting engagement counts: %w", err)
	}

	result := make([]reports.CoreWorkspaceEngagementCount, len(rows))
	for i, row := range rows {
		result[i] = reports.CoreWorkspaceEngagementCount{
			Name:  row.Name,
			Count: row.Count,
		}
	}
	return result, nil
}

func (r *repo) getWorkspaceEngagementTopUsers(ctx context.Context, where string, namedParams map[string]any) ([]reports.CoreWorkspaceEngagementUser, error) {
	query := fmt.Sprintf(`
		SELECT
			u.user_id,
			COALESCE(u.full_name, '') AS full_name,
			u.username,
			COALESCE(u.avatar_url, '') AS avatar_url,
			CAST(COUNT(*) AS int) AS events
		FROM workspace_analytics_events wae
		JOIN users u ON u.user_id = wae.user_id
		WHERE %s
		GROUP BY u.user_id, u.full_name, u.username, u.avatar_url
		ORDER BY events DESC, u.username ASC
		LIMIT 10
	`, where)

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("preparing engagement top users query: %w", err)
	}
	defer stmt.Close()

	var rows []dbWorkspaceEngagementUser
	if err := stmt.SelectContext(ctx, &rows, namedParams); err != nil {
		return nil, fmt.Errorf("selecting engagement top users: %w", err)
	}

	result := make([]reports.CoreWorkspaceEngagementUser, len(rows))
	for i, row := range rows {
		result[i] = reports.CoreWorkspaceEngagementUser{
			UserID:    row.UserID,
			FullName:  row.FullName,
			Username:  row.Username,
			AvatarURL: row.AvatarURL,
			Events:    row.Events,
		}
	}
	return result, nil
}
