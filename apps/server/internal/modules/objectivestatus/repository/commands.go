package objectivestatusrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
		resetParams := map[string]any{
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
	params := map[string]any{
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
		OrderIndex: maxOrder + 10,
		Workspace:  workspaceId,
		IsDefault:  ns.IsDefault,
		Color:      ns.Color,
	}

	params = map[string]any{
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
		resetParams := map[string]any{
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

	params := map[string]any{
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

	params := map[string]any{
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
