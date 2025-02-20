package usersrepo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

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
			username,
			email,
			full_name,
			avatar_url,
			last_login_at
		)
		VALUES (
			:username,
			:email,
			:full_name,
			:avatar_url,
			:last_login_at
		)
		RETURNING
			username,
			email,
			full_name,
			avatar_url,
			last_login_at,
			last_used_workspace_id
	`

	params := map[string]interface{}{
		"username":      user.Username,
		"email":         user.Email,
		"full_name":     user.FullName,
		"avatar_url":    user.AvatarURL,
		"last_login_at": user.LastLoginAt,
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, currentUserID uuid.UUID, teamID *uuid.UUID) ([]users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.List")
	defer span.End()

	var users []dbUser
	params := map[string]interface{}{
		"workspace_id":    workspaceId,
		"current_user_id": currentUserID,
	}

	baseQuery := `
		SELECT DISTINCT
			u.user_id,
			u.username,
			u.email,
			u.full_name,
			u.avatar_url,
			u.is_active,
			u.last_login_at,
			u.created_at,
			u.updated_at
		FROM users u
		INNER JOIN workspace_members wm ON u.user_id = wm.user_id
	`

	var q string
	if teamID != nil {
		// Query for specific team members
		params["team_id"] = *teamID
		q = baseQuery + `
			INNER JOIN team_members tm ON u.user_id = tm.user_id
			WHERE wm.workspace_id = :workspace_id
				AND tm.team_id = :team_id
				AND u.is_active = TRUE
			ORDER BY u.created_at DESC
		`
	} else {
		// Query for members from all teams the current user belongs to
		q = baseQuery + `
			INNER JOIN team_members tm ON u.user_id = tm.user_id
			INNER JOIN team_members my_teams 
				ON tm.team_id = my_teams.team_id 
				AND my_teams.user_id = :current_user_id
			WHERE wm.workspace_id = :workspace_id
				AND u.is_active = TRUE
			ORDER BY u.created_at DESC
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

// generateToken creates a cryptographically secure token
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateVerificationToken creates a new verification token in the database
func (r *repo) CreateVerificationToken(ctx context.Context, email, tokenType string) (users.CoreVerificationToken, error) {
	r.log.Info(ctx, "repository.users.CreateVerificationToken")
	ctx, span := web.AddSpan(ctx, "repository.users.CreateVerificationToken")
	defer span.End()

	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Check existing valid tokens
	var count int
	checkQ := `
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

	params := map[string]interface{}{
		"email":          email,
		"min_created_at": time.Now().Add(-time.Hour),
	}

	stmt, err := tx.PrepareNamedContext(ctx, checkQ)
	if err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("preparing check statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &count, params); err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("checking token count: %w", err)
	}

	if count >= 3 {
		return users.CoreVerificationToken{}, users.ErrTooManyAttempts
	}

	// Generate new token
	token, err := generateToken()
	if err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("generating token: %w", err)
	}

	// Create new token
	var dbToken dbVerificationToken
	createQ := `
		INSERT INTO verification_tokens (
			token, email, expires_at, token_type, created_at, updated_at
		) VALUES (
			:token, :email, :expires_at, :token_type, :created_at, :updated_at
		)
		RETURNING id, token, email, user_id, expires_at, used_at, token_type, created_at, updated_at
	`

	now := time.Now()
	createParams := map[string]interface{}{
		"token":      token,
		"email":      email,
		"expires_at": now.Add(10 * time.Minute),
		"token_type": tokenType,
		"created_at": now,
		"updated_at": now,
	}

	stmt, err = tx.PrepareNamedContext(ctx, createQ)
	if err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("preparing create statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &dbToken, createParams); err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("creating token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return users.CoreVerificationToken{}, fmt.Errorf("committing transaction: %w", err)
	}

	span.AddEvent("token created")
	return toCoreVerificationToken(dbToken), nil
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

	params := map[string]interface{}{
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

// MarkTokenUsed marks a verification token as used
func (r *repo) MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	r.log.Info(ctx, "repository.users.MarkTokenUsed")
	ctx, span := web.AddSpan(ctx, "repository.users.MarkTokenUsed")
	defer span.End()

	q := `
		UPDATE verification_tokens
		SET used_at = NOW(), updated_at = NOW()
		WHERE id = :token_id
	`

	params := map[string]interface{}{
		"token_id": tokenID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return fmt.Errorf("marking token used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}

	if rows == 0 {
		return users.ErrInvalidToken
	}

	return nil
}

// InvalidateTokens invalidates all unused tokens for an email
func (r *repo) InvalidateTokens(ctx context.Context, email string) error {
	r.log.Info(ctx, "repository.users.InvalidateTokens")
	ctx, span := web.AddSpan(ctx, "repository.users.InvalidateTokens")
	defer span.End()

	q := `
		UPDATE verification_tokens
		SET used_at = NOW(), updated_at = NOW()
		WHERE email = :email AND used_at IS NULL
	`

	params := map[string]interface{}{
		"email": email,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		return fmt.Errorf("invalidating tokens: %w", err)
	}

	return nil
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

	params := map[string]interface{}{
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
