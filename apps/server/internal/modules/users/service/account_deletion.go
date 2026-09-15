package users

import (
	"context"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/google/uuid"
)

type AccountDeletionService interface {
	Delete(context.Context, usersdomain.AccountDeletion) (pending bool, err error)
}

func WithAccountDeletion(deletion AccountDeletionService) Option {
	return func(service *Service) { service.accountDeletion = deletion }
}

// DeleteAccount permanently removes the account. Pending means the personal
// data and access have already been erased, while limited provider cleanup data
// remains temporarily. Calendar cleanup also temporarily needs an anonymous FK.
func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) (bool, error) {
	if s.accountDeletion == nil {
		return false, usersdomain.ErrAccountDeletionUnavailable
	}
	return s.accountDeletion.Delete(ctx, usersdomain.AccountDeletion{UserID: userID, RequestedAt: s.clock.Now().UTC()})
}
