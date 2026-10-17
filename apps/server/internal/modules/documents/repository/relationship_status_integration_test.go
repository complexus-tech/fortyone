//go:build integration

package documentsrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestDocumentRelationshipReturnsCurrentStoryStatusColor(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	fixture := newDocumentFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	document := createDocument(t, ctx, repository, fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Live story status")
	statusID, completedStatusID := uuid.New(), uuid.New()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO statuses (status_id, name, category, workspace_id, team_id, color)
		VALUES ($1, 'Started', 'started', $2, $3, '#f59e0b'),
		       ($4, 'Done', 'completed', $2, $3, '#3b82f6')
	`, statusID, fixture.workspaceA, fixture.teamA, completedStatusID); err != nil {
		t.Fatalf("insert statuses: %v", err)
	}
	setStatus := func(id *uuid.UUID) {
		t.Helper()
		if _, err := postgres.Pool.Exec(ctx, "UPDATE stories SET status_id = $1 WHERE id = $2", id, fixture.storyA); err != nil {
			t.Fatalf("set story status: %v", err)
		}
	}
	assertColor := func(want *string) {
		t.Helper()
		current, err := repository.Get(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
		if err != nil || len(current.RelatedWork) != 1 {
			t.Fatalf("Get() related work = %#v, %v", current.RelatedWork, err)
		}
		got := current.RelatedWork[0].StatusColor
		if (got == nil) != (want == nil) || (want != nil && *got != *want) {
			t.Fatalf("status color = %v, want %v", got, want)
		}
	}
	setStatus(&statusID)
	input := documentdomain.RelationshipInput{
		WorkspaceID: fixture.workspaceA, UserID: fixture.ownerA, DocumentID: document.ID,
		EntityType: documentdomain.RelationshipStory, EntityID: fixture.storyA,
	}
	related, err := repository.AddRelationship(ctx, input)
	if err != nil || related.StatusColor == nil || *related.StatusColor != "#f59e0b" {
		t.Fatalf("AddRelationship() = %#v, %v; want current status color", related, err)
	}
	assertColor(stringPointer("#f59e0b"))
	if _, err := postgres.Pool.Exec(ctx, "UPDATE statuses SET color = '#22c55e' WHERE status_id = $1", statusID); err != nil {
		t.Fatalf("change status color: %v", err)
	}
	assertColor(stringPointer("#22c55e"))
	setStatus(&completedStatusID)
	assertColor(stringPointer("#3b82f6"))
	setStatus(nil)
	assertColor(nil)
	setStatus(&statusID)

	viewer, err := repository.Get(ctx, fixture.workspaceA, fixture.viewerA, document.ID)
	if err != nil || len(viewer.RelatedWork) != 0 {
		t.Fatalf("non-team viewer related work = %#v, %v; want none", viewer.RelatedWork, err)
	}
	input.UserID = fixture.viewerA
	if _, err := repository.AddRelationship(ctx, input); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("non-team AddRelationship() = %v; want hidden not found", err)
	}
}
