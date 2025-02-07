package usersrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository errors
var (
	ErrNotFound = users.ErrNotFound
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

// GetUser retrieves a user by ID
func (r *repo) GetUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.GetUser")
	defer span.End()

	var user dbUser
	q := `
		SELECT
			user_id,
			username,
			email,
			password_hash,
			full_name,
			avatar_url,
			is_active,
			last_login_at,
			last_used_workspace_id,
			created_at,
			updated_at
		FROM
			users
		WHERE 
			user_id = :user_id
			AND is_active = true
	`

	params := map[string]interface{}{
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
			user_id,
			username,
			email,
			password_hash,
			full_name,
			avatar_url,
			is_active,
			last_login_at,
			last_used_workspace_id,
			created_at,
			updated_at
		FROM
			users
		WHERE 
			email = :email
			AND is_active = true
	`

	params := map[string]interface{}{
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

// UpdateUser updates user profile information
func (r *repo) UpdateUser(ctx context.Context, userID uuid.UUID, updates users.CoreUpdateUser) error {
	ctx, span := web.AddSpan(ctx, "business.repository.users.UpdateUser")
	defer span.End()

	q := `
		UPDATE users
		SET 
			username = CASE WHEN :username = '' THEN username ELSE :username END,
			full_name = CASE WHEN :full_name = '' THEN full_name ELSE :full_name END,
			avatar_url = CASE WHEN :avatar_url = '' THEN avatar_url ELSE :avatar_url END,
			updated_at = NOW()
		WHERE 
			user_id = :user_id
			AND is_active = true
		RETURNING user_id
	`

	params := map[string]interface{}{
		"user_id":    userID,
		"username":   updates.Username,
		"full_name":  updates.FullName,
		"avatar_url": updates.AvatarURL,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var returnedID uuid.UUID
	if err := stmt.GetContext(ctx, &returnedID, params); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to update user: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update user"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// DeleteUser marks a user as inactive
func (r *repo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.users.DeleteUser")
	defer span.End()

	q := `
		UPDATE users
		SET 
			is_active = false,
			updated_at = NOW()
		WHERE 
			user_id = :user_id
			AND is_active = true
		RETURNING user_id
	`

	params := map[string]interface{}{
		"user_id": userID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var returnedID uuid.UUID
	if err := stmt.GetContext(ctx, &returnedID, params); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to delete user: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete user"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// UpdateUserWorkspace updates the user's last used workspace
func (r *repo) UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.users.UpdateUserWorkspace")
	defer span.End()

	q := `
		UPDATE users u
		SET 
			last_used_workspace_id = :workspace_id,
			updated_at = NOW()
		FROM workspace_members wm
		WHERE 
			u.user_id = :user_id
			AND u.is_active = true
			AND wm.user_id = u.user_id
			AND wm.workspace_id = :workspace_id
		RETURNING u.user_id
	`

	params := map[string]interface{}{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var returnedID uuid.UUID
	if err := stmt.GetContext(ctx, &returnedID, params); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found or not a member of workspace")
		}
		errMsg := fmt.Sprintf("failed to update user workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update user workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *repo) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.users.UpdatePassword")
	defer span.End()

	q := `
		UPDATE users
		SET 
			password_hash = :password_hash,
			updated_at = NOW()
		WHERE 
			user_id = :user_id
			AND is_active = true
		RETURNING user_id
	`

	params := map[string]interface{}{
		"user_id":       userID,
		"password_hash": hashedPassword,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var returnedID uuid.UUID
	if err := stmt.GetContext(ctx, &returnedID, params); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to update password: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update password"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// Create registers a new user
func (r *repo) Create(ctx context.Context, user users.CoreUser) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.Create")
	defer span.End()

	q := `
		INSERT INTO users (
			user_id,
			username,
			email,
			password_hash,
			full_name,
			avatar_url,
			is_active,
			last_login_at,
			created_at,
			updated_at
		)
		VALUES (
			:user_id,
			:username,
			:email,
			:password_hash,
			:full_name,
			:avatar_url,
			:is_active,
			:last_login_at,
			:created_at,
			:updated_at
		)
		RETURNING
			user_id,
			username,
			email,
			password_hash,
			full_name,
			avatar_url,
			is_active,
			last_login_at,
			last_used_workspace_id,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"user_id":       user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"password_hash": user.Password,
		"full_name":     user.FullName,
		"avatar_url":    user.AvatarURL,
		"is_active":     user.IsActive,
		"last_login_at": user.LastLoginAt,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}
	defer stmt.Close()

	var dbUser dbUser
	if err := stmt.GetContext(ctx, &dbUser, params); err != nil {
		errMsg := fmt.Sprintf("failed to create user: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create user"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}

	return toCoreUser(dbUser), nil
}

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.List")
	defer span.End()

	var users []dbUser

	params := map[string]interface{}{
		"workspace_id": workspaceId,
	}

	q := `
		SELECT
			u.user_id,
			u.username,
			u.email,
			u.password_hash,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.last_login_at,
			u.created_at,
			u.updated_at
		FROM
			users u
		INNER JOIN 
			workspace_members wm ON u.user_id = wm.user_id
		WHERE 
			wm.workspace_id = :workspace_id
		AND u.is_active = TRUE
	`
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
