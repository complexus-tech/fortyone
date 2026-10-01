//go:build integration

package storiesrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
)

func TestImportReceiptsRetainRetriesMetadataAndBoundedTenantPages(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	mustMutationExec(t, ctx, postgres.Pool, "UPDATE workspace_members SET role = 'admin' WHERE workspace_id = $1 AND user_id = $2", fixture.workspaceID, fixture.actorID)
	repository := NewMutationRepository(nil, postgres.Pool)
	storyID := createSecondaryMutationStory(t, ctx, repository, fixture, time.Now().UTC())
	actor := mutationScopeForFixture(t, fixture).Actor
	namespace := "trello:board:receipt-test"
	metadata := json.RawMessage(`{"createdAt":"2025-01-01T12:00:00Z","estimateValue":13,"estimatedDurationMinutes":6000}`)
	receipt := storydomain.ImportReceipt{WorkspaceID: fixture.workspaceID, TeamID: fixture.teamID, CreationKey: "receipt-0000", Provider: "file", SourceDigest: strings.Repeat("a", 64), SourceNamespace: &namespace, SourceKey: "source-0", StoryID: &storyID, Created: true, SourceMetadata: metadata}
	if err := repository.RecordImportReceipt(ctx, receipt); err != nil {
		t.Fatal(err)
	}
	receipt.Created = false
	if err := repository.RecordImportReceipt(ctx, receipt); err != nil {
		t.Fatal(err)
	}
	for index := 1; index <= 500; index++ {
		receipt.CreationKey = fmt.Sprintf("receipt-%04d", index)
		receipt.SourceKey = fmt.Sprintf("source-%d", index)
		if err := repository.RecordImportReceipt(ctx, receipt); err != nil {
			t.Fatal(err)
		}
	}
	page, err := repository.ListImportReceipts(ctx, actor, "file", receipt.SourceDigest, &namespace, 500, 0)
	if err != nil || len(page) != 500 {
		t.Fatalf("first page = %d, %v", len(page), err)
	}
	if !page[0].Created || page[0].StoryID == nil || *page[0].StoryID != storyID {
		t.Fatalf("retry lost original creation: %#v", page[0])
	}
	var original, restored map[string]any
	if err := json.Unmarshal(metadata, &original); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(page[0].SourceMetadata, &restored); err != nil {
		t.Fatal(err)
	}
	if restored["estimateValue"] != original["estimateValue"] || restored["estimatedDurationMinutes"] != original["estimatedDurationMinutes"] || restored["createdAt"] != original["createdAt"] {
		t.Fatalf("source metadata changed: %s", page[0].SourceMetadata)
	}
	tail, err := repository.ListImportReceipts(ctx, actor, "file", receipt.SourceDigest, &namespace, 500, 500)
	if err != nil || len(tail) != 1 {
		t.Fatalf("continuation page = %d, %v", len(tail), err)
	}
	foreignActor, err := actor.WithWorkspace(fixture.foreignWorkspaceID)
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := repository.ListImportReceipts(ctx, foreignActor, "file", receipt.SourceDigest, &namespace, 500, 0)
	if err != nil || len(foreign) != 0 {
		t.Fatalf("cross-workspace page = %d, %v", len(foreign), err)
	}
	mustMutationExec(t, ctx, postgres.Pool, "UPDATE teams SET is_private = TRUE WHERE team_id = $1", fixture.teamID)
	mustMutationExec(t, ctx, postgres.Pool, "DELETE FROM team_members WHERE team_id = $1 AND user_id = $2", fixture.teamID, fixture.actorID)
	private, err := repository.ListImportReceipts(ctx, actor, "file", receipt.SourceDigest, &namespace, 500, 0)
	if err != nil || len(private) != 0 {
		t.Fatalf("inaccessible private team page = %d, %v", len(private), err)
	}
}
