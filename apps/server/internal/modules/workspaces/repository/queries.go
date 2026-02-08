package workspacesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// List returns a list of workspaces a user is a member of.
func (r *repo) List(ctx context.Context, userID uuid.UUID) ([]workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.List")
	defer span.End()

	var workspaces []dbWorkspaceWithRole

	params := map[string]any{
		"user_id": userID,
	}

	query := `
		SELECT DISTINCT
			w.workspace_id,
			w.slug,
			w.name,
			w.avatar_url,
			CASE 
				WHEN u.last_used_workspace_id = w.workspace_id THEN TRUE
				WHEN u.last_used_workspace_id IS NULL AND w.workspace_id = (
					SELECT w2.workspace_id 
					FROM workspaces w2 
					INNER JOIN workspace_members wm2 ON w2.workspace_id = wm2.workspace_id 
					WHERE wm2.user_id = :user_id
					ORDER BY w2.created_at ASC
					LIMIT 1
				) THEN TRUE
				ELSE FALSE
			END as is_active,
			wm.role as user_role,
			w.created_at,
			w.color,
			w.updated_at,
			w.trial_ends_on,
			w.deleted_at,
			w.deleted_by
		FROM
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		INNER JOIN
			users u ON wm.user_id = u.user_id
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

	return toCoreWorkspacesWithRole(workspaces), nil
}

func (r *repo) Get(ctx context.Context, workspaceID, userID uuid.UUID) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Get")
	defer span.End()

	var workspace dbWorkspaceWithRole

	query := `
		SELECT
			w.workspace_id,
			w.name,
			w.slug,
			w.color,
			w.avatar_url,
			CASE 
				WHEN u.last_used_workspace_id = w.workspace_id THEN TRUE
				WHEN u.last_used_workspace_id IS NULL AND w.workspace_id = (
					SELECT w2.workspace_id 
					FROM workspaces w2 
					INNER JOIN workspace_members wm2 ON w2.workspace_id = wm2.workspace_id 
					WHERE wm2.user_id = :user_id
					ORDER BY w2.created_at ASC
					LIMIT 1
				) THEN TRUE
				ELSE FALSE
			END as is_active,
			wm.role as user_role,
			w.created_by,
			w.created_at,
			w.updated_at,
			w.trial_ends_on,
			w.deleted_at,
			w.deleted_by
		FROM
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		INNER JOIN
			users u ON wm.user_id = u.user_id
		WHERE
			w.workspace_id = :workspace_id
			AND wm.user_id = :user_id
	`

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
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
			return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspaceWithRole(workspace), nil
}

func (r *repo) GetByID(ctx context.Context, workspaceID uuid.UUID) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.GetByID")
	defer span.End()

	var workspace dbWorkspace

	query := `
		SELECT
			w.workspace_id,
			w.name,
			w.slug,
			w.color,
			w.avatar_url,
			w.is_active,
			w.created_by,
			w.created_at,
			w.updated_at,
			w.trial_ends_on,
			w.deleted_at,
			w.deleted_by
		FROM
			workspaces w
		WHERE
			w.workspace_id = :workspace_id
			AND w.deleted_at IS NULL
	`

	params := map[string]any{
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
			return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get workspace by ID: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace by ID"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	span.AddEvent("workspace retrieved by ID.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))

	return toCoreWorkspace(workspace), nil
}

func (r *repo) GetBySlug(ctx context.Context, slug string, userID uuid.UUID) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.GetBySlug")
	defer span.End()

	var workspace dbWorkspaceWithRole

	query := `
		SELECT
			w.workspace_id,
			w.name,
			w.slug,
			w.color,
			w.avatar_url,
			CASE 
				WHEN u.last_used_workspace_id = w.workspace_id THEN TRUE
				WHEN u.last_used_workspace_id IS NULL AND w.workspace_id = (
					SELECT w2.workspace_id 
					FROM workspaces w2 
					INNER JOIN workspace_members wm2 ON w2.workspace_id = wm2.workspace_id 
					WHERE wm2.user_id = :user_id
					ORDER BY w2.created_at ASC
					LIMIT 1
				) THEN TRUE
				ELSE FALSE
			END as is_active,
			wm.role as user_role,
			w.created_by,
			w.created_at,
			w.updated_at,
			w.trial_ends_on,
			w.deleted_at,
			w.deleted_by
		FROM
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		INNER JOIN
			users u ON wm.user_id = u.user_id
		WHERE
			w.slug = :slug
			AND wm.user_id = :user_id
	`

	params := map[string]any{
		"slug":    slug,
		"user_id": userID,
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
			return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get workspace by slug: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspaceWithRole(workspace), nil
}

func (r *repo) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.CheckSlugAvailability")
	defer span.End()

	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM workspaces 
			WHERE slug = :slug
		)
	`

	params := map[string]any{
		"slug": slug,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false, err
	}
	defer stmt.Close()

	var exists bool
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		errMsg := fmt.Sprintf("failed to check slug availability: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to check slug"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false, err
	}

	return !exists, nil
}

// GetWorkspaceSettings retrieves the settings for a workspace
func (r *repo) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (workspaces.CoreWorkspaceSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.GetWorkspaceSettings")
	defer span.End()

	var result dbWorkspaceSettings
	query := `
		SELECT 
			workspace_id,
			story_term,
			sprint_term,
			objective_term,
			key_result_term,
			objective_enabled,
			key_result_enabled,
			created_at,
			updated_at
		FROM 
			workspace_settings
		WHERE 
			workspace_id = :workspace_id
	`

	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspaceSettings{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspaceSettings{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get workspace settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace settings"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspaceSettings{}, err
	}

	return toCoreWorkspaceSettings(result), nil
}

// GetWorkspaceAdminEmails retrieves email addresses of workspace admins (excluding the actor)
func (r *repo) GetWorkspaceAdminEmails(ctx context.Context, workspaceID, actorID uuid.UUID) ([]string, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.GetWorkspaceAdminEmails")
	defer span.End()

	query := `
		SELECT u.email
		FROM users u
		INNER JOIN workspace_members wm ON u.user_id = wm.user_id
		WHERE wm.workspace_id = :workspace_id
		AND wm.role = 'admin'
		AND u.user_id != :actor_id
		AND u.is_active = TRUE
	`

	params := map[string]any{
		"workspace_id": workspaceID,
		"actor_id":     actorID,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to query workspace admin emails: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to query workspace admin emails"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			errMsg := fmt.Sprintf("failed to scan email: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to scan email"), trace.WithAttributes(attribute.String("error", errMsg)))
			return nil, err
		}
		emails = append(emails, email)
	}

	span.AddEvent("workspace admin emails retrieved", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.Int("admin_count", len(emails)),
	))

	return emails, nil
}
