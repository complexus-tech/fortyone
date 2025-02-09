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

	params := map[string]interface{}{
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

func (r *repo) Create(ctx context.Context, workspaceId uuid.UUID, ns states.CoreNewState) (states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.Create")
	defer span.End()

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

	stmt1, err := r.db.PrepareNamedContext(ctx, q1)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare max order statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt1.Close()

	if err := stmt1.GetContext(ctx, &maxOrder, params); err != nil {
		errMsg := fmt.Sprintf("failed to get max order: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	// Create the new state
	state := dbState{
		ID:         uuid.New(),
		Name:       ns.Name,
		Category:   ns.Category,
		OrderIndex: maxOrder + 1,
		Team:       ns.Team,
		Workspace:  workspaceId,
	}

	params = map[string]interface{}{
		"status_id":    state.ID,
		"name":         state.Name,
		"category":     state.Category,
		"order_index":  state.OrderIndex,
		"team_id":      state.Team,
		"workspace_id": state.Workspace,
	}

	q2 := `
		INSERT INTO statuses (
			status_id, name, category, order_index,
			team_id, workspace_id, created_at, updated_at
		) VALUES (
			:status_id, :name, :category, :order_index,
			:team_id, :workspace_id, NOW(), NOW()
		)
		RETURNING *
	`

	stmt2, err := r.db.PrepareNamedContext(ctx, q2)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare insert statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt2.Close()

	var created dbState
	if err := stmt2.GetContext(ctx, &created, params); err != nil {
		errMsg := fmt.Sprintf("failed to create state: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}

	return toCoreState(created), nil
}

func (r *repo) Update(ctx context.Context, workspaceId, stateId uuid.UUID, us states.CoreUpdateState) (states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.Update")
	defer span.End()

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

	if len(setClauses) == 0 {
		// No fields to update
		return states.CoreState{}, errors.New("no fields to update")
	}

	setClause := "SET " + strings.Join(setClauses, ", ") + ", updated_at = NOW()"

	q := fmt.Sprintf(`
		UPDATE statuses
		%s
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
		RETURNING *
	`, setClause)

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare update statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return states.CoreState{}, err
	}
	defer stmt.Close()

	var updated dbState
	if err := stmt.GetContext(ctx, &updated, params); err != nil {
		errMsg := fmt.Sprintf("failed to update state: %s", err)
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
