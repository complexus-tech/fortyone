package stories

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// GetForSystem is an explicit background read. The mutation snapshot checks
// that actorID is an active system user and the story belongs to workspaceID.
// Get deliberately continues to require the requesting user's read context.
func (s *Service) GetForSystem(ctx context.Context, actorID, storyID, workspaceID uuid.UUID) (CoreSingleStory, error) {
	if actorID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil {
		return CoreSingleStory{}, ErrStoryReadForbidden
	}
	repository, ok := s.mutationRepository()
	if !ok {
		return CoreSingleStory{}, errors.New("story repository does not support system snapshots")
	}
	scope, err := mutationScope(ctx, workspaceID, actorID, auth.PrincipalSystem)
	if err != nil {
		return CoreSingleStory{}, err
	}
	if scope.Actor.Kind != auth.PrincipalSystem {
		return CoreSingleStory{}, ErrStoryReadForbidden
	}
	story, err := repository.GetStoryForMutation(ctx, scope, storyID)
	if err != nil {
		return CoreSingleStory{}, err
	}
	applySingleStoryEstimateLabels(&story)
	return story, nil
}

// RecordSystemActivity is used by durable automation events, whose actor must
// be verified as an active system user by the activity write transaction.
func (s *Service) RecordSystemActivity(ctx context.Context, activity CoreActivity) error {
	scope, err := mutationScope(ctx, activity.WorkspaceID, activity.UserID, auth.PrincipalSystem)
	if err != nil {
		return err
	}
	if scope.Actor.Kind != auth.PrincipalSystem {
		return ErrStoryMutationForbidden
	}
	ctx, err = auth.SetActor(ctx, scope.Actor)
	if err != nil {
		return err
	}
	return s.recordActivities(ctx, []CoreActivity{activity})
}

type storyEventTitleRepository interface {
	GetEventStoryTitle(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (string, error)
}

// GetEventStoryTitle resolves notification metadata using the event's explicit
// actor. SQL rechecks current human membership or an active system identity;
// notification creation independently checks each recipient's access.
func (s *Service) GetEventStoryTitle(ctx context.Context, actorID, storyID, workspaceID uuid.UUID) (string, error) {
	if actorID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil {
		return "", ErrStoryReadForbidden
	}
	repository, ok := s.repo.(storyEventTitleRepository)
	if !ok {
		return "", errors.New("story repository does not support event title reads")
	}
	title, err := repository.GetEventStoryTitle(ctx, actorID, storyID, workspaceID)
	if err != nil {
		return "", fmt.Errorf("read story event title: %w", err)
	}
	return title, nil
}
