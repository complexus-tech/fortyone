package stories

import (
	"context"
	"errors"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// Delete deletes the story with the specified ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, authorization BulkDeleteAuthorization) error {
	s.log.Info(ctx, "business.core.stories.Delete")
	ctx, span := storyServiceTracer.Start(ctx, "business.core.stories.Delete")
	defer span.End()

	if mutationRepo, ok := s.mutationRepository(); ok {
		scope, err := mutationScope(ctx, workspaceId, authorization.ActorID, auth.PrincipalHumanUser)
		if err != nil {
			return mapStoryMutationError(err)
		}
		story, err := mutationRepo.GetStoryForMutation(ctx, scope, id)
		if err != nil {
			return mapStoryMutationError(err)
		}
		mutationTime := time.Now().UTC()
		event, err := newStoryMutationEvent(
			scope,
			id,
			storydomain.MutationEventStoryDeleted,
			storyDeletedIntegrationPayload{StoryID: id, WorkspaceID: workspaceId},
			mutationTime,
		)
		if err != nil {
			return err
		}
		activity, err := newStoryMutationActivity(
			scope, id, "delete", "story", story.Title, story.Title, nil, nil, mutationTime,
		)
		if err != nil {
			return err
		}
		_, err = mutationRepo.DeleteStoryMutation(ctx, storydomain.DeleteStoryCommand{
			Scope: scope, StoryID: id, ExpectedUpdatedAt: story.UpdatedAt.UTC(),
			Event: event, Activity: activity,
		})
		if errors.Is(err, storydomain.ErrMutationForbidden) {
			err = ErrDeleteForbidden
		} else {
			err = mapStoryMutationError(err)
		}
		if err != nil {
			span.RecordError(err)
			return err
		}
	} else {
		legacy, ok := s.repo.(legacyStoryDeleteRepository)
		if !ok {
			return errors.New("story repository does not support legacy deletion")
		}
		if err := legacy.Delete(ctx, id, workspaceId, authorization); err != nil {
			span.RecordError(err)
			return err
		}
	}
	s.enqueueStoryScheduleReconcile(ctx, id, workspaceId)
	return nil
}
