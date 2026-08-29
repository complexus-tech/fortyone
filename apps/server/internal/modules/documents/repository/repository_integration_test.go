//go:build integration

package documentsrepository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryEnforcesTenantVisibilityAndEditPolicy(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)

	workspaceDocument := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Workspace plan")
	guestView, err := repository.Get(ctx, fixture.workspaceA, fixture.guestA, workspaceDocument.ID)
	if err != nil || guestView.CanEdit {
		t.Fatalf("guest workspace read = %#v, %v; want read-only", guestView, err)
	}
	workspaceList, err := repository.List(ctx, documentdomain.ListInput{
		WorkspaceID: fixture.workspaceA,
		UserID:      fixture.outsiderA,
	})
	if err != nil || len(workspaceList) != 1 || workspaceList[0].ID != workspaceDocument.ID {
		t.Fatalf("workspace List() = %#v, %v; want visible workspace document", workspaceList, err)
	}
	if _, err := repository.Create(ctx, documentdomain.CreateInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.guestA,
		Title: "Denied", Visibility: documentdomain.VisibilityWorkspace,
	}); !errors.Is(err, documentdomain.ErrForbidden) {
		t.Fatalf("guest Create() error = %v, want ErrForbidden", err)
	}

	restricted := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityRestricted, "Restricted plan")
	updated, err := repository.SetAccess(ctx, documentdomain.AccessInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: restricted.ID,
		Visibility: documentdomain.VisibilityRestricted,
		Members: []documentdomain.Member{
			{UserID: fixture.viewerA, Role: "viewer"},
			{UserID: fixture.editorA, Role: "editor"},
			{UserID: fixture.guestA, Role: "editor"},
		},
	})
	if err != nil {
		t.Fatalf("SetAccess() error = %v", err)
	}
	if !membersEqualUnordered(updated.SharedWith, []documentdomain.Member{
		{UserID: fixture.viewerA, Role: "viewer"},
		{UserID: fixture.editorA, Role: "editor"},
		{UserID: fixture.guestA, Role: "viewer"},
	}) {
		t.Fatalf("SetAccess() members = %#v", updated.SharedWith)
	}

	viewerDocument, err := repository.Get(ctx, fixture.workspaceA, fixture.viewerA, restricted.ID)
	if err != nil || viewerDocument.CanEdit {
		t.Fatalf("viewer Get() = %#v, %v; want read-only", viewerDocument, err)
	}
	newTitle := "viewer overwrite"
	if _, err := repository.Update(ctx, documentdomain.UpdateInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.viewerA,
		DocumentID: restricted.ID, Title: &newTitle,
	}); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("viewer Update() error = %v, want hidden not found", err)
	}
	editorTitle := "Editor revision"
	if _, err := repository.Update(ctx, documentdomain.UpdateInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.editorA,
		DocumentID: restricted.ID, Title: &editorTitle,
	}); err != nil {
		t.Fatalf("editor Update() error = %v", err)
	}
	guestRestricted, err := repository.Get(ctx, fixture.workspaceA, fixture.guestA, restricted.ID)
	if err != nil || guestRestricted.CanEdit {
		t.Fatalf("guest restricted Get() = %#v, %v; want downgraded viewer", guestRestricted, err)
	}
	sharedList, err := repository.List(ctx, documentdomain.ListInput{
		WorkspaceID: fixture.workspaceA,
		UserID:      fixture.viewerA,
		Scope:       "shared",
	})
	if err != nil || len(sharedList) != 1 || sharedList[0].ID != restricted.ID {
		t.Fatalf("shared List() = %#v, %v; want only explicitly shared document", sharedList, err)
	}

	for name, identity := range map[string]struct {
		workspaceID uuid.UUID
		userID      uuid.UUID
	}{
		"unshared member": {workspaceID: fixture.workspaceA, userID: fixture.outsiderA},
		"inactive member": {workspaceID: fixture.workspaceA, userID: fixture.inactiveA},
		"cross tenant":    {workspaceID: fixture.workspaceB, userID: fixture.ownerB},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := repository.Get(ctx, identity.workspaceID, identity.userID, restricted.ID); !errors.Is(err, documentdomain.ErrNotFound) {
				t.Fatalf("Get() error = %v, want ErrNotFound", err)
			}
		})
	}

	// Invalid targets must roll back both the visibility update and member
	// replacement. A cross-workspace identifier cannot partially erase access.
	if _, err := repository.SetAccess(ctx, documentdomain.AccessInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: restricted.ID,
		Visibility: documentdomain.VisibilityRestricted,
		Members: []documentdomain.Member{
			{UserID: fixture.viewerA, Role: "viewer"},
			{UserID: fixture.ownerB, Role: "viewer"},
		},
	}); !errors.Is(err, documentdomain.ErrInvalidInput) {
		t.Fatalf("invalid SetAccess() error = %v, want ErrInvalidInput", err)
	}
	afterRollback, err := repository.Get(ctx, fixture.workspaceA, fixture.editorA, restricted.ID)
	if err != nil || afterRollback.Visibility != documentdomain.VisibilityRestricted || !afterRollback.CanEdit {
		t.Fatalf("access rollback = %#v, %v", afterRollback, err)
	}

	// Private is creator-only even if a legacy or interrupted write leaves a
	// stale document_members row behind. Authorization must be derived from the
	// current document state, not merely the existence of an old grant.
	if _, err := postgres.Pool.Exec(ctx, `
		UPDATE documents SET visibility = 'private' WHERE document_id = $1
	`, restricted.ID); err != nil {
		t.Fatalf("make document private with stale members: %v", err)
	}
	if _, err := repository.Get(ctx, fixture.workspaceA, fixture.editorA, restricted.ID); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("stale private editor Get() error = %v, want ErrNotFound", err)
	}
	privateOwnerView, err := repository.Get(ctx, fixture.workspaceA, fixture.ownerA, restricted.ID)
	if err != nil || privateOwnerView.Visibility != documentdomain.VisibilityPrivate || len(privateOwnerView.SharedWith) != 0 {
		t.Fatalf("private owner Get() = %#v, %v; want creator-only view without grants", privateOwnerView, err)
	}

	if err := repository.Archive(ctx, fixture.workspaceA, fixture.editorA, restricted.ID); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("non-owner Archive() error = %v, want ErrNotFound", err)
	}
	if _, err := repository.Delete(ctx, fixture.workspaceB, fixture.ownerB, restricted.ID); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v, want ErrNotFound", err)
	}
}

func TestRepositoryScopesRelationshipsToDocumentAndTeamAccess(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	document := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Linked work")
	input := documentdomain.RelationshipInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID,
		EntityType: documentdomain.RelationshipStory, EntityID: fixture.storyA,
	}
	if _, err := repository.AddRelationship(ctx, input); err != nil {
		t.Fatalf("AddRelationship() error = %v", err)
	}
	if _, err := repository.AddRelationship(ctx, input); err != nil {
		t.Fatalf("idempotent AddRelationship() error = %v", err)
	}

	ownerView, err := repository.Get(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
	if err != nil || len(ownerView.RelatedWork) != 1 || ownerView.RelatedWorkCount != 1 {
		t.Fatalf("owner related work = %#v, %v", ownerView.RelatedWork, err)
	}
	viewerView, err := repository.Get(ctx, fixture.workspaceA, fixture.viewerA, document.ID)
	if err != nil || len(viewerView.RelatedWork) != 0 || viewerView.RelatedWorkCount != 0 {
		t.Fatalf("non-team viewer related work = %#v, %v", viewerView.RelatedWork, err)
	}
	if _, err := repository.ListRelatedDocuments(
		ctx, fixture.workspaceA, fixture.viewerA,
		documentdomain.RelationshipStory, fixture.storyA,
	); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("non-team ListRelatedDocuments() error = %v, want ErrNotFound", err)
	}

	crossTarget := input
	crossTarget.EntityID = fixture.storyB
	if _, err := repository.AddRelationship(ctx, crossTarget); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("cross-tenant AddRelationship() error = %v, want ErrNotFound", err)
	}

	// Defend against legacy-corrupt rows where an entity claims workspace A
	// while pointing at a team owned by workspace B. Team membership alone must
	// never make that target visible or linkable.
	insertDocumentTeamMember(t, ctx, postgres.Pool, fixture.teamB, fixture.ownerA)
	misScopedStory := insertDocumentStory(
		t, ctx, postgres.Pool, fixture.workspaceA, fixture.teamB, fixture.ownerA, 2,
	)
	misScopedTarget := input
	misScopedTarget.EntityID = misScopedStory
	if _, err := repository.AddRelationship(ctx, misScopedTarget); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("mis-scoped team AddRelationship() error = %v, want ErrNotFound", err)
	}
	if err := repository.RemoveRelationship(ctx, documentdomain.RelationshipInput{
		WorkspaceID: fixture.workspaceB, UserID: fixture.ownerB, DocumentID: document.ID,
		EntityType: documentdomain.RelationshipStory, EntityID: fixture.storyA,
	}); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("cross-tenant RemoveRelationship() error = %v, want ErrNotFound", err)
	}
	if err := repository.RemoveRelationship(ctx, input); err != nil {
		t.Fatalf("RemoveRelationship() error = %v", err)
	}
}

func TestRepositoryMediaTransactionsPreserveOrphanRules(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	documentA := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Media A")
	documentB := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Media B")
	shared := insertDocumentAttachment(t, ctx, postgres.Pool, fixture.workspaceA, fixture.ownerA, "shared")
	crossTenant := insertDocumentAttachment(t, ctx, postgres.Pool, fixture.workspaceB, fixture.ownerB, "cross")

	for _, documentID := range []uuid.UUID{documentA.ID, documentB.ID} {
		if err := repository.LinkMedia(ctx, documentdomain.MediaInput{
			WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
			DocumentID: documentID, AttachmentID: shared,
		}); err != nil {
			t.Fatalf("LinkMedia(%s) error = %v", documentID, err)
		}
	}
	if err := repository.LinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, AttachmentID: shared,
	}); err != nil {
		t.Fatalf("idempotent LinkMedia() error = %v", err)
	}
	if err := repository.LinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, AttachmentID: crossTenant,
	}); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("cross-tenant LinkMedia() error = %v, want ErrNotFound", err)
	}

	orphaned, err := repository.UnlinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, AttachmentID: shared,
	})
	if err != nil || orphaned {
		t.Fatalf("first UnlinkMedia() = %t, %v; want referenced", orphaned, err)
	}
	orphaned, err = repository.UnlinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentB.ID, AttachmentID: shared,
	})
	if err != nil || !orphaned {
		t.Fatalf("final UnlinkMedia() = %t, %v; want orphan candidate", orphaned, err)
	}
	assertAttachmentExists(t, ctx, postgres.Pool, shared, fixture.workspaceA)

	// Generic story files prevent document unlink/delete from reporting an
	// attachment as orphaned.
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)
	`, fixture.storyA, shared); err != nil {
		t.Fatalf("insert story attachment reference: %v", err)
	}
	if err := repository.LinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, AttachmentID: shared,
	}); err != nil {
		t.Fatalf("relink shared media: %v", err)
	}
	orphanOnly := insertDocumentAttachment(t, ctx, postgres.Pool, fixture.workspaceA, fixture.ownerA, "orphan")
	if err := repository.LinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, AttachmentID: orphanOnly,
	}); err != nil {
		t.Fatalf("link orphan-only media: %v", err)
	}

	newHTML := fmt.Sprintf(`<img src="/documents/%s/media/%s">`, documentA.ID, orphanOnly)
	if _, err := repository.Update(ctx, documentdomain.UpdateInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, ContentHTML: &newHTML,
	}); err != nil {
		t.Fatalf("update stable media URL: %v", err)
	}
	duplicate, err := repository.Duplicate(ctx, fixture.workspaceA, fixture.ownerA, documentA.ID)
	if err != nil {
		t.Fatalf("Duplicate() error = %v", err)
	}
	wantDuplicatePath := fmt.Sprintf("/documents/%s/media/%s", duplicate.ID, orphanOnly)
	if duplicate.Visibility != documentdomain.VisibilityPrivate || !strings.Contains(duplicate.ContentHTML, wantDuplicatePath) {
		t.Fatalf("duplicate = %#v, want rewritten private media path", duplicate)
	}
	if !documentMediaLinked(t, ctx, postgres.Pool, duplicate.ID, orphanOnly) {
		t.Fatal("duplicate did not copy media relation")
	}

	// The duplicate still references orphanOnly, so deleting documentA should
	// not return it. A document-only attachment proves the positive case.
	deleteCandidate := insertDocumentAttachment(t, ctx, postgres.Pool, fixture.workspaceA, fixture.ownerA, "delete")
	if err := repository.LinkMedia(ctx, documentdomain.MediaInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA,
		DocumentID: documentA.ID, AttachmentID: deleteCandidate,
	}); err != nil {
		t.Fatalf("link delete candidate: %v", err)
	}
	candidates, err := repository.Delete(ctx, fixture.workspaceA, fixture.ownerA, documentA.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !slices.Equal(candidates, []uuid.UUID{deleteCandidate}) {
		t.Fatalf("Delete() orphan candidates = %v, want [%s]", candidates, deleteCandidate)
	}
	assertAttachmentExists(t, ctx, postgres.Pool, deleteCandidate, fixture.workspaceA)
	if documentMediaLinked(t, ctx, postgres.Pool, documentA.ID, deleteCandidate) {
		t.Fatal("document delete did not cascade media relation")
	}
}

func TestConcurrentPartialUpdatesComposeWithoutLostFields(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	document := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityRestricted, "Original")
	if _, err := repository.SetAccess(ctx, documentdomain.AccessInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID,
		Visibility: documentdomain.VisibilityRestricted,
		Members:    []documentdomain.Member{{UserID: fixture.editorA, Role: "editor"}},
	}); err != nil {
		t.Fatalf("grant editor access: %v", err)
	}

	start := make(chan struct{})
	result := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	title := "Concurrent title"
	content := "<p>Concurrent content</p>"
	for _, input := range []documentdomain.UpdateInput{
		{WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID, Title: &title},
		{WorkspaceID: fixture.workspaceA, UserID: fixture.editorA, DocumentID: document.ID, ContentHTML: &content},
	} {
		input := input
		go func() {
			ready.Done()
			<-start
			_, err := repository.Update(ctx, input)
			result <- err
		}()
	}
	ready.Wait()
	close(start)
	for range 2 {
		if err := <-result; err != nil {
			t.Fatalf("concurrent Update() error = %v", err)
		}
	}

	stored, err := repository.Get(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
	if err != nil || stored.Title != title || stored.ContentHTML != content {
		t.Fatalf("concurrent document = %#v, %v", stored, err)
	}
}

type documentFixture struct {
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	ownerA     uuid.UUID
	editorA    uuid.UUID
	viewerA    uuid.UUID
	guestA     uuid.UUID
	outsiderA  uuid.UUID
	inactiveA  uuid.UUID
	ownerB     uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	storyA     uuid.UUID
	storyB     uuid.UUID
}

func newDocumentFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) documentFixture {
	t.Helper()
	workspaceA := insertDocumentWorkspace(t, ctx, pool, "a")
	workspaceB := insertDocumentWorkspace(t, ctx, pool, "b")
	ownerA := insertDocumentUser(t, ctx, pool, "owner-a", true)
	editorA := insertDocumentUser(t, ctx, pool, "editor-a", true)
	viewerA := insertDocumentUser(t, ctx, pool, "viewer-a", true)
	guestA := insertDocumentUser(t, ctx, pool, "guest-a", true)
	outsiderA := insertDocumentUser(t, ctx, pool, "outsider-a", true)
	inactiveA := insertDocumentUser(t, ctx, pool, "inactive-a", false)
	ownerB := insertDocumentUser(t, ctx, pool, "owner-b", true)
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceA, ownerA, "admin")
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceA, editorA, "member")
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceA, viewerA, "member")
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceA, guestA, "guest")
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceA, outsiderA, "member")
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceA, inactiveA, "member")
	insertDocumentWorkspaceMember(t, ctx, pool, workspaceB, ownerB, "admin")
	teamA := insertDocumentTeam(t, ctx, pool, workspaceA, "DA")
	teamB := insertDocumentTeam(t, ctx, pool, workspaceB, "DB")
	insertDocumentTeamMember(t, ctx, pool, teamA, ownerA)
	insertDocumentTeamMember(t, ctx, pool, teamA, editorA)
	insertDocumentTeamMember(t, ctx, pool, teamB, ownerB)
	return documentFixture{
		workspaceA: workspaceA, workspaceB: workspaceB, ownerA: ownerA, editorA: editorA,
		viewerA: viewerA, guestA: guestA, outsiderA: outsiderA, inactiveA: inactiveA,
		ownerB: ownerB, teamA: teamA, teamB: teamB,
		storyA: insertDocumentStory(t, ctx, pool, workspaceA, teamA, ownerA, 1),
		storyB: insertDocumentStory(t, ctx, pool, workspaceB, teamB, ownerB, 1),
	}
}

func createDocument(
	t *testing.T,
	ctx context.Context,
	repository *Repository,
	workspaceID, userID uuid.UUID,
	visibility documentdomain.Visibility,
	title string,
) documentdomain.Document {
	t.Helper()
	document, err := repository.Create(ctx, documentdomain.CreateInput{
		WorkspaceID: workspaceID, UserID: userID, Title: title, Visibility: visibility,
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	return document
}

func insertDocumentWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)
	`, id, "Documents "+label, "documents-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertDocumentUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, FALSE)
	`, id, label+"-"+id.String(), fmt.Sprintf("%s-%s@example.com", label, id), "Documents "+label, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertDocumentWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, userID uuid.UUID,
	role string,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS public.user_role))
	`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertDocumentTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, code string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Documents "+code, workspaceID, code+uuid.NewString()[:6]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertDocumentTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func insertDocumentStory(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, teamID, reporterID uuid.UUID,
	sequence int32,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO stories (id, sequence_id, team_id, title, workspace_id, reporter_id)
		VALUES ($1, $2, $3, 'Document relationship', $4, $5)
	`, id, sequence, teamID, workspaceID, reporterID); err != nil {
		t.Fatalf("insert story: %v", err)
	}
	return id
}

func insertDocumentAttachment(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, uploaderID uuid.UUID,
	label string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO attachments (
			attachment_id, filename, blob_name, size, mime_type, uploaded_by, workspace_id,
			scan_status, optimization_status
		)
		VALUES ($1, $2, $3, 128, 'image/png', $4, $5, 'unscanned', 'not_requested')
	`, id, label+".png", label+"-"+id.String()+".png", uploaderID, workspaceID); err != nil {
		t.Fatalf("insert attachment: %v", err)
	}
	return id
}

func assertAttachmentExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, attachmentID, workspaceID uuid.UUID) {
	t.Helper()
	var exists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM attachments WHERE attachment_id = $1 AND workspace_id = $2
		)
	`, attachmentID, workspaceID).Scan(&exists); err != nil || !exists {
		t.Fatalf("attachment exists = %t, %v", exists, err)
	}
}

func documentMediaLinked(t *testing.T, ctx context.Context, pool *pgxpool.Pool, documentID, attachmentID uuid.UUID) bool {
	t.Helper()
	var linked bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM document_attachments WHERE document_id = $1 AND attachment_id = $2
		)
	`, documentID, attachmentID).Scan(&linked); err != nil {
		t.Fatalf("check document media relation: %v", err)
	}
	return linked
}

func membersEqualUnordered(left, right []documentdomain.Member) bool {
	if len(left) != len(right) {
		return false
	}
	rolesByUser := make(map[uuid.UUID]string, len(left))
	for _, member := range left {
		rolesByUser[member.UserID] = member.Role
	}
	for _, member := range right {
		if rolesByUser[member.UserID] != member.Role {
			return false
		}
	}
	return true
}
