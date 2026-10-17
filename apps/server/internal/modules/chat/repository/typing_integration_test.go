//go:build integration

package chatrepository

import (
	"context"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	service "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type capturedTyping struct {
	allowCalls int
	activities []domain.TypingActivity
}

func (c *capturedTyping) AllowTyping(context.Context, domain.TypingCommand) (time.Time, error) {
	c.allowCalls++
	return time.Now().UTC().Add(6 * time.Second), nil
}

func (c *capturedTyping) PublishTyping(_ context.Context, activity domain.TypingActivity) error {
	c.activities = append(c.activities, activity)
	return nil
}

func TestTypingRechecksMembershipAndThreadScopeWithoutDurableWrites(t *testing.T) {
	f := newFixture(t)
	broadcast := &capturedTyping{}
	chat := service.New(f.repo, service.WithTypingBroadcaster(broadcast))
	root := f.create(t, "Thread root", nil)
	reply := f.create(t, "Thread reply", &root.ID)
	command := domain.TypingCommand{Scope: f.scope, ConversationID: f.room, IsTyping: true}
	require.NoError(t, chat.Typing(f.ctx, command))
	require.Equal(t, f.author, broadcast.activities[0].Actor.ID)
	require.Equal(t, "Person "+f.author.String(), broadcast.activities[0].Actor.FullName)
	require.Equal(t, []uuid.UUID{f.peer}, broadcast.activities[0].RecipientIDs)
	require.Nil(t, broadcast.activities[0].ThreadRootID)

	command.ThreadRootID = &root.ID
	require.NoError(t, chat.Typing(f.ctx, command))
	require.Equal(t, &root.ID, broadcast.activities[1].ThreadRootID)
	command.ThreadRootID = &reply.ID
	require.ErrorIs(t, chat.Typing(f.ctx, command), domain.ErrNotFound, "a reply cannot become another thread root")
	dm, err := f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.NoError(t, err)
	command.ThreadRootID, command.ConversationID = &root.ID, dm.ID
	require.ErrorIs(t, chat.Typing(f.ctx, command), domain.ErrNotFound, "thread cannot cross conversation scope")
	command.ConversationID = f.room
	command.Scope = humanScope(t, uuid.New(), f.author)
	require.ErrorIs(t, chat.Typing(f.ctx, command), domain.ErrNotFound)
	command.Scope = f.scope
	command.Scope.Actor.TeamAccess, err = auth.RestrictedTeamAccess(uuid.New())
	require.NoError(t, err)
	require.ErrorIs(t, chat.Typing(f.ctx, command), domain.ErrNotFound)
	command.Scope = f.scope
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: root.ID, Version: root.Version})
	require.NoError(t, err)
	require.NoError(t, chat.Typing(f.ctx, command), "a tombstone with remaining replies is a valid thread")
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: reply.ID, Version: reply.Version})
	require.NoError(t, err)
	require.ErrorIs(t, chat.Typing(f.ctx, command), domain.ErrNotFound, "a completely removed thread cannot publish activity")
	command.ThreadRootID = nil
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
	require.NoError(t, err)
	require.NoError(t, chat.Typing(f.ctx, command))
	require.Empty(t, broadcast.activities[len(broadcast.activities)-1].RecipientIDs)
	priorCalls := broadcast.allowCalls
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.author)
	require.NoError(t, err)
	require.ErrorIs(t, chat.Typing(f.ctx, command), domain.ErrNotFound)
	require.Equal(t, priorCalls, broadcast.allowCalls, "revoked sender never reaches cached throttle")
	recipients, err := f.repo.TypingRecipientIDs(f.ctx, command)
	require.NoError(t, err)
	require.Empty(t, recipients, "fanout independently checks sender membership")
	var count int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE workspace_id=$1 AND event_type NOT IN ('message.created','message.deleted')`, f.workspace).Scan(&count))
	require.Zero(t, count, "typing creates no durable event")
}
