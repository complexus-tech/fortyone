package searchrepository

import (
	"context"
	"fmt"
	"strings"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// SearchStories searches for stories based on the provided parameters.
func (r *repo) SearchStories(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, params search.SearchParams) ([]search.CoreSearchStory, int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.search.SearchStories")
	defer span.End()

	span.SetAttributes(
		attribute.String("workspace.id", workspaceID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("search.query", params.Query),
	)

	// Build the main search query
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.parent_id,
			s.objective_id,
			s.status_id,
			s.assignee_id,
			s.reporter_id,
			s.priority,
			s.sprint_id,
			s.team_id,
			s.workspace_id,
			s.start_date,
			s.end_date,
			s.created_at,
			s.updated_at,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), '[]'
			) AS labels
		FROM
			stories s
			INNER JOIN team_members tm ON tm.team_id = s.team_id AND tm.user_id = :user_id
		WHERE
			s.workspace_id = :workspace_id
			AND s.deleted_at IS NULL
	`)

	// Add search condition if query is provided
	if params.Query != "" {
		queryBuilder.WriteString(`
			AND (
				s.search_vector @@ plainto_tsquery('english', :query)
				OR s.title ILIKE '%' || :query || '%'
			)
		`)
	}

	// Add filters
	if params.TeamID != nil {
		queryBuilder.WriteString(`
			AND s.team_id = :team_id
		`)
	}

	if params.AssigneeID != nil {
		queryBuilder.WriteString(`
			AND s.assignee_id = :assignee_id
		`)
	}

	if params.StatusID != nil {
		queryBuilder.WriteString(`
			AND s.status_id = :status_id
		`)
	}

	if params.Priority != nil {
		queryBuilder.WriteString(`
			AND s.priority = :priority
		`)
	}

	if params.LabelID != nil {
		queryBuilder.WriteString(`
			AND EXISTS (
				SELECT 1 FROM story_labels sl
				WHERE sl.story_id = s.id AND sl.label_id = :label_id
			)
		`)
	}

	// Add order by based on sort option
	switch params.SortBy {
	case search.SortByUpdated:
		queryBuilder.WriteString(`ORDER BY s.updated_at DESC`)
	case search.SortByCreated:
		queryBuilder.WriteString(`ORDER BY s.created_at DESC`)
	default: // SortByRelevance or empty
		if params.Query != "" {
			queryBuilder.WriteString(`
				ORDER BY 
					ts_rank(s.search_vector, plainto_tsquery('english', :query)) DESC,
					s.created_at DESC
			`)
		} else {
			queryBuilder.WriteString(`ORDER BY s.created_at DESC`)
		}
	}

	// Add pagination
	queryBuilder.WriteString(`
		LIMIT :page_size OFFSET :offset
	`)

	// Build count query for pagination
	countQuery := `
		SELECT COUNT(*)
		FROM stories s
		INNER JOIN team_members tm ON tm.team_id = s.team_id AND tm.user_id = :user_id
		WHERE s.workspace_id = :workspace_id
		AND s.deleted_at IS NULL
	`

	if params.Query != "" {
		countQuery += `
			AND (
				s.search_vector @@ plainto_tsquery('english', :query)
				OR s.title ILIKE '%' || :query || '%'
			)
		`
	}

	if params.TeamID != nil {
		countQuery += `
			AND s.team_id = :team_id
		`
	}

	if params.AssigneeID != nil {
		countQuery += `
			AND s.assignee_id = :assignee_id
		`
	}

	if params.StatusID != nil {
		countQuery += `
			AND s.status_id = :status_id
		`
	}

	if params.Priority != nil {
		countQuery += `
			AND s.priority = :priority
		`
	}

	if params.LabelID != nil {
		countQuery += `
			AND EXISTS (
				SELECT 1 FROM story_labels sl
				WHERE sl.story_id = s.id AND sl.label_id = :label_id
			)
		`
	}

	// Prepare named parameters
	namedParams := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"query":        params.Query,
		"team_id":      params.TeamID,
		"assignee_id":  params.AssigneeID,
		"status_id":    params.StatusID,
		"priority":     params.Priority,
		"label_id":     params.LabelID,
		"page_size":    params.PageSize,
		"offset":       (params.Page - 1) * params.PageSize,
	}

	// Execute count query
	var totalStories int
	countStmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare count query: %s", err))
		return nil, 0, err
	}
	defer countStmt.Close()

	err = countStmt.GetContext(ctx, &totalStories, namedParams)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to execute count query: %s", err))
		return nil, 0, err
	}

	// If no stories found, return empty result
	if totalStories == 0 {
		return []search.CoreSearchStory{}, 0, nil
	}

	// Execute main query
	stmt, err := r.db.PrepareNamedContext(ctx, queryBuilder.String())
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare search query: %s", err))
		return nil, 0, err
	}
	defer stmt.Close()

	var stories []dbStory
	err = stmt.SelectContext(ctx, &stories, namedParams)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to execute search query: %s", err))
		return nil, 0, err
	}

	// Convert to core model
	return toCoreSearchStories(stories), totalStories, nil
}

// SearchObjectives searches for objectives based on the provided parameters.
func (r *repo) SearchObjectives(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, params search.SearchParams) ([]search.CoreSearchObjective, int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.search.SearchObjectives")
	defer span.End()

	span.SetAttributes(
		attribute.String("workspace.id", workspaceID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("search.query", params.Query),
	)

	// Build the main search query
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT
			o.objective_id,
			o.name,
			o.description,
			o.lead_user_id,
			o.team_id,
			o.workspace_id,
			o.start_date,
			o.end_date,
			o.status_id,
			o.priority,
			o.health,
			o.created_at,
			o.updated_at
		FROM
			objectives o
			INNER JOIN team_members tm ON tm.team_id = o.team_id AND tm.user_id = :user_id
		WHERE
			o.workspace_id = :workspace_id
	`)

	// Add search condition if query is provided
	if params.Query != "" {
		queryBuilder.WriteString(`
			AND (
				o.search_vector @@ plainto_tsquery('english', :query)
				OR o.name ILIKE '%' || :query || '%'
			)
		`)
	}

	// Add filters
	if params.TeamID != nil {
		queryBuilder.WriteString(`
			AND o.team_id = :team_id
		`)
	}

	if params.StatusID != nil {
		queryBuilder.WriteString(`
			AND o.status_id = :status_id
		`)
	}

	// Add order by based on sort option
	switch params.SortBy {
	case search.SortByUpdated:
		queryBuilder.WriteString(`ORDER BY o.updated_at DESC`)
	case search.SortByCreated:
		queryBuilder.WriteString(`ORDER BY o.created_at DESC`)
	default: // SortByRelevance or empty
		if params.Query != "" {
			queryBuilder.WriteString(`
				ORDER BY 
					ts_rank(o.search_vector, plainto_tsquery('english', :query)) DESC,
					o.created_at DESC
			`)
		} else {
			queryBuilder.WriteString(`ORDER BY o.created_at DESC`)
		}
	}

	// Add pagination
	queryBuilder.WriteString(`
		LIMIT :page_size OFFSET :offset
	`)

	// Build count query for pagination
	countQuery := `
		SELECT COUNT(*)
		FROM objectives o
		INNER JOIN team_members tm ON tm.team_id = o.team_id AND tm.user_id = :user_id
		WHERE o.workspace_id = :workspace_id
	`

	if params.Query != "" {
		countQuery += `
			AND (
				o.search_vector @@ plainto_tsquery('english', :query)
				OR o.name ILIKE '%' || :query || '%'
			)
		`
	}

	if params.TeamID != nil {
		countQuery += `
			AND o.team_id = :team_id
		`
	}

	if params.StatusID != nil {
		countQuery += `
			AND o.status_id = :status_id
		`
	}

	// Prepare named parameters
	namedParams := map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"query":        params.Query,
		"team_id":      params.TeamID,
		"status_id":    params.StatusID,
		"page_size":    params.PageSize,
		"offset":       (params.Page - 1) * params.PageSize,
	}

	// Execute count query
	var totalObjectives int
	countStmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare count query: %s", err))
		return nil, 0, err
	}
	defer countStmt.Close()

	err = countStmt.GetContext(ctx, &totalObjectives, namedParams)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to execute count query: %s", err))
		return nil, 0, err
	}

	// If no objectives found, return empty result
	if totalObjectives == 0 {
		return []search.CoreSearchObjective{}, 0, nil
	}

	// Execute main query
	stmt, err := r.db.PrepareNamedContext(ctx, queryBuilder.String())
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare search query: %s", err))
		return nil, 0, err
	}
	defer stmt.Close()

	var objectives []dbObjective
	err = stmt.SelectContext(ctx, &objectives, namedParams)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to execute search query: %s", err))
		return nil, 0, err
	}

	// Convert to core model
	return toCoreSearchObjectives(objectives), totalObjectives, nil
}
