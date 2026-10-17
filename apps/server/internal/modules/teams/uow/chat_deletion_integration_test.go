//go:build integration

package teamuow

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTeamDeletionRetiresChatUploadsAndPreservesOtherConversations(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	fixture := seedTeamDeletionFixture(t, ctx, postgres.Pool, 1)
	var room, otherRoom uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT id FROM chat_conversations WHERE team_id=$1`, fixture.teamID).Scan(&room))
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT id FROM chat_conversations WHERE team_id=$1`, fixture.otherTeamID).Scan(&otherRoom))
	message := uuid.New()
	mustTeamDeletionExec(t, ctx, postgres.Pool, `INSERT INTO chat_conversation_messages(id,workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text) VALUES ($1,$2,$3,1,$4,$1,decode(repeat('ab',32),'hex'),'Team attachment')`, message, fixture.workspaceID, room, fixture.adminID)
	direct, pending := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{direct, pending} {
		mustTeamDeletionExec(t, ctx, postgres.Pool, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name) VALUES ($1,$2,$3,'team.txt',5,'text/plain',$4)`, id, fixture.workspaceID, fixture.adminID, id.String())
	}
	mustTeamDeletionExec(t, ctx, postgres.Pool, `INSERT INTO chat_attachments(attachment_id,workspace_id,conversation_id,uploaded_by,message_id) VALUES ($1,$3,$4,$5,$6),($2,$3,$4,$5,NULL)`, direct, pending, fixture.workspaceID, room, fixture.adminID, message)
	// An existing story attachment becomes shared with the surviving team chat.
	sharedWithOtherRoom := fixture.orphanIDs[0]
	mustTeamDeletionExec(t, ctx, postgres.Pool, `INSERT INTO chat_attachments(attachment_id,workspace_id,conversation_id,uploaded_by) VALUES ($1,$2,$3,$4)`, sharedWithOtherRoom, fixture.workspaceID, otherRoom, fixture.adminID)

	require.NoError(t, newTeamDeletionManager(t, postgres.Pool).Delete(ctx, fixture.command(t)))
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT count(*) FROM chat_conversations WHERE id=$1`, 0, room)
	for _, id := range []uuid.UUID{direct, pending} {
		assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, 0, id)
		assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1 AND container_name='team-deletion-test'`, 1, id)
	}
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, 1, sharedWithOtherRoom)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND conversation_id=$2`, 1, sharedWithOtherRoom, otherRoom)
	assertTeamDeletionCount(t, ctx, postgres.Pool, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, 0, sharedWithOtherRoom)
}
