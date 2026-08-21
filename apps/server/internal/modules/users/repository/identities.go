package usersrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/jmoiron/sqlx"
)

const externalIdentityUserColumns = `
	u.user_id,
	u.username,
	u.email,
	u.full_name,
	u.avatar_url,
	u.is_active,
	u.is_system,
	u.is_internal,
	u.has_seen_walkthrough,
	u.timezone,
	u.working_days,
	u.working_start_minute,
	u.working_end_minute,
	u.last_login_at,
	u.last_used_workspace_id,
	u.github_username,
	u.created_at,
	u.updated_at
`

func (r *repo) ResolveExternalIdentity(ctx context.Context, input users.CoreExternalIdentityInput) (users.CoreExternalIdentityResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return users.CoreExternalIdentityResult{}, fmt.Errorf("begin external identity transaction: %w", err)
	}
	defer tx.Rollback()

	identityLock := strings.Join([]string{"external-identity", input.Provider, input.Issuer, input.Subject}, ":")
	emailLock := "external-identity-email:" + input.Email
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, identityLock); err != nil {
		return users.CoreExternalIdentityResult{}, fmt.Errorf("lock external identity: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, emailLock); err != nil {
		return users.CoreExternalIdentityResult{}, fmt.Errorf("lock external identity email: %w", err)
	}

	user, err := getExternalIdentityUser(ctx, tx, input.Provider, input.Issuer, input.Subject)
	if err == nil {
		if _, err := tx.ExecContext(ctx, `
			UPDATE public.user_external_identities
			SET email_at_link = $1, last_authenticated_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			WHERE provider = $2 AND issuer = $3 AND subject = $4
		`, input.Email, input.Provider, input.Issuer, input.Subject); err != nil {
			return users.CoreExternalIdentityResult{}, fmt.Errorf("update external identity authentication: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return users.CoreExternalIdentityResult{}, fmt.Errorf("commit external identity authentication: %w", err)
		}
		return users.CoreExternalIdentityResult{User: user}, nil
	}
	if !errors.Is(err, users.ErrNotFound) {
		return users.CoreExternalIdentityResult{}, err
	}

	created := false
	user, err = getUserByEmailForIdentity(ctx, tx, input.Email)
	if errors.Is(err, users.ErrNotFound) {
		user, err = createExternalIdentityUser(ctx, tx, input)
		created = err == nil
	}
	if err != nil {
		return users.CoreExternalIdentityResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO public.user_external_identities (
			user_id, provider, issuer, subject, email_at_link, last_authenticated_at
		) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`, user.ID, input.Provider, input.Issuer, input.Subject, input.Email); err != nil {
		return users.CoreExternalIdentityResult{}, fmt.Errorf("link external identity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return users.CoreExternalIdentityResult{}, fmt.Errorf("commit external identity link: %w", err)
	}
	return users.CoreExternalIdentityResult{User: user, Created: created}, nil
}

func getExternalIdentityUser(ctx context.Context, tx *sqlx.Tx, provider, issuer, subject string) (users.CoreUser, error) {
	query := `
		SELECT ` + externalIdentityUserColumns + `
		FROM public.user_external_identities identity
		INNER JOIN public.users u ON u.user_id = identity.user_id
		WHERE identity.provider = $1 AND identity.issuer = $2 AND identity.subject = $3
	`
	return getIdentityUser(ctx, tx, query, provider, issuer, subject)
}

func getUserByEmailForIdentity(ctx context.Context, tx *sqlx.Tx, email string) (users.CoreUser, error) {
	query := `
		SELECT ` + externalIdentityUserColumns + `
		FROM public.users u
		WHERE u.email = $1
	`
	return getIdentityUser(ctx, tx, query, email)
}

func getIdentityUser(ctx context.Context, tx *sqlx.Tx, query string, args ...any) (users.CoreUser, error) {
	var user dbUser
	if err := tx.GetContext(ctx, &user, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.CoreUser{}, users.ErrNotFound
		}
		return users.CoreUser{}, fmt.Errorf("load external identity user: %w", err)
	}
	return toCoreUser(user), nil
}

func createExternalIdentityUser(ctx context.Context, tx *sqlx.Tx, input users.CoreExternalIdentityInput) (users.CoreUser, error) {
	username := strings.TrimSpace(strings.Split(input.Email, "@")[0])
	if username == "" {
		username = "user"
	}

	var user dbUser
	err := tx.GetContext(ctx, &user, `
		INSERT INTO public.users (
			username, email, full_name, avatar_url, timezone, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			user_id, username, email, full_name, avatar_url, is_active, is_system,
			is_internal, has_seen_walkthrough, timezone, working_days,
			working_start_minute, working_end_minute, last_login_at,
			last_used_workspace_id, github_username, created_at, updated_at
	`, username, input.Email, input.FullName, input.AvatarURL, input.Timezone, time.Now())
	if err != nil {
		return users.CoreUser{}, fmt.Errorf("create external identity user: %w", err)
	}
	return toCoreUser(user), nil
}
