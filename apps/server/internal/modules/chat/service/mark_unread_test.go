package chatservice

import (
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMarkUnreadValidatesScopeAndRequiredIDsBeforeRepository(t *testing.T) {
	service := New(nil)
	_, err := service.MarkUnread(t.Context(), domain.MarkUnreadCommand{})
	require.ErrorIs(t, err, domain.ErrForbidden)
	emptyID := uuid.Nil
	for _, command := range []domain.MarkUnreadCommand{
		{Scope: testScope(t), FromMessageID: uuid.New()},
		{Scope: testScope(t), ConversationID: uuid.New()},
		{Scope: testScope(t), ConversationID: uuid.New(), FromMessageID: uuid.New(), ThreadRootID: &emptyID},
	} {
		_, err = service.MarkUnread(t.Context(), command)
		require.ErrorIs(t, err, domain.ErrInvalid)
	}
}
