package usersrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GetUser retrieves a user by ID
func (r *repo) GetUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.GetUser")
	defer span.End()

	var user dbUser
	q := `
		SELECT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.has_seen_walkthrough,
			u.timezone,
			u.last_login_at,
			u.last_used_workspace_id,
			u.created_at,
			u.updated_at
		FROM
			users u
		WHERE 
			u.user_id = :user_id
			AND u.is_active = true
	`

	params := map[string]any{
		"user_id": userID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &user, params); err != nil {
		if err == sql.ErrNoRows {
			return users.CoreUser{}, ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get user: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get user"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}

	return toCoreUser(user), nil
}

// GetUserByEmail retrieves a user by email
func (r *repo) GetUserByEmail(ctx context.Context, email string) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.GetUserByEmail")
	defer span.End()

	var user dbUser
	q := `
		SELECT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.has_seen_walkthrough,
			u.timezone,
			u.last_login_at,
			u.last_used_workspace_id,
			u.created_at,
			u.updated_at
		FROM
			users u
		WHERE 
			u.email = :email
			AND u.is_active = true
	`

	params := map[string]any{
		"email": email,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &user, params); err != nil {
		if err == sql.ErrNoRows {
			return users.CoreUser{}, ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get user: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get user"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}

	return toCoreUser(user), nil
}

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, teamID *uuid.UUID) ([]users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.List")
	defer span.End()

	var users []dbUser
	params := map[string]any{
		"workspace_id": workspaceId,
	}

	var q string
	if teamID != nil {
		// Query for specific team members
		params["team_id"] = *teamID
		q = `
		SELECT DISTINCT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.timezone,
			u.last_login_at,
			u.created_at,
			u.updated_at,
			wm.role as role
		FROM users u
		INNER JOIN workspace_members wm ON u.user_id = wm.user_id
		INNER JOIN team_members tm ON u.user_id = tm.user_id
		WHERE wm.workspace_id = :workspace_id
			AND tm.team_id = :team_id
			AND u.is_active = TRUE
		ORDER BY u.full_name
		`
	} else {
		// Query for all workspace members
		q = `
		SELECT DISTINCT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.timezone,
			u.last_login_at,
			u.created_at,
			u.updated_at,
			wm.role as role
		FROM users u
		INNER JOIN workspace_members wm ON u.user_id = wm.user_id
		WHERE wm.workspace_id = :workspace_id
			AND u.is_active = TRUE
		ORDER BY u.full_name
		`
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg, workspaceId)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &users, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve users from the database: %s", err)
		span.RecordError(errors.New("failed to retrieve users"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to retrieve users %w", err)
	}

	return toCoreUsers(users), nil
}

// GetVerificationToken retrieves a verification token from the database
func (r *repo) GetVerificationToken(ctx context.Context, token string) (users.CoreVerificationToken, error) {
	r.log.Info(ctx, "repository.users.GetVerificationToken")
	ctx, span := web.AddSpan(ctx, "repository.users.GetVerificationToken")
	defer span.End()

	var dbToken dbVerificationToken
	q := `
		SELECT id, token, email, user_id, expires_at, used_at, token_type, created_at, updated_at
		FROM verification_tokens
		WHERE token = :token
	`

	params := map[string]any{
		"token": token,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &dbToken, params); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.CoreVerificationToken{}, users.ErrInvalidToken
		}
		return users.CoreVerificationToken{}, fmt.Errorf("getting token: %w", err)
	}

	return toCoreVerificationToken(dbToken), nil
}

// GetValidTokenCount gets the count of valid tokens for an email within a duration
func (r *repo) GetValidTokenCount(ctx context.Context, email string, duration time.Duration) (int, error) {
	r.log.Info(ctx, "repository.users.GetValidTokenCount")
	ctx, span := web.AddSpan(ctx, "repository.users.GetValidTokenCount")
	defer span.End()

	var count int
	q := `
		WITH recent_tokens AS (
			SELECT *
			FROM verification_tokens
			WHERE email = :email
			AND created_at > :min_created_at
			AND used_at IS NULL
			ORDER BY created_at DESC
			LIMIT 3
		)
		SELECT COUNT(*)
		FROM recent_tokens
	`

	params := map[string]any{
		"email":          email,
		"min_created_at": time.Now().Add(-duration),
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &count, params); err != nil {
		return 0, fmt.Errorf("getting token count: %w", err)
	}

	return count, nil
}

// GetAutomationPreferences retrieves automation preferences for a user in a workspace
func (r *repo) GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (users.CoreAutomationPreferences, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.GetAutomationPreferences")
	defer span.End()

	query := `
		SELECT user_id, workspace_id, auto_assign_self, assign_self_on_branch_copy, 
			   move_story_to_started_on_branch, open_story_in_dialog, created_at, updated_at
		FROM user_automation_preferences
		WHERE user_id = :user_id AND workspace_id = :workspace_id;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreAutomationPreferences{}, err
	}
	defer stmt.Close()

	var prefs dbAutomationPreferences
	err = stmt.GetContext(ctx, &prefs, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default preferences if not found
			return r.createDefaultAutomationPreferences(ctx, userID, workspaceID)
		}
		errMsg := fmt.Sprintf("failed to get automation preferences: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get automation preferences"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreAutomationPreferences{}, err
	}

	span.AddEvent("automation preferences retrieved")
	return toCoreAutomationPreferences(prefs), nil
}

// ListUserMemories retrieves all memory items for a user in a workspace.
func (r *repo) ListUserMemories(ctx context.Context, userID, workspaceID uuid.UUID) ([]users.CoreUserMemoryItem, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.ListUserMemories")
	defer span.End()

	q := `
		SELECT id, workspace_id, user_id, content, created_at, updated_at
		FROM user_memories
		WHERE user_id = :user_id AND workspace_id = :workspace_id
		ORDER BY created_at DESC
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	var dbItems []dbUserMemoryItem
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("preparing list statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbItems, params); err != nil {
		return nil, fmt.Errorf("listing user memories: %w", err)
	}

	return toCoreUserMemoryItems(dbItems), nil
}
