package chatservice

import (
	"context"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type unreadRepositoryStub struct {
	Repository
	query domain.UnreadThreadQuery
}

func (r *unreadRepositoryStub) UnreadThreads(_ context.Context, query domain.UnreadThreadQuery) (domain.UnreadThreadPage, error) {
	r.query = query
	return domain.UnreadThreadPage{Items: []domain.UnreadThread{}}, nil
}

func TestUnreadThreadsRejectsInvalidScopeAndBounds(t *testing.T) {
	scope := testScope(t)
	service := New(nil)
	_, err := service.Unread(t.Context(), domain.ActorScope{})
	require.ErrorIs(t, err, domain.ErrForbidden)
	for _, limit := range []int{-1, 101} {
		_, err = service.UnreadThreads(t.Context(), domain.UnreadThreadQuery{Scope: scope, Limit: limit})
		require.ErrorIs(t, err, domain.ErrInvalid)
	}
	emptyID := uuid.Nil
	_, err = service.UnreadThreads(t.Context(), domain.UnreadThreadQuery{Scope: scope, ConversationID: &emptyID})
	require.ErrorIs(t, err, domain.ErrInvalid)
	repo := &unreadRepositoryStub{}
	service = New(repo)
	_, err = service.UnreadThreads(t.Context(), domain.UnreadThreadQuery{Scope: scope})
	require.NoError(t, err)
	require.Equal(t, 50, repo.query.Limit)
}
