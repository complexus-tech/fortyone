//go:build integration

package documentsrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationrepo "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestDocumentCommentMentionsReplayAuthorizationAndDelivery(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, db.Pool)
	repo := New(db.Pool)
	doc := createDocument(t, ctx, repo, f.workspaceA, f.ownerA, domain.VisibilityRestricted, "Private plan")
	_, err := repo.SetAccess(ctx, domain.AccessInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, Visibility: domain.VisibilityRestricted, Members: []domain.Member{{UserID: f.editorA, Role: "editor"}}})
	require.NoError(t, err)
	creation := uuid.New()
	input := domain.CreateCommentInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, CreationID: &creation, Body: fmt.Sprintf("[@Spoof](mention:%s) [@Hidden](mention:%s)", f.editorA, f.viewerA)}
	thread, err := repo.CreateComment(ctx, input)
	require.NoError(t, err)
	require.Contains(t, thread.Comments[0].Body, "[@Documents editor-a]")
	require.Contains(t, thread.Comments[0].Body, "@Hidden")
	require.NotContains(t, thread.Comments[0].Body, "mention:"+f.viewerA.String())
	replay, err := repo.CreateComment(ctx, input)
	require.NoError(t, err)
	require.Equal(t, thread.ID, replay.ID)
	input.Body = "changed"
	_, err = repo.CreateComment(ctx, input)
	require.ErrorIs(t, err, domain.ErrConflict)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1", doc.ID).Scan(&count))
	require.Equal(t, 1, count)
	dispatch := notificationrepo.New(db.Pool)
	inbox, err := dispatch.List(ctx, notifications.ListQuery{Access: notifications.WorkspaceAccess{WorkspaceID: f.workspaceA, ActorID: f.editorA}, Limit: 20})
	require.NoError(t, err)
	require.Len(t, inbox, 1)
	attempts := 0
	_, err = dispatch.DispatchDocumentNotifications(ctx, 10, func(context.Context, notifications.Notification) error {
		attempts++
		return errors.New("redis unavailable")
	})
	require.Error(t, err)
	delivered, err := dispatch.DispatchDocumentNotifications(ctx, 10, func(_ context.Context, n notifications.Notification) error {
		attempts++
		require.Equal(t, f.editorA, n.RecipientID)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, delivered)
	require.Equal(t, 2, attempts)
	delivered, err = dispatch.DispatchDocumentNotifications(ctx, 10, func(context.Context, notifications.Notification) error { t.Fatal("delivered twice"); return nil })
	require.NoError(t, err)
	require.Zero(t, delivered)
	replyID := uuid.New()
	parent := thread.Comments[0].ID
	_, err = repo.ReplyToComment(ctx, domain.ReplyCommentInput{WorkspaceID: f.workspaceA, UserID: f.editorA, DocumentID: doc.ID, ThreadID: thread.ID, ReplyToCommentID: &parent, CreationID: &replyID, Body: fmt.Sprintf("[@Owner](mention:%s)", f.ownerA)})
	require.NoError(t, err)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1 AND recipient_id=$2", doc.ID, f.ownerA).Scan(&count))
	require.Equal(t, 1, count, "mention and reply notification deduplicate")
	_, err = repo.SetAccess(ctx, domain.AccessInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, Visibility: domain.VisibilityPrivate})
	require.NoError(t, err)
	_, err = repo.ListCommentMentions(ctx, f.workspaceA, f.editorA, doc.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	inbox, err = dispatch.List(ctx, notifications.ListQuery{Access: notifications.WorkspaceAccess{WorkspaceID: f.workspaceA, ActorID: f.editorA}, Limit: 20})
	require.NoError(t, err)
	require.Empty(t, inbox, "revoked document access removes notification visibility")
	// Newly queued notification to a now-revoked reader must never be fanned out.
	_, err = db.Pool.Exec(ctx, "UPDATE document_notification_deliveries SET delivered_at=NULL WHERE notification_id IN (SELECT notification_id FROM notifications WHERE recipient_id=$1)", f.editorA)
	require.NoError(t, err)
	_, err = dispatch.DispatchDocumentNotifications(ctx, 10, func(_ context.Context, n notifications.Notification) error {
		require.NotEqual(t, f.editorA, n.RecipientID)
		return nil
	})
	require.NoError(t, err)
}

func TestDocumentCommentCreationIsAtomicAndConcurrentReplaySafe(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, db.Pool)
	repo := New(db.Pool)
	doc := createDocument(t, ctx, repo, f.workspaceA, f.ownerA, domain.VisibilityWorkspace, "Atomic")
	input := domain.CreateCommentInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, Body: fmt.Sprintf("[@Editor](mention:%s)", f.editorA)}
	failure := errors.New("invocation failed")
	err := database.WithinUnitOfWork(ctx, db.Pool, func(txctx context.Context, _ pgx.Tx) error {
		_, err := repo.CreateComment(txctx, input)
		require.NoError(t, err)
		return failure
	})
	require.ErrorIs(t, err, failure)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM document_comment_threads WHERE document_id=$1", doc.ID).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1", doc.ID).Scan(&count))
	require.Zero(t, count)
	id := uuid.New()
	input.CreationID = &id
	var wg sync.WaitGroup
	results := make(chan domain.CommentThread, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); thread, err := repo.CreateComment(ctx, input); results <- thread; errs <- err }()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var first uuid.UUID
	for thread := range results {
		if first == uuid.Nil {
			first = thread.ID
		}
		require.Equal(t, first, thread.ID)
	}
}

func TestDocumentBodyReferencesAndMentionsTrackFinalOccurrence(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, db.Pool)
	repo := New(db.Pool)
	doc := createDocument(t, ctx, repo, f.workspaceA, f.ownerA, domain.VisibilityWorkspace, "References")
	atom := fmt.Sprintf(`<span data-type="document-reference" data-entity-type="story" data-id="%s">story</span>`, f.storyA)
	mention := fmt.Sprintf(`<span data-type="mention" data-mention-kind="user" data-id="%s">Editor</span>`, f.editorA)
	body := "<p>" + atom + atom + mention + "</p>"
	updated, err := repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: doc.Revision, ContentHTML: &body})
	require.NoError(t, err)
	require.Len(t, updated.RelatedWork, 1)
	require.True(t, updated.RelatedWork[0].ReferencedInContent)
	relation := domain.RelationshipInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, EntityType: domain.RelationshipStory, EntityID: f.storyA}
	require.ErrorIs(t, repo.RemoveRelationship(ctx, relation), domain.ErrInlineReference)
	_, err = repo.AddRelationship(ctx, relation)
	require.NoError(t, err)
	body = "<p>" + mention + " ordinary prose edit</p>"
	updated, err = repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: updated.Revision, ContentHTML: &body})
	require.NoError(t, err)
	require.Len(t, updated.RelatedWork, 1)
	require.False(t, updated.RelatedWork[0].ReferencedInContent, "manual reference survives body removal")
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1", doc.ID).Scan(&count))
	require.Equal(t, 1, count, "unchanged mention is not repeated")
	require.NoError(t, repo.RemoveRelationship(ctx, relation))
	foreign := fmt.Sprintf(`<p><span data-type="document-reference" data-entity-type="story" data-id="%s">secret</span></p>`, f.storyB)
	_, err = repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: updated.Revision, ContentHTML: &foreign})
	require.Error(t, err)
	current, err := repo.Get(ctx, f.workspaceA, f.ownerA, doc.ID)
	require.NoError(t, err)
	require.Equal(t, body, current.ContentHTML, "unauthorized reference rolls back content")
}

func TestDocumentRestoreAndDuplicateReconcileBodyReferences(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, db.Pool)
	repo := New(db.Pool)
	body := fmt.Sprintf(`<p><span data-type="document-reference" data-entity-type="story" data-id="%s">Visible story</span></p>`, f.storyA)
	doc, err := repo.Create(ctx, domain.CreateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, Title: "History", Visibility: domain.VisibilityWorkspace, ContentHTML: body})
	require.NoError(t, err)
	copy, err := repo.Duplicate(ctx, f.workspaceA, f.ownerA, doc.ID)
	require.NoError(t, err)
	require.Len(t, copy.RelatedWork, 1)
	require.True(t, copy.RelatedWork[0].ReferencedInContent)
	copy, err = repo.Duplicate(ctx, f.workspaceA, f.viewerA, doc.ID)
	require.NoError(t, err)
	require.Empty(t, copy.RelatedWork)
	require.NotContains(t, copy.ContentHTML, "data-type")
	require.Contains(t, copy.ContentHTML, "Visible story")
	empty := "<p>Removed</p>"
	updated, err := repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: doc.Revision, ContentHTML: &empty})
	require.NoError(t, err)
	require.Empty(t, updated.RelatedWork)
	restored, err := repo.RestoreRevision(ctx, f.workspaceA, f.ownerA, doc.ID, doc.Revision, updated.Revision)
	require.NoError(t, err)
	require.Len(t, restored.RelatedWork, 1)
	// Restore the no-reference version after forcing a checkpoint boundary.
	_, err = db.Pool.Exec(ctx, "UPDATE document_revisions SET created_at=created_at-interval '10 minutes' WHERE document_id=$1", doc.ID)
	require.NoError(t, err)
	removed, err := repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: restored.Revision, ContentHTML: &empty})
	require.NoError(t, err)
	require.Empty(t, removed.RelatedWork)
	_, err = db.Pool.Exec(ctx, "DELETE FROM team_members WHERE team_id=$1 AND user_id=$2", f.teamA, f.ownerA)
	require.NoError(t, err)
	_, err = repo.RestoreRevision(ctx, f.workspaceA, f.ownerA, doc.ID, doc.Revision, removed.Revision)
	require.Error(t, err)
	current, err := repo.Get(ctx, f.workspaceA, f.ownerA, doc.ID)
	require.NoError(t, err)
	require.Equal(t, removed.Revision, current.Revision)
	require.Equal(t, empty, current.ContentHTML)
}

func TestDocumentMentionObservationDoesNotNotifyWhenAccessChanges(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, db.Pool)
	repo := New(db.Pool)
	body := fmt.Sprintf(`<p><span data-type="mention" data-mention-kind="user" data-id="%s">Editor</span></p>`, f.editorA)
	source, err := repo.Create(ctx, domain.CreateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, Title: "Source", Visibility: domain.VisibilityWorkspace, ContentHTML: body})
	require.NoError(t, err)
	copy, err := repo.Duplicate(ctx, f.workspaceA, f.ownerA, source.ID)
	require.NoError(t, err)
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM document_body_mentions WHERE document_id=$1 AND user_id=$2", copy.ID, f.editorA).Scan(&count))
	require.Equal(t, 1, count, "private copy remembers the existing mention without notifying")
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1", copy.ID).Scan(&count))
	require.Zero(t, count)
	shared, err := repo.SetAccess(ctx, domain.AccessInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: copy.ID, Visibility: domain.VisibilityWorkspace})
	require.NoError(t, err)
	edited := body + "<p>Unrelated edit</p>"
	updated, err := repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: copy.ID, ExpectedRevision: shared.Revision, ContentHTML: &edited})
	require.NoError(t, err)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1", copy.ID).Scan(&count))
	require.Zero(t, count, "granting access does not make an existing mention new")
	empty := "<p>Removed mention</p>"
	updated, err = repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: copy.ID, ExpectedRevision: updated.Revision, ContentHTML: &empty})
	require.NoError(t, err)
	_, err = repo.Update(ctx, domain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: copy.ID, ExpectedRevision: updated.Revision, ContentHTML: &body})
	require.NoError(t, err)
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM notifications WHERE entity_id=$1", copy.ID).Scan(&count))
	require.Equal(t, 1, count, "an explicit later insertion still notifies")
}
