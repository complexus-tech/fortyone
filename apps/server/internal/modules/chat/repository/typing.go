package chatrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) AuthorizeTyping(ctx context.Context, command domain.TypingCommand) (domain.Person, error) {
	if _, err := r.authorize(ctx, command.Scope, command.ConversationID, false); err != nil {
		return domain.Person{}, err
	}
	raw, err := r.query(ctx).AuthorizeTyping(ctx, chatsql.AuthorizeTypingParams{
		WorkspaceID: command.Scope.WorkspaceID, ConversationID: command.ConversationID,
		UserID: command.Scope.Actor.PrincipalID, ThreadRootID: command.ThreadRootID,
	})
	if err != nil {
		return domain.Person{}, scopedError(err)
	}
	return decode[domain.Person](raw)
}

func (r *Repository) TypingRecipientIDs(ctx context.Context, command domain.TypingCommand) ([]uuid.UUID, error) {
	return r.query(ctx).TypingRecipientIDs(ctx, chatsql.TypingRecipientIDsParams{
		WorkspaceID: command.Scope.WorkspaceID, ConversationID: command.ConversationID,
		UserID: command.Scope.Actor.PrincipalID, ThreadRootID: command.ThreadRootID,
	})
}
