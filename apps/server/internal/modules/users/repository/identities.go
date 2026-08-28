package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func (r *repo) ResolveExternalIdentity(
	ctx context.Context,
	input usersdomain.ExternalIdentityInput,
) (usersdomain.ExternalIdentityResult, error) {
	var result usersdomain.ExternalIdentityResult
	err := r.withinTransaction(ctx, func(queries usersql.Querier) error {
		if err := acquireExternalIdentityLocks(ctx, queries, input); err != nil {
			return err
		}

		row, err := queries.GetUserByExternalIdentity(ctx, usersql.GetUserByExternalIdentityParams{
			Provider: input.Provider,
			Issuer:   input.Issuer,
			Subject:  input.Subject,
		})
		if err == nil {
			rows, touchErr := queries.TouchExternalIdentity(ctx, usersql.TouchExternalIdentityParams{
				Email:    input.Email,
				Provider: input.Provider,
				Issuer:   input.Issuer,
				Subject:  input.Subject,
			})
			if touchErr != nil {
				return fmt.Errorf("touch external identity: %w", touchErr)
			}
			if rows != 1 {
				return usersdomain.ErrNotFound
			}
			result.User = mapExternalIdentityUser(row)
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("get external identity user: %w", err)
		}

		user, created, err := resolveExternalIdentityUser(ctx, queries, input)
		if err != nil {
			return err
		}
		if err := queries.LinkExternalIdentity(ctx, usersql.LinkExternalIdentityParams{
			UserID:   user.ID,
			Provider: input.Provider,
			Issuer:   input.Issuer,
			Subject:  input.Subject,
			Email:    input.Email,
		}); err != nil {
			return fmt.Errorf("link external identity: %w", err)
		}

		result = usersdomain.ExternalIdentityResult{User: user, Created: created}
		return nil
	})
	if err != nil {
		return usersdomain.ExternalIdentityResult{}, err
	}
	return result, nil
}

func acquireExternalIdentityLocks(
	ctx context.Context,
	queries usersql.Querier,
	input usersdomain.ExternalIdentityInput,
) error {
	identityLock := strings.Join(
		[]string{"external-identity", input.Provider, input.Issuer, input.Subject},
		":",
	)
	if err := queries.AcquireExternalIdentityLock(ctx, usersql.AcquireExternalIdentityLockParams{
		LockIdentity: identityLock,
	}); err != nil {
		return fmt.Errorf("lock external identity: %w", err)
	}
	if err := queries.AcquireExternalIdentityLock(ctx, usersql.AcquireExternalIdentityLockParams{
		LockIdentity: "external-identity-email:" + input.Email,
	}); err != nil {
		return fmt.Errorf("lock external identity email: %w", err)
	}
	return nil
}

func resolveExternalIdentityUser(
	ctx context.Context,
	queries usersql.Querier,
	input usersdomain.ExternalIdentityInput,
) (usersdomain.User, bool, error) {
	existing, err := queries.GetUserByEmailAnyStatus(ctx, usersql.GetUserByEmailAnyStatusParams{
		Email: input.Email,
	})
	if err == nil {
		return mapUserByEmailAnyStatus(existing), false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return usersdomain.User{}, false, fmt.Errorf("get external identity user by email: %w", err)
	}

	username := strings.TrimSpace(strings.SplitN(input.Email, "@", 2)[0])
	if username == "" {
		username = "user"
	}
	created, err := queries.CreateExternalIdentityUser(ctx, usersql.CreateExternalIdentityUserParams{
		Username:  username,
		Email:     input.Email,
		FullName:  input.FullName,
		AvatarURL: input.AvatarURL,
		Timezone:  input.Timezone,
	})
	if err != nil {
		if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
			return usersdomain.User{}, false, usersdomain.ErrEmailTaken
		}
		return usersdomain.User{}, false, fmt.Errorf("create external identity user: %w", err)
	}
	return mapCreatedExternalIdentityUser(created), true, nil
}

func mapExternalIdentityUser(row usersql.GetUserByExternalIdentityRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}

func mapCreatedExternalIdentityUser(row usersql.CreateExternalIdentityUserRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}
