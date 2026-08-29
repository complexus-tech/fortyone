package stories

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

// Update updates a story.
func (s *Service) Update(ctx context.Context, storyID, workspaceID uuid.UUID, updates map[string]any) error {
	actorID, _ := auth.GetUserID(ctx)
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: false,
		actorKind:                auth.PrincipalHumanUser,
	})
}

// UpdatePatch is the strongly typed story mutation entry point used by HTTP
// and new integrations. Update remains as a compatibility adapter for existing
// in-process providers while they migrate to this contract.
func (s *Service) UpdatePatch(ctx context.Context, storyID, workspaceID uuid.UUID, patch StoryPatch) error {
	actorID, _ := auth.GetUserID(ctx)
	return s.updatePatchWithOptions(ctx, storyID, workspaceID, actorID, patch, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: false,
		actorKind:                auth.PrincipalHumanUser,
	})
}

// UpdateWithMediaReconciliation applies an authoritative description snapshot.
// The caller must only use this after all editor uploads have settled. Ordinary
// updates intentionally do not reconcile media so an older autosave that omits
// a pending upload cannot unlink it.
func (s *Service) UpdateWithMediaReconciliation(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	updates map[string]any,
	referencedMediaIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	if storyID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, ErrInvalidStoryMediaReference
	}
	if _, hasDescriptionHTML := updates["description_html"].(string); !hasDescriptionHTML {
		return nil, ErrInvalidStoryMediaReference
	}
	seenMediaIDs := make(map[uuid.UUID]struct{}, len(referencedMediaIDs))
	deduplicatedMediaIDs := make([]uuid.UUID, 0, len(referencedMediaIDs))
	for _, attachmentID := range referencedMediaIDs {
		if attachmentID == uuid.Nil {
			return nil, ErrInvalidStoryMediaReference
		}
		if _, seen := seenMediaIDs[attachmentID]; seen {
			continue
		}
		seenMediaIDs[attachmentID] = struct{}{}
		deduplicatedMediaIDs = append(deduplicatedMediaIDs, attachmentID)
	}

	actorID, _ := auth.GetUserID(ctx)
	orphanedMediaIDs := []uuid.UUID{}
	err := s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: false,
		reconcileMedia:           true,
		referencedMediaIDs:       deduplicatedMediaIDs,
		orphanedMediaIDs:         &orphanedMediaIDs,
		actorKind:                auth.PrincipalHumanUser,
	})
	if err != nil {
		return nil, err
	}
	return orphanedMediaIDs, nil
}

// UpdatePatchWithMediaReconciliation is the typed equivalent of
// UpdateWithMediaReconciliation.
func (s *Service) UpdatePatchWithMediaReconciliation(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	patch StoryPatch,
	referencedMediaIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	updates := storyPatchToUpdates(patch)
	return s.UpdateWithMediaReconciliation(ctx, storyID, workspaceID, updates, referencedMediaIDs)
}

func (s *Service) UpdateExternal(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error {
	return s.UpdateExternalWithReason(ctx, actorID, storyID, workspaceID, updates, "")
}

// UpdateExternalWithReason applies a provider-originated update. Actual status
// transitions publish a domain event, but never enqueue a GitHub write-back.
func (s *Service) UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error {
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		recordDescriptionUpdates: true,
		activityReason:           reason,
		publishStatusEvents:      true,
		actorKind:                auth.PrincipalSystem,
	})
}

// UpdateExternalIfUnchanged applies an integration-originated update only when
// the story still has the version the caller inspected. This closes the race
// between confirmation-time validation and the repository write.
func (s *Service) UpdateExternalIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	return s.UpdateExternalWithReasonIfUnchanged(ctx, actorID, storyID, workspaceID, expectedUpdatedAt, updates, "")
}

// UpdateExternalWithReasonIfUnchanged preserves the inspected story version
// across a reason-aware integration update. It prevents an automated actor
// from overwriting a user edit that committed after planning.
func (s *Service) UpdateExternalWithReasonIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
	reason string,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected story update time is required")
	}
	expectedUpdatedAt = expectedUpdatedAt.UTC()
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		recordDescriptionUpdates: true,
		activityReason:           reason,
		expectedUpdatedAt:        &expectedUpdatedAt,
		publishStatusEvents:      true,
		actorKind:                auth.PrincipalSystem,
	})
}

// UpdateAutomationIfUnchanged applies a Maya-owned story mutation with
// compare-and-swap protection and publishes the full StoryUpdated contract.
// It is intentionally separate from provider-originated updates: Maya's
// assignment decisions must be observable to scheduling and notification
// consumers, while provider ingestion keeps its narrower event behavior.
func (s *Service) UpdateAutomationIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
	reason string,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected story update time is required")
	}
	expectedUpdatedAt = expectedUpdatedAt.UTC()
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		recordDescriptionUpdates: true,
		activityReason:           reason,
		expectedUpdatedAt:        &expectedUpdatedAt,
		eventSource:              events.StoryUpdateSourceMaya,
		eventReason:              reason,
		automationMutation:       true,
		actorKind:                auth.PrincipalSystem,
	})
}

// UpdateExternalUserActionIfUnchanged applies a user-initiated external edit
// with compare-and-swap protection and the same downstream event and GitHub
// synchronization behavior as an in-app update.
func (s *Service) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected story update time is required")
	}
	expectedUpdatedAt = expectedUpdatedAt.UTC()
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: true,
		expectedUpdatedAt:        &expectedUpdatedAt,
		actorKind:                auth.PrincipalHumanUser,
	})
}

func (s *Service) RecordActivity(ctx context.Context, activity CoreActivity) error {
	return s.recordActivities(ctx, []CoreActivity{activity})
}

func (s *Service) updateWithOptions(ctx context.Context, storyID, workspaceID, actorID uuid.UUID, updates map[string]any, options updateOptions) error {
	patch, err := storyPatchFromUpdates(updates)
	if err != nil {
		return err
	}
	return s.updatePatchWithOptions(ctx, storyID, workspaceID, actorID, patch, options)
}
