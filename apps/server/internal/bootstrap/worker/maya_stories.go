package workerbootstrap

import (
	"context"
	"errors"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type workerMayaStoryReader interface {
	GetStoryForMutation(context.Context, storydomain.MutationScope, uuid.UUID) (storydomain.Story, error)
}

// workerMayaStories keeps background scheduling on the system-authorized story
// snapshot path. Ordinary API reads still require the requesting user's actor
// and current membership; the database validates the active system identity here.
type workerMayaStories struct {
	maya.StoriesService
	reader  workerMayaStoryReader
	actorID uuid.UUID
}

var _ maya.StoriesService = workerMayaStories{}

func (s workerMayaStories) Get(ctx context.Context, storyID, workspaceID uuid.UUID) (storydomain.Story, error) {
	if s.reader == nil || s.actorID == uuid.Nil || workspaceID == uuid.Nil || storyID == uuid.Nil {
		return storydomain.Story{}, errors.New("Maya worker story reader, system actor, workspace, and story are required")
	}
	actor, err := auth.NewActor(s.actorID, auth.PrincipalSystem, uuid.Nil,
		auth.MustScopeSet(auth.ScopeStoriesWrite), auth.UnrestrictedTeamAccess())
	if err != nil {
		return storydomain.Story{}, err
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		return storydomain.Story{}, err
	}
	scope := storydomain.MutationScope{Actor: actor, WorkspaceID: workspaceID, ActivityUser: &s.actorID}
	if err := scope.Validate(); err != nil {
		return storydomain.Story{}, err
	}
	story, err := s.reader.GetStoryForMutation(ctx, scope, storyID)
	if err != nil {
		return storydomain.Story{}, err
	}
	story.EstimateLabel = stories.EstimateLabelFromValue(story.EstimateScheme, story.EstimateValue)
	return story, nil
}
