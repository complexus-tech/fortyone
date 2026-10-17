//go:build integration

package chatrepository

import (
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/stretchr/testify/require"
)

func TestMemberSearchReachesBeyondBoundedAvatarPayloadAndRechecksAccess(t *testing.T) {
	f := newFixture(t)
	_, err := f.pool.Exec(f.ctx, `WITH people AS (
        INSERT INTO users(username,email,full_name)
        SELECT 'member-' || gen_random_uuid(),gen_random_uuid() || '@example.test','Member ' || n FROM generate_series(1,60) n RETURNING user_id
    ), workspace_memberships AS (
        INSERT INTO workspace_members(workspace_id,user_id,role) SELECT $1,user_id,'member' FROM people RETURNING user_id
    ) INSERT INTO team_members(team_id,user_id) SELECT $2,user_id FROM workspace_memberships`, f.workspace, f.team)
	require.NoError(t, err)
	conversation, err := f.service.GetConversation(f.ctx, f.scope, f.room)
	require.NoError(t, err)
	require.Len(t, conversation.Members, 50)
	require.Equal(t, 63, conversation.MemberCount)
	require.True(t, conversation.Members[0].IsSystem)
	members, err := f.service.ListMembers(f.ctx, domain.MemberSearch{Scope: f.scope, ConversationID: f.room, Query: "Member 60"})
	require.NoError(t, err)
	require.Len(t, members.Items, 1)
	require.Equal(t, "Member 60", members.Items[0].FullName)
	for _, person := range conversation.Members {
		require.NotEqual(t, members.Items[0].ID, person.ID, "searched member must be beyond the avatar cap")
	}
	limited, err := f.service.ListMembers(f.ctx, domain.MemberSearch{Scope: f.scope, ConversationID: f.room, Limit: 2})
	require.NoError(t, err)
	require.Len(t, limited.Items, 2)
	_, err = f.service.ListMembers(f.ctx, domain.MemberSearch{Scope: humanScope(t, f.workspace, f.outsider), ConversationID: f.room, Query: "Member"})
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, members.Items[0].ID)
	require.NoError(t, err)
	removed, err := f.service.ListMembers(f.ctx, domain.MemberSearch{Scope: f.scope, ConversationID: f.room, Query: "Member 60"})
	require.NoError(t, err)
	require.Empty(t, removed.Items)
	_, err = f.pool.Exec(f.ctx, `UPDATE users SET is_active=false WHERE user_id=$1`, f.author)
	require.NoError(t, err)
	_, err = f.service.ListMembers(f.ctx, domain.MemberSearch{Scope: f.scope, ConversationID: f.room, Query: "Member"})
	require.ErrorIs(t, err, domain.ErrNotFound)
}
