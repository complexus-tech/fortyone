package objectivestatusrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]objectivestatus.CoreObjectiveStatus, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.List")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceId,
	}

	var statuses []dbObjectiveStatus
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			workspace_id,
			is_default,
			color,
			created_at,
			updated_at
		FROM
			objective_statuses
		WHERE workspace_id = :workspace_id
		ORDER BY order_index ASC;
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objective statuses.")
	if err := stmt.SelectContext(ctx, &statuses, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve objective statuses from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("statuses not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Objective statuses retrieved successfully.")
	span.AddEvent("Objective statuses retrieved.", trace.WithAttributes(
		attribute.Int("statuses.count", len(statuses)),
		attribute.String("query", q),
	))

	return toCoreObjectiveStatuses(statuses), nil
}

func (r *repo) Create(ctx context.Context, workspaceId uuid.UUID, ns objectivestatus.CoreNewObjectiveStatus) (objectivestatus.CoreObjectiveStatus, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.Create")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	// If this status is going to be default, reset any existing defaults for this workspace
	if ns.IsDefault {
		resetQuery := `
			UPDATE objective_statuses 
			SET is_default = false, updated_at = NOW() 
			WHERE workspace_id = :workspace_id 
			AND is_default = true
		`
		resetParams := map[string]interface{}{
			"workspace_id": workspaceId,
		}

		resetStmt, err := tx.PrepareNamedContext(ctx, resetQuery)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to prepare reset query: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return objectivestatus.CoreObjectiveStatus{}, err
		}
		defer resetStmt.Close()

		if _, err := resetStmt.ExecContext(ctx, resetParams); err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to reset existing default statuses: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return objectivestatus.CoreObjectiveStatus{}, err
		}
	}

	// Get the next order index based on category
	params := map[string]interface{}{
		"workspace_id": workspaceId,
		"category":     ns.Category,
	}

	var maxOrder int
	q1 := `
		SELECT COALESCE(MAX(order_index), -1)
		FROM objective_statuses
		WHERE workspace_id = :workspace_id
		AND category = :category
	`

	stmt1, err := tx.PrepareNamedContext(ctx, q1)
	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to prepare max order statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}
	defer stmt1.Close()

	if err := stmt1.GetContext(ctx, &maxOrder, params); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to get max order: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	// Create the new status
	status := dbObjectiveStatus{
		Name:       ns.Name,
		Category:   ns.Category,
		OrderIndex: maxOrder + 1,
		Workspace:  workspaceId,
		IsDefault:  ns.IsDefault,
		Color:      ns.Color,
	}

	params = map[string]interface{}{
		"name":         status.Name,
		"category":     status.Category,
		"order_index":  status.OrderIndex,
		"workspace_id": status.Workspace,
		"is_default":   status.IsDefault,
		"color":        status.Color,
	}

	q2 := `
		INSERT INTO objective_statuses (
			name, category, order_index, color,
			workspace_id, is_default
		) VALUES (
			:name, :category, :order_index, :color,
			:workspace_id, :is_default
		)
		RETURNING 
			status_id, name, category, order_index, workspace_id, is_default, color, created_at, updated_at
	`

	stmt2, err := tx.PrepareNamedContext(ctx, q2)
	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to prepare insert statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}
	defer stmt2.Close()

	var created dbObjectiveStatus
	if err := stmt2.GetContext(ctx, &created, params); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to create objective status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	return toCoreObjectiveStatus(created), nil
}

func (r *repo) Update(ctx context.Context, workspaceId, statusId uuid.UUID, us objectivestatus.CoreUpdateObjectiveStatus) (objectivestatus.CoreObjectiveStatus, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.Update")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	// If we're setting this as default, we need to handle that in a transaction
	if us.IsDefault != nil && *us.IsDefault {
		// Reset all existing default statuses for this workspace
		resetQuery := `
			UPDATE objective_statuses 
			SET is_default = false, updated_at = NOW() 
			WHERE workspace_id = :workspace_id 
			AND is_default = true
		`
		resetParams := map[string]interface{}{
			"workspace_id": workspaceId,
		}

		resetStmt, err := tx.PrepareNamedContext(ctx, resetQuery)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to prepare reset query: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return objectivestatus.CoreObjectiveStatus{}, err
		}
		defer resetStmt.Close()

		if _, err := resetStmt.ExecContext(ctx, resetParams); err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to reset existing default statuses: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return objectivestatus.CoreObjectiveStatus{}, err
		}
	}

	params := map[string]interface{}{
		"status_id":    statusId,
		"workspace_id": workspaceId,
	}

	// Build dynamic update query
	setClauses := []string{}
	if us.Name != nil {
		params["name"] = *us.Name
		setClauses = append(setClauses, "name = :name")
	}
	if us.OrderIndex != nil {
		params["order_index"] = *us.OrderIndex
		setClauses = append(setClauses, "order_index = :order_index")
	}
	if us.IsDefault != nil {
		params["is_default"] = *us.IsDefault
		setClauses = append(setClauses, "is_default = :is_default")
	}
	if us.Color != nil {
		params["color"] = *us.Color
		setClauses = append(setClauses, "color = :color")
	}

	if len(setClauses) == 0 {
		tx.Rollback()
		// No fields to update
		return objectivestatus.CoreObjectiveStatus{}, errors.New("no fields to update")
	}

	setClause := "SET " + strings.Join(setClauses, ", ") + ", updated_at = NOW()"

	q := fmt.Sprintf(`
		UPDATE objective_statuses
		%s
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
		RETURNING status_id, name, category, order_index, workspace_id, is_default, color, created_at, updated_at
	`, setClause)

	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to prepare update statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}
	defer stmt.Close()

	var updated dbObjectiveStatus
	if err := stmt.GetContext(ctx, &updated, params); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to update objective status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	return toCoreObjectiveStatus(updated), nil
}

func (r *repo) Delete(ctx context.Context, workspaceId, statusId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.Delete")
	defer span.End()

	params := map[string]interface{}{
		"status_id":    statusId,
		"workspace_id": workspaceId,
	}

	q := `
		DELETE FROM objective_statuses
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare delete statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete objective status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get affected rows: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if rows == 0 {
		return errors.New("objective status not found or does not belong to workspace")
	}

	return nil
}

func (r *repo) CountObjectivesWithStatus(ctx context.Context, statusID uuid.UUID, workspaceID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.CountObjectivesWithStatus")
	defer span.End()

	params := map[string]interface{}{
		"status_id":    statusID,
		"workspace_id": workspaceID,
	}

	q := `
		SELECT COUNT(*)
		FROM objectives
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count objectives statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count objectives with status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	return count, nil
}

func (r *repo) CountStatusesInCategory(ctx context.Context, workspaceID uuid.UUID, category string) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.CountStatusesInCategory")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"category":     category,
	}

	q := `
		SELECT COUNT(*)
		FROM objective_statuses
		WHERE workspace_id = :workspace_id
		AND category = :category
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count statuses statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count statuses in category: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	return count, nil
}

func (r *repo) Get(ctx context.Context, workspaceId, statusId uuid.UUID) (objectivestatus.CoreObjectiveStatus, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.Get")
	defer span.End()

	params := map[string]interface{}{
		"status_id":    statusId,
		"workspace_id": workspaceId,
	}

	var status dbObjectiveStatus
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			workspace_id,
			is_default,
			color,
			created_at,
			updated_at
		FROM
			objective_statuses
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objective status.")
	if err := stmt.GetContext(ctx, &status, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve objective status from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("status not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	r.log.Info(ctx, "Objective status retrieved successfully.")
	span.AddEvent("Objective status retrieved.", trace.WithAttributes(
		attribute.String("status.id", statusId.String()),
		attribute.String("query", q),
	))

	return toCoreObjectiveStatus(status), nil
}
