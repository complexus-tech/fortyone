package statesrepository

import (
	"context"
	"errors"
	"fmt"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID) ([]states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.List")
	defer span.End()

	params := map[string]any{
		"workspace_id": workspaceId,
		"user_id":      userID,
	}

	var statuses []dbState
	q := `
		SELECT
			s.status_id,
			s.name,
			s.category,
			s.order_index,
			s.team_id,
			s.workspace_id,
			s.is_default,
			s.color,
			s.created_at,
			s.updated_at
		FROM
			statuses s
		WHERE s.workspace_id = :workspace_id
		AND EXISTS (
			SELECT 1 
			FROM team_members tm 
			WHERE tm.team_id = s.team_id 
			AND tm.user_id = :user_id
		)
		ORDER BY s.order_index ASC;
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

	params := map[string]any{
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
			color,
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

func (r *repo) CountStoriesWithStatus(ctx context.Context, statusID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.CountStoriesWithStatus")
	defer span.End()

	params := map[string]any{
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

	params := map[string]any{
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
			color,
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
