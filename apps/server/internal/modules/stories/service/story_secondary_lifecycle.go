package stories

import (
	"context"
	"errors"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// BulkDelete soft-deletes the stories with the specified IDs.
func (s *Service) BulkDelete(
	ctx context.Context,
	ids []uuid.UUID,
	workspaceID uuid.UUID,
	authorization BulkDeleteAuthorization,
) ([]uuid.UUID, error) {
	s.log.Info(ctx, "business.core.stories.BulkDelete")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.BulkDelete")
	defer span.End()

	var deletedIDs []uuid.UUID
	var err error
	if _, migrated := s.secondaryMutationRepository(); migrated {
		result, mutationErr := s.applySecondaryLifecycle(
			ctx, ids, workspaceID, authorization.ActorID, storydomain.SecondaryMutationSoftDelete,
		)
		deletedIDs, err = result.StoryIDs, mutationErr
	} else {
		legacy, ok := s.repo.(legacyBulkDeleteRepository)
		if !ok {
			return nil, errors.New("story repository does not support bulk deletion")
		}
		deletedIDs, err = legacy.BulkDelete(ctx, ids, workspaceID, authorization)
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	for _, storyID := range deletedIDs {
		s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceID)
	}
	return deletedIDs, nil
}

// HardBulkDelete permanently removes the stories with the specified IDs.
func (s *Service) HardBulkDelete(
	ctx context.Context,
	ids []uuid.UUID,
	workspaceID uuid.UUID,
	authorization BulkDeleteAuthorization,
) (HardBulkDeleteResult, error) {
	s.log.Info(ctx, "hard bulk deleting stories", "story_ids", ids)
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.HardBulkDelete")
	defer span.End()

	var result HardBulkDeleteResult
	var err error
	if _, migrated := s.secondaryMutationRepository(); migrated {
		mutationResult, mutationErr := s.applySecondaryLifecycle(
			ctx, ids, workspaceID, authorization.ActorID, storydomain.SecondaryMutationHardDelete,
		)
		result = HardBulkDeleteResult{
			StoryIDs:                         mutationResult.StoryIDs,
			OrphanedAttachmentIDs:            mutationResult.OrphanedAttachmentIDs,
			AttachmentObjectDeletionDeferred: mutationResult.AttachmentObjectDeletionDeferred,
		}
		err = mutationErr
	} else {
		legacy, ok := s.repo.(legacyHardBulkDeleteRepository)
		if !ok {
			return HardBulkDeleteResult{}, errors.New("story repository does not support hard deletion")
		}
		result, err = legacy.HardBulkDelete(ctx, ids, workspaceID, authorization)
	}
	if err != nil {
		s.log.Error(ctx, "failed to hard bulk delete stories", "story_ids", ids, "error", err)
		span.RecordError(err)
		return HardBulkDeleteResult{}, err
	}

	s.log.Info(ctx, "hard bulk deleted stories", "story_ids", ids)
	span.AddEvent("Stories hard bulk deleted.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))
	return result, nil
}

// Restore restores one soft-deleted story.
func (s *Service) Restore(ctx context.Context, id, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Restore")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.Restore")
	defer span.End()

	var err error
	if _, migrated := s.secondaryMutationRepository(); migrated {
		_, err = s.applySecondaryLifecycle(
			ctx, []uuid.UUID{id}, workspaceID, uuid.Nil, storydomain.SecondaryMutationRestore,
		)
	} else {
		legacy, ok := s.repo.(legacyRestoreRepository)
		if !ok {
			return errors.New("story repository does not support restore")
		}
		err = legacy.Restore(ctx, id, workspaceID)
	}
	if err != nil {
		span.RecordError(err)
		return err
	}
	s.enqueueRestoredAutoScheduling(ctx, []uuid.UUID{id}, workspaceID)
	return nil
}

// BulkRestore restores the soft-deleted stories with the specified IDs.
func (s *Service) BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkRestore")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.BulkRestore")
	defer span.End()

	var err error
	if _, migrated := s.secondaryMutationRepository(); migrated {
		_, err = s.applySecondaryLifecycle(ctx, ids, workspaceID, uuid.Nil, storydomain.SecondaryMutationRestore)
	} else {
		legacy, ok := s.repo.(legacyBulkRestoreRepository)
		if !ok {
			return errors.New("story repository does not support bulk restore")
		}
		err = legacy.BulkRestore(ctx, ids, workspaceID)
	}
	if err != nil {
		span.RecordError(err)
		return err
	}
	s.enqueueRestoredAutoScheduling(ctx, ids, workspaceID)
	return nil
}

// BulkArchive archives the active stories with the specified IDs.
func (s *Service) BulkArchive(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "bulk archiving stories", "story_ids", ids)
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.BulkArchive")
	defer span.End()

	var err error
	if _, migrated := s.secondaryMutationRepository(); migrated {
		_, err = s.applySecondaryLifecycle(ctx, ids, workspaceID, uuid.Nil, storydomain.SecondaryMutationArchive)
	} else {
		legacy, ok := s.repo.(legacyBulkArchiveRepository)
		if !ok {
			return errors.New("story repository does not support bulk archive")
		}
		err = legacy.BulkArchive(ctx, ids, workspaceID)
	}
	if err != nil {
		s.log.Error(ctx, "failed to bulk archive stories", "story_ids", ids, "error", err)
		span.RecordError(err)
		return err
	}

	s.log.Info(ctx, "bulk archived stories", "story_ids", ids)
	span.AddEvent("Stories bulk archived.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))
	for _, storyID := range ids {
		s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceID)
	}
	return nil
}

// BulkUnarchive unarchives the active stories with the specified IDs.
func (s *Service) BulkUnarchive(ctx context.Context, ids []uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "bulk unarchiving stories", "story_ids", ids)
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.BulkUnarchive")
	defer span.End()

	var err error
	if _, migrated := s.secondaryMutationRepository(); migrated {
		_, err = s.applySecondaryLifecycle(ctx, ids, workspaceID, uuid.Nil, storydomain.SecondaryMutationUnarchive)
	} else {
		legacy, ok := s.repo.(legacyBulkUnarchiveRepository)
		if !ok {
			return errors.New("story repository does not support bulk unarchive")
		}
		err = legacy.BulkUnarchive(ctx, ids, workspaceID)
	}
	if err != nil {
		s.log.Error(ctx, "failed to bulk unarchive stories", "story_ids", ids, "error", err)
		span.RecordError(err)
		return err
	}
	s.enqueueRestoredAutoScheduling(ctx, ids, workspaceID)

	s.log.Info(ctx, "bulk unarchived stories", "story_ids", ids)
	span.AddEvent("Stories bulk unarchived.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))
	return nil
}

func (s *Service) enqueueRestoredAutoScheduling(
	ctx context.Context,
	storyIDs []uuid.UUID,
	workspaceID uuid.UUID,
) {
	mutationRepository, typed := s.mutationRepository()
	var scope storydomain.MutationScope
	if typed {
		var err error
		scope, err = mutationScope(ctx, workspaceID, uuid.Nil, auth.PrincipalHumanUser)
		if err != nil {
			s.log.Error(ctx, "failed to authorize restored story auto-scheduling lookup",
				"workspace_id", workspaceID, "error", err)
			return
		}
	}
	for _, storyID := range storyIDs {
		autoSchedulingEnabled := false
		if typed {
			story, err := mutationRepository.GetStoryForMutation(ctx, scope, storyID)
			if err != nil {
				s.log.Error(ctx, "failed to load restored story for auto-scheduling",
					"story_id", storyID, "workspace_id", workspaceID, "error", err)
				continue
			}
			autoSchedulingEnabled = story.AutoSchedulingEnabled
		} else {
			story, err := s.getVisibleStory(ctx, storyID, workspaceID)
			if err != nil {
				s.log.Error(ctx, "failed to load restored story for auto-scheduling",
					"story_id", storyID, "workspace_id", workspaceID, "error", err)
				continue
			}
			autoSchedulingEnabled = story.AutoSchedulingEnabled
		}
		if autoSchedulingEnabled {
			s.enqueueStoryScheduleReconcile(ctx, storyID, workspaceID)
		}
	}
}
