//go:build integration

package useruow

import (
	"testing"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccountDeletionRetainsSharedHistoryAndScrubsIdentityReferences(t *testing.T) {
	f := newDeletionFixture(t)
	ctx := t.Context()
	objective, activity, assignment, collaborators, progress := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	otherInbox, assignmentInbox, ownInbox, nestedReply := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	legacyInbox := uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO objectives(objective_id,workspace_id,name,created_by,lead_user_id,sequence_id) VALUES ($1,$2,'Shared objective',$3,$3,1)`, objective, f.workspace, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO story_comments(comment_id,story_id,commenter_id,content,parent_id) VALUES ($1,$2,$3,'Shared nested reply',$4)`, nestedReply, f.story, f.user, f.reply)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO story_activities(activity_id,story_id,workspace_id,user_id,activity_type,field_changed,current_value,old_value,new_value,reason) VALUES ($1,$2,$3,$4,'update','priority','High',CAST('"Low"' AS jsonb),CAST('"High"' AS jsonb),'Shared explanation')`, activity, f.story, f.workspace, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO story_activities(activity_id,story_id,workspace_id,user_id,activity_type,field_changed,current_value,old_value,new_value) VALUES ($1,$2,$3,$4,'update','assignee_id','departing',to_jsonb(CAST(CAST($4 AS uuid) AS text)),to_jsonb(CAST(CAST($5 AS uuid) AS text)))`, assignment, f.story, f.workspace, f.other, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO story_activities(activity_id,story_id,workspace_id,user_id,activity_type,field_changed,current_value,old_value,new_value) VALUES ($1,$2,$3,$4,'update','collaborator_ids','[' || CAST(CAST($5 AS uuid) AS text) || ' ' || CAST(CAST($4 AS uuid) AS text) || ']',jsonb_build_array(CAST(CAST($5 AS uuid) AS text)),jsonb_build_array(CAST(CAST($5 AS uuid) AS text),CAST(CAST($4 AS uuid) AS text)))`, collaborators, f.story, f.workspace, f.other, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO okr_activities(activity_id,objective_id,workspace_id,user_id,activity_type,update_type,field_changed,current_value,comment) VALUES ($1,$2,$3,$4,'update','objective','progress','75','Shared progress explanation')`, progress, objective, f.workspace, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO okr_activities(objective_id,workspace_id,user_id,activity_type,update_type,field_changed,current_value) VALUES ($1,$2,$3,'update','key_result','contributors',CAST(CAST($4 AS uuid) AS text) || ',' || CAST(CAST($3 AS uuid) AS text))`, objective, f.workspace, f.other, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO notifications(notification_id,recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message) VALUES ($1,$2,$3,'story_comment','comment',$4,$5,'Shared discussion',CAST('{"template":"{actor} commented","variables":{"actor":{"value":"departing","type":"actor"},"comment":{"value":"Shared prose","type":"value"}}}' AS jsonb))`, otherInbox, f.other, f.workspace, f.comment, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO notifications(notification_id,recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message) VALUES ($1,$2,$3,'story_update','story',$4,$2,'Task assignment',CAST('{"template":"{actor} assigned {assignee}","variables":{"actor":{"value":"remaining","type":"actor"},"assignee":{"value":"departing","type":"assignee"}}}' AS jsonb) || jsonb_build_object('identityReferences',jsonb_build_object('assignee',CAST($5 AS text))))`, assignmentInbox, f.other, f.workspace, f.story, f.user.String())
	// A currently unique name still cannot identify who a historical snapshot meant.
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO notifications(notification_id,recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message) VALUES ($1,$2,$3,'story_update','story',$4,$2,'Legacy assignment',CAST('{"template":"Assigned {assignee}","variables":{"assignee":{"value":"departing","type":"assignee"}}}' AS jsonb))`, legacyInbox, f.other, f.workspace, f.story)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO notifications(notification_id,recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message) VALUES ($1,$2,$3,'story_update','story',$4,$5,'Personal inbox entry',CAST('{}' AS jsonb))`, ownInbox, f.user, f.workspace, f.story, f.other)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO audit_events(workspace_id,actor_id,actor_type,entity_type,event_type,metadata) VALUES ($1,$2,'user','story','account_snapshot',CAST('{"email":"departing@example.com"}' AS jsonb))`, f.workspace, f.user)

	request, thread, integrationComment := uuid.New(), uuid.New(), uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO integration_requests(id,workspace_id,team_id,provider,source_type,source_external_id,title) SELECT $1,$2,team_id,'slack','message','source-1','Shared request' FROM teams WHERE workspace_id=$2 LIMIT 1`, request, f.workspace)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO integration_request_threads(id,workspace_id,integration_request_id,provider,external_workspace_id,external_channel_id,external_thread_id) VALUES ($1,$2,$3,'slack','external-workspace','channel','thread')`, thread, f.workspace, request)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO integration_request_comments(id,workspace_id,thread_id,direction,author_user_id,external_author_id,external_message_id,body) VALUES ($1,$2,$3,'inbound',$4,'private-slack-user','message','Shared imported reply')`, integrationComment, f.workspace, thread, f.user)

	_, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	former := usersdomain.DeletedUserID
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1`, 0, f.user)
	deletionCount(t, f, `SELECT count(*) FROM integration_request_comments WHERE id=$1 AND author_user_id=$2 AND external_author_id IS NULL AND body='Shared imported reply' AND thread_id=$3`, 1, integrationComment, former, thread)
	deletionCount(t, f, `SELECT count(*) FROM story_comments WHERE comment_id=$1 AND parent_id=$2 AND commenter_id=$3 AND content='Shared nested reply'`, 1, nestedReply, f.reply, former)
	deletionCount(t, f, `SELECT count(*) FROM story_comments WHERE comment_id=$1 AND parent_id=$2 AND commenter_id=$3`, 1, f.reply, f.comment, f.other)
	deletionCount(t, f, `SELECT count(*) FROM story_activities WHERE activity_id=$1 AND user_id=$2 AND current_value='High' AND old_value=CAST('"Low"' AS jsonb) AND new_value=CAST('"High"' AS jsonb) AND reason='Shared explanation'`, 1, activity, former)
	deletionCount(t, f, `SELECT count(*) FROM story_activities WHERE activity_id=$1 AND user_id=$2 AND old_value=to_jsonb(CAST(CAST($2 AS uuid) AS text)) AND new_value=to_jsonb(CAST(CAST($3 AS uuid) AS text)) AND current_value='Former user'`, 1, assignment, f.other, former)
	deletionCount(t, f, `SELECT count(*) FROM story_activities WHERE activity_id=$1 AND new_value=jsonb_build_array(CAST(CAST($2 AS uuid) AS text),CAST(CAST($3 AS uuid) AS text)) AND current_value='[' || CAST(CAST($2 AS uuid) AS text) || ' ' || CAST(CAST($3 AS uuid) AS text) || ']'`, 1, collaborators, former, f.other)
	deletionCount(t, f, `SELECT count(*) FROM okr_activities WHERE activity_id=$1 AND user_id=$2 AND current_value='75' AND comment='Shared progress explanation'`, 1, progress, former)
	deletionCount(t, f, `SELECT count(*) FROM okr_activities WHERE objective_id=$1 AND field_changed='contributors' AND current_value=CAST(CAST($2 AS uuid) AS text) || ',' || CAST(CAST($3 AS uuid) AS text)`, 1, objective, former, f.other)
	deletionCount(t, f, `SELECT count(*) FROM objectives WHERE objective_id=$1 AND created_by=$2 AND lead_user_id IS NULL`, 1, objective, former)
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1 AND actor_id=$2 AND message #>> '{variables,actor,value}'='Former user' AND message #>> '{variables,comment,value}'='Shared prose'`, 1, otherInbox, former)
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1 AND actor_id=$2 AND message #>> '{variables,actor,value}'='remaining' AND message #>> '{variables,assignee,value}'='Former user'`, 1, assignmentInbox, f.other)
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1 AND actor_id=$2 AND message #>> '{variables,assignee,value}'='departing'`, 1, legacyInbox, f.other)
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1`, 0, ownInbox)
	deletionCount(t, f, `SELECT count(*) FROM audit_events WHERE actor_id=$1`, 0, f.user)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND full_name='Former user' AND NOT is_active AND is_system AND login_reactivation_policy='admin_only'`, 1, former)
	_, err = f.db.Pool.Exec(ctx, `UPDATE users SET is_active=TRUE WHERE user_id=$1`, former)
	require.ErrorContains(t, err, "cannot be changed or reactivated")
}

func TestAccountDeletionDoesNotGuessAtSameNamedNotificationAssignees(t *testing.T) {
	f := newDeletionFixture(t)
	ctx := t.Context()
	execDeletion(t, ctx, f.db.Pool, `UPDATE users SET username='departing' WHERE user_id=$1`, f.other)
	notification := uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO notifications(notification_id,recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message) VALUES ($1,$2,$3,'story_update','story',$4,$2,'Task assignment',CAST('{"template":"Assigned {assignee}","variables":{"assignee":{"value":"departing","type":"assignee"}}}' AS jsonb))`, notification, f.other, f.workspace, f.story)
	exactTarget, sameNamedOther := uuid.New(), uuid.New()
	for _, entry := range []struct{ id, assignee uuid.UUID }{{exactTarget, f.user}, {sameNamedOther, f.other}} {
		execDeletion(t, ctx, f.db.Pool, `INSERT INTO notifications(notification_id,recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message) VALUES ($1,$2,$3,'story_update','story',$4,$2,'Exact assignment',CAST('{"template":"Assigned {assignee}","variables":{"assignee":{"value":"departing","type":"assignee"}}}' AS jsonb) || jsonb_build_object('identityReferences',jsonb_build_object('assignee',CAST($5 AS text))))`, entry.id, f.other, f.workspace, f.story, entry.assignee.String())
	}
	_, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1 AND actor_id=$2 AND message #>> '{variables,assignee,value}'='departing'`, 1, notification, f.other)
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1 AND message #>> '{variables,assignee,value}'='Former user' AND message #>> '{identityReferences,assignee}'=$2`, 1, exactTarget, usersdomain.DeletedUserID.String())
	deletionCount(t, f, `SELECT count(*) FROM notifications WHERE notification_id=$1 AND message #>> '{variables,assignee,value}'='departing' AND message #>> '{identityReferences,assignee}'=$2`, 1, sameNamedOther, f.other.String())
}
