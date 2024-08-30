package usersrepo

import (
	"context"
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

func (r *repo) Login(ctx context.Context, email, password string) (users.CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.users.Login")
	defer span.End()

	var user dbUser
	q := `
		SELECT
			user_id,
			username,
			email,
			password,
			full_name,
			role,
			avatar_url,
			is_active,
			last_login_at,
			created_at,
			updated_at
		FROM
			users
		WHERE email = :email AND password = :password;
	`

	var params = map[string]interface{}{
		"email":    email,
		"password": password,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching user.")
	if err := stmt.GetContext(ctx, &user, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve user from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("user not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return users.CoreUser{}, err
	}

	r.log.Info(ctx, "User retrieved successfully.")
	span.AddEvent("User retrieved.", trace.WithAttributes(
		attribute.String("user.email", user.Email),
	))

	return toCoreUser(user), nil
}
