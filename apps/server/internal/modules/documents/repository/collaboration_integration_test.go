//go:build integration

package documentsrepository

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

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
