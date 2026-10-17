package chatservice

import (
	"context"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDisabledWorkspaceCannotReachChatStorageOrTyping(t *testing.T) {
	scope := testScope(t)
	for _, policy := range []chatrollout.Policy{{}, chatrollout.New(chatrollout.ReaderFunc(func(_ context.Context, id uuid.UUID) (bool, error) { return false, nil }))} {
		service := New(nil, WithWorkspaceSettings(policy))
		_, err := service.ListConversations(t.Context(), domain.ListQuery{Scope: scope})
		require.ErrorIs(t, err, domain.ErrForbidden)
		_, err = service.GetConversation(t.Context(), scope, uuid.New())
		require.ErrorIs(t, err, domain.ErrForbidden)
		_, err = service.CreateMessage(t.Context(), domain.CreateMessageCommand{Scope: scope})
		require.ErrorIs(t, err, domain.ErrForbidden)
		_, err = service.CreateAgentReply(t.Context(), domain.AgentReplyCommand{WorkspaceID: scope.WorkspaceID, Requester: scope.Actor})
		require.ErrorIs(t, err, domain.ErrForbidden)
		require.ErrorIs(t, service.Typing(t.Context(), domain.TypingCommand{Scope: scope}), domain.ErrForbidden)
		attachment := domain.AttachmentScope{ActorScope: scope, ConversationID: uuid.New()}
		require.ErrorIs(t, service.AuthorizeAttachment(t.Context(), attachment, uuid.New()), domain.ErrForbidden)
		require.ErrorIs(t, service.LinkPendingAttachment(t.Context(), attachment, uuid.New()), domain.ErrForbidden)
	}
}

func TestRolloutDoesNotReplaceActorAuthorization(t *testing.T) {
	scope := testScope(t)
	service := New(nil, WithWorkspaceSettings(chatrollout.New(chatrollout.ReaderFunc(func(_ context.Context, id uuid.UUID) (bool, error) { return id == scope.WorkspaceID, nil }))))
	require.NoError(t, service.validateScope(t.Context(), scope, false))
	scope.Actor.WorkspaceID = uuid.New()
	require.ErrorIs(t, service.validateScope(t.Context(), scope, false), domain.ErrForbidden)
}

func TestWorkspaceSettingDatabaseConflictRemainsRetryable(t *testing.T) {
	conflict := &pgconn.PgError{Code: "40001"}
	service := New(nil, WithWorkspaceSettings(chatrollout.New(chatrollout.ReaderFunc(func(context.Context, uuid.UUID) (bool, error) { return false, conflict }))))
	_, err := service.CreateMessage(t.Context(), domain.CreateMessageCommand{Scope: testScope(t)})
	require.ErrorIs(t, err, conflict)
}
