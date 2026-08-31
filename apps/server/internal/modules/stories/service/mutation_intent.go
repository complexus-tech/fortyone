package stories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// storyMutationRepository is a caller-owned port. The service depends on the
// mutation capability it needs, while the SQLC/pgx adapter remains in the
// repository package.
type storyMutationRepository interface {
	PrepareStoryMutation(
		context.Context,
		storydomain.MutationScope,
		uuid.UUID,
		*uuid.UUID,
	) (storydomain.MutationPreconditions, error)
	GetStoryForMutation(context.Context, storydomain.MutationScope, uuid.UUID) (storydomain.Story, error)
	CreateStoryMutation(context.Context, storydomain.CreateStoryCommand) (storydomain.CreateStoryResult, error)
	ApplyStoryMutation(context.Context, storydomain.UpdateStoryCommand) (storydomain.UpdateStoryResult, error)
	DeleteStoryMutation(context.Context, storydomain.DeleteStoryCommand) (storydomain.DeleteStoryResult, error)
}

func (s *Service) mutationRepository() (storyMutationRepository, bool) {
	repository, ok := s.repo.(storyMutationRepository)
	return repository, ok
}

func mutationScope(
	ctx context.Context,
	workspaceID, suppliedActorID uuid.UUID,
	fallbackKind platformauth.PrincipalKind,
) (storydomain.MutationScope, error) {
	actor, err := platformauth.GetActor(ctx)
	if err == nil {
		if suppliedActorID != uuid.Nil && actor.PrincipalID != suppliedActorID {
			return storydomain.MutationScope{}, fmt.Errorf("%w: actor identity mismatch", ErrStoryMutationForbidden)
		}
		if actor.WorkspaceID == uuid.Nil {
			actor, err = actor.WithWorkspace(workspaceID)
			if err != nil {
				return storydomain.MutationScope{}, fmt.Errorf("%w: bind actor workspace: %v", ErrStoryMutationForbidden, err)
			}
		}
	} else if errors.Is(err, platformauth.ErrActorNotFound) {
		if suppliedActorID == uuid.Nil {
			return storydomain.MutationScope{}, fmt.Errorf("%w: actor is required", ErrStoryMutationForbidden)
		}
		actor, err = platformauth.NewActor(
			suppliedActorID,
			fallbackKind,
			uuid.Nil,
			platformauth.MustScopeSet(platformauth.ScopeFirstParty),
			platformauth.UnrestrictedTeamAccess(),
		)
		if err == nil {
			actor, err = actor.WithWorkspace(workspaceID)
		}
		if err != nil {
			return storydomain.MutationScope{}, fmt.Errorf("%w: construct actor: %v", ErrStoryMutationForbidden, err)
		}
	} else {
		return storydomain.MutationScope{}, fmt.Errorf("%w: read actor: %v", ErrStoryMutationForbidden, err)
	}
	if actor.WorkspaceID != workspaceID || !actor.Scopes.Has(platformauth.ScopeStoriesWrite) {
		return storydomain.MutationScope{}, ErrStoryMutationForbidden
	}

	var activityUser *uuid.UUID
	if actor.IsUserActor() || actor.Kind == platformauth.PrincipalSystem {
		userID := actor.PrincipalID
		activityUser = &userID
	}
	scope := storydomain.MutationScope{Actor: actor, WorkspaceID: workspaceID, ActivityUser: activityUser}
	if err := scope.Validate(); err != nil {
		return storydomain.MutationScope{}, err
	}
	return scope, nil
}

func newStoryMutationEvent(
	scope storydomain.MutationScope,
	storyID uuid.UUID,
	eventType storydomain.MutationEventType,
	payload any,
	occurredAt time.Time,
) (storydomain.MutationEvent, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return storydomain.MutationEvent{}, fmt.Errorf("encode story mutation event: %w", err)
	}
	event := storydomain.MutationEvent{
		ID: uuid.New(), WorkspaceID: scope.WorkspaceID, StoryID: storyID,
		Type: eventType, Actor: scope.Actor, Payload: body, OccurredAt: occurredAt.UTC(),
	}
	if err := event.Validate(); err != nil {
		return storydomain.MutationEvent{}, err
	}
	return event, nil
}

func newStoryMutationActivity(
	scope storydomain.MutationScope,
	storyID uuid.UUID,
	activityType, field, currentValue string,
	oldValue, newValue any,
	reason *string,
	createdAt time.Time,
) (*storydomain.MutationActivity, error) {
	if scope.ActivityUser == nil {
		return nil, nil
	}
	oldBody, err := json.Marshal(oldValue)
	if err != nil {
		return nil, fmt.Errorf("encode old activity value: %w", err)
	}
	newBody, err := json.Marshal(newValue)
	if err != nil {
		return nil, fmt.Errorf("encode new activity value: %w", err)
	}
	activity := &storydomain.MutationActivity{
		ID: uuid.New(), StoryID: storyID, UserID: *scope.ActivityUser,
		Type: activityType, Field: field, CurrentValue: currentValue,
		OldValue: oldBody, NewValue: newBody, Reason: reason,
		WorkspaceID: scope.WorkspaceID, CreatedAt: createdAt.UTC(),
	}
	if err := activity.Validate(scope, storyID); err != nil {
		return nil, err
	}
	return activity, nil
}

type mutationEventDelivery string

const mutationEventDeliveryInternalOnly mutationEventDelivery = "internal_only"

type storyCreatedIntegrationPayload struct {
	StoryID     uuid.UUID             `json:"storyId"`
	WorkspaceID uuid.UUID             `json:"workspaceId"`
	TeamID      uuid.UUID             `json:"teamId"`
	Title       string                `json:"title"`
	AssigneeID  *uuid.UUID            `json:"assigneeId"`
	ReporterID  *uuid.UUID            `json:"reporterId"`
	Delivery    mutationEventDelivery `json:"_delivery,omitempty"`
}

type storyUpdatedIntegrationPayload struct {
	StoryID     uuid.UUID      `json:"storyId"`
	WorkspaceID uuid.UUID      `json:"workspaceId"`
	Changes     map[string]any `json:"changes"`
}

type storyDeletedIntegrationPayload struct {
	StoryID     uuid.UUID `json:"storyId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
}

func mapStoryMutationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, storydomain.ErrMutationConflict):
		return ErrStoryChanged
	case errors.Is(err, storydomain.ErrMutationForbidden):
		return ErrStoryMutationForbidden
	case errors.Is(err, storydomain.ErrInvalidMutation):
		return errors.Join(ErrInvalidStoryMutation, err)
	default:
		return err
	}
}
