package stories

import (
	"context"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type storyDuplicationRepository interface {
	GetStoryForMutation(context.Context, storydomain.MutationScope, uuid.UUID) (storydomain.Story, error)
	DuplicateStoryMutation(
		context.Context,
		storydomain.DuplicateStoryCommand,
	) (storydomain.DuplicateStoryResult, error)
}

func (s *Service) duplicateStoryTyped(
	ctx context.Context,
	repository storyDuplicationRepository,
	sourceStoryID, workspaceID, suppliedActorID uuid.UUID,
) (CoreSingleStory, error) {
	scope, err := mutationScope(
		ctx, workspaceID, suppliedActorID, platformauth.PrincipalHumanUser,
	)
	if err != nil {
		return CoreSingleStory{}, mapStoryMutationError(err)
	}
	if scope.ActivityUser == nil {
		return CoreSingleStory{}, ErrStoryMutationForbidden
	}
	source, err := repository.GetStoryForMutation(ctx, scope, sourceStoryID)
	if err != nil {
		return CoreSingleStory{}, mapStoryMutationError(err)
	}

	targetStoryID := uuid.New()
	createdAt := time.Now().UTC()
	title := "Copy of " + source.Title
	reporterID := *scope.ActivityUser
	event, err := newStoryMutationEvent(
		scope,
		targetStoryID,
		storydomain.MutationEventStoryCreated,
		storyCreatedIntegrationPayload{
			StoryID: targetStoryID, WorkspaceID: workspaceID, TeamID: source.Team,
			Title: title, AssigneeID: source.Assignee, ReporterID: &reporterID,
		},
		createdAt,
	)
	if err != nil {
		return CoreSingleStory{}, err
	}
	reason := "story_duplicated"
	activity, err := newStoryMutationActivity(
		scope, targetStoryID, "create", "story", title,
		nil, targetStoryID, &reason, createdAt,
	)
	if err != nil {
		return CoreSingleStory{}, err
	}
	if activity == nil {
		return CoreSingleStory{}, fmt.Errorf("%w: duplication requires user activity", ErrStoryMutationForbidden)
	}

	result, err := repository.DuplicateStoryMutation(ctx, storydomain.DuplicateStoryCommand{
		Scope: scope, SourceStoryID: sourceStoryID, TargetStoryID: targetStoryID,
		ExpectedSourceUpdatedAt: source.UpdatedAt, ReporterID: reporterID, OccurredAt: createdAt,
		Event: event, Activity: *activity,
	})
	if err != nil {
		return CoreSingleStory{}, mapStoryMutationError(err)
	}
	return result.Story, nil
}
