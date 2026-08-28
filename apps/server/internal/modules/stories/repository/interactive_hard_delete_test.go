package storiesrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestApplyInteractiveHardDeleteEnqueuesOnlyRetiredAttachments(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 15, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	workspaceID, storyID := uuid.New(), uuid.New()
	sharedAttachmentID, retiredAttachmentID := uuid.New(), uuid.New()
	queries := &interactiveHardDeleteQueryStub{
		attachmentIDs:   []uuid.UUID{sharedAttachmentID, retiredAttachmentID},
		deletedStoryIDs: []uuid.UUID{storyID},
		retiredRows: []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{{
			AttachmentID: retiredAttachmentID,
			WorkspaceID:  workspaceID,
			BlobName:     "opaque-retired-object",
		}},
		insertRows: 1,
	}

	result, err := applyInteractiveHardDelete(
		context.Background(),
		queries,
		attachmentObjectStorageRoute{provider: "aws", container: "attachments"},
		workspaceID,
		[]uuid.UUID{storyID},
		now,
	)

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{storyID}, result.storyIDs)
	require.Equal(t, []uuid.UUID{retiredAttachmentID}, result.retiredAttachmentIDs)
	require.Equal(t, []string{"list_attachments", "delete_stories", "delete_attachments", "insert_outbox"}, queries.calls)
	require.Equal(t, int32(maximumInteractiveHardDeleteAttachmentCount+1), queries.listParams.MaximumAttachmentCount)
	require.Equal(t, workspaceID, queries.listParams.WorkspaceID)
	require.Equal(t, []uuid.UUID{storyID}, queries.listParams.StoryIds)
	require.Equal(t, []uuid.UUID{sharedAttachmentID, retiredAttachmentID}, queries.deleteAttachmentParams.AttachmentIds)
	require.Equal(t, []storyreadsql.InsertAttachmentObjectDeletionOutboxParams{{
		AttachmentID:    retiredAttachmentID,
		WorkspaceID:     workspaceID,
		StorageProvider: "aws",
		ContainerName:   "attachments",
		BlobName:        "opaque-retired-object",
		EnqueuedAt:      now.UTC(),
	}}, queries.insertParams)
}

func TestApplyInteractiveHardDeleteReturnsNoResultWhenOutboxEnqueueFails(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("outbox unavailable")
	workspaceID, storyID, attachmentID := uuid.New(), uuid.New(), uuid.New()
	queries := &interactiveHardDeleteQueryStub{
		attachmentIDs:   []uuid.UUID{attachmentID},
		deletedStoryIDs: []uuid.UUID{storyID},
		retiredRows: []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{{
			AttachmentID: attachmentID,
			WorkspaceID:  workspaceID,
			BlobName:     "opaque-object",
		}},
		insertErr: wantErr,
	}

	result, err := applyInteractiveHardDelete(
		context.Background(),
		queries,
		attachmentObjectStorageRoute{provider: "azure", container: "attachments"},
		workspaceID,
		[]uuid.UUID{storyID},
		time.Now(),
	)

	require.ErrorIs(t, err, wantErr)
	require.Zero(t, result, "a failed transaction operation must not expose a successful deletion receipt")
	require.Equal(t, []string{"list_attachments", "delete_stories", "delete_attachments", "insert_outbox"}, queries.calls)
}

func TestApplyInteractiveHardDeleteRejectsLookaheadBeforeMutation(t *testing.T) {
	t.Parallel()

	queries := &interactiveHardDeleteQueryStub{
		attachmentIDs: make([]uuid.UUID, maximumInteractiveHardDeleteAttachmentCount+1),
	}
	for index := range queries.attachmentIDs {
		queries.attachmentIDs[index] = uuid.New()
	}

	result, err := applyInteractiveHardDelete(
		context.Background(),
		queries,
		attachmentObjectStorageRoute{provider: "aws", container: "attachments"},
		uuid.New(),
		[]uuid.UUID{uuid.New()},
		time.Now(),
	)

	require.ErrorIs(t, err, storydomain.ErrInvalidMutation)
	require.Zero(t, result)
	require.Equal(t, []string{"list_attachments"}, queries.calls)
}

func TestInteractiveHardDeleteStorageRouteFailsClosedAndStoresValuesOnly(t *testing.T) {
	t.Parallel()

	_, err := (&repo{}).interactiveHardDeleteStorageRoute()
	require.ErrorIs(t, err, errInteractiveHardDeleteStorageRoute)

	repository := &repo{}
	WithAttachmentObjectStorage(" aws ", " attachments ")(repository)
	route, err := repository.interactiveHardDeleteStorageRoute()
	require.NoError(t, err)
	require.Equal(t, attachmentObjectStorageRoute{provider: "aws", container: "attachments"}, route)

	WithAttachmentObjectStorage("", "attachments")(repository)
	_, err = repository.interactiveHardDeleteStorageRoute()
	require.ErrorIs(t, err, errInteractiveHardDeleteStorageRoute)
}

func TestApplySecondaryHardDeleteRejectsMissingOrInvalidRouteBeforeStartingTransaction(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, storyID := uuid.New(), uuid.New(), uuid.New()
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	require.NoError(t, err)
	changedAt := time.Date(2026, time.August, 28, 16, 0, 0, 0, time.UTC)
	command := storydomain.SecondaryLifecycleCommand{
		Scope: storydomain.MutationScope{
			Actor: actor, WorkspaceID: workspaceID, ActivityUser: &actorID,
		},
		Action:    storydomain.SecondaryMutationHardDelete,
		StoryIDs:  []uuid.UUID{storyID},
		ChangedAt: changedAt,
		Events: []storydomain.MutationEvent{{
			ID: uuid.New(), WorkspaceID: workspaceID, StoryID: storyID,
			Type: storydomain.MutationEventStoryDeleted, Actor: actor,
			Payload: []byte("{}"), OccurredAt: changedAt,
		}},
	}
	require.NoError(t, command.Validate())

	tests := []struct {
		name    string
		options []Option
	}{
		{name: "missing"},
		{name: "blank provider", options: []Option{WithAttachmentObjectStorage("", "attachments")}},
		{name: "blank container", options: []Option{WithAttachmentObjectStorage("aws", "")}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := NewMutationRepository(nil, &pgxpool.Pool{}, test.options...)
			result, err := repository.ApplySecondaryStoryLifecycle(context.Background(), command)
			require.ErrorIs(t, err, errInteractiveHardDeleteStorageRoute)
			require.Empty(t, result.StoryIDs)
			require.Empty(t, result.ChangedStoryIDs)
			require.Empty(t, result.OrphanedAttachmentIDs)
			require.False(t, result.AttachmentObjectDeletionDeferred)
		})
	}
}

type interactiveHardDeleteQueryStub struct {
	listParams             storyreadsql.ListInteractiveHardDeleteAttachmentCandidatesParams
	attachmentIDs          []uuid.UUID
	attachmentErr          error
	deletedStoryIDs        []uuid.UUID
	deleteStoryErr         error
	deleteAttachmentParams storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams
	retiredRows            []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow
	deleteAttachmentErr    error
	insertParams           []storyreadsql.InsertAttachmentObjectDeletionOutboxParams
	insertRows             int64
	insertErr              error
	calls                  []string
}

func (queries *interactiveHardDeleteQueryStub) ListInteractiveHardDeleteAttachmentCandidates(
	_ context.Context,
	params storyreadsql.ListInteractiveHardDeleteAttachmentCandidatesParams,
) ([]uuid.UUID, error) {
	queries.calls = append(queries.calls, "list_attachments")
	queries.listParams = params
	return queries.attachmentIDs, queries.attachmentErr
}

func (queries *interactiveHardDeleteQueryStub) HardDeleteSecondaryStories(
	_ context.Context,
	_ storyreadsql.HardDeleteSecondaryStoriesParams,
) ([]uuid.UUID, error) {
	queries.calls = append(queries.calls, "delete_stories")
	return queries.deletedStoryIDs, queries.deleteStoryErr
}

func (queries *interactiveHardDeleteQueryStub) DeleteUnreferencedStoryRetentionAttachments(
	_ context.Context,
	params storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams,
) ([]storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow, error) {
	queries.calls = append(queries.calls, "delete_attachments")
	queries.deleteAttachmentParams = params
	return queries.retiredRows, queries.deleteAttachmentErr
}

func (queries *interactiveHardDeleteQueryStub) InsertAttachmentObjectDeletionOutbox(
	_ context.Context,
	params storyreadsql.InsertAttachmentObjectDeletionOutboxParams,
) (int64, error) {
	queries.calls = append(queries.calls, "insert_outbox")
	queries.insertParams = append(queries.insertParams, params)
	return queries.insertRows, queries.insertErr
}
