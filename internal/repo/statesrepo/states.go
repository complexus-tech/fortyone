package statesrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/states"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.List")
	defer span.End()

	params := map[string]any{
		"workspace_id": workspaceId,
	}

	var statuses []dbState
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			team_id,
			workspace_id,
			is_default,
			created_at,
			updated_at
		FROM
			statuses
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

	r.log.Info(ctx, "Fetching statuses.")
	if err := stmt.SelectContext(ctx, &statuses, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve statuses from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("statuses not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Statuses retrieved successfully.")
	span.AddEvent("Statuses retrieved.", trace.WithAttributes(
		attribute.Int("statuses.count", len(statuses)),
		attribute.String("query", q),
	))

	return toCoreStates(statuses), nil
}

func (r *repo) TeamList(ctx context.Context, workspaceId uuid.UUID, teamId uuid.UUID) ([]states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.TeamList")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceId,
		"team_id":      teamId,
	}

	var statuses []dbState
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			team_id,
			workspace_id,
			is_default,
			created_at,
			updated_at
		FROM
			statuses
		WHERE workspace_id = :workspace_id
		AND team_id = :team_id
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

	r.log.Info(ctx, "Fetching statuses.")
	if err := stmt.SelectContext(ctx, &statuses, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve statuses from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("statuses not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Statuses retrieved successfully.")
	span.AddEvent("Statuses retrieved.", trace.WithAttributes(
		attribute.Int("statuses.count", len(statuses)),
		attribute.String("query", q),
	))

	return toCoreStates(statuses), nil
}

func (r *repo) Create(ctx context.Context, workspaceId uuid.UUID, ns states.CoreNewState) (states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.Create")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	// If this status is going to be default, reset any existing defaults for this team
	if ns.IsDefault {
		resetQuery := `
			UPDATE statuses 
			SET is_default = false, updated_at = NOW() 
			WHERE team_id = :team_id 
			AND workspace_id = :workspace_id 
			AND is_default = true
		`
		resetParams := map[string]interface{}{
			"team_id":      ns.Team,
			"workspace_id": workspaceId,
		}

		resetStmt, err := tx.PrepareNamedContext(ctx, resetQuery)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to prepare reset query: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return states.CoreState{}, err
		}
		defer resetStmt.Close()

		if _, err := resetStmt.ExecContext(ctx, resetParams); err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to reset existing default statuses: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return states.CoreState{}, err
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
		FROM statuses
		WHERE workspace_id = :workspace_id
		AND category = :category
	`

	stmt1, err := tx.PrepareNamedContext(ctx, q1)
	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to prepare max order statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt1.Close()

	if err := stmt1.GetContext(ctx, &maxOrder, params); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to get max order: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	// Create the new state
	state := dbState{
		Name:       ns.Name,
		Category:   ns.Category,
		OrderIndex: maxOrder + 1,
		Team:       ns.Team,
		Workspace:  workspaceId,
		IsDefault:  ns.IsDefault,
	}

	params = map[string]interface{}{
		"name":         state.Name,
		"category":     state.Category,
		"order_index":  state.OrderIndex,
		"team_id":      state.Team,
		"workspace_id": state.Workspace,
		"is_default":   state.IsDefault,
	}

	q2 := `
		INSERT INTO statuses (
			name, category, order_index,
			team_id, workspace_id, is_default
		) VALUES (
			:name, :category, :order_index,
			:team_id, :workspace_id, :is_default
		)
		RETURNING status_id, name, category, order_index, team_id, workspace_id, is_default, created_at, updated_at
	`

	stmt2, err := tx.PrepareNamedContext(ctx, q2)
	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to prepare insert statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt2.Close()

	var created dbState
	if err := stmt2.GetContext(ctx, &created, params); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to create state: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	return toCoreState(created), nil
}

func (r *repo) Update(ctx context.Context, workspaceId, stateId uuid.UUID, us states.CoreUpdateState) (states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.Update")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	// If we're setting this as default, we need to handle that in a transaction
	if us.IsDefault != nil && *us.IsDefault {
		// First, get the team_id for this status
		var teamID uuid.UUID
		teamQuery := `
			SELECT team_id 
			FROM statuses 
			WHERE status_id = :status_id 
			AND workspace_id = :workspace_id
		`
		teamParams := map[string]interface{}{
			"status_id":    stateId,
			"workspace_id": workspaceId,
		}

		teamStmt, err := tx.PrepareNamedContext(ctx, teamQuery)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to prepare team query: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return states.CoreState{}, err
		}
		defer teamStmt.Close()

		if err := teamStmt.GetContext(ctx, &teamID, teamParams); err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to get team for status: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return states.CoreState{}, err
		}

		// Reset all existing default statuses for this team
		resetQuery := `
			UPDATE statuses 
			SET is_default = false, updated_at = NOW() 
			WHERE team_id = :team_id 
			AND workspace_id = :workspace_id 
			AND is_default = true
		`
		resetParams := map[string]interface{}{
			"team_id":      teamID,
			"workspace_id": workspaceId,
		}

		resetStmt, err := tx.PrepareNamedContext(ctx, resetQuery)
		if err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to prepare reset query: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return states.CoreState{}, err
		}
		defer resetStmt.Close()

		if _, err := resetStmt.ExecContext(ctx, resetParams); err != nil {
			tx.Rollback()
			errMsg := fmt.Sprintf("failed to reset existing default statuses: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return states.CoreState{}, err
		}
	}

	params := map[string]interface{}{
		"status_id":    stateId,
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

	if len(setClauses) == 0 {
		tx.Rollback()
		// No fields to update
		return states.CoreState{}, errors.New("no fields to update")
	}

	setClause := "SET " + strings.Join(setClauses, ", ") + ", updated_at = NOW()"

	q := fmt.Sprintf(`
		UPDATE statuses
		%s
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
		RETURNING status_id, name, category, order_index, team_id, workspace_id, is_default, created_at, updated_at
	`, setClause)

	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to prepare update statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt.Close()

	var updated dbState
	if err := stmt.GetContext(ctx, &updated, params); err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("failed to update state: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	return toCoreState(updated), nil
}

func (r *repo) Delete(ctx context.Context, workspaceId, stateId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.states.Delete")
	defer span.End()

	params := map[string]interface{}{
		"status_id":    stateId,
		"workspace_id": workspaceId,
	}

	q := `
		DELETE FROM statuses
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
		errMsg := fmt.Sprintf("failed to delete state: %s", err)
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
		return errors.New("state not found or does not belong to workspace")
	}

	return nil
}

func (r *repo) CountStoriesWithStatus(ctx context.Context, statusID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.CountStoriesWithStatus")
	defer span.End()

	params := map[string]interface{}{
		"status_id": statusID,
	}

	q := `
		SELECT COUNT(*)
		FROM stories
		WHERE status_id = :status_id
		AND deleted_at IS NULL
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count stories statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count stories with status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	return count, nil
}

func (r *repo) CountStatusesInCategory(ctx context.Context, teamID uuid.UUID, category string) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.CountStatusesInCategory")
	defer span.End()

	params := map[string]any{
		"team_id":  teamID,
		"category": category,
	}

	q := `
		SELECT COUNT(*)
		FROM statuses
		WHERE team_id = :team_id
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

func (r *repo) Get(ctx context.Context, workspaceId, stateId uuid.UUID) (states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.Get")
	defer span.End()

	params := map[string]interface{}{
		"status_id":    stateId,
		"workspace_id": workspaceId,
	}

	var status dbState
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			team_id,
			workspace_id,
			is_default,
			created_at,
			updated_at
		FROM
			statuses
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching status.")
	if err := stmt.GetContext(ctx, &status, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve status from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("status not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	r.log.Info(ctx, "Status retrieved successfully.")
	span.AddEvent("Status retrieved.", trace.WithAttributes(
		attribute.String("status.id", stateId.String()),
		attribute.String("query", q),
	))

	return toCoreState(status), nil
}
