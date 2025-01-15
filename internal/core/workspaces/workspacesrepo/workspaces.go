package workspacesrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
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

// List returns a list of workspaces a user is a member of.
func (r *repo) List(ctx context.Context, userID uuid.UUID) ([]workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.List")
	defer span.End()

	var workspaces []dbWorkspace

	params := map[string]interface{}{
		"user_id": userID,
	}

	query := `
		SELECT DISTINCT
			w.workspace_id,
			w.slug,
			w.name,
			w.created_at,
			w.updated_at
		FROM
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		WHERE
			wm.user_id = :user_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg, userID)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &workspaces, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve workspaces from the database: %s", err)
		span.RecordError(errors.New("failed to retrieve workspaces"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to retrieve workspaces %w", err)
	}

	return toCoreWorkspaces(workspaces), nil
}

func (r *repo) Create(ctx context.Context, workspace workspaces.CoreWorkspace) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Create")
	defer span.End()

	var result dbWorkspace
	query := `
		INSERT INTO workspaces (
			name,
			slug,
			created_at,
			updated_at
		)
		VALUES (
			:name,
			:slug,
			NOW(),
			NOW()
		)
		RETURNING
			workspace_id,
			name,
			slug,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"name": workspace.Name,
		"slug": workspace.Slug,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		errMsg := fmt.Sprintf("failed to create workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspace(result), nil
}

func (r *repo) Update(ctx context.Context, workspaceID uuid.UUID, updates workspaces.CoreWorkspace) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Update")
	defer span.End()

	var result dbWorkspace
	query := `
		UPDATE workspaces
		SET 
			name = CASE WHEN :name = '' THEN name ELSE :name END,
			updated_at = NOW()
		WHERE 
			workspace_id = :workspace_id
		RETURNING
			workspace_id,
			name,
			slug,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"name":         updates.Name,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspace{}, errors.New("workspace not found")
		}
		errMsg := fmt.Sprintf("failed to update workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspace(result), nil
}

func (r *repo) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Delete")
	defer span.End()

	query := `
		DELETE FROM workspaces
		WHERE 
			workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (r *repo) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.AddMember")
	defer span.End()

	query := `
		INSERT INTO workspace_members (
			workspace_id,
			user_id,
			role,
			created_at
		)
		VALUES (
			:workspace_id,
			:user_id,
			:role,
			NOW()
		)
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"role":         role,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to add workspace member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to add workspace member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

func (r *repo) Get(ctx context.Context, workspaceID uuid.UUID) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Get")
	defer span.End()

	var workspace dbWorkspace
	query := `
		SELECT 
			workspace_id,
			slug,
			name,
			created_at,
			updated_at
		FROM 
			workspaces
		WHERE 
			workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &workspace, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspace{}, errors.New("workspace not found")
		}
		errMsg := fmt.Sprintf("failed to get workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspace(workspace), nil
}
