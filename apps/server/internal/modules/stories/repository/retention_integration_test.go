//go:build integration

package storiesrepository

import (
	"context"
	"slices"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestStoryRetentionAtomicallySeparatesSharedAndOrphanedMedia(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)

	now := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	deletedBefore := now.Add(-30 * 24 * time.Hour)
	expiredStoryID := createSecondaryMutationStory(t, ctx, repository, fixture, now.Add(-40*24*time.Hour))
	liveStoryID := createSecondaryMutationStory(t, ctx, repository, fixture, now.Add(-24*time.Hour))
	mustMutationExec(
		t, ctx, postgres.Pool,
		"UPDATE stories SET deleted_at = $2 WHERE id = $1",
		expiredStoryID, deletedBefore.Add(-time.Hour),
	)

	orphanGenericID := insertRetentionAttachment(t, ctx, postgres, fixture, "orphan-generic.png")
	orphanInlineID := insertRetentionAttachment(t, ctx, postgres, fixture, "orphan-inline.png")
	sharedGenericID := insertRetentionAttachment(t, ctx, postgres, fixture, "shared-generic.png")
	sharedInlineID := insertRetentionAttachment(t, ctx, postgres, fixture, "shared-inline.png")
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)",
		expiredStoryID, orphanGenericID,
	)
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_inline_attachments (story_id, attachment_id, created_by) VALUES ($1, $2, $3)",
		expiredStoryID, orphanInlineID, fixture.actorID,
	)
	for _, storyID := range []uuid.UUID{expiredStoryID, liveStoryID} {
		mustMutationExec(t, ctx, postgres.Pool,
			"INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)",
			storyID, sharedGenericID,
		)
		mustMutationExec(t, ctx, postgres.Pool,
			"INSERT INTO story_inline_attachments (story_id, attachment_id, created_by) VALUES ($1, $2, $3)",
			storyID, sharedInlineID, fixture.actorID,
		)
	}

	result, err := repository.PurgeDeletedStoriesBatch(ctx, storydomain.StoryRetentionBatch{
		DeletedBefore: deletedBefore, EnqueuedAt: now,
		BatchSize: 10, MaximumAttachmentCount: 20,
		StorageProvider: "aws", ContainerName: "attachments",
	})
	if err != nil {
		t.Fatalf("purge expired story retention batch: %v", err)
	}
	if result.CandidateCount != 1 || result.DeletedStoryCount != 1 || result.EnqueuedObjectCount != 2 {
		t.Fatalf("retention result = %#v, want one story and two orphan objects", result)
	}
	assertRetentionRowExists(t, ctx, postgres, "stories", "id", expiredStoryID, false)
	assertRetentionRowExists(t, ctx, postgres, "stories", "id", liveStoryID, true)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", orphanGenericID, false)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", orphanInlineID, false)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", sharedGenericID, true)
	assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", sharedInlineID, true)

	rows, err := postgres.Pool.Query(ctx,
		"SELECT attachment_id FROM attachment_object_deletion_outbox ORDER BY attachment_id",
	)
	if err != nil {
		t.Fatalf("list attachment object deletion outbox: %v", err)
	}
	deferredIDs := make([]uuid.UUID, 0, 2)
	for rows.Next() {
		var attachmentID uuid.UUID
		if err := rows.Scan(&attachmentID); err != nil {
			rows.Close()
			t.Fatalf("scan attachment object deletion outbox: %v", err)
		}
		deferredIDs = append(deferredIDs, attachmentID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate attachment object deletion outbox: %v", err)
	}
	wantDeferredIDs := []uuid.UUID{orphanGenericID, orphanInlineID}
	slices.SortFunc(wantDeferredIDs, func(left, right uuid.UUID) int {
		return slices.Compare(left[:], right[:])
	})
	if !slices.Equal(deferredIDs, wantDeferredIDs) {
		t.Fatalf("deferred attachment IDs = %v, want %v", deferredIDs, wantDeferredIDs)
	}

	t.Run("outbox conflict rolls back story and attachment deletion without advancing a cursor", func(t *testing.T) {
		rollbackStoryID := createSecondaryMutationStory(t, ctx, repository, fixture, now.Add(-39*24*time.Hour))
		mustMutationExec(
			t, ctx, postgres.Pool,
			"UPDATE stories SET deleted_at = $2 WHERE id = $1",
			rollbackStoryID, deletedBefore.Add(-30*time.Minute),
		)
		rollbackAttachmentID := insertRetentionAttachment(t, ctx, postgres, fixture, "rollback.png")
		mustMutationExec(t, ctx, postgres.Pool,
			"INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)",
			rollbackStoryID, rollbackAttachmentID,
		)
		mustMutationExec(t, ctx, postgres.Pool, `
			INSERT INTO attachment_object_deletion_outbox (
				attachment_id, workspace_id, storage_provider, container_name, blob_name
			) VALUES ($1, $2, 'aws', 'attachments', 'preexisting.png')
		`, rollbackAttachmentID, fixture.workspaceID)

		result, err := repository.PurgeDeletedStoriesBatch(ctx, storydomain.StoryRetentionBatch{
			DeletedBefore: deletedBefore, EnqueuedAt: now,
			BatchSize: 10, MaximumAttachmentCount: 20,
			StorageProvider: "aws", ContainerName: "attachments",
		})
		if err == nil {
			t.Fatal("purge with a conflicting outbox row unexpectedly succeeded")
		}
		if result != (storydomain.StoryRetentionResult{}) {
			t.Fatalf("failed retention result = %#v, want no cursor or committed counts", result)
		}
		assertRetentionRowExists(t, ctx, postgres, "stories", "id", rollbackStoryID, true)
		assertRetentionRowExists(t, ctx, postgres, "attachments", "attachment_id", rollbackAttachmentID, true)
		assertRetentionRowExists(t, ctx, postgres, "story_attachments", "story_id", rollbackStoryID, true)
	})
}

func insertRetentionAttachment(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	fixture storyMutationFixture,
	blobName string,
) uuid.UUID {
	t.Helper()
	attachmentID := uuid.New()
	mustMutationExec(t, ctx, postgres.Pool, `
		INSERT INTO attachments (
			attachment_id, filename, blob_name, size, mime_type, uploaded_by, workspace_id
		) VALUES ($1, $2, $2, 128, 'image/png', $3, $4)
	`, attachmentID, blobName, fixture.actorID, fixture.workspaceID)
	return attachmentID
}

func assertRetentionRowExists(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	table string,
	column string,
	id uuid.UUID,
	want bool,
) {
	t.Helper()
	statement := "SELECT EXISTS (SELECT 1 FROM " +
		pgx.Identifier{table}.Sanitize() + " WHERE " + pgx.Identifier{column}.Sanitize() + " = $1)"
	var exists bool
	if err := postgres.Pool.QueryRow(ctx, statement, id).Scan(&exists); err != nil {
		t.Fatalf("check %s.%s retention row: %v", table, column, err)
	}
	if exists != want {
		t.Fatalf("%s.%s row exists = %t, want %t", table, column, exists, want)
	}
}
