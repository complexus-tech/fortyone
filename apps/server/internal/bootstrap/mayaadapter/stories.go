package mayaadapter

import (
	"context"
	"errors"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

type StoryReader interface {
	GetStoryForMutation(context.Context, storydomain.MutationScope, uuid.UUID) (storydomain.Story, error)
}

// storyService keeps background scheduling on the system-authorized story
// snapshot path. Ordinary API reads still require the requesting user's actor
// and current membership; the database validates the active system identity here.
type storyService struct {
	maya.StoriesService
	reader  StoryReader
	actorID uuid.UUID
}

// New wires Maya's authorized operations to its explicit system identity.
// Interactive reads keep the user's access checks; only automation writes
// deliberately switch to Maya after the Maya service validates the operation.
func New(source maya.StoriesService, reader StoryReader, actorID uuid.UUID) maya.StoriesService {
	return storyService{StoriesService: source, reader: reader, actorID: actorID}
}

var _ maya.StoriesService = storyService{}

func (s storyService) Get(ctx context.Context, storyID, workspaceID uuid.UUID) (storydomain.Story, error) {
	if s.reader == nil || s.actorID == uuid.Nil || workspaceID == uuid.Nil || storyID == uuid.Nil {
		return storydomain.Story{}, errors.New("Maya worker story reader, system actor, workspace, and story are required")
	}
	if contextualActor, err := auth.GetActor(ctx); err == nil {
		if contextualActor.WorkspaceID != uuid.Nil && contextualActor.WorkspaceID != workspaceID {
			return storydomain.Story{}, stories.ErrStoryReadForbidden
		}
		if contextualActor.Kind != auth.PrincipalSystem {
			return s.StoriesService.Get(ctx, storyID, workspaceID)
		}
		if contextualActor.PrincipalID != s.actorID || !contextualActor.TeamAccess.IsUnrestricted() || !contextualActor.Scopes.Has(auth.ScopeStoriesWrite) {
			return storydomain.Story{}, stories.ErrStoryReadForbidden
		}
	} else if !errors.Is(err, auth.ErrActorNotFound) {
		return storydomain.Story{}, err
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

func (s storyService) automationContext(ctx context.Context, actorID, workspaceID uuid.UUID) (context.Context, error) {
	if actorID == uuid.Nil || actorID != s.actorID || workspaceID == uuid.Nil {
		return nil, stories.ErrStoryMutationForbidden
	}
	teamAccess := auth.UnrestrictedTeamAccess()
	if caller, err := auth.GetActor(ctx); err == nil {
		if (caller.WorkspaceID != uuid.Nil && caller.WorkspaceID != workspaceID) ||
			!caller.Scopes.Has(auth.ScopeStoriesWrite) ||
			(!caller.IsUserActor() && (caller.Kind != auth.PrincipalSystem || caller.PrincipalID != actorID)) {
			return nil, stories.ErrStoryMutationForbidden
		}
		teamAccess = caller.TeamAccess
	} else if !errors.Is(err, auth.ErrActorNotFound) {
		return nil, err
	}
	actor, err := auth.NewActor(actorID, auth.PrincipalSystem, uuid.Nil,
		auth.MustScopeSet(auth.ScopeStoriesRead, auth.ScopeStoriesWrite), teamAccess)
	if err != nil {
		return nil, err
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}
	return auth.SetActor(ctx, actor)
}

func (s storyService) UpdateAutomationIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error {
	ctx, err := s.automationContext(ctx, actorID, workspaceID)
	if err != nil {
		return err
	}
	return s.StoriesService.UpdateAutomationIfUnchanged(ctx, actorID, storyID, workspaceID, expectedUpdatedAt, updates, reason)
}

func (s storyService) UpdateAutomationStateIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, status string, reason *string, locked *bool, schedule *events.StoryScheduleTransition) error {
	ctx, err := s.automationContext(ctx, actorID, workspaceID)
	if err != nil {
		return err
	}
	return s.StoriesService.UpdateAutomationStateIfUnchanged(ctx, actorID, storyID, workspaceID, expectedUpdatedAt, status, reason, locked, schedule)
}
