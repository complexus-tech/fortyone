//go:build integration

package storiesrepository

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestInteractiveHardDeleteAtomicallyDefersOnlySameWorkspaceOrphanedMedia(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(
		nil,
		postgres.Pool,
		WithAttachmentObjectStorage("aws", "attachments"),
	)
	now := time.Date(2026, time.August, 28, 18, 0, 0, 0, time.UTC)
	scope := mutationScopeForFixture(t, fixture)

	deletedStoryID := createSecondaryMutationStory(t, ctx, repository, fixture, now)
	liveStoryID := createSecondaryMutationStory(t, ctx, repository, fixture, now.Add(time.Minute))
	orphanGenericID := insertRetentionAttachment(t, ctx, postgres, fixture, "interactive-orphan-generic")
	orphanInlineID := insertRetentionAttachment(t, ctx, postgres, fixture, "interactive-orphan-inline")
	sharedGenericID := insertRetentionAttachment(t, ctx, postgres, fixture, "interactive-shared-generic")
	sharedInlineID := insertRetentionAttachment(t, ctx, postgres, fixture, "interactive-shared-inline")
	foreignAttachmentID := insertRetentionAttachment(
		t,
		ctx,
		postgres,
		foreignSecondaryMutationFixture(fixture),
		"interactive-foreign-workspace",
	)
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2), ($1, $3), ($1, $4)",
		deletedStoryID, orphanGenericID, sharedGenericID, foreignAttachmentID,
	)
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_inline_attachments (story_id, attachment_id, created_by) VALUES ($1, $2, $3), ($1, $4, $3)",
		deletedStoryID, orphanInlineID, fixture.actorID, sharedInlineID,
	)
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)",
		liveStoryID, sharedGenericID,
	)
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_inline_attachments (story_id, attachment_id, created_by) VALUES ($1, $2, $3)",
		liveStoryID, sharedInlineID, fixture.actorID,
	)

	command := secondaryLifecycleCommand(
		t,
		scope,
		[]uuid.UUID{deletedStoryID},
		storydomain.SecondaryMutationHardDelete,
		now.Add(2*time.Minute),
	)
	result, err := repository.ApplySecondaryStoryLifecycle(ctx, command)
	if err != nil {
		t.Fatalf("interactive hard delete: %v", err)
	}
	if !result.AttachmentObjectDeletionDeferred {
		t.Fatal("interactive hard delete did not transfer object deletion to the durable outbox")
	}
	wantRetired := []uuid.UUID{orphanGenericID, orphanInlineID}
	slices.SortFunc(wantRetired, func(left, right uuid.UUID) int {
		return slices.Compare(left[:], right[:])
	})
	if !slices.Equal(result.OrphanedAttachmentIDs, wantRetired) {
		t.Fatalf("retired attachments = %v, want %v", result.OrphanedAttachmentIDs, wantRetired)
	}
	assertRetentionRowExists(t, ctx, postgres, "stories", "id", deletedStoryID, false)
	assertRetentionRowExists(t, ctx, postgres, "stories", "id", liveStoryID, true)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", orphanGenericID, false)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", orphanInlineID, false)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", sharedGenericID, true)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", sharedInlineID, true)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", foreignAttachmentID, true)
	assertMutationEventCount(t, ctx, postgres, deletedStoryID, "story.deleted", 1)

	rows, err := postgres.Pool.Query(ctx, `
		SELECT attachment_id, storage_provider, container_name, status
		FROM attachment_object_deletion_outbox
		ORDER BY attachment_id
	`)
	if err != nil {
		t.Fatalf("list interactive attachment deletion outbox: %v", err)
	}
	deferredIDs := make([]uuid.UUID, 0, len(wantRetired))
	for rows.Next() {
		var attachmentID uuid.UUID
		var provider, container, status string
		if err := rows.Scan(&attachmentID, &provider, &container, &status); err != nil {
			rows.Close()
			t.Fatalf("scan interactive attachment deletion outbox: %v", err)
		}
		if provider != "aws" || container != "attachments" || status != "pending" {
			rows.Close()
			t.Fatalf("outbox route/status = %q/%q/%q", provider, container, status)
		}
		deferredIDs = append(deferredIDs, attachmentID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate interactive attachment deletion outbox: %v", err)
	}
	if !slices.Equal(deferredIDs, wantRetired) {
		t.Fatalf("outbox attachments = %v, want %v", deferredIDs, wantRetired)
	}

	t.Run("enqueue failure rolls back story attachment and event deletion together", func(t *testing.T) {
		storyID := createSecondaryMutationStory(t, ctx, repository, fixture, now.Add(3*time.Minute))
		attachmentID := insertRetentionAttachment(t, ctx, postgres, fixture, "interactive-rollback")
		mustMutationExec(t, ctx, postgres.Pool,
			"INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)",
			storyID, attachmentID,
		)
		mustMutationExec(t, ctx, postgres.Pool, `
			INSERT INTO attachment_object_deletion_outbox (
				attachment_id, workspace_id, storage_provider, container_name, blob_name
			) VALUES ($1, $2, 'aws', 'attachments', 'preexisting-object')
		`, attachmentID, fixture.workspaceID)

		result, err := repository.ApplySecondaryStoryLifecycle(
			ctx,
			secondaryLifecycleCommand(
				t,
				scope,
				[]uuid.UUID{storyID},
				storydomain.SecondaryMutationHardDelete,
				now.Add(4*time.Minute),
			),
		)
		if !errors.Is(err, storydomain.ErrMutationConflict) {
			t.Fatalf("conflicting enqueue error = %v, want mutation conflict", err)
		}
		if len(result.StoryIDs) != 0 || len(result.ChangedStoryIDs) != 0 ||
			len(result.OrphanedAttachmentIDs) != 0 || result.AttachmentObjectDeletionDeferred {
			t.Fatalf("failed hard-delete result = %#v, want empty", result)
		}
		assertRetentionRowExists(t, ctx, postgres, "stories", "id", storyID, true)
		assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", attachmentID, true)
		assertRetentionRowExists(t, ctx, postgres, "story_attachments", "story_id", storyID, true)
		assertMutationEventCount(t, ctx, postgres, storyID, "story.deleted", 0)
	})

	t.Run("missing route fails closed before mutation", func(t *testing.T) {
		unconfiguredRepository := NewMutationRepository(nil, postgres.Pool)
		storyID := createSecondaryMutationStory(t, ctx, unconfiguredRepository, fixture, now.Add(5*time.Minute))
		result, err := unconfiguredRepository.ApplySecondaryStoryLifecycle(
			ctx,
			secondaryLifecycleCommand(
				t,
				scope,
				[]uuid.UUID{storyID},
				storydomain.SecondaryMutationHardDelete,
				now.Add(6*time.Minute),
			),
		)
		if !errors.Is(err, errInteractiveHardDeleteStorageRoute) {
			t.Fatalf("missing hard-delete route error = %v, want fail-closed route error", err)
		}
		if len(result.StoryIDs) != 0 || len(result.ChangedStoryIDs) != 0 ||
			len(result.OrphanedAttachmentIDs) != 0 || result.AttachmentObjectDeletionDeferred {
			t.Fatalf("missing-route hard-delete result = %#v, want empty", result)
		}
		assertRetentionRowExists(t, ctx, postgres, "stories", "id", storyID, true)
		assertMutationEventCount(t, ctx, postgres, storyID, "story.deleted", 0)
	})
}
