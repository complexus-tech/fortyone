package chatrepository

import (
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnreadThreadCursorCannotCrossViewerWorkspaceOrFilter(t *testing.T) {
	query := domain.UnreadThreadQuery{Scope: domain.ActorScope{WorkspaceID: uuid.New()}}
	query.Scope.Actor.PrincipalID = uuid.New()
	now, root := time.Now().UTC(), uuid.New()
	encoded := encodeCursor(cursor{Time: &now, ID: &root, ResourceScope: unreadThreadCursorScope(query)})
	_, err := decodeUnreadThreadCursor(encoded, unreadThreadCursorScope(query))
	require.NoError(t, err)
	for _, change := range []func(*domain.UnreadThreadQuery){
		func(q *domain.UnreadThreadQuery) { q.Scope.WorkspaceID = uuid.New() },
		func(q *domain.UnreadThreadQuery) { q.Scope.Actor.PrincipalID = uuid.New() },
		func(q *domain.UnreadThreadQuery) { id := uuid.New(); q.ConversationID = &id },
	} {
		other := query
		change(&other)
		_, err = decodeUnreadThreadCursor(encoded, unreadThreadCursorScope(other))
		require.ErrorIs(t, err, domain.ErrInvalid)
	}
	_, err = decodeUnreadThreadCursor("invalid", unreadThreadCursorScope(query))
	require.ErrorIs(t, err, domain.ErrInvalid)
}
