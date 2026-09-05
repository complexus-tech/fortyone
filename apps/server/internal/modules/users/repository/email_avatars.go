package usersrepository

import (
	"context"
	"errors"
	"fmt"

	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) EnsureEmailAvatarHandle(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	if userID == uuid.Nil {
		return uuid.Nil, ErrNotFound
	}
	handle, err := r.queries.EnsureEmailAvatarHandle(ctx, usersql.EnsureEmailAvatarHandleParams{UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("ensure email avatar handle: %w", err)
	}
	return handle, nil
}

func (r *repo) GetEmailAvatar(ctx context.Context, handle uuid.UUID) (string, error) {
	if handle == uuid.Nil {
		return "", ErrNotFound
	}
	avatar, err := r.queries.GetEmailAvatar(ctx, usersql.GetEmailAvatarParams{Handle: handle})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("resolve email avatar handle: %w", err)
	}
	if avatar == nil {
		return "", ErrNotFound
	}
	return *avatar, nil
}
