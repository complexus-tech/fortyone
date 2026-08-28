package stories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) enqueueGitHubStorySync(ctx context.Context, storyID, workspaceID uuid.UUID) {
	if s.tasksService == nil {
		return
	}
	if _, err := s.tasksService.EnqueueGitHubStorySync(tasks.GitHubStorySyncPayload{
		StoryID:     storyID,
		WorkspaceID: workspaceID,
	}); err != nil {
		s.log.Error(ctx, "failed to enqueue github story sync task", "story_id", storyID, "workspace_id", workspaceID, "error", err)
	}
}

func (s *Service) enqueueStoryScheduleReconcile(ctx context.Context, storyID, workspaceID uuid.UUID) {
	if s.tasksService == nil {
		return
	}
	enqueueCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.tasksService.EnqueueStoryScheduleReconcile(enqueueCtx, workspaceID, storyID); err != nil {
		s.log.Error(ctx, "failed to enqueue story schedule reconciliation", "story_id", storyID, "workspace_id", workspaceID, "error", err)
	}
}

func (s *Service) enqueueWorkspaceScheduleBatch(ctx context.Context, workspaceID uuid.UUID) {
	if s.tasksService == nil {
		return
	}
	enqueueCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.tasksService.EnqueueCalendarWorkspaceScheduleBatch(enqueueCtx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to enqueue workspace schedule batch", "workspace_id", workspaceID, "error", err)
	}
}

func scheduleReconcileMustRunImmediately(updates map[string]any) bool {
	if status, ok := updates["auto_scheduling_status"].(string); ok && status == AutoSchedulingStatusOff {
		return true
	}
	if enabled, ok := updates["auto_scheduling_enabled"].(bool); ok && !enabled {
		return true
	}
	if rawAssignee, changed := updates["assignee_id"]; changed {
		assignee, valid := optionalUUIDUpdate(rawAssignee)
		if valid && assignee == nil {
			return true
		}
	}
	if archivedAt, changed := updates["archived_at"]; changed && archivedAt != nil {
		return true
	}
	return false
}

// handleCompletionStatusChange handles auto-setting completed_at based on status category changes
func (s *Service) handleCompletionStatusChange(ctx context.Context, story CoreSingleStory,
	newStatusID any, updates map[string]any) error {

	// Convert status ID to string
	newStatus, ok := newStatusID.(uuid.UUID)
	if !ok {
		return fmt.Errorf("status ID is not a string: %T", newStatusID)
	}

	// Get old status category
	oldCategory, err := s.getStoryStatusCategory(ctx, story.Workspace, *story.Status)
	if err != nil {
		s.log.Error(ctx, "failed to get old status category", "statusId", *story.Status, "error", err)
		// Continue without old category info
	}

	// Get new status category
	newCategory, err := s.getStoryStatusCategory(ctx, story.Workspace, newStatus)
	if err != nil {
		return fmt.Errorf("failed to get new status category: %w", err)
	}

	// Auto-completion logic
	now := time.Now()
	if newCategory == "completed" && oldCategory != "completed" {
		updates["completed_at"] = &now
		s.log.Info(ctx, "auto-completing story", "storyId", story.ID, "oldCategory", oldCategory, "newCategory", newCategory)
	} else if oldCategory == "completed" && newCategory != "completed" {
		updates["completed_at"] = nil
		s.log.Info(ctx, "auto-uncompleting story", "storyId", story.ID, "oldCategory", oldCategory, "newCategory", newCategory)
	}
	// If both old and new are in completed category, do nothing (keep existing completed_at)

	return nil
}

// BulkUpdate updates multiple stories with the same updates in parallel and
// returns an ordered receipt for every requested story.
func (s *Service) BulkUpdate(ctx context.Context, storyIDs []uuid.UUID, workspaceID uuid.UUID, updates map[string]any) (BulkUpdateResult, error) {
	patch, err := storyPatchFromUpdates(updates)
	if err != nil {
		return BulkUpdateResult{}, err
	}
	return s.BulkUpdatePatch(ctx, storyIDs, workspaceID, patch)
}

// BulkUpdatePatch applies one immutable typed patch to each target. Every item
// independently enforces its own current authorization and version boundary.
func (s *Service) BulkUpdatePatch(ctx context.Context, storyIDs []uuid.UUID, workspaceID uuid.UUID, patch StoryPatch) (BulkUpdateResult, error) {
	s.log.Info(ctx, "business.core.stories.BulkUpdate")
	ctx, span := storyServiceTracer.Start(ctx, "business.core.stories.BulkUpdate")
	defer span.End()

	if err := patch.Validate(); err != nil {
		return BulkUpdateResult{}, err
	}
	if len(storyIDs) == 0 {
		return BulkUpdateResult{}, fmt.Errorf("no story IDs provided")
	}

	span.AddEvent("bulk update started", trace.WithAttributes(
		attribute.Int("story.count", len(storyIDs)),
		attribute.String("workspace.id", workspaceID.String()),
	))
	result := executeBulkStoryUpdates(ctx, storyIDs, func(updateCtx context.Context, storyID uuid.UUID) error {
		return s.UpdatePatch(updateCtx, storyID, workspaceID, patch)
	})

	span.AddEvent("bulk update completed", trace.WithAttributes(
		attribute.Int("stories.updated", result.SucceededCount),
		attribute.Int("stories.failed", result.FailedCount),
	))
	if result.FailedCount > 0 {
		span.RecordError(fmt.Errorf("bulk update completed with %d item failures", result.FailedCount))
	}

	return result, nil
}

type bulkUpdateJob struct {
	index   int
	storyID uuid.UUID
}

func executeBulkStoryUpdates(
	ctx context.Context,
	storyIDs []uuid.UUID,
	update func(context.Context, uuid.UUID) error,
) BulkUpdateResult {
	const maxConcurrentUpdates = 10
	result := BulkUpdateResult{
		TotalCount: len(storyIDs),
		Items:      make([]BulkUpdateItemResult, len(storyIDs)),
	}
	jobs := make(chan bulkUpdateJob)
	workerCount := min(len(storyIDs), maxConcurrentUpdates)
	var wg sync.WaitGroup

	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				item := BulkUpdateItemResult{StoryID: job.storyID, Success: true}
				if err := ctx.Err(); err != nil {
					item.Success = false
					item.Error = err.Error()
				} else if err := update(ctx, job.storyID); err != nil {
					item.Success = false
					item.Error = err.Error()
				}
				result.Items[job.index] = item
			}
		}()
	}

	for index, storyID := range storyIDs {
		jobs <- bulkUpdateJob{index: index, storyID: storyID}
	}
	close(jobs)
	wg.Wait()

	for _, item := range result.Items {
		if item.Success {
			result.SucceededCount++
		} else {
			result.FailedCount++
		}
	}
	result.Partial = result.SucceededCount > 0 && result.FailedCount > 0
	return result
}
