//go:build integration

package documentsrepository

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/complexus-tech/projects-api/internal/migrations"
	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestDocumentHistoryRestoreAndPublicLinkRevocation(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, postgres.Pool)
	r := New(postgres.Pool)
	document := createDocument(t, ctx, r, f.workspaceA, f.ownerA, documentdomain.VisibilityPrivate, "Original")
	revisions, err := r.ListRevisions(ctx, f.workspaceA, f.ownerA, document.ID, 0)
	require.NoError(t, err)
	require.Len(t, revisions, 1)
	title := "Updated"
	updated, err := r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: document.ID, ExpectedRevision: 1, Title: &title})
	require.NoError(t, err)
	require.EqualValues(t, 2, updated.Revision)
	_, err = r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: document.ID, ExpectedRevision: 1, Title: &title})
	require.ErrorIs(t, err, documentdomain.ErrConflict)
	_, err = r.GetRevision(ctx, f.workspaceA, f.outsiderA, document.ID, 1)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	_, err = r.SetPublicLink(ctx, f.workspaceA, f.outsiderA, document.ID, true)
	require.Error(t, err)
	shared, err := r.SetPublicLink(ctx, f.workspaceA, f.ownerA, document.ID, true)
	require.NoError(t, err)
	require.NotNil(t, shared.PublicToken)
	token := *shared.PublicToken
	public, err := r.GetPublicDocument(ctx, token)
	require.NoError(t, err)
	require.Equal(t, title, public.Title)

	restored, err := r.RestoreRevision(ctx, f.workspaceA, f.ownerA, document.ID, 1, updated.Revision)
	require.NoError(t, err)
	require.EqualValues(t, 3, restored.Revision)
	require.EqualValues(t, 2, restored.CollaborationEpoch)
	require.Equal(t, "Original", restored.Title)
	require.Equal(t, documentdomain.VisibilityPrivate, restored.Visibility)
	require.Equal(t, token, *restored.PublicToken)
	revisions, err = r.ListRevisions(ctx, f.workspaceA, f.ownerA, document.ID, 0)
	require.NoError(t, err)
	require.Len(t, revisions, 3)
	_, err = r.SetPublicLink(ctx, f.workspaceA, f.ownerA, document.ID, false)
	require.NoError(t, err)
	_, err = r.GetPublicDocument(ctx, token)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	reshared, err := r.SetPublicLink(ctx, f.workspaceA, f.ownerA, document.ID, true)
	require.NoError(t, err)
	require.NotEqual(t, token, *reshared.PublicToken)
	require.NoError(t, r.Archive(ctx, f.workspaceA, f.ownerA, document.ID))
	_, err = r.GetPublicDocument(ctx, *reshared.PublicToken)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
}

func TestCollaborationTicketsAndLegacyWriterFence(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, postgres.Pool)
	r := New(postgres.Pool)
	document := createDocument(t, ctx, r, f.workspaceA, f.ownerA, documentdomain.VisibilityPrivate, "Plan")
	_, err := r.CreateCollaborationSession(ctx, f.workspaceB, f.ownerB, document.ID)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	ticket, err := r.CreateCollaborationSession(ctx, f.workspaceA, f.ownerA, document.ID)
	require.NoError(t, err)
	digest := sha256.Sum256([]byte(ticket.Token))
	var stored string
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT token_hash FROM document_collaboration_sessions WHERE document_id=$1", document.ID).Scan(&stored))
	require.Equal(t, hex.EncodeToString(digest[:]), stored)
	_, err = postgres.Pool.Exec(ctx, "UPDATE documents SET collaboration_state=$2 WHERE document_id=$1", document.ID, []byte{0, 0})
	require.NoError(t, err)
	title := "Stale HTML writer"
	_, err = r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: document.ID, ExpectedRevision: 2, Title: &title})
	require.True(t, errors.Is(err, documentdomain.ErrConflict))
	_, err = postgres.Pool.Exec(ctx, "UPDATE documents SET title=$2 WHERE document_id=$1", document.ID, title)
	require.Error(t, err, "database must also fence an old deployed API")
	restored, err := r.RestoreRevision(ctx, f.workspaceA, f.ownerA, document.ID, 1, 2)
	require.NoError(t, err)
	require.False(t, restored.Collaborative)
	require.EqualValues(t, 2, restored.CollaborationEpoch)
	duplicate, err := r.Duplicate(ctx, f.workspaceA, f.ownerA, document.ID)
	require.NoError(t, err)
	revisions, err := r.ListRevisions(ctx, f.workspaceA, f.ownerA, duplicate.ID, 0)
	require.NoError(t, err)
	require.Len(t, revisions, 1)
}

func TestPublicMediaIsScopedToCurrentContentAndHistoryRetainsFiles(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, postgres.Pool)
	r := New(postgres.Pool)
	document := createDocument(t, ctx, r, f.workspaceA, f.ownerA, documentdomain.VisibilityPrivate, "Media")
	attachment := insertDocumentAttachment(t, ctx, postgres.Pool, f.workspaceA, f.ownerA, "retained")
	foreign := insertDocumentAttachment(t, ctx, postgres.Pool, f.workspaceB, f.ownerB, "foreign")
	media := documentdomain.MediaInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: document.ID, AttachmentID: attachment}
	require.NoError(t, r.LinkMedia(ctx, media))
	html := fmt.Sprintf(`<p>Image</p><img src="/workspaces/team/documents/%s/media/%s">`, document.ID, attachment)
	updated, err := r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: document.ID, ExpectedRevision: 1, ContentHTML: &html})
	require.NoError(t, err)
	shared, err := r.SetPublicLink(ctx, f.workspaceA, f.ownerA, document.ID, true)
	require.NoError(t, err)
	token := *shared.PublicToken
	_, err = r.AuthorizePublicMedia(ctx, token, attachment)
	require.NoError(t, err)
	_, err = r.AuthorizePublicMedia(ctx, token, foreign)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	_, err = postgres.Pool.Exec(ctx, "UPDATE document_revisions SET created_at=NOW()-INTERVAL '3 minutes' WHERE document_id=$1 AND revision=$2", document.ID, updated.Revision)
	require.NoError(t, err)
	empty := "<p>Removed image</p>"
	_, err = r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: document.ID, ExpectedRevision: updated.Revision, ContentHTML: &empty})
	require.NoError(t, err)
	_, err = r.AuthorizePublicMedia(ctx, token, attachment)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	_, err = r.UnlinkMedia(ctx, media)
	require.ErrorIs(t, err, documentdomain.ErrConflict)
	require.NoError(t, r.AuthorizeMedia(ctx, media), "historical version can still load its image")
	media.UserID = f.outsiderA
	_, err = r.UnlinkMedia(ctx, media)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
}

func TestHistoryGroupsAutosavesAndBoundsRetention(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, postgres.Pool)
	r := New(postgres.Pool)
	doc := createDocument(t, ctx, r, f.workspaceA, f.ownerA, documentdomain.VisibilityWorkspace, "Original")
	save := func(title string) {
		t.Helper()
		next, err := r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: doc.Revision, Title: &title})
		require.NoError(t, err)
		doc = next
	}
	list := func() []documentdomain.Revision {
		t.Helper()
		rows, err := r.ListRevisions(ctx, f.workspaceA, f.ownerA, doc.ID, 0)
		require.NoError(t, err)
		return rows
	}
	for i := 0; i < 30; i++ {
		save(fmt.Sprintf("Typing %d", i))
	}
	rows := list()
	require.Len(t, rows, 2, "initial content plus one rolling editing checkpoint")
	require.EqualValues(t, 31, doc.Revision, "conflict detection still advances on every save")
	snapshot, err := r.GetRevision(ctx, f.workspaceA, f.ownerA, doc.ID, rows[0].Revision)
	require.NoError(t, err)
	require.Equal(t, "Typing 29", snapshot.Title)
	_, err = r.GetRevision(ctx, f.workspaceA, f.ownerA, doc.ID, 2)
	require.ErrorIs(t, err, documentdomain.ErrNotFound, "superseded previews cannot restore different content")
	save("Typing 29")
	require.Len(t, list(), 2, "unchanged saves do not make checkpoints")
	_, err = postgres.Pool.Exec(ctx, "UPDATE document_revisions SET created_at=NOW()-INTERVAL '3 minutes' WHERE document_id=$1 AND revision=$2", doc.ID, doc.Revision)
	require.NoError(t, err)
	save("After a pause")
	require.Len(t, list(), 3)
	_, err = postgres.Pool.Exec(ctx, "UPDATE document_revisions SET started_at=NOW()-INTERVAL '11 minutes' WHERE document_id=$1 AND revision=$2", doc.ID, doc.Revision)
	require.NoError(t, err)
	save("Ten minute checkpoint")
	require.Len(t, list(), 4)
	other := f.editorA
	nextTitle := "Another editor"
	doc, err = r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: other, DocumentID: doc.ID, ExpectedRevision: doc.Revision, Title: &nextTitle})
	require.NoError(t, err)
	require.Len(t, list(), 5)
	for i := 0; i < 14; i++ {
		_, err = postgres.Pool.Exec(ctx, "UPDATE document_revisions SET created_at=NOW()-INTERVAL '3 minutes' WHERE document_id=$1 AND revision=$2", doc.ID, doc.Revision)
		require.NoError(t, err)
		save(fmt.Sprintf("Session %d", i))
	}
	rows = list()
	require.Len(t, rows, 10)
	var storedCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT count(*) FROM document_revisions WHERE document_id=$1", doc.ID).Scan(&storedCount))
	require.Equal(t, 10, storedCount)
	require.Equal(t, "Session 13", rows[0].Title)
	_, err = r.GetRevision(ctx, f.workspaceA, f.ownerA, doc.ID, 1)
	require.ErrorIs(t, err, documentdomain.ErrNotFound, "oldest checkpoint expires")
	oldLatest := doc.Revision
	restored, err := r.RestoreRevision(ctx, f.workspaceA, f.ownerA, doc.ID, rows[8].Revision, doc.Revision)
	require.NoError(t, err)
	doc = restored
	rows = list()
	require.Len(t, rows, 10)
	require.Equal(t, oldLatest, rows[1].Revision, "restore keeps pre-restore state")
	save("Edit after restore")
	rows = list()
	require.Len(t, rows, 10)
	require.Equal(t, restored.Revision, rows[1].Revision, "first edit cannot overwrite restore checkpoint")
}

func TestCheckpointMigrationConsolidatesAndCapsExistingHistory(t *testing.T) {
	postgres := testkit.NewPostgresAtMigration(t, 194)
	ctx := t.Context()
	f := newDocumentFixture(t, ctx, postgres.Pool)
	r := New(postgres.Pool)
	doc := createDocument(t, ctx, r, f.workspaceA, f.ownerA, documentdomain.VisibilityPrivate, "Baseline")
	for i := 0; i < 25; i++ {
		title := fmt.Sprintf("Old autosave %d", i)
		next, err := r.Update(ctx, documentdomain.UpdateInput{WorkspaceID: f.workspaceA, UserID: f.ownerA, DocumentID: doc.ID, ExpectedRevision: doc.Revision, Title: &title})
		require.NoError(t, err)
		doc = next
	}
	// Create distinct old sessions plus a burst of five saves in the newest bucket.
	_, err := postgres.Pool.Exec(ctx, `UPDATE document_revisions SET created_at = CASE WHEN revision <=21 THEN date_trunc('hour',NOW()) - (26-revision)*INTERVAL '20 minutes' ELSE date_trunc('hour',NOW()) END WHERE document_id=$1`, doc.ID)
	require.NoError(t, err)
	script, err := migrations.FS.ReadFile("000195_document_history_checkpoints.up.sql")
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, string(script))
	require.NoError(t, err)
	var count, burst int
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT count(*), count(*) FILTER (WHERE revision>21) FROM document_revisions WHERE document_id=$1", doc.ID).Scan(&count, &burst))
	require.Equal(t, 10, count)
	require.Equal(t, 1, burst)
	latest, err := r.GetRevision(ctx, f.workspaceA, f.ownerA, doc.ID, doc.Revision)
	require.NoError(t, err)
	require.Equal(t, "Old autosave 24", latest.Title)
}
