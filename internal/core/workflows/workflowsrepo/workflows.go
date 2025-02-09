package workflowsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/workflows"
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

func (r *repo) Create(ctx context.Context, workspaceId uuid.UUID, workflow workflows.CoreNewWorkflow) (workflows.CoreWorkflow, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workflows.Create")
	defer span.End()

	params := map[string]interface{}{
		"name":         workflow.Name,
		"team_id":      workflow.Team,
		"workspace_id": workspaceId,
	}

	var created dbWorkflow
	q := `
		INSERT INTO workflows (
			name, team_id, workspace_id
		) VALUES (
			:name, :team_id, :workspace_id
		)
		RETURNING *
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workflows.CoreWorkflow{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &created, params); err != nil {
		errMsg := fmt.Sprintf("failed to create workflow: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workflows.CoreWorkflow{}, err
	}

	return toCoreWorkflow(created), nil
}

func (r *repo) Update(ctx context.Context, workspaceId, workflowId uuid.UUID, workflow workflows.CoreUpdateWorkflow) (workflows.CoreWorkflow, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workflows.Update")
	defer span.End()

	params := map[string]interface{}{
		"workflow_id":  workflowId,
		"workspace_id": workspaceId,
	}

	setClauses := []string{}
	if workflow.Name != nil {
		params["name"] = *workflow.Name
		setClauses = append(setClauses, "name = :name")
	}

	if len(setClauses) == 0 {
		return workflows.CoreWorkflow{}, errors.New("no fields to update")
	}

	setClause := "SET " + strings.Join(setClauses, ", ") + ", updated_at = NOW()"

	q := fmt.Sprintf(`
		UPDATE workflows
		%s
		WHERE workflow_id = :workflow_id
		AND workspace_id = :workspace_id
		RETURNING *
	`, setClause)

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workflows.CoreWorkflow{}, err
	}
	defer stmt.Close()

	var updated dbWorkflow
	if err := stmt.GetContext(ctx, &updated, params); err != nil {
		errMsg := fmt.Sprintf("failed to update workflow: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workflows.CoreWorkflow{}, err
	}

	return toCoreWorkflow(updated), nil
}

func (r *repo) Delete(ctx context.Context, workspaceId, workflowId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workflows.Delete")
	defer span.End()

	params := map[string]interface{}{
		"workflow_id":  workflowId,
		"workspace_id": workspaceId,
	}

	q := `
		DELETE FROM workflows
		WHERE workflow_id = :workflow_id
		AND workspace_id = :workspace_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete workflow: %s", err)
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
		return errors.New("workflow not found or does not belong to workspace")
	}

	return nil
}

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]workflows.CoreWorkflow, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workflows.List")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceId,
	}

	var workflows []dbWorkflow
	q := `
		SELECT
			workflow_id,
			name,
			team_id,
			workspace_id,
			created_at,
			updated_at
		FROM
			workflows
		WHERE workspace_id = :workspace_id
		ORDER BY created_at ASC
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &workflows, params); err != nil {
		errMsg := fmt.Sprintf("failed to list workflows: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreWorkflows(workflows), nil
}

func (r *repo) ListByTeam(ctx context.Context, workspaceId, teamId uuid.UUID) ([]workflows.CoreWorkflow, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workflows.ListByTeam")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceId,
		"team_id":      teamId,
	}

	var workflows []dbWorkflow
	q := `
		SELECT
			workflow_id,
			name,
			team_id,
			workspace_id,
			created_at,
			updated_at
		FROM
			workflows
		WHERE workspace_id = :workspace_id
		AND team_id = :team_id
		ORDER BY created_at ASC
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &workflows, params); err != nil {
		errMsg := fmt.Sprintf("failed to list team workflows: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreWorkflows(workflows), nil
}
