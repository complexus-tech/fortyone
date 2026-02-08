package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

// Role represents a user role in the workspace
type Role string

const (
	RoleGuest  Role = "guest"
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
)

// RoleLevel defines role hierarchy (higher number = more permissions)
var RoleLevel = map[Role]int{
	RoleGuest:  1,
	RoleMember: 2,
	RoleAdmin:  3,
}

// RequireMinimumRole ensures the user has at least the specified role level or higher.
// Example: RequireMinimumRole(log, RoleMember) allows member and admin users.
func RequireMinimumRole(log *logger.Logger, minimumRole Role) web.Middleware {
	// Validate minimum role at startup (fail fast)
	minimumLevel, exists := RoleLevel[minimumRole]
	if !exists {
		panic(fmt.Sprintf("invalid minimum role: %s", minimumRole))
	}

	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			workspace, err := GetWorkspace(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}

			userRole := Role(workspace.UserRole)
			userLevel, exists := RoleLevel[userRole]
			if !exists {
				log.Error(ctx, "invalid role in database", "role", userRole)
				return web.RespondError(ctx, w, errors.New("invalid user role"), http.StatusInternalServerError)
			}

			if userLevel >= minimumLevel {
				return next(ctx, w, r)
			}

			return web.RespondError(ctx, w, errors.New("you don't have permission to access this resource"), http.StatusForbidden)
		}
	}
}

// RequireRoles ensures user has exactly one of the specified roles
func RequireRoles(log *logger.Logger, roles ...Role) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			workspace, err := GetWorkspace(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}

			userRole := Role(workspace.UserRole)
			_, exists := RoleLevel[userRole]
			if !exists {
				log.Error(ctx, "invalid role in database", "role", userRole)
				return web.RespondError(ctx, w, errors.New("invalid user role"), http.StatusInternalServerError)
			}

			if slices.Contains(roles, userRole) {
				return next(ctx, w, r)
			}

			return web.RespondError(ctx, w, errors.New("you don't have permission to access this resource"), http.StatusForbidden)
		}
	}
}
