package stories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) updatePatchWithOptions(ctx context.Context, storyID, workspaceID, actorID uuid.UUID, patch StoryPatch, options updateOptions) error {
	s.log.Info(ctx, "business.core.stories.Update")
	ctx, span := storyServiceTracer.Start(ctx, "business.core.stories.Update")
	defer span.End()

	if err := patch.Validate(); err != nil {
		return err
	}
	// The current scheduling and activity policies still operate on a local map.
	// This map is derived from the finite typed patch and never crosses the
	// persistence boundary.
	updates := storyPatchToUpdates(patch)
	if !options.automationMutation {
		for _, field := range []string{"auto_scheduling_status", "auto_scheduling_reason", "auto_scheduling_updated_at"} {
			if _, exists := updates[field]; exists {
				return fmt.Errorf("%s is managed by Maya", field)
			}
		}
	}

	mutationRepo, useTypedMutation := s.mutationRepository()
	var scope storydomain.MutationScope
	var story CoreSingleStory
	var err error
	if useTypedMutation {
		scope, err = mutationScope(ctx, workspaceID, actorID, options.actorKind)
		if err == nil {
			story, err = mutationRepo.GetStoryForMutation(ctx, scope, storyID)
		}
	} else {
		legacy, ok := s.repo.(legacyStoryRepository)
		if !ok {
			return errors.New("story repository does not support legacy story reads")
		}
		story, err = legacy.Get(ctx, storyID, workspaceID)
	}
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to get story", "error", err)
		return mapStoryMutationError(err)
	}
	if useTypedMutation && options.expectedUpdatedAt != nil && !story.UpdatedAt.Equal(options.expectedUpdatedAt.UTC()) {
		return ErrStoryChanged
	}
	activityDisplayValues := make(map[string]string)

	keyResultValue, hasKeyResultUpdate := updates["key_result_id"]
	if hasKeyResultUpdate {
		keyResultID, validKeyResultUpdate := optionalUUIDUpdate(keyResultValue)
		if validKeyResultUpdate && keyResultID != nil {
			var keyResult CoreKeyResultReference
			if useTypedMutation {
				preconditions, prepareErr := mutationRepo.PrepareStoryMutation(ctx, scope, story.Team, keyResultID)
				if prepareErr != nil {
					span.RecordError(prepareErr)
					return mapStoryMutationError(prepareErr)
				}
				if preconditions.KeyResult == nil {
					return ErrInvalidStoryReference
				}
				keyResult = CoreKeyResultReference{
					ObjectiveID: preconditions.KeyResult.ObjectiveID,
					Name:        preconditions.KeyResult.Name,
				}
			} else {
				var resolveErr error
				keyResult, resolveErr = s.resolveStoryKeyResult(ctx, workspaceID, *keyResultID)
				if resolveErr != nil {
					span.RecordError(resolveErr)
					return fmt.Errorf("resolve key result objective: %w", resolveErr)
				}
			}
			activityDisplayValues["key_result_id"] = keyResult.Name

			if objectiveValue, hasObjectiveUpdate := updates["objective_id"]; hasObjectiveUpdate {
				requestedObjectiveID, validObjectiveUpdate := optionalUUIDUpdate(objectiveValue)
				if !validObjectiveUpdate || requestedObjectiveID == nil || *requestedObjectiveID != keyResult.ObjectiveID {
					return fmt.Errorf(
						"%w: key result %s belongs to objective %s",
						ErrObjectiveKeyResultMismatch,
						*keyResultID,
						keyResult.ObjectiveID,
					)
				}
			}

			updates["objective_id"] = &keyResult.ObjectiveID
		}
	} else if objectiveValue, hasObjectiveUpdate := updates["objective_id"]; hasObjectiveUpdate {
		if objectiveID, validObjectiveUpdate := optionalUUIDUpdate(objectiveValue); validObjectiveUpdate && !sameOptionalUUID(objectiveID, story.Objective) {
			updates["key_result_id"] = nil
		}
	}

	if err := s.applyEstimateUpdate(ctx, workspaceID, story, updates); err != nil {
		span.RecordError(err)
		return err
	}
	if err := applyStoryTimeContractUpdate(story, updates); err != nil {
		span.RecordError(err)
		return err
	}

	// Handle auto-completion logic if status is being updated
	if newStatusID, hasStatusUpdate := updates["status_id"]; hasStatusUpdate {
		if err := s.handleCompletionStatusChange(ctx, story, newStatusID, updates); err != nil {
			s.log.Error(ctx, "failed to handle completion status change", "error", err)
			// Don't fail the entire update - log and continue
		}
	}

	reconcileAutoScheduling, err := s.prepareAutoSchedulingUpdate(ctx, story, updates)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.validateMayaSchedulingUpdate(story, updates); err != nil {
		span.RecordError(err)
		return err
	}
	enableRequested, err := autoSchedulingEnableRequested(story, updates)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if enableRequested {
		if err := s.validateAutoSchedulingEligibility(ctx, workspaceID); err != nil {
			span.RecordError(err)
			return err
		}
	}
	if assigneeID, ok := mayaAssignmentUpdateAssignee(updates); ok {
		updatedStory, err := storyWithAssignee(story, assigneeID)
		if err != nil {
			s.log.Error(ctx, "failed to prepare story for Maya assignment automation", "story_id", storyID, "workspace_id", workspaceID, "error", err)
			return err
		}
		if err := s.validateMayaAssignment(ctx, updatedStory, story.Assignee, actorID); err != nil {
			span.RecordError(err)
			return err
		}
	}
	for field, value := range updates {
		if s.valuesEqual(s.getOldValue(story, field), value) {
			delete(updates, field)
		}
	}
	if len(updates) == 0 && !options.reconcileMedia {
		return nil
	}
	var normalizedPatch StoryPatch
	if len(updates) > 0 {
		normalizedPatch, err = storyPatchFromUpdates(updates)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	mutationTime := time.Now().UTC()
	activityReason := normalizeActivityReason(options.activityReason)
	coreActivities := make([]CoreActivity, 0, len(updates))
	durableActivities := make([]storydomain.MutationActivity, 0, len(updates))
	for _, field := range normalizedPatch.Fields() {
		value := updates[field]
		if strings.Contains(field, "description") && !options.recordDescriptionUpdates {
			continue
		}
		if isAutoSchedulingStateField(field) {
			continue
		}

		currentValue := s.formatValue(value)
		if displayValue, ok := activityDisplayValues[field]; ok {
			currentValue = displayValue
		}
		oldValue := s.getOldValue(story, field)
		coreActivities = append(coreActivities, CoreActivity{
			ID: uuid.New(), StoryID: storyID, Type: "update", Field: field,
			CurrentValue: currentValue, OldValue: oldValue, NewValue: value,
			Reason: activityReason, UserID: actorID, WorkspaceID: workspaceID,
			CreatedAt: mutationTime,
		})
		activity, activityErr := newStoryMutationActivity(
			scope, storyID, "update", field, currentValue, oldValue, value,
			activityReason, mutationTime,
		)
		if activityErr != nil && useTypedMutation {
			return activityErr
		}
		if activity != nil {
			durableActivities = append(durableActivities, *activity)
		}
	}

	// Update the story, reconciling inline media only for an explicitly
	// authoritative editor snapshot.
	var updateErr error
	if useTypedMutation {
		event, eventErr := newStoryMutationEvent(
			scope,
			storyID,
			storydomain.MutationEventStoryUpdated,
			storyUpdatedIntegrationPayload{StoryID: storyID, WorkspaceID: workspaceID, Changes: updates},
			mutationTime,
		)
		if eventErr != nil {
			return eventErr
		}
		result, mutationErr := mutationRepo.ApplyStoryMutation(ctx, storydomain.UpdateStoryCommand{
			Scope: scope, StoryID: storyID, ExpectedUpdatedAt: story.UpdatedAt.UTC(),
			Patch: normalizedPatch, Event: event, Activities: durableActivities,
			ReferencedMedia: options.referencedMediaIDs, ReconcileMedia: options.reconcileMedia,
		})
		updateErr = mapStoryMutationError(mutationErr)
		if updateErr == nil && options.orphanedMediaIDs != nil {
			*options.orphanedMediaIDs = append((*options.orphanedMediaIDs)[:0], result.OrphanedAttachmentIDs...)
		}
	} else if len(updates) == 0 {
		legacy, ok := s.repo.(legacyStoryMediaRepository)
		if !ok {
			return errors.New("story repository does not support legacy media reconciliation")
		}
		var orphanedMediaIDs []uuid.UUID
		orphanedMediaIDs, updateErr = legacy.UpdateWithMediaReconciliation(
			ctx,
			storyID,
			workspaceID,
			updates,
			options.referencedMediaIDs,
		)
		if updateErr == nil && options.orphanedMediaIDs != nil {
			*options.orphanedMediaIDs = append((*options.orphanedMediaIDs)[:0], orphanedMediaIDs...)
		}
	} else if options.reconcileMedia {
		legacy, ok := s.repo.(legacyStoryMediaRepository)
		if !ok {
			return errors.New("story repository does not support legacy media reconciliation")
		}
		var orphanedMediaIDs []uuid.UUID
		orphanedMediaIDs, updateErr = legacy.UpdateWithMediaReconciliation(
			ctx,
			storyID,
			workspaceID,
			updates,
			options.referencedMediaIDs,
		)
		if updateErr == nil && options.orphanedMediaIDs != nil {
			*options.orphanedMediaIDs = append((*options.orphanedMediaIDs)[:0], orphanedMediaIDs...)
		}
	} else {
		if options.expectedUpdatedAt == nil {
			legacy, ok := s.repo.(legacyStoryUpdateRepository)
			if !ok {
				return errors.New("story repository does not support legacy updates")
			}
			updateErr = legacy.Update(ctx, storyID, workspaceID, updates)
		} else {
			conditionalRepo, ok := s.repo.(conditionalUpdateRepository)
			if !ok {
				return errors.New("story repository does not support conditional updates")
			}
			var updated bool
			updated, updateErr = conditionalRepo.UpdateIfUnchanged(
				ctx,
				storyID,
				workspaceID,
				*options.expectedUpdatedAt,
				updates,
			)
			if updateErr == nil && !updated {
				updateErr = ErrStoryChanged
			}
		}
	}
	if updateErr != nil {
		span.RecordError(updateErr)
		return updateErr
	}
	if !useTypedMutation && len(coreActivities) > 0 {
		if err := s.recordActivities(ctx, coreActivities); err != nil {
			span.RecordError(err)
		}
	}

	span.AddEvent("story updated", trace.WithAttributes(
		attribute.String("story.id", storyID.String()),
	))

	_, statusChanged := updates["status_id"]
	if (options.publishEvents || (options.publishStatusEvents && statusChanged)) && s.publisher != nil {
		audienceIDs, audienceErr := s.GetNotificationAudience(ctx, storyID, workspaceID)
		audienceResolved := audienceErr == nil
		if audienceErr != nil {
			s.log.Error(ctx, "failed to load story notification audience", "error", audienceErr, "story_id", storyID)
		}
		payload := events.StoryUpdatedPayload{
			StoryID:          storyID,
			WorkspaceID:      workspaceID,
			Updates:          updates,
			AssigneeID:       story.Assignee, // Current assignee before update
			AudienceIDs:      audienceIDs,
			AudienceResolved: audienceResolved,
			Source:           options.eventSource,
			Reason:           options.eventReason,
			Schedule:         options.eventSchedule,
		}
		if statusChanged {
			payload.PreviousStatusID = story.Status
		}

		event := events.Event{
			Type:      events.StoryUpdated,
			Payload:   payload,
			Timestamp: time.Now(),
			ActorID:   actorID,
		}

		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish story updated event", "error", err)
			// Don't return error as this is not critical
		}
	}
	if options.enqueueGitHubSync {
		s.enqueueGitHubStorySync(ctx, storyID, workspaceID)
	}
	if reconcileAutoScheduling {
		if scheduleReconcileMustRunImmediately(updates) {
			s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceID)
		} else {
			s.enqueueWorkspaceScheduleBatch(ctx, workspaceID)
		}
	}

	return nil
}

func isAutoSchedulingStateField(field string) bool {
	switch field {
	case "auto_scheduling_status", "auto_scheduling_reason", "auto_scheduling_updated_at":
		return true
	default:
		return false
	}
}

func normalizeActivityReason(reason string) *string {
	const maxActivityReasonRunes = 180

	normalized := strings.Join(strings.Fields(reason), " ")
	if normalized == "" {
		return nil
	}

	runes := []rune(normalized)
	if len(runes) > maxActivityReasonRunes {
		normalized = strings.TrimSpace(string(runes[:maxActivityReasonRunes-3])) + "..."
	}
	return &normalized
}
