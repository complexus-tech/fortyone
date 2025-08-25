package keyresultsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository errors
var (
	ErrNotFound = errors.New("key result not found")
)

// Repository defines the repository for key results
type Repository interface {
	Create(ctx context.Context, kr *CoreKeyResult) error
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error)
	List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error)
	ListPaginated(ctx context.Context, filters CoreKeyResultFilters) (CoreKeyResultListResponse, error)
	AddContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error
	UpdateContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error
	GetContributors(ctx context.Context, keyResultID uuid.UUID) ([]uuid.UUID, error)
}

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

// New creates a new key results repository
func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// Create inserts a new key result into the database
func (r *repo) Create(ctx context.Context, kr *CoreKeyResult) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Create")
	defer span.End()

	const q = `
		INSERT INTO key_results (
			objective_id, name, measurement_type,
			start_value, current_value, target_value,
			lead, start_date, end_date, created_by
		) VALUES (
			:objective_id, :name, :measurement_type,
			:start_value, :current_value, :target_value,
			:lead, :start_date, :end_date, :created_by
		)
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, "creating key result")
	if _, err := stmt.ExecContext(ctx, toDBKeyResult(*kr)); err != nil {
		errMsg := fmt.Sprintf("failed to create key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "key result created successfully")
	span.AddEvent("key result created", trace.WithAttributes(
		attribute.String("key_result.name", kr.Name),
	))

	return nil
}

// Update modifies data about a key result
func (r *repo) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Update")
	defer span.End()

	// Verify the key result exists and belongs to the workspace
	if _, err := r.getKeyResultById(ctx, id, workspaceId); err != nil {
		return err
	}

	query := "UPDATE key_results SET "
	var setClauses []string
	params := map[string]any{"id": id}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE id = :id"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating key result #%s", id), "id", id)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to update key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Key result #%s updated successfully", id), "id", id)
	span.AddEvent("key result updated", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

// Delete deletes a key result
func (r *repo) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Delete")
	defer span.End()

	// Verify the key result exists and belongs to the workspace
	if _, err := r.getKeyResultById(ctx, id, workspaceId); err != nil {
		return err
	}

	const q = `
		DELETE FROM key_results
		WHERE id = :id
	`

	params := map[string]any{
		"id": id,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, "deleting key result")
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to delete key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "key result deleted successfully")
	span.AddEvent("key result deleted", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

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

// AddContributors adds contributors to a key result
func (r *repo) AddContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.AddContributors")
	defer span.End()

	if len(contributorIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO key_result_contributors (
			key_result_id,
			user_id,
			created_at,
			updated_at
		) VALUES (
			:key_result_id,
			:user_id,
			NOW(),
			NOW()
		)
	`

	for _, contributorID := range contributorIDs {
		params := map[string]any{
			"key_result_id": keyResultID,
			"user_id":       contributorID,
		}

		if _, err := r.db.NamedExecContext(ctx, query, params); err != nil {
			errMsg := fmt.Sprintf("failed to add contributor: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to add contributor"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}
	}

	span.AddEvent("contributors added", trace.WithAttributes(
		attribute.Int("contributors.count", len(contributorIDs)),
		attribute.String("key_result.id", keyResultID.String()),
	))

	return nil
}

// UpdateContributors replaces all contributors for a key result
func (r *repo) UpdateContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.UpdateContributors")
	defer span.End()

	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing contributors
	deleteQuery := `DELETE FROM key_result_contributors WHERE key_result_id = :key_result_id`
	deleteParams := map[string]any{"key_result_id": keyResultID}

	if _, err := tx.NamedExecContext(ctx, deleteQuery, deleteParams); err != nil {
		errMsg := fmt.Sprintf("failed to delete existing contributors: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete contributors"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	// Add new contributors if any
	if len(contributorIDs) > 0 {
		insertQuery := `
			INSERT INTO key_result_contributors (
				key_result_id,
				user_id,
				created_at,
				updated_at
			) VALUES (
				:key_result_id,
				:user_id,
				NOW(),
				NOW()
			)
		`

		for _, contributorID := range contributorIDs {
			params := map[string]any{
				"key_result_id": keyResultID,
				"user_id":       contributorID,
			}

			if _, err := tx.NamedExecContext(ctx, insertQuery, params); err != nil {
				errMsg := fmt.Sprintf("failed to add contributor: %s", err)
				r.log.Error(ctx, errMsg)
				span.RecordError(errors.New("failed to add contributor"), trace.WithAttributes(attribute.String("error", errMsg)))
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to commit transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	span.AddEvent("contributors updated", trace.WithAttributes(
		attribute.Int("contributors.count", len(contributorIDs)),
		attribute.String("key_result.id", keyResultID.String()),
	))

	return nil
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
