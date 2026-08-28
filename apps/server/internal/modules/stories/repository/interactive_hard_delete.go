package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

const maximumInteractiveHardDeleteAttachmentCount = 25_000

var errInteractiveHardDeleteStorageRoute = errors.New("interactive hard delete storage route is not configured")

// Option configures the PostgreSQL story adapter without exposing
// infrastructure credentials to the story domain.
type Option func(*repo)

type attachmentObjectStorageRoute struct {
	provider  string
	container string
}

// WithAttachmentObjectStorage records only the credential-free routing values
// that an atomic hard delete must copy into the durable deletion outbox.
func WithAttachmentObjectStorage(provider, container string) Option {
	return func(repository *repo) {
		if repository == nil {
			return
		}
		repository.attachmentObjectStorage = &attachmentObjectStorageRoute{
			provider:  strings.TrimSpace(provider),
			container: strings.TrimSpace(container),
		}
	}
}

type interactiveHardDeleteQueries interface {
	ListInteractiveHardDeleteAttachmentCandidates(
		context.Context,
		storyreadsql.ListInteractiveHardDeleteAttachmentCandidatesParams,
	) ([]uuid.UUID, error)
	HardDeleteSecondaryStories(
		context.Context,
		storyreadsql.HardDeleteSecondaryStoriesParams,
	) ([]uuid.UUID, error)
	DeleteUnreferencedStoryRetentionAttachments(
		context.Context,
		storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams,
	) ([]storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow, error)
	InsertAttachmentObjectDeletionOutbox(
		context.Context,
		storyreadsql.InsertAttachmentObjectDeletionOutboxParams,
	) (int64, error)
}

type interactiveHardDeleteResult struct {
	storyIDs             []uuid.UUID
	retiredAttachmentIDs []uuid.UUID
}

func (r *repo) interactiveHardDeleteStorageRoute() (attachmentObjectStorageRoute, error) {
	if r == nil || r.attachmentObjectStorage == nil {
		return attachmentObjectStorageRoute{}, errInteractiveHardDeleteStorageRoute
	}
	return validateInteractiveHardDeleteStorageRoute(*r.attachmentObjectStorage)
}

func validateInteractiveHardDeleteStorageRoute(
	route attachmentObjectStorageRoute,
) (attachmentObjectStorageRoute, error) {
	normalized := attachmentObjectStorageRoute{
		provider:  strings.TrimSpace(route.provider),
		container: strings.TrimSpace(route.container),
	}
	if normalized.provider == "" || len(normalized.provider) > maximumStorageProviderLength ||
		normalized.container == "" || len(normalized.container) > maximumContainerNameLength {
		return attachmentObjectStorageRoute{}, errInteractiveHardDeleteStorageRoute
	}
	return normalized, nil
}

// applyInteractiveHardDelete atomically removes locked stories, retires only
// attachment metadata that is no longer referenced anywhere, and schedules the
// corresponding object deletions. Its caller owns the surrounding transaction.
func applyInteractiveHardDelete(
	ctx context.Context,
	queries interactiveHardDeleteQueries,
	route attachmentObjectStorageRoute,
	workspaceID uuid.UUID,
	storyIDs []uuid.UUID,
	enqueuedAt time.Time,
) (interactiveHardDeleteResult, error) {
	route, err := validateInteractiveHardDeleteStorageRoute(route)
	if err != nil {
		return interactiveHardDeleteResult{}, err
	}
	lookahead := int32(maximumInteractiveHardDeleteAttachmentCount + 1)
	attachmentIDs, err := queries.ListInteractiveHardDeleteAttachmentCandidates(
		ctx,
		storyreadsql.ListInteractiveHardDeleteAttachmentCandidatesParams{
			MaximumAttachmentCount: lookahead,
			WorkspaceID:            workspaceID,
			StoryIds:               storyIDs,
		},
	)
	if err != nil {
		return interactiveHardDeleteResult{}, fmt.Errorf("list interactive hard-delete attachment candidates: %w", err)
	}
	if len(attachmentIDs) > maximumInteractiveHardDeleteAttachmentCount {
		return interactiveHardDeleteResult{}, fmt.Errorf(
			"%w: hard delete exceeds the supported attachment count",
			storydomain.ErrInvalidMutation,
		)
	}
	attachmentSet := make(map[uuid.UUID]struct{}, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		if attachmentID == uuid.Nil {
			return interactiveHardDeleteResult{}, errors.New("list interactive hard-delete attachment candidates: database returned an invalid attachment")
		}
		if _, duplicate := attachmentSet[attachmentID]; duplicate {
			return interactiveHardDeleteResult{}, errors.New("list interactive hard-delete attachment candidates: database returned a duplicate attachment")
		}
		attachmentSet[attachmentID] = struct{}{}
	}

	deletedStoryIDs, err := queries.HardDeleteSecondaryStories(
		ctx,
		storyreadsql.HardDeleteSecondaryStoriesParams{
			StoryIds:    storyIDs,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		return interactiveHardDeleteResult{}, fmt.Errorf("hard delete stories: %w", err)
	}
	if len(deletedStoryIDs) != len(storyIDs) {
		return interactiveHardDeleteResult{}, fmt.Errorf(
			"%w: hard delete removed %d of %d locked stories",
			storydomain.ErrMutationConflict,
			len(deletedStoryIDs),
			len(storyIDs),
		)
	}
	expectedStories := make(map[uuid.UUID]struct{}, len(storyIDs))
	for _, storyID := range storyIDs {
		expectedStories[storyID] = struct{}{}
	}
	seenDeletedStories := make(map[uuid.UUID]struct{}, len(deletedStoryIDs))
	for _, storyID := range deletedStoryIDs {
		_, expected := expectedStories[storyID]
		_, duplicate := seenDeletedStories[storyID]
		if storyID == uuid.Nil || !expected || duplicate {
			return interactiveHardDeleteResult{}, fmt.Errorf(
				"%w: hard delete returned an invalid story receipt",
				storydomain.ErrMutationConflict,
			)
		}
		seenDeletedStories[storyID] = struct{}{}
	}

	var retired []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow
	if len(attachmentIDs) > 0 {
		retired, err = queries.DeleteUnreferencedStoryRetentionAttachments(
			ctx,
			storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams{AttachmentIds: attachmentIDs},
		)
		if err != nil {
			return interactiveHardDeleteResult{}, fmt.Errorf("delete unreferenced hard-delete attachments: %w", err)
		}
	}
	slices.SortFunc(retired, func(left, right storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow) int {
		return slices.Compare(left.AttachmentID[:], right.AttachmentID[:])
	})

	retiredAttachmentIDs := make([]uuid.UUID, 0, len(retired))
	seenRetired := make(map[uuid.UUID]struct{}, len(retired))
	enqueuedAt = enqueuedAt.UTC()
	for _, attachment := range retired {
		_, wasCandidate := attachmentSet[attachment.AttachmentID]
		_, duplicate := seenRetired[attachment.AttachmentID]
		if attachment.AttachmentID == uuid.Nil || attachment.WorkspaceID != workspaceID ||
			!wasCandidate || duplicate || strings.TrimSpace(attachment.BlobName) == "" ||
			len(attachment.BlobName) > maximumBlobNameLength {
			return interactiveHardDeleteResult{}, errors.New("delete unreferenced hard-delete attachments: database returned invalid routing metadata")
		}
		seenRetired[attachment.AttachmentID] = struct{}{}
		inserted, insertErr := queries.InsertAttachmentObjectDeletionOutbox(
			ctx,
			storyreadsql.InsertAttachmentObjectDeletionOutboxParams{
				AttachmentID:    attachment.AttachmentID,
				WorkspaceID:     attachment.WorkspaceID,
				StorageProvider: route.provider,
				ContainerName:   route.container,
				BlobName:        attachment.BlobName,
				EnqueuedAt:      enqueuedAt,
			},
		)
		if insertErr != nil {
			return interactiveHardDeleteResult{}, fmt.Errorf("enqueue hard-delete attachment object: %w", insertErr)
		}
		if inserted != 1 {
			return interactiveHardDeleteResult{}, fmt.Errorf("enqueue hard-delete attachment object: inserted %d rows, want 1", inserted)
		}
		retiredAttachmentIDs = append(retiredAttachmentIDs, attachment.AttachmentID)
	}

	return interactiveHardDeleteResult{
		storyIDs:             deletedStoryIDs,
		retiredAttachmentIDs: retiredAttachmentIDs,
	}, nil
}
