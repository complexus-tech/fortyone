package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type EmailAvatarRepository interface {
	GetEmailAvatar(context.Context, uuid.UUID) (string, error)
}

func (s *Service) GetEmailAvatar(ctx context.Context, handle uuid.UUID) (string, error) {
	store, ok := s.repo.(EmailAvatarRepository)
	if !ok {
		return "", errors.New("email avatar repository is unavailable")
	}
	return store.GetEmailAvatar(ctx, handle)
}
