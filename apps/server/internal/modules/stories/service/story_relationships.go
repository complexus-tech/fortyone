package stories

import (
	"context"
	"errors"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
)

// UpdateLabels replaces the labels for a story.
func (s *Service) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID, labels []uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.UpdateLabels")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.UpdateLabels")
	defer span.End()

	if repository, migrated := s.secondaryMutationRepository(); migrated {
		scope, err := mutationScope(ctx, workspaceID, uuid.Nil, auth.PrincipalHumanUser)
		if err != nil {
			return err
		}
		next, err := storydomain.NormalizeSecondaryReplacementIDs(labels)
		if err != nil {
			return mapStoryMutationError(err)
		}
		mutationTime := time.Now().UTC()
		var activity *storydomain.MutationActivity
		if scope.ActivityUser != nil {
			activity, err = newStoryMutationActivity(
				scope, id, "update", "labels", s.formatLabelActivityValue(next), nil, next, nil, mutationTime,
			)
			if err != nil {
				return err
			}
		}
		event, err := newStoryMutationEvent(scope, id, storydomain.MutationEventStoryUpdated,
			storyUpdatedIntegrationPayload{StoryID: id, WorkspaceID: workspaceID, Changes: map[string]any{"label_ids": "changed"}}, mutationTime)
		if err != nil {
			return err
		}
		result, err := repository.ReplaceStoryLabels(ctx, storydomain.ReplaceStoryLabelsCommand{
			Scope: scope, StoryID: id, LabelIDs: next, Event: event, Activity: activity,
		})
		if err != nil {
			span.RecordError(err)
			return mapStoryMutationError(err)
		}
		if !result.Changed {
			return nil
		}
		return nil
	} else if legacy, ok := s.repo.(legacyStoryLabelsRepository); !ok {
		return errors.New("story repository does not support label replacement")
	} else if err := legacy.UpdateLabels(ctx, id, workspaceID, labels); err != nil {
		span.RecordError(err)
		return err
	}
	actorID, _ := auth.GetUserID(ctx)
	if err := s.RecordActivity(ctx, CoreActivity{
		StoryID: id, Type: "update", Field: "labels",
		CurrentValue: s.formatLabelActivityValue(labels), NewValue: labels,
		UserID: actorID, WorkspaceID: workspaceID,
	}); err != nil {
		span.RecordError(err)
	}
	return nil
}

// UpdateCollaborators replaces the active collaborators for a story.
func (s *Service) UpdateCollaborators(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	collaboratorIDs []uuid.UUID,
) error {
	if _, migrated := s.secondaryMutationRepository(); migrated {
		return s.updateCollaboratorsTyped(ctx, storyID, workspaceID, collaboratorIDs)
	}
	actorID, _ := auth.GetUserID(ctx)

	story, err := s.getVisibleStory(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}

	legacy, ok := s.repo.(legacyStoryCollaborationRepository)
	if !ok {
		return errors.New("story repository does not support collaborator replacement")
	}
	previousIDs, err := legacy.GetCollaborators(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}
	nextIDs := uniqueUUIDs(collaboratorIDs)
	if sameUUIDSet(previousIDs, nextIDs) {
		return nil
	}

	if err := legacy.SetCollaborators(ctx, storyID, workspaceID, nextIDs); err != nil {
		return err
	}

	activity := CoreActivity{
		StoryID: storyID, Type: "update", Field: "collaborator_ids",
		CurrentValue: s.formatValue(nextIDs), OldValue: previousIDs, NewValue: nextIDs,
		UserID: actorID, WorkspaceID: workspaceID,
	}
	if err := s.recordActivities(ctx, []CoreActivity{activity}); err != nil {
		s.log.Error(ctx, "failed to record collaborator activity", "error", err, "story_id", storyID)
	}

	audienceIDs, err := legacy.GetNotificationAudience(ctx, storyID, workspaceID)
	audienceResolved := err == nil
	if err != nil {
		s.log.Error(ctx, "failed to load story notification audience", "error", err, "story_id", storyID)
		audienceIDs = nil
	}
	if s.publisher != nil {
		event := events.Event{
			Type: events.StoryUpdated,
			Payload: events.StoryUpdatedPayload{
				StoryID: storyID, WorkspaceID: workspaceID,
				Updates: map[string]any{"collaborator_ids": nextIDs}, AssigneeID: story.Assignee,
				AudienceIDs: audienceIDs, AudienceResolved: audienceResolved,
				PreviousCollaboratorIDs: previousIDs,
			},
			Timestamp: time.Now(), ActorID: actorID,
		}
		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish collaborators updated event", "error", err)
		}
	}
	return nil
}

// SetWatching updates the current user's explicit or muted story subscription.
func (s *Service) SetWatching(
	ctx context.Context,
	storyID, workspaceID, userID uuid.UUID,
	watching bool,
) error {
	if repository, migrated := s.secondaryMutationRepository(); migrated {
		scope, err := mutationScope(ctx, workspaceID, userID, auth.PrincipalHumanUser)
		if err != nil {
			return err
		}
		return mapStoryMutationError(repository.SetStoryWatching(ctx, scope, storyID, userID, watching))
	}
	legacy, ok := s.repo.(legacyStoryCollaborationRepository)
	if !ok {
		return errors.New("story repository does not support watch preferences")
	}
	return legacy.SetWatching(ctx, storyID, workspaceID, userID, watching)
}

// GetNotificationAudience returns current, active, non-muted users following a story.
func (s *Service) GetNotificationAudience(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
) ([]uuid.UUID, error) {
	if repository, migrated := s.secondaryMutationRepository(); migrated {
		return repository.ListStoryNotificationAudience(ctx, storyID, workspaceID)
	}
	legacy, ok := s.repo.(legacyStoryCollaborationRepository)
	if !ok {
		return nil, errors.New("story repository does not support notification audiences")
	}
	return legacy.GetNotificationAudience(ctx, storyID, workspaceID)
}
