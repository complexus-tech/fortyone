package storiesrepository

import (
	"context"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const maximumStoryMutationEventBatch = 100

func (r *repo) ClaimStoryMutationEvents(
	ctx context.Context,
	batchSize int,
	now, staleBefore time.Time,
) ([]storydomain.MutationEvent, error) {
	if err := r.mutationConfigured(); err != nil {
		return nil, err
	}
	if batchSize < 1 || batchSize > maximumStoryMutationEventBatch || now.IsZero() || staleBefore.IsZero() || !staleBefore.Before(now) {
		return nil, fmt.Errorf("%w: invalid mutation event claim", storydomain.ErrInvalidMutation)
	}
	rows, err := r.reads.ClaimStoryMutationEvents(ctx, storyreadsql.ClaimStoryMutationEventsParams{
		Now: now.UTC(), StaleBefore: staleBefore.UTC(), BatchSize: int32(batchSize),
	})
	if err != nil {
		return nil, fmt.Errorf("claim story mutation events: %w", err)
	}
	events := make([]storydomain.MutationEvent, 0, len(rows))
	for _, row := range rows {
		if row.ClaimToken == nil || *row.ClaimToken == uuid.Nil || row.ClaimedAt == nil {
			return nil, fmt.Errorf("%w: claimed event is missing its lease", storydomain.ErrInvalidMutation)
		}
		kind := platformauth.PrincipalKind(row.ActorKind)
		credentialID := uuid.Nil
		if row.ActorCredentialID != nil {
			credentialID = *row.ActorCredentialID
		}
		actor, actorErr := platformauth.NewActor(
			row.ActorID,
			kind,
			credentialID,
			platformauth.MustScopeSet(),
			platformauth.UnrestrictedTeamAccess(),
		)
		if actorErr != nil {
			return nil, fmt.Errorf("decode story mutation event actor: %w", actorErr)
		}
		actor, actorErr = actor.WithWorkspace(row.WorkspaceID)
		if actorErr != nil {
			return nil, fmt.Errorf("decode story mutation event workspace: %w", actorErr)
		}
		event := storydomain.MutationEvent{
			ID: row.EventID, WorkspaceID: row.WorkspaceID, StoryID: row.StoryID,
			Type: storydomain.MutationEventType(row.EventType), Actor: actor,
			Payload: append([]byte(nil), row.Payload...), OccurredAt: row.OccurredAt.UTC(),
			AttemptCount: int(row.AttemptCount), ClaimToken: *row.ClaimToken,
			ClaimedAt: row.ClaimedAt, CreatedAt: row.CreatedAt.UTC(),
		}
		if err := event.Validate(); err != nil {
			return nil, fmt.Errorf("decode story mutation event: %w", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *repo) CompleteStoryMutationEvent(
	ctx context.Context,
	eventID, claimToken uuid.UUID,
	completedAt time.Time,
) error {
	if err := r.mutationConfigured(); err != nil {
		return err
	}
	if eventID == uuid.Nil || claimToken == uuid.Nil || completedAt.IsZero() {
		return fmt.Errorf("%w: invalid mutation event completion", storydomain.ErrInvalidMutation)
	}
	rows, err := r.reads.CompleteStoryMutationEvent(ctx, storyreadsql.CompleteStoryMutationEventParams{
		CompletedAt: completedAt.UTC(), EventID: eventID, ClaimToken: claimToken,
	})
	if err != nil {
		return fmt.Errorf("complete story mutation event: %w", err)
	}
	if rows != 1 {
		return storydomain.ErrMutationEventNotFound
	}
	return nil
}

func (r *repo) RetryStoryMutationEvent(
	ctx context.Context,
	eventID, claimToken uuid.UUID,
	nextAttemptAt, updatedAt time.Time,
	lastError string,
) error {
	if err := r.mutationConfigured(); err != nil {
		return err
	}
	if eventID == uuid.Nil || claimToken == uuid.Nil || nextAttemptAt.IsZero() || updatedAt.IsZero() || nextAttemptAt.Before(updatedAt) {
		return fmt.Errorf("%w: invalid mutation event retry", storydomain.ErrInvalidMutation)
	}
	rows, err := r.reads.RetryStoryMutationEvent(ctx, storyreadsql.RetryStoryMutationEventParams{
		NextAttemptAt: nextAttemptAt.UTC(), UpdatedAt: updatedAt.UTC(), LastError: lastError,
		EventID: eventID, ClaimToken: claimToken,
	})
	if err != nil {
		return fmt.Errorf("retry story mutation event: %w", err)
	}
	if rows != 1 {
		return storydomain.ErrMutationEventNotFound
	}
	return nil
}
