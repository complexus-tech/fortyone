package chatservice

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
)

func (s *Service) MarkUnread(ctx context.Context, command domain.MarkUnreadCommand) (result domain.MarkUnreadState, err error) {
	if err = s.validateScope(ctx, command.Scope, true); err != nil {
		return
	}
	if command.ConversationID == uuid.Nil || command.FromMessageID == uuid.Nil ||
		(command.ThreadRootID != nil && *command.ThreadRootID == uuid.Nil) {
		return result, domain.ErrInvalid
	}
	err = s.mutate(ctx, func(unit context.Context) error {
		result, err = s.repo.MarkUnread(unit, command)
		if err != nil {
			return err
		}
		result.Summary, err = s.repo.Unread(unit, command.Scope)
		return err
	})
	return
}
