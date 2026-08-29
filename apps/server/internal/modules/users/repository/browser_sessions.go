package usersrepository

import (
	"context"
	"errors"

	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ResolveActiveBrowserSessionVersion loads the authoritative epoch for an
// active account. Inactive and unknown accounts intentionally share the same
// not-found result so authentication callers cannot distinguish them.
func (r *repo) ResolveActiveBrowserSessionVersion(ctx context.Context, userID uuid.UUID) (int64, bool, error) {
	version, err := r.queries.GetActiveBrowserSessionVersion(ctx, usersql.GetActiveBrowserSessionVersionParams{
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, mapUserNotFound("resolve active browser session", err)
	}
	return version, true, nil
}
