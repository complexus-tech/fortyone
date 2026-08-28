package stories

import (
	"context"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"

	"github.com/complexus-tech/projects-api/pkg/events"
)

type secondaryStoryMutationRepository interface {
	ApplySecondaryStoryLifecycle(context.Context, storydomain.SecondaryLifecycleCommand) (storydomain.SecondaryLifecycleResult, error)
	ReplaceStoryLabels(context.Context, storydomain.ReplaceStoryLabelsCommand) (storydomain.ReplacementResult, error)
	ReplaceStoryCollaborators(context.Context, storydomain.ReplaceStoryCollaboratorsCommand) (storydomain.ReplacementResult, error)
	SetStoryWatching(context.Context, storydomain.MutationScope, uuid.UUID, uuid.UUID, bool) error
	ListStoryNotificationAudience(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
}

type legacyStoryLabelsRepository interface {
	UpdateLabels(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) error
}

type legacyStoryCollaborationRepository interface {
	GetCollaborators(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
	SetCollaborators(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) error
	SetWatching(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, bool) error
	GetNotificationAudience(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
}

type legacyBulkDeleteRepository interface {
	BulkDelete(context.Context, []uuid.UUID, uuid.UUID, BulkDeleteAuthorization) ([]uuid.UUID, error)
}

type legacyHardBulkDeleteRepository interface {
	HardBulkDelete(context.Context, []uuid.UUID, uuid.UUID, BulkDeleteAuthorization) (HardBulkDeleteResult, error)
}

type legacyRestoreRepository interface {
	Restore(context.Context, uuid.UUID, uuid.UUID) error
}

type legacyBulkRestoreRepository interface {
	BulkRestore(context.Context, []uuid.UUID, uuid.UUID) error
}

type legacyBulkArchiveRepository interface {
	BulkArchive(context.Context, []uuid.UUID, uuid.UUID) error
}

type legacyBulkUnarchiveRepository interface {
	BulkUnarchive(context.Context, []uuid.UUID, uuid.UUID) error
}

type legacyStoryDeleteRepository interface {
	Delete(context.Context, uuid.UUID, uuid.UUID, BulkDeleteAuthorization) error
}

func (s *Service) secondaryMutationRepository() (secondaryStoryMutationRepository, bool) {
	repository, ok := s.repo.(secondaryStoryMutationRepository)
	return repository, ok
}

func buildSecondaryLifecycleCommand(
	scope storydomain.MutationScope,
	storyIDs []uuid.UUID,
	action storydomain.SecondaryMutationAction,
	changedAt time.Time,
) (storydomain.SecondaryLifecycleCommand, error) {
	storyIDs, err := storydomain.NormalizeSecondaryMutationIDs(storyIDs)
	if err != nil {
		return storydomain.SecondaryLifecycleCommand{}, err
	}
	events := make([]storydomain.MutationEvent, 0, len(storyIDs))
	for _, storyID := range storyIDs {
		payload := any(storyUpdatedIntegrationPayload{
			StoryID: storyID, WorkspaceID: scope.WorkspaceID,
			Changes: secondaryLifecycleChanges(action, changedAt.UTC()),
		})
		if action.EventType() == storydomain.MutationEventStoryDeleted {
			payload = storyDeletedIntegrationPayload{StoryID: storyID, WorkspaceID: scope.WorkspaceID}
		}
		event, err := newStoryMutationEvent(scope, storyID, action.EventType(), payload, changedAt)
		if err != nil {
			return storydomain.SecondaryLifecycleCommand{}, err
		}
		events = append(events, event)
	}
	command := storydomain.SecondaryLifecycleCommand{
		Scope: scope, Action: action, StoryIDs: storyIDs, ChangedAt: changedAt.UTC(), Events: events,
	}
	if err := command.Validate(); err != nil {
		return storydomain.SecondaryLifecycleCommand{}, err
	}
	return command, nil
}

func secondaryLifecycleChanges(action storydomain.SecondaryMutationAction, changedAt time.Time) map[string]any {
	switch action {
	case storydomain.SecondaryMutationRestore:
		return map[string]any{"deleted_at": nil}
	case storydomain.SecondaryMutationArchive:
		return map[string]any{"archived_at": changedAt}
	case storydomain.SecondaryMutationUnarchive:
		return map[string]any{"archived_at": nil}
	default:
		return map[string]any{}
	}
}

func (s *Service) applySecondaryLifecycle(
	ctx context.Context,
	storyIDs []uuid.UUID,
	workspaceID, suppliedActorID uuid.UUID,
	action storydomain.SecondaryMutationAction,
) (storydomain.SecondaryLifecycleResult, error) {
	repository, ok := s.secondaryMutationRepository()
	if !ok {
		return storydomain.SecondaryLifecycleResult{}, fmt.Errorf("story repository does not support secondary mutations")
	}
	scope, err := mutationScope(ctx, workspaceID, suppliedActorID, platformauth.PrincipalHumanUser)
	if err != nil {
		return storydomain.SecondaryLifecycleResult{}, err
	}
	command, err := buildSecondaryLifecycleCommand(scope, storyIDs, action, time.Now().UTC())
	if err != nil {
		return storydomain.SecondaryLifecycleResult{}, err
	}
	result, err := repository.ApplySecondaryStoryLifecycle(ctx, command)
	if err != nil {
		return storydomain.SecondaryLifecycleResult{}, mapStoryMutationError(err)
	}
	return result, nil
}

func (s *Service) updateCollaboratorsTyped(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	collaboratorIDs []uuid.UUID,
) error {
	repository, _ := s.secondaryMutationRepository()
	scope, err := mutationScope(ctx, workspaceID, uuid.Nil, platformauth.PrincipalHumanUser)
	if err != nil {
		return err
	}
	next, err := storydomain.NormalizeSecondaryReplacementIDs(collaboratorIDs)
	if err != nil {
		return mapStoryMutationError(err)
	}
	mutationTime := time.Now().UTC()
	var activity *storydomain.MutationActivity
	if scope.ActivityUser != nil {
		activity, err = newStoryMutationActivity(
			scope, storyID, "update", "collaborator_ids", s.formatValue(next), nil, next, nil, mutationTime,
		)
		if err != nil {
			return err
		}
	}
	event, err := newStoryMutationEvent(
		scope,
		storyID,
		storydomain.MutationEventStoryUpdated,
		storyUpdatedIntegrationPayload{
			StoryID: storyID, WorkspaceID: workspaceID,
			Changes: map[string]any{"collaborator_ids": "changed"},
		},
		mutationTime,
	)
	if err != nil {
		return err
	}
	result, err := repository.ReplaceStoryCollaborators(ctx, storydomain.ReplaceStoryCollaboratorsCommand{
		Scope: scope, StoryID: storyID, CollaboratorIDs: next, Event: event, Activity: activity,
	})
	if err != nil {
		return mapStoryMutationError(err)
	}
	if !result.Changed {
		return nil
	}

	audienceIDs, audienceErr := repository.ListStoryNotificationAudience(ctx, storyID, workspaceID)
	audienceResolved := audienceErr == nil
	if audienceErr != nil && s.log != nil {
		s.log.Error(ctx, "failed to load story notification audience", "error", audienceErr, "story_id", storyID)
	}
	if s.publisher != nil {
		event := events.Event{
			Type: events.StoryUpdated,
			Payload: events.StoryUpdatedPayload{
				StoryID: storyID, WorkspaceID: workspaceID,
				Updates: map[string]any{"collaborator_ids": result.CurrentIDs}, AssigneeID: result.AssigneeID,
				AudienceIDs: audienceIDs, AudienceResolved: audienceResolved,
				PreviousCollaboratorIDs: result.PreviousIDs,
			},
			Timestamp: mutationTime, ActorID: scope.Actor.PrincipalID,
		}
		if err := s.publisher.Publish(context.WithoutCancel(ctx), event); err != nil && s.log != nil {
			s.log.Error(ctx, "failed to publish collaborators updated event", "error", err)
		}
	}
	return nil
}
