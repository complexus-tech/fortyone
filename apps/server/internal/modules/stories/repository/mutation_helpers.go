package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func insertMutationEvent(ctx context.Context, queries *storyreadsql.Queries, event storydomain.MutationEvent) error {
	credentialID := event.Actor.CredentialID
	var optionalCredentialID *uuid.UUID
	if credentialID != uuid.Nil {
		optionalCredentialID = &credentialID
	}
	if err := queries.InsertStoryMutationEvent(ctx, storyreadsql.InsertStoryMutationEventParams{
		EventID: event.ID, WorkspaceID: event.WorkspaceID, StoryID: event.StoryID,
		EventType: string(event.Type), ActorKind: string(event.Actor.Kind), ActorID: event.Actor.PrincipalID,
		ActorCredentialID: optionalCredentialID, Payload: event.Payload,
		OccurredAt: event.OccurredAt.UTC(), CreatedAt: event.OccurredAt.UTC(),
	}); err != nil {
		return fmt.Errorf("insert story mutation event: %w", err)
	}
	return nil
}

func upsertMutationActivity(
	ctx context.Context,
	queries *storyreadsql.Queries,
	activity storydomain.MutationActivity,
	compact bool,
) error {
	workspaceID := activity.WorkspaceID
	_, err := queries.UpsertStoryMutationActivity(ctx, storyreadsql.UpsertStoryMutationActivityParams{
		CurrentValue: activity.CurrentValue, NewValue: activity.NewValue, Reason: activity.Reason,
		CreatedAt: activity.CreatedAt.UTC(), StoryID: activity.StoryID, UserID: activity.UserID,
		WorkspaceID: &workspaceID, ActivityType: activity.Type, FieldChanged: activity.Field,
		Compact: compact, ActivityID: activity.ID, OldValue: activity.OldValue,
	})
	if err != nil {
		return fmt.Errorf("persist story mutation activity: %w", err)
	}
	return nil
}

func reconcileMutationMedia(
	ctx context.Context,
	queries *storyreadsql.Queries,
	command storydomain.UpdateStoryCommand,
) ([]uuid.UUID, error) {
	linked, err := queries.LockStoryMediaLinks(ctx, storyreadsql.LockStoryMediaLinksParams{
		WorkspaceID: command.Scope.WorkspaceID, StoryID: command.StoryID,
	})
	if err != nil {
		return nil, fmt.Errorf("lock story media links: %w", err)
	}
	linkedByID := make(map[uuid.UUID]string, len(linked))
	for _, media := range linked {
		linkedByID[media.AttachmentID] = media.MimeType
	}
	referenced := make(map[uuid.UUID]struct{}, len(command.ReferencedMedia))
	for _, attachmentID := range command.ReferencedMedia {
		mimeType, exists := linkedByID[attachmentID]
		if attachmentID == uuid.Nil || !exists || !isStoryInlineMediaType(mimeType) {
			return nil, storydomain.ErrInvalidMutation
		}
		referenced[attachmentID] = struct{}{}
	}
	orphaned := make([]uuid.UUID, 0)
	for _, media := range linked {
		if _, retained := referenced[media.AttachmentID]; retained {
			continue
		}
		rows, err := queries.DeleteStoryMediaLink(ctx, storyreadsql.DeleteStoryMediaLinkParams{
			StoryID: command.StoryID, TargetAttachmentID: media.AttachmentID,
		})
		if err != nil {
			return nil, fmt.Errorf("unlink story media: %w", err)
		}
		if rows == 0 {
			continue
		}
		isOrphaned, err := queries.StoryMediaAttachmentIsOrphaned(ctx, storyreadsql.StoryMediaAttachmentIsOrphanedParams{
			TargetAttachmentID: media.AttachmentID,
		})
		if err != nil {
			return nil, fmt.Errorf("check story media references: %w", err)
		}
		if isOrphaned != nil && *isOrphaned {
			orphaned = append(orphaned, media.AttachmentID)
		}
	}
	return orphaned, nil
}

func normalizeMutationLabelIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return nil, fmt.Errorf("%w: label id cannot be empty", storydomain.ErrInvalidMutation)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	slices.SortFunc(result, func(left, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	return result, nil
}

func isCreationKeyConflict(err error) bool {
	var postgresErr *pgconn.PgError
	return errors.As(err, &postgresErr) && postgresErr.Code == "23505" &&
		postgresErr.ConstraintName == "stories_external_creation_key_key"
}

func isStorySequenceUniqueConflict(err error) bool {
	var postgresErr *pgconn.PgError
	return errors.As(err, &postgresErr) && postgresErr.Code == "23505" &&
		postgresErr.ConstraintName == "unique_team_sequence"
}

func mapMutationDatabaseError(err error) error {
	if err == nil || errors.Is(err, storydomain.ErrNotFound) || errors.Is(err, storydomain.ErrMutationForbidden) ||
		errors.Is(err, storydomain.ErrMutationConflict) || errors.Is(err, storydomain.ErrInvalidMutation) {
		return err
	}
	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassForeignKeyViolation, platformdatabase.ErrorClassCheckViolation,
		platformdatabase.ErrorClassNotNullViolation:
		return errors.Join(storydomain.ErrInvalidMutation, err)
	case platformdatabase.ErrorClassUniqueViolation:
		return errors.Join(storydomain.ErrMutationConflict, err)
	default:
		return err
	}
}
