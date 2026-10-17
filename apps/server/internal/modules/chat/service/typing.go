package chatservice

import (
	"context"
	"errors"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
)

// TypingBroadcaster is ephemeral. It never writes message history or durable events.
type TypingBroadcaster interface {
	// A zero expiry means repeated activity was throttled.
	AllowTyping(context.Context, domain.TypingCommand) (time.Time, error)
	PublishTyping(context.Context, domain.TypingActivity) error
}

func WithTypingBroadcaster(broadcaster TypingBroadcaster) Option {
	return func(s *Service) { s.typing = broadcaster }
}

func (s *Service) Typing(ctx context.Context, command domain.TypingCommand) error {
	if err := s.validateScope(ctx, command.Scope, true); err != nil {
		return err
	}
	if command.ConversationID == uuid.Nil || (command.ThreadRootID != nil && *command.ThreadRootID == uuid.Nil) {
		return domain.ErrInvalid
	}
	// Recheck access before consulting the throttle: a cached typing state is
	// never an authorization grant after membership or credential changes.
	actor, err := s.repo.AuthorizeTyping(ctx, command)
	if err != nil {
		return err
	}
	if s.typing == nil {
		return errors.New("chat typing delivery is unavailable")
	}
	expiresAt, err := s.typing.AllowTyping(ctx, command)
	if err != nil || expiresAt.IsZero() {
		return err
	}
	recipients, err := s.repo.TypingRecipientIDs(ctx, command)
	if err != nil {
		return err
	}
	return s.typing.PublishTyping(ctx, domain.TypingActivity{
		WorkspaceID: command.Scope.WorkspaceID, ConversationID: command.ConversationID,
		ThreadRootID: command.ThreadRootID, Actor: actor, RecipientIDs: recipients, IsTyping: command.IsTyping, ExpiresAt: expiresAt,
	})
}
