package stories

import (
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Create creates a new story.
func (s *Service) Create(ctx context.Context, ns CoreNewStory, workspaceId uuid.UUID) (CoreSingleStory, error) {
	actorID := uuid.Nil
	if actor, err := auth.GetActor(ctx); err == nil {
		actorID = actor.PrincipalID
	} else if ns.Reporter != nil {
		actorID = *ns.Reporter
	}
	return s.createWithOptions(ctx, ns, workspaceId, actorID, createOptions{
		publishEvents:     true,
		enqueueGitHubSync: true,
		traceStoryTitle:   true,
		actorKind:         auth.PrincipalHumanUser,
	})
}

func (s *Service) CreateExternal(ctx context.Context, actorID uuid.UUID, ns CoreNewStory, workspaceID uuid.UUID) (CoreSingleStory, error) {
	if ns.Reporter == nil {
		ns.Reporter = &actorID
	}
	var delivery mutationEventDelivery
	switch ns.ExternalDelivery {
	case storydomain.ExternalStoryDeliveryDefault:
	case storydomain.ExternalStoryDeliveryInternalOnly:
		delivery = mutationEventDeliveryInternalOnly
	default:
		return CoreSingleStory{}, fmt.Errorf("%w: unsupported external story delivery", ErrInvalidStoryMutation)
	}
	return s.createWithOptions(ctx, ns, workspaceID, actorID, createOptions{
		mutationEventDelivery: delivery,
		actorKind:             auth.PrincipalSystem,
	})
}

// CreateExternalUserAction creates a story from a user-initiated external
// surface such as Slack. Unlike provider ingestion, this preserves the normal
// FortyOne event and GitHub-sync side effects while retaining the explicit
// actor and external idempotency key.
func (s *Service) CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, ns CoreNewStory, workspaceID uuid.UUID) (CoreSingleStory, error) {
	if ns.Reporter == nil {
		ns.Reporter = &actorID
	}
	return s.createWithOptions(ctx, ns, workspaceID, actorID, createOptions{
		publishEvents:     true,
		enqueueGitHubSync: true,
		traceStoryTitle:   true,
		actorKind:         auth.PrincipalHumanUser,
	})
}

func (s *Service) createWithOptions(ctx context.Context, ns CoreNewStory, workspaceId, actorID uuid.UUID, options createOptions) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.create")
	ctx, span := storyServiceTracer.Start(ctx, "business.core.stories.Create")
	defer span.End()

	if actorID == uuid.Nil {
		if ns.Reporter == nil {
			return CoreSingleStory{}, fmt.Errorf("actor ID is required to create a story")
		}
		actorID = *ns.Reporter
	}
	if ns.Reporter == nil {
		if actor, actorErr := auth.GetActor(ctx); actorErr != nil || actor.IsUserActor() || actor.Kind == auth.PrincipalSystem {
			ns.Reporter = &actorID
		}
	}
	mutationRepo, useTypedMutation := s.mutationRepository()
	var scope storydomain.MutationScope
	var mutationPreconditions storydomain.MutationPreconditions
	if useTypedMutation {
		var err error
		scope, err = mutationScope(ctx, workspaceId, actorID, options.actorKind)
		if err != nil {
			return CoreSingleStory{}, mapStoryMutationError(err)
		}
		mutationPreconditions, err = mutationRepo.PrepareStoryMutation(ctx, scope, ns.Team, ns.KeyResult)
		if err != nil {
			return CoreSingleStory{}, mapStoryMutationError(err)
		}
	}
	if ns.KeyResult != nil {
		var keyResult CoreKeyResultReference
		if useTypedMutation {
			if mutationPreconditions.KeyResult == nil {
				return CoreSingleStory{}, ErrInvalidStoryReference
			}
			keyResult = CoreKeyResultReference{
				ObjectiveID: mutationPreconditions.KeyResult.ObjectiveID,
				Name:        mutationPreconditions.KeyResult.Name,
			}
		} else {
			var err error
			keyResult, err = s.resolveStoryKeyResult(ctx, workspaceId, *ns.KeyResult)
			if err != nil {
				return CoreSingleStory{}, fmt.Errorf("resolve key result objective: %w", err)
			}
		}
		if ns.Objective != nil && *ns.Objective != keyResult.ObjectiveID {
			return CoreSingleStory{}, fmt.Errorf(
				"%w: key result %s belongs to objective %s, not %s",
				ErrObjectiveKeyResultMismatch,
				*ns.KeyResult,
				keyResult.ObjectiveID,
				*ns.Objective,
			)
		}
		ns.Objective = &keyResult.ObjectiveID
	}

	story := toCoreSingleStory(ns, workspaceId)
	estimateScheme := mutationPreconditions.EstimateScheme
	if !useTypedMutation {
		var err error
		estimateScheme, err = s.getTeamEstimateScheme(ctx, workspaceId, ns.Team)
		if err != nil {
			span.RecordError(err)
			return CoreSingleStory{}, err
		}
	}

	story.EstimateScheme = estimateScheme
	if err := ValidateEstimateValue(estimateScheme, ns.EstimateValue); err != nil {
		span.RecordError(err)
		if ns.EstimateValue != nil {
			return CoreSingleStory{}, fmt.Errorf("%w. If this work is larger than the max estimate, split it into smaller stories", err)
		}
		return CoreSingleStory{}, err
	}
	story.EstimateValue = ns.EstimateValue
	story.EstimateLabel = EstimateLabelFromValue(estimateScheme, ns.EstimateValue)
	if err := ValidateStoryTimeContract(ns.EstimatedDurationMinutes, ns.MinimumFocusBlockMinutes); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	if err := s.prepareAutoSchedulingCreate(&story); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	if err := s.validateMayaSchedulingCreate(story); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	if story.AutoSchedulingEnabled {
		if err := s.validateAutoSchedulingEligibility(ctx, workspaceId); err != nil {
			span.RecordError(err)
			return CoreSingleStory{}, err
		}
	}
	if err := s.validateMayaAssignment(ctx, story, nil, actorID); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	created := true
	var cs CoreSingleStory
	var err error
	if useTypedMutation {
		story.ID = uuid.New()
		mutationTime := story.CreatedAt.UTC()
		event, eventErr := newStoryMutationEvent(
			scope,
			story.ID,
			storydomain.MutationEventStoryCreated,
			storyCreatedIntegrationPayload{
				StoryID: story.ID, WorkspaceID: workspaceId, TeamID: story.Team,
				Title: story.Title, AssigneeID: story.Assignee, ReporterID: story.Reporter,
				Delivery: options.mutationEventDelivery,
			},
			mutationTime,
		)
		if eventErr != nil {
			return CoreSingleStory{}, eventErr
		}
		activity, activityErr := newStoryMutationActivity(
			scope, story.ID, "create", "story", story.Title, nil, story.Title, nil, mutationTime,
		)
		if activityErr != nil {
			return CoreSingleStory{}, activityErr
		}
		result, mutationErr := mutationRepo.CreateStoryMutation(ctx, storydomain.CreateStoryCommand{
			Scope: scope, Story: story, LabelIDs: ns.LabelIDs, Event: event, Activity: activity,
		})
		if mutationErr != nil {
			err = mapStoryMutationError(mutationErr)
		} else {
			cs, created = result.Story, result.Created
		}
	} else if ns.CreationKey != nil {
		idempotentRepo, ok := s.repo.(idempotentCreateRepository)
		if !ok {
			return CoreSingleStory{}, errors.New("story repository does not support idempotent creation")
		}
		cs, created, err = idempotentRepo.CreateIdempotent(ctx, &story)
	} else {
		legacy, ok := s.repo.(legacyStoryCreateRepository)
		if !ok {
			return CoreSingleStory{}, errors.New("story repository does not support legacy creation")
		}
		cs, err = legacy.Create(ctx, &story)
	}
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	cs.EstimateScheme = estimateScheme
	cs.EstimateLabel = EstimateLabelFromValue(estimateScheme, cs.EstimateValue)
	cs.CreatedNow = created
	if !created {
		return cs, nil
	}

	// Record in the activity log
	ca := CoreActivity{
		StoryID:      cs.ID,
		Type:         "create",
		Field:        "story",
		CurrentValue: cs.Title,
		NewValue:     cs.Title,
		UserID:       actorID,
		WorkspaceID:  workspaceId,
	}
	if !useTypedMutation {
		if err := s.recordActivities(ctx, []CoreActivity{ca}); err != nil {
			span.RecordError(err)
		}
	}

	if options.publishEvents && s.publisher != nil {
		reporterID := uuid.Nil
		if ns.Reporter != nil {
			reporterID = *ns.Reporter
		}
		payload := events.StoryCreatedPayload{
			StoryID:     cs.ID,
			WorkspaceID: workspaceId,
			Title:       cs.Title,
			AssigneeID:  cs.Assignee,
			ReporterID:  reporterID,
		}

		event := events.Event{
			Type:      events.StoryCreated,
			Payload:   payload,
			Timestamp: time.Now(),
			ActorID:   actorID,
		}

		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish story created event", "error", err)
			// Don't return error as this is not critical
		}
	}
	if options.enqueueGitHubSync {
		s.enqueueGitHubStorySync(ctx, cs.ID, workspaceId)
	}
	if cs.AutoSchedulingEnabled {
		s.enqueueWorkspaceScheduleBatch(ctx, workspaceId)
	}
	if options.traceStoryTitle {
		span.AddEvent("story created.", trace.WithAttributes(
			attribute.String("story.title", cs.Title),
		))
	} else {
		span.AddEvent("story created.")
	}
	return cs, nil
}
