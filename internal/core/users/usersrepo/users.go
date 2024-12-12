package usersrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
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

func (r *repo) GetByEmail(ctx context.Context, email string) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.Login")
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
			created_at,
			updated_at
		FROM
			users
		WHERE email = :email;
	`

	var params = map[string]interface{}{
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

	r.log.Info(ctx, "Fetching user.")
	if err := stmt.GetContext(ctx, &user, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve user from the database: %s", err)
		span.RecordError(errors.New("user not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		if err == sql.ErrNoRows {
			return users.CoreUser{}, errors.New("email or password is incorrect")
		}
		return users.CoreUser{}, fmt.Errorf("failed to retrieve user %w", err)
	}

	r.log.Info(ctx, "User retrieved successfully.")
	span.AddEvent("User retrieved.", trace.WithAttributes(
		attribute.String("user.email", user.Email),
	))

	return toCoreUser(user), nil
}

func (r *repo) Create(ctx context.Context, user users.CoreUser) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.Create")
	defer span.End()

	// q := `
	// 	INSERT INTO users (username, email, password_hash, full_name, avatar_url, is_active, last_login_at, created_at, updated_at)
	// 	VALUES (:username, :email, :password_hash, :full_name, :avatar_url, :is_active, :last_login_at, :created_at, :updated_at)
	// `

	return users.CoreUser{}, nil

}
