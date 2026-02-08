package keyresultsrepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Get retrieves a key result by ID
func (r *repo) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Get")
	defer span.End()

	kr, err := r.getKeyResultById(ctx, id, workspaceId)
	if err != nil {
		return CoreKeyResult{}, err
	}

	// Contributors are now handled in getKeyResultById and toCoreKeyResult
	return toCoreKeyResult(kr), nil
}

// List retrieves all key results for an objective
func (r *repo) List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.List")
	defer span.End()

	const q = `
		SELECT
			kr.id,
			kr.objective_id,
			kr.name,
			kr.measurement_type,
			kr.start_value,
			kr.current_value,
			kr.target_value,
			kr.lead,
			kr.start_date,
			kr.end_date,
			kr.created_at,
			kr.updated_at,
			kr.created_by,
			-- Aggregate contributors from junction table into JSON array
			COALESCE(
				(
					SELECT json_agg(krc.user_id)
					FROM key_result_contributors krc
					WHERE krc.key_result_id = kr.id
				), '[]'
			) AS contributors
		FROM key_results kr
		INNER JOIN objectives o ON kr.objective_id = o.objective_id
		WHERE kr.objective_id = :objective_id
		AND o.workspace_id = :workspace_id
		ORDER BY kr.created_at DESC
	`

	params := map[string]any{
		"objective_id": objectiveId,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	var keyResults []dbKeyResult
	if err := stmt.SelectContext(ctx, &keyResults, params); err != nil {
		errMsg := fmt.Sprintf("failed to list key results: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to list key results"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	// Convert to core models (contributors are now handled in toCoreKeyResult)
	coreKeyResults := make([]CoreKeyResult, len(keyResults))
	for i, kr := range keyResults {
		coreKeyResults[i] = toCoreKeyResult(kr)
	}

	return coreKeyResults, nil
}

// getKeyResultById retrieves a key result by ID and verifies it belongs to the workspace
func (r *repo) getKeyResultById(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (dbKeyResult, error) {
	const q = `
		SELECT
			kr.id,
			kr.objective_id,
			kr.name,
			kr.measurement_type,
			kr.start_value,
			kr.current_value,
			kr.target_value,
			kr.lead,
			kr.start_date,
			kr.end_date,
			kr.created_at,
			kr.updated_at,
			kr.created_by,
			-- Aggregate contributors from junction table into JSON array
			COALESCE(
				(
					SELECT json_agg(krc.user_id)
					FROM key_result_contributors krc
					WHERE krc.key_result_id = kr.id
				), '[]'
			) AS contributors
		FROM key_results kr
		INNER JOIN objectives o ON kr.objective_id = o.objective_id
		WHERE kr.id = :id
		AND o.workspace_id = :workspace_id
	`

	params := map[string]any{
		"id":           id,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err))
		return dbKeyResult{}, err
	}
	defer stmt.Close()

	var kr dbKeyResult
	if err := stmt.GetContext(ctx, &kr, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get key result: %s", err))
		return dbKeyResult{}, ErrNotFound
	}

	return kr, nil
}

// ListPaginated retrieves paginated key results with filters
func (r *repo) ListPaginated(ctx context.Context, filters CoreKeyResultFilters) (CoreKeyResultListResponse, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.ListPaginated")
	defer span.End()

	// Build query with filters and sorting
	query, params := r.buildPaginatedQuery(filters)

	// Get total count
	countQuery := r.buildCountQuery(filters)
	var totalCount int

	// Prepare and execute count query with named parameters
	countStmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare count query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare count query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return CoreKeyResultListResponse{}, err
	}
	defer countStmt.Close()

	if err := countStmt.GetContext(ctx, &totalCount, params); err != nil {
		errMsg := fmt.Sprintf("Failed to execute count query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to execute count query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return CoreKeyResultListResponse{}, err
	}

	// Get paginated results
	var dbResults []dbKeyResultWithObjective

	// Prepare and execute main query with named parameters
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare main query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare main query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return CoreKeyResultListResponse{}, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbResults, params); err != nil {
		errMsg := fmt.Sprintf("Failed to execute main query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to execute main query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return CoreKeyResultListResponse{}, err
	}

	// Convert to core models (contributors are now handled in toCoreKeyResult)
	keyResults := make([]CoreKeyResultWithObjective, len(dbResults))
	for i, dbResult := range dbResults {
		keyResults[i] = toCoreKeyResultWithObjective(dbResult)
	}

	hasMore := (filters.Page * filters.PageSize) < totalCount

	return CoreKeyResultListResponse{
		KeyResults: keyResults,
		TotalCount: totalCount,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		HasMore:    hasMore,
	}, nil
}

func (r *repo) buildPaginatedQuery(filters CoreKeyResultFilters) (string, map[string]any) {
	query := `
		SELECT
			kr.id, kr.objective_id, kr.name, kr.measurement_type,
			kr.start_value, kr.current_value, kr.target_value,
			kr.lead, kr.start_date, kr.end_date,
			kr.created_at, kr.updated_at, kr.created_by,
			o.name as objective_name, o.team_id, t.name as team_name, o.workspace_id,
			-- Aggregate contributors from junction table into JSON array
			COALESCE(
				(
					SELECT json_agg(krc.user_id)
					FROM key_result_contributors krc
					WHERE krc.key_result_id = kr.id
				), '[]'
			) AS contributors
		FROM key_results kr
		INNER JOIN objectives o ON kr.objective_id = o.objective_id
		INNER JOIN teams t ON o.team_id = t.team_id
		INNER JOIN team_members tm ON t.team_id = tm.team_id AND tm.user_id = :current_user_id
		WHERE o.workspace_id = :workspace_id
	`

	params := map[string]any{
		"workspace_id":    filters.WorkspaceID,
		"current_user_id": filters.CurrentUserID,
		"offset":          (filters.Page - 1) * filters.PageSize,
		"limit":           filters.PageSize,
	}

	// Add team filters (if specified, otherwise shows all teams user is member of)
	if len(filters.TeamIDs) > 0 {
		query += " AND o.team_id = ANY(:team_ids)"
		params["team_ids"] = filters.TeamIDs
	}

	// Add other filters
	if len(filters.ObjectiveIDs) > 0 {
		query += " AND kr.objective_id = ANY(:objective_ids)"
		params["objective_ids"] = filters.ObjectiveIDs
	}

	if len(filters.MeasurementTypes) > 0 {
		query += " AND kr.measurement_type = ANY(:measurement_types)"
		params["measurement_types"] = filters.MeasurementTypes
	}

	if filters.CreatedAfter != nil {
		query += " AND kr.created_at >= CAST(:created_after AS TIMESTAMP)"
		params["created_after"] = filters.CreatedAfter.Format("2006-01-02 15:04:05")
	}

	if filters.CreatedBefore != nil {
		query += " AND kr.created_at <= CAST(:created_before AS TIMESTAMP)"
		params["created_before"] = filters.CreatedBefore.Format("2006-01-02 15:04:05")
	}

	// Add sorting
	orderBy := r.getOrderByClause(filters.OrderBy, filters.OrderDirection)
	query += " ORDER BY " + orderBy

	// Add pagination
	query += " LIMIT :limit OFFSET :offset"

	return query, params
}

func (r *repo) buildCountQuery(filters CoreKeyResultFilters) string {
	query := `
		SELECT COUNT(*)
		FROM key_results kr
		INNER JOIN objectives o ON kr.objective_id = o.objective_id
		INNER JOIN teams t ON o.team_id = t.team_id
		INNER JOIN team_members tm ON t.team_id = tm.team_id AND tm.user_id = :current_user_id
		WHERE o.workspace_id = :workspace_id
	`

	// Add same filters as main query
	if len(filters.TeamIDs) > 0 {
		query += " AND o.team_id = ANY(:team_ids)"
	}

	if len(filters.ObjectiveIDs) > 0 {
		query += " AND kr.objective_id = ANY(:objective_ids)"
	}

	if len(filters.MeasurementTypes) > 0 {
		query += " AND kr.measurement_type = ANY(:measurement_types)"
	}

	if filters.CreatedAfter != nil {
		query += " AND kr.created_at >= CAST(:created_after AS TIMESTAMP)"
	}

	if filters.CreatedBefore != nil {
		query += " AND kr.created_at <= CAST(:created_before AS TIMESTAMP)"
	}

	return query
}

// GetContributors retrieves all contributor user IDs for a key result
func (r *repo) GetContributors(ctx context.Context, keyResultID uuid.UUID) ([]uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.GetContributors")
	defer span.End()

	query := `
		SELECT user_id
		FROM key_result_contributors
		WHERE key_result_id = :key_result_id
		ORDER BY created_at ASC
	`

	params := map[string]any{"key_result_id": keyResultID}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	var contributorIDs []uuid.UUID
	if err := stmt.SelectContext(ctx, &contributorIDs, params); err != nil {
		errMsg := fmt.Sprintf("failed to get contributors: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get contributors"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return contributorIDs, nil
}

func (r *repo) getOrderByClause(orderBy, orderDirection string) string {
	if orderDirection != "asc" && orderDirection != "desc" {
		orderDirection = "desc"
	}

	switch orderBy {
	case "name":
		return "kr.name " + orderDirection
	case "created_at":
		return "kr.created_at " + orderDirection
	case "updated_at":
		return "kr.updated_at " + orderDirection
	case "objective_name":
		return "o.name " + orderDirection
	default:
		return "kr.created_at desc"
	}
}
