package chatservice

import (
	"context"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type typingRepositoryStub struct {
	Repository
	authorizations int
	recipientCalls int
	authorizeErr   error
	actor          domain.Person
	recipients     []uuid.UUID
}

func (r *typingRepositoryStub) AuthorizeTyping(context.Context, domain.TypingCommand) (domain.Person, error) {
	r.authorizations++
	return r.actor, r.authorizeErr
}

func (r *typingRepositoryStub) TypingRecipientIDs(context.Context, domain.TypingCommand) ([]uuid.UUID, error) {
	r.recipientCalls++
	return r.recipients, nil
}

type typingBroadcasterStub struct {
	allowed    bool
	allowCalls int
	published  []domain.TypingActivity
}

func (b *typingBroadcasterStub) AllowTyping(context.Context, domain.TypingCommand) (time.Time, error) {
	b.allowCalls++
	if !b.allowed {
		return time.Time{}, nil
	}
	return time.Now().Add(6 * time.Second), nil
}

func (b *typingBroadcasterStub) PublishTyping(_ context.Context, activity domain.TypingActivity) error {
	b.published = append(b.published, activity)
	return nil
}

func TestTypingChecksAccessBeforeThrottleAndSkipsFanoutForRepeatedActivity(t *testing.T) {
	scope := testScope(t)
	command := domain.TypingCommand{Scope: scope, ConversationID: uuid.New(), IsTyping: true}
	repo := &typingRepositoryStub{actor: domain.Person{ID: scope.Actor.PrincipalID, FullName: "Server profile"}, recipients: []uuid.UUID{uuid.New()}}
	broadcaster := &typingBroadcasterStub{allowed: true}
	service := New(repo, WithTypingBroadcaster(broadcaster))
	require.NoError(t, service.Typing(t.Context(), command))
	require.Len(t, broadcaster.published, 1)
	require.Equal(t, repo.actor, broadcaster.published[0].Actor)
	require.Equal(t, repo.recipients, broadcaster.published[0].RecipientIDs)

	broadcaster.allowed = false
	require.NoError(t, service.Typing(t.Context(), command))
	require.Equal(t, 2, repo.authorizations)
	require.Equal(t, 1, repo.recipientCalls)
	require.Len(t, broadcaster.published, 1)

	repo.authorizeErr = domain.ErrNotFound
	require.ErrorIs(t, service.Typing(t.Context(), command), domain.ErrNotFound)
	require.Equal(t, 2, broadcaster.allowCalls, "revoked access is not revived by the cached throttle")
}

func TestTypingRejectsInvalidScopeAndDoesNotSilentlyDropMissingDelivery(t *testing.T) {
	scope := testScope(t)
	repo := &typingRepositoryStub{}
	service := New(repo)
	require.ErrorIs(t, service.Typing(t.Context(), domain.TypingCommand{Scope: scope}), domain.ErrInvalid)
	command := domain.TypingCommand{Scope: scope, ConversationID: uuid.New()}
	err := service.Typing(t.Context(), command)
	require.ErrorContains(t, err, "delivery is unavailable")
	command.Scope.WorkspaceID = uuid.New()
	require.ErrorIs(t, service.Typing(t.Context(), command), domain.ErrForbidden)
}
