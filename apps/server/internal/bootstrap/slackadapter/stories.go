package slackadapter

import (
	"context"
	"errors"
	"time"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type StoryBackend interface {
	Create(context.Context, stories.CoreNewStory, uuid.UUID) (stories.CoreSingleStory, error)
	QueryByRef(context.Context, uuid.UUID, string) (stories.CoreSingleStory, error)
	QueryByRefForIntegration(context.Context, uuid.UUID, string) (stories.CoreSingleStory, error)
	UpdateExternalUserActionIfUnchanged(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time, map[string]any) error
}

type StoryService struct {
	backend StoryBackend
}

func NewStoryService(backend StoryBackend) *StoryService {
	if backend == nil {
		return nil
	}
	return &StoryService{backend: backend}
}

func (adapter *StoryService) Create(ctx context.Context, input slack.NewStory, workspaceID uuid.UUID) (slack.Story, error) {
	story, err := adapter.backend.Create(ctx, stories.CoreNewStory{
		Title: input.Title, EstimateValue: input.EstimateValue,
		EstimatedDurationMinutes: input.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: input.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    input.AutoSchedulingEnabled, AutoSchedulingLocked: input.AutoSchedulingLocked,
		Description: input.Description, DescriptionHTML: input.DescriptionHTML,
		Objective: input.Objective, Status: input.Status, Assignee: input.Assignee, Reporter: input.Reporter,
		Priority: input.Priority, Sprint: input.Sprint, KeyResult: input.KeyResult,
		LabelIDs: append([]uuid.UUID(nil), input.LabelIDs...), StartDate: input.StartDate,
		EndDate: input.EndDate, Team: input.Team, CreationKey: input.CreationKey,
	}, workspaceID)
	return mapStory(story), mapStoryError(err)
}

func (adapter *StoryService) QueryByRef(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	reference string,
) (slack.Story, error) {
	return adapter.QueryByRefForUser(ctx, workspaceID, actorID, reference)
}

func (adapter *StoryService) QueryByRefForUser(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	reference string,
) (slack.Story, error) {
	if actorID == uuid.Nil {
		return slack.Story{}, errors.New("slack story actor is required")
	}

	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		return slack.Story{}, err
	}
	ctx, err = platformauth.SetActor(ctx, actor)
	if err != nil {
		return slack.Story{}, err
	}

	story, err := adapter.backend.QueryByRef(ctx, workspaceID, reference)
	return mapStory(story), mapStoryError(err)
}

func (adapter *StoryService) QueryByRefForInstallation(
	ctx context.Context,
	workspaceID, installationID uuid.UUID,
	allowedTeamIDs []uuid.UUID,
	reference string,
) (slack.Story, error) {
	if workspaceID == uuid.Nil || installationID == uuid.Nil || len(allowedTeamIDs) == 0 {
		return slack.Story{}, errors.New("slack installation story scope is required")
	}
	teamAccess, err := platformauth.RestrictedTeamAccess(allowedTeamIDs...)
	if err != nil {
		return slack.Story{}, err
	}
	actor, err := platformauth.NewActor(
		installationID,
		platformauth.PrincipalServiceAccount,
		installationID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
		teamAccess,
	)
	if err != nil {
		return slack.Story{}, err
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		return slack.Story{}, err
	}
	ctx, err = platformauth.SetActor(ctx, actor)
	if err != nil {
		return slack.Story{}, err
	}
	story, err := adapter.backend.QueryByRefForIntegration(ctx, workspaceID, reference)
	return mapStory(story), mapStoryError(err)
}

func (adapter *StoryService) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	return mapStoryError(adapter.backend.UpdateExternalUserActionIfUnchanged(
		ctx,
		actorID,
		storyID,
		workspaceID,
		expectedUpdatedAt,
		cloneAnyMap(updates),
	))
}

func mapStory(story stories.CoreSingleStory) slack.Story {
	return slack.Story{
		ID: story.ID, SequenceID: story.SequenceID, Title: story.Title, TeamCode: story.TeamCode,
		Description: story.Description, DescriptionHTML: story.DescriptionHTML,
		Status: story.Status, Assignee: story.Assignee,
		Reporter: story.Reporter, Priority: story.Priority, Team: story.Team, Workspace: story.Workspace,
		EndDate: story.EndDate, CreatedAt: story.CreatedAt, UpdatedAt: story.UpdatedAt,
		CreatedNow: story.CreatedNow,
	}
}

func mapStoryError(err error) error {
	switch {
	case errors.Is(err, stories.ErrNotFound):
		return errors.Join(slack.ErrStoryNotFound, err)
	case errors.Is(err, stories.ErrInvalidStoryReference):
		return errors.Join(slack.ErrInvalidStoryReference, err)
	case errors.Is(err, stories.ErrStoryChanged):
		return errors.Join(slack.ErrStoryChanged, err)
	default:
		return err
	}
}

var (
	_ slack.StoryService     = (*StoryService)(nil)
	_ slack.SlackStoryReader = (*StoryService)(nil)
)
