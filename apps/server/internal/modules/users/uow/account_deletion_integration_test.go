//go:build integration

package useruow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type deletionCleanup struct{ pending bool }

func (*deletionCleanup) StageAccountDeletion(context.Context, pgx.Tx, uuid.UUID) error { return nil }
func (cleanup *deletionCleanup) PendingAccountDeletion(context.Context, pgx.Tx, uuid.UUID) (bool, error) {
	return cleanup.pending, nil
}

type deletionFixture struct {
	db                                                                                                        *testkit.Postgres
	manager                                                                                                   *Manager
	cleanup                                                                                                   *deletionCleanup
	user, other, workspace, story, comment, reply, privateDocument, sharedDocument, orphan, shared, principal uuid.UUID
}

func newDeletionFixture(t *testing.T) deletionFixture {
	t.Helper()
	db := testkit.NewPostgres(t)
	f := deletionFixture{db: db, cleanup: &deletionCleanup{}, user: uuid.New(), other: uuid.New(), workspace: uuid.New(), story: uuid.New(), comment: uuid.New(), reply: uuid.New(), privateDocument: uuid.New(), sharedDocument: uuid.New(), orphan: uuid.New(), shared: uuid.New(), principal: uuid.New()}
	repo, err := usersrepository.NewAccountDeletionRepository(db.Pool, "aws", "attachments", "profiles")
	require.NoError(t, err)
	f.manager, err = New(db.Pool, repo, f.cleanup)
	require.NoError(t, err)
	ctx := t.Context()
	execDeletion(t, ctx, db.Pool, `INSERT INTO users(user_id,username,email,full_name,avatar_url) VALUES ($1,'departing','departing@example.com','Private Full Name','private-avatar.png'),($2,'remaining','remaining@example.com','Other Member',NULL)`, f.user, f.other)
	execDeletion(t, ctx, db.Pool, `INSERT INTO workspaces(workspace_id,name,slug,created_by) VALUES ($1,'Shared workspace',$2,$3)`, f.workspace, uuid.NewString(), f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES ($1,$2,'admin'),($1,$3,'admin')`, f.workspace, f.user, f.other)
	team := uuid.New()
	execDeletion(t, ctx, db.Pool, `INSERT INTO teams(team_id,workspace_id,name,color,code) VALUES ($1,$2,'Team','#cccccc','DEL')`, team, f.workspace)
	execDeletion(t, ctx, db.Pool, `INSERT INTO stories(id,workspace_id,team_id,title,reporter_id,assignee_id) VALUES ($1,$2,$3,'Organization task',$4,$4)`, f.story, f.workspace, team, f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO story_comments(comment_id,story_id,commenter_id,content) VALUES ($1,$2,$3,'private comment')`, f.comment, f.story, f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO story_comments(comment_id,story_id,commenter_id,content,parent_id) VALUES ($1,$2,$3,'other reply',$4)`, f.reply, f.story, f.other, f.comment)
	execDeletion(t, ctx, db.Pool, `INSERT INTO documents(document_id,workspace_id,title,content_text,visibility,created_by,updated_by) VALUES ($1,$3,'Private','Private personal text','private',$4,$4),($2,$3,'Shared','Team knowledge','workspace',$4,$4)`, f.privateDocument, f.sharedDocument, f.workspace, f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name) VALUES ($1,$3,$4,'private.txt',5,'text/plain','private-file'),($2,$3,$4,'shared.txt',6,'text/plain','shared-file')`, f.orphan, f.shared, f.workspace, f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO document_attachments(document_id,attachment_id,created_by) VALUES ($1,$2,$3),($4,$5,$3)`, f.privateDocument, f.orphan, f.user, f.sharedDocument, f.shared)
	execDeletion(t, ctx, db.Pool, `INSERT INTO user_memories(workspace_id,user_id,content) VALUES ($1,$2,'private memory')`, f.workspace, f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO user_external_identities(identity_id,user_id,provider,issuer,subject,email_at_link,created_at,updated_at,last_authenticated_at) VALUES ($1,$2,'microsoft','https://login.microsoftonline.com/common/v2.0','personal-ms-subject','departing@example.com',now(),now(),now())`, uuid.New(), f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO principals(principal_id,workspace_id,kind,name,subject_user_id,created_at,updated_at) VALUES ($1,$2,'human_user','Personal credential',$3,now(),now())`, f.principal, f.workspace, f.user)
	execDeletion(t, ctx, db.Pool, `INSERT INTO api_credentials(credential_id,workspace_id,principal_id,kind,name,lookup_prefix,secret_digest,token_version,digest_key_id,digest_key_version,created_at,expires_at) VALUES ($1,$2,$3,'personal_access_token','Private token','aaaaaaaaaaaa',decode(repeat('ab',32),'hex'),1,'test',1,now(),now()+interval '1 day')`, uuid.New(), f.workspace, f.principal)
	return f
}

func execDeletion(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	_, err := pool.Exec(ctx, query, args...)
	require.NoError(t, err)
}
func deletionCount(t *testing.T, f deletionFixture, query string, want int, args ...any) {
	t.Helper()
	var count int
	require.NoError(t, f.db.Pool.QueryRow(t.Context(), query, args...).Scan(&count))
	require.Equal(t, want, count, query)
}
func (f deletionFixture) command() usersdomain.AccountDeletion {
	return usersdomain.AccountDeletion{UserID: f.user, RequestedAt: time.Now().UTC()}
}

func TestAccountDeletionErasesPrivateGraphAndPreservesOtherPeopleContent(t *testing.T) {
	f := newDeletionFixture(t)
	pending, err := f.manager.Delete(t.Context(), f.command())
	require.NoError(t, err)
	require.True(t, pending)
	for _, table := range []string{"users", "workspace_members", "user_memories", "user_external_identities"} {
		deletionCount(t, f, "SELECT count(*) FROM "+table+" WHERE user_id=$1", 0, f.user)
	}
	deletionCount(t, f, `SELECT count(*) FROM principals WHERE principal_id=$1`, 0, f.principal)
	deletionCount(t, f, `SELECT count(*) FROM api_credentials WHERE principal_id=$1`, 0, f.principal)
	deletionCount(t, f, `SELECT count(*) FROM stories WHERE id=$1 AND reporter_id=$2 AND assignee_id IS NULL`, 1, f.story, usersdomain.DeletedUserID)
	deletionCount(t, f, `SELECT count(*) FROM story_comments WHERE comment_id=$1 AND commenter_id=$2 AND content='private comment'`, 1, f.comment, usersdomain.DeletedUserID)
	deletionCount(t, f, `SELECT count(*) FROM story_comments WHERE comment_id=$1 AND content='other reply' AND parent_id=$2`, 1, f.reply, f.comment)
	deletionCount(t, f, `SELECT count(*) FROM documents WHERE document_id=$1`, 0, f.privateDocument)
	deletionCount(t, f, `SELECT count(*) FROM documents WHERE document_id=$1 AND content_text='Team knowledge' AND created_by=$2 AND updated_by=$2`, 1, f.sharedDocument, usersdomain.DeletedUserID)
	deletionCount(t, f, `SELECT count(*) FROM attachments WHERE attachment_id=$1 AND uploaded_by IS NULL`, 1, f.shared)
	deletionCount(t, f, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, 0, f.orphan)
	deletionCount(t, f, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1 AND container_name='attachments'`, 1, f.orphan)
	deletionCount(t, f, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE blob_name='private-avatar.png' AND container_name='profiles'`, 1)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND is_active`, 1, f.other)
	deletionCount(t, f, `SELECT count(*) FROM account_subscriber_deletions WHERE email='departing@example.com'`, 1)
	_, active, err := usersrepository.New(f.db.Pool).ResolveActiveBrowserSessionVersion(t.Context(), f.user)
	require.NoError(t, err)
	require.False(t, active)
}

func TestAccountDeletionConflictDoesNotEraseAnythingAndDeletedWorkspaceDoesNotBlock(t *testing.T) {
	f := newDeletionFixture(t)
	execDeletion(t, t.Context(), f.db.Pool, `UPDATE workspace_members SET role='member' WHERE user_id=$1`, f.other)
	_, err := f.manager.Delete(t.Context(), f.command())
	var conflict *usersdomain.AccountDeletionConflict
	require.ErrorAs(t, err, &conflict)
	require.Contains(t, conflict.Workspaces, "Shared workspace")
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND is_active AND email='departing@example.com'`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM story_comments WHERE comment_id=$1`, 1, f.comment)
	execDeletion(t, t.Context(), f.db.Pool, `UPDATE workspaces SET deleted_at=now(),deleted_by=$1 WHERE workspace_id=$2`, f.user, f.workspace)
	_, err = f.manager.Delete(t.Context(), f.command())
	require.NoError(t, err)
}

func TestAccountDeletionPendingKeyIsAnonymousCannotReactivateAndFinalizesOnce(t *testing.T) {
	f := newDeletionFixture(t)
	f.cleanup.pending = true
	pending, err := f.manager.Delete(t.Context(), f.command())
	require.NoError(t, err)
	require.True(t, pending)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND NOT is_active AND full_name IS NULL AND avatar_url IS NULL AND email<> 'departing@example.com' AND github_access_token IS NULL`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM workspace_members WHERE user_id=$1`, 0, f.user)
	_, err = f.db.Pool.Exec(t.Context(), `UPDATE users SET is_active=TRUE WHERE user_id=$1`, f.user)
	require.ErrorContains(t, err, "cannot be changed or reactivated")
	count, err := f.manager.FinalizePending(t.Context(), 50)
	require.NoError(t, err)
	require.Zero(t, count)
	f.cleanup.pending = false
	count, err = f.manager.FinalizePending(t.Context(), 50)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = f.manager.FinalizePending(t.Context(), 50)
	require.NoError(t, err)
	require.Zero(t, count)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1`, 0, f.user)
	deletionCount(t, f, `SELECT count(*) FROM account_deletion_requests`, 0)
}

func TestAccountDeletionRollsBackPersonalDataRevocationAndObjectQueueOnLateFailure(t *testing.T) {
	f := newDeletionFixture(t)
	execDeletion(t, t.Context(), f.db.Pool, `CREATE FUNCTION reject_account_delete_test() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced late failure'; END $$; CREATE TRIGGER reject_account_delete_test BEFORE DELETE ON users FOR EACH ROW EXECUTE FUNCTION reject_account_delete_test()`)
	_, err := f.manager.Delete(t.Context(), f.command())
	require.ErrorContains(t, err, "forced late failure")
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND is_active AND email='departing@example.com'`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM api_credentials WHERE principal_id=$1`, 1, f.principal)
	deletionCount(t, f, `SELECT count(*) FROM story_comments WHERE comment_id=$1`, 1, f.comment)
	deletionCount(t, f, `SELECT count(*) FROM documents WHERE document_id=$1`, 1, f.privateDocument)
	deletionCount(t, f, `SELECT count(*) FROM attachment_object_deletion_outbox`, 0)
	deletionCount(t, f, `SELECT count(*) FROM account_subscriber_deletions`, 0)
}

func TestConcurrentAdministratorsCannotBothDeleteTheirAccounts(t *testing.T) {
	f := newDeletionFixture(t)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, id := range []uuid.UUID{f.user, f.other} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.manager.Delete(t.Context(), usersdomain.AccountDeletion{UserID: id, RequestedAt: time.Now().UTC()})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	succeeded, conflicted := 0, 0
	for err := range results {
		if err == nil {
			succeeded++
			continue
		}
		var conflict *usersdomain.AccountDeletionConflict
		if errors.As(err, &conflict) {
			conflicted++
		} else {
			t.Fatal(err)
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, conflicted)
}

func TestAccountDeletionErasesFeedbackIdentityAndOwnedOAuthApplication(t *testing.T) {
	f := newDeletionFixture(t)
	portal, board, contributor, otherContributor, item, comment, reply, application := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	ctx := t.Context()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_portals(id,workspace_id) VALUES ($1,$2)`, portal, f.workspace)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_boards(id,workspace_id,portal_id,team_id,name,slug) SELECT $1,$2,$3,team_id,'Feedback','feedback' FROM teams WHERE workspace_id=$2 LIMIT 1`, board, f.workspace, portal)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_contributors(id,portal_id,kind,user_id,display_name) VALUES ($1,$3,'account',$4,'Private display name'),($2,$3,'account',$5,'Other member')`, contributor, otherContributor, portal, f.user, f.other)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_items(id,workspace_id,board_id,portal_id,contributor_id,author_id,title,slug,description,description_html) VALUES ($1,$2,$3,$4,$5,$6,'Private idea','idea','Personal description','<p>Personal description</p>')`, item, f.workspace, board, portal, contributor, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_comments(id,workspace_id,item_id,author_id,contributor_id,body) VALUES ($1,$2,$3,$4,$5,'Private feedback comment')`, comment, f.workspace, item, f.user, contributor)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_comments(id,workspace_id,item_id,author_id,contributor_id,parent_id,body) VALUES ($1,$2,$3,$4,$5,$6,'Other feedback reply')`, reply, f.workspace, item, f.other, otherContributor, comment)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO oauth_applications(application_id,client_id,registration_kind,name,owner_workspace_id,owner_user_id,created_at,updated_at,expires_at) VALUES ($1,$2,'confidential','Personal integration',$3,$4,now(),now(),now()+interval '1 year')`, application, uuid.NewString(), f.workspace, f.user)
	installation, applicationPrincipal, organizationPrincipal := uuid.New(), uuid.New(), uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO principals(principal_id,workspace_id,kind,name,workspace_role,created_at,updated_at) VALUES ($1,$3,'oauth_application','Owned integration','member',now(),now()),($2,$3,'service_account','Organization bot','member',now(),now())`, applicationPrincipal, organizationPrincipal, f.workspace)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO oauth_application_installations(installation_id,application_id,workspace_id,principal_id,resource,installed_by_user_id,created_at,updated_at) VALUES ($1,$2,$3,$4,'https://api.example.com',$5,now(),now())`, installation, application, f.workspace, applicationPrincipal, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO oauth_client_secrets(secret_id,application_id,lookup_prefix,secret_digest,digest_key_id,expires_at,created_by_user_id,created_at) VALUES ($1,$2,'bbbbbbbbbbbb',decode(repeat('ab',32),'hex'),'test',now()+interval '1 day',$3,now())`, uuid.New(), application, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO feedback_votes(item_id,user_id,contributor_id,workspace_id) VALUES ($1,$2,$3,$4)`, item, f.user, contributor, f.workspace)
	execDeletion(t, ctx, f.db.Pool, `UPDATE feedback_contributors SET blocked_at=now(),blocked_reason='Private moderation details' WHERE id=$1`, contributor)
	_, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	deletionCount(t, f, `SELECT count(*) FROM feedback_items WHERE id=$1 AND title='Private idea' AND description='Personal description' AND description_html='<p>Personal description</p>' AND author_id=$2`, 1, item, usersdomain.DeletedUserID)
	deletionCount(t, f, `SELECT count(*) FROM feedback_contributors WHERE id=$1 AND kind='anonymous' AND user_id IS NULL AND display_name IS NULL AND email IS NULL`, 1, contributor)
	deletionCount(t, f, `SELECT count(*) FROM feedback_comments WHERE id=$1 AND author_id=$2 AND body='Private feedback comment'`, 1, comment, usersdomain.DeletedUserID)
	deletionCount(t, f, `SELECT count(*) FROM feedback_comments WHERE id=$1 AND parent_id=$2 AND body='Other feedback reply'`, 1, reply, comment)
	deletionCount(t, f, `SELECT count(*) FROM feedback_votes WHERE item_id=$1 AND user_id IS NULL AND contributor_id=$2`, 1, item, contributor)
	deletionCount(t, f, `SELECT count(*) FROM feedback_contributors WHERE id=$1 AND blocked_at IS NOT NULL AND blocked_reason IS NULL`, 1, contributor)
	deletionCount(t, f, `SELECT count(*) FROM oauth_applications WHERE application_id=$1`, 0, application)
	deletionCount(t, f, `SELECT count(*) FROM oauth_application_installations WHERE installation_id=$1`, 0, installation)
	deletionCount(t, f, `SELECT count(*) FROM oauth_client_secrets WHERE application_id=$1`, 0, application)
	deletionCount(t, f, `SELECT count(*) FROM principals WHERE principal_id=$1`, 0, applicationPrincipal)
	deletionCount(t, f, `SELECT count(*) FROM principals WHERE principal_id=$1`, 1, organizationPrincipal)
}
