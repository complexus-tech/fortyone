//go:build integration

package documentsrepository

import (
	"testing"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestArchivedDocumentsRequireActiveOwnerAndRestoreWithoutRepublishing(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	document := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityRestricted, "Launch brief")
	members := []documentdomain.Member{{UserID: fixture.editorA, Role: "editor"}, {UserID: fixture.viewerA, Role: "viewer"}}
	_, err := repository.SetAccess(ctx, documentdomain.AccessInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID,
		Visibility: documentdomain.VisibilityRestricted, Members: members,
	})
	require.NoError(t, err)
	_, err = repository.AddRelationship(ctx, documentdomain.RelationshipInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID,
		EntityType: documentdomain.RelationshipStory, EntityID: fixture.storyA,
	})
	require.NoError(t, err)
	shared, err := repository.SetPublicLink(ctx, fixture.workspaceA, fixture.ownerA, document.ID, true)
	require.NoError(t, err)
	require.NotNil(t, shared.PublicToken)
	token := *shared.PublicToken
	require.NoError(t, repository.Archive(ctx, fixture.workspaceA, fixture.ownerA, document.ID))
	var archivedToken *string
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT public_token FROM documents WHERE document_id = $1", document.ID).Scan(&archivedToken))
	require.Nil(t, archivedToken, "archiving revokes the old public link")
	_, err = repository.Get(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	_, err = repository.GetPublicDocument(ctx, token)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	active, err := repository.List(ctx, documentdomain.ListInput{WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA})
	require.NoError(t, err)
	require.Empty(t, active)
	archived, err := repository.List(ctx, documentdomain.ListInput{WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, Scope: "archived"})
	require.NoError(t, err)
	require.Len(t, archived, 1)
	require.Equal(t, document.ID, archived[0].ID)
	require.NotNil(t, archived[0].ArchivedAt)
	require.False(t, archived[0].CanEdit)
	require.False(t, archived[0].PublicEnabled)
	require.Len(t, archived[0].RelatedWork, 1)
	require.Equal(t, fixture.storyA, archived[0].RelatedWork[0].EntityID)

	for _, identity := range []struct {
		name            string
		workspace, user uuid.UUID
	}{
		{name: "editor", workspace: fixture.workspaceA, user: fixture.editorA},
		{name: "viewer", workspace: fixture.workspaceA, user: fixture.viewerA},
		{name: "guest", workspace: fixture.workspaceA, user: fixture.guestA},
		{name: "inactive member", workspace: fixture.workspaceA, user: fixture.inactiveA},
		{name: "another workspace", workspace: fixture.workspaceB, user: fixture.ownerB},
	} {
		t.Run(identity.name, func(t *testing.T) {
			results, err := repository.List(ctx, documentdomain.ListInput{WorkspaceID: identity.workspace, UserID: identity.user, Scope: "archived"})
			require.NoError(t, err)
			require.Empty(t, results)
			_, err = repository.Unarchive(ctx, identity.workspace, identity.user, document.ID)
			require.ErrorIs(t, err, documentdomain.ErrNotFound)
		})
	}

	assertOwnerDenied := func() {
		t.Helper()
		results, err := repository.List(ctx, documentdomain.ListInput{WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, Scope: "archived"})
		require.NoError(t, err)
		require.Empty(t, results)
		_, err = repository.Unarchive(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
		require.ErrorIs(t, err, documentdomain.ErrNotFound)
	}
	_, err = postgres.Pool.Exec(ctx, "UPDATE workspace_members SET role = 'guest' WHERE workspace_id = $1 AND user_id = $2", fixture.workspaceA, fixture.ownerA)
	require.NoError(t, err)
	assertOwnerDenied()
	_, err = postgres.Pool.Exec(ctx, "UPDATE workspace_members SET role = 'admin' WHERE workspace_id = $1 AND user_id = $2", fixture.workspaceA, fixture.ownerA)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, "UPDATE users SET is_active = FALSE WHERE user_id = $1", fixture.ownerA)
	require.NoError(t, err)
	assertOwnerDenied()
	_, err = postgres.Pool.Exec(ctx, "UPDATE users SET is_active = TRUE WHERE user_id = $1", fixture.ownerA)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, "DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2", fixture.workspaceA, fixture.ownerA)
	require.NoError(t, err)
	assertOwnerDenied()
	insertDocumentWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.ownerA, "admin")

	// Older archives retained their token. Restoring one must revoke it too.
	_, err = postgres.Pool.Exec(ctx, "UPDATE documents SET public_token = $2 WHERE document_id = $1", document.ID, token)
	require.NoError(t, err)
	restored, err := repository.Unarchive(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
	require.NoError(t, err)
	require.Nil(t, restored.ArchivedAt)
	require.Nil(t, restored.PublicToken)
	require.False(t, restored.PublicEnabled)
	require.Equal(t, document.Title, restored.Title)
	require.Equal(t, documentdomain.VisibilityRestricted, restored.Visibility)
	require.ElementsMatch(t, members, restored.SharedWith)
	require.Len(t, restored.RelatedWork, 1)
	_, err = repository.GetPublicDocument(ctx, token)
	require.ErrorIs(t, err, documentdomain.ErrNotFound)
	_, err = repository.Unarchive(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
	require.ErrorIs(t, err, documentdomain.ErrConflict)
	archived, err = repository.List(ctx, documentdomain.ListInput{WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, Scope: "archived"})
	require.NoError(t, err)
	require.Empty(t, archived)
	active, err = repository.List(ctx, documentdomain.ListInput{WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA})
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.False(t, active[0].PublicEnabled)
}

func TestDocumentSummariesRespectRelationshipVisibilityAndExposePublicStatus(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	document := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Project brief")
	_, err := repository.AddRelationship(ctx, documentdomain.RelationshipInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID,
		EntityType: documentdomain.RelationshipStory, EntityID: fixture.storyA,
	})
	require.NoError(t, err)
	_, err = repository.SetPublicLink(ctx, fixture.workspaceA, fixture.ownerA, document.ID, true)
	require.NoError(t, err)
	for _, userID := range []uuid.UUID{fixture.ownerA, fixture.editorA, fixture.viewerA} {
		results, err := repository.List(ctx, documentdomain.ListInput{WorkspaceID: fixture.workspaceA, UserID: userID})
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.True(t, results[0].PublicEnabled)
		require.Nil(t, results[0].ArchivedAt)
		if userID == fixture.viewerA {
			require.Empty(t, results[0].RelatedWork, "non-team reader must not learn linked work details")
			require.Zero(t, results[0].RelatedWorkCount)
		} else {
			require.Len(t, results[0].RelatedWork, 1)
			require.Equal(t, 1, results[0].RelatedWorkCount)
			require.Equal(t, fixture.storyA, results[0].RelatedWork[0].EntityID)
			require.NotEmpty(t, results[0].RelatedWork[0].Title)
		}
	}
	reader, err := repository.Get(ctx, fixture.workspaceA, fixture.viewerA, document.ID)
	require.NoError(t, err)
	require.True(t, reader.PublicEnabled)
	require.Nil(t, reader.PublicToken, "public status must not reveal the owner's link token")
	related, err := repository.ListRelatedDocuments(ctx, fixture.workspaceA, fixture.ownerA, documentdomain.RelationshipStory, fixture.storyA)
	require.NoError(t, err)
	require.Len(t, related, 1)
	require.True(t, related[0].PublicEnabled)
	require.Len(t, related[0].RelatedWork, 1)
	require.Equal(t, 1, related[0].RelatedWorkCount)
}
