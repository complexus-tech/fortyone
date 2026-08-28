package taskhandlers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	mayaAssignmentBatchPageSize      = 25
	mayaScheduleRecoveryBatchSize    = 100
	mayaWorkspaceSchedulePageSize    = 100
	mayaWorkspaceAssignmentBatchSize = 25
)

type mayaAssignmentCandidateStory = mayadomain.AssignmentCandidateStory

type mayaAssignmentGroupKey struct {
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
}

func (h *handlers) processWorkspaceScheduleBatch(ctx context.Context, workspaceID uuid.UUID) error {
	if h.mayaAssignments == nil {
		return fmt.Errorf("maya assignment store is not configured")
	}

	humanOwnerIDs := make(map[uuid.UUID]struct{})
	var batchErr error
	windowStart := time.Now().UTC()
	var cursor uuid.UUID
	for {
		stories, err := h.listWorkspaceScheduleCandidates(
			ctx,
			workspaceID,
			cursor,
			mayaWorkspaceSchedulePageSize,
		)
		if err != nil {
			return err
		}
		if len(stories) == 0 {
			break
		}

		nextCursor := stories[len(stories)-1].ID
		if nextCursor == uuid.Nil || nextCursor == cursor {
			return fmt.Errorf("list workspace schedule candidates: cursor did not advance")
		}

		mayaAssigned := make([]mayaAssignmentCandidateStory, 0, len(stories))
		for _, story := range stories {
			if story.AssigneeID == h.systemUserID {
				mayaAssigned = append(mayaAssigned, story)
				continue
			}
			if _, alreadyReconciled := humanOwnerIDs[story.AssigneeID]; alreadyReconciled {
				continue
			}
			humanOwnerIDs[story.AssigneeID] = struct{}{}
			ownerID := story.AssigneeID
			if err := h.mayaService.ReconcileSchedule(ctx, maya.ReconcileScheduleInput{
				WorkspaceID: &workspaceID,
				UserID:      &ownerID,
			}); err != nil {
				batchErr = errors.Join(batchErr, fmt.Errorf("reconcile owner %s: %w", ownerID, err))
			}
		}

		groups := groupMayaAssignmentCandidates(mayaAssigned)
		for _, key := range sortedMayaAssignmentGroupKeys(groups) {
			storyIDs := groups[key]
			for start := 0; start < len(storyIDs); start += mayaWorkspaceAssignmentBatchSize {
				end := min(start+mayaWorkspaceAssignmentBatchSize, len(storyIDs))
				if _, err := h.mayaService.ProcessAssignmentBatch(ctx, maya.ProcessAssignmentBatchInput{
					WorkspaceID: key.WorkspaceID,
					TeamID:      key.TeamID,
					StoryIDs:    storyIDs[start:end],
					TriggeredBy: h.systemUserID,
					WindowStart: windowStart,
					WindowEnd:   windowStart.Add(14 * 24 * time.Hour),
					AutoApply:   true,
				}); err != nil {
					batchErr = errors.Join(batchErr, fmt.Errorf("process Maya team %s assignment batch: %w", key.TeamID, err))
				}
			}
		}

		cursor = nextCursor
		if len(stories) < mayaWorkspaceSchedulePageSize {
			break
		}
	}

	if batchErr != nil {
		return fmt.Errorf("process workspace %s schedule batch: %w", workspaceID, batchErr)
	}
	return nil
}

func (h *handlers) HandleMayaBatchAssignment(ctx context.Context, t *asynq.Task) error {
	h.log.Info(ctx, "HANDLER: Processing MayaBatchAssignment task", "task_id", t.ResultWriter().TaskID())

	if h.mayaService == nil || h.mayaAssignments == nil || h.systemUserID == uuid.Nil {
		return fmt.Errorf("maya batch assignment worker is not configured")
	}

	var cursor uuid.UUID
	totalProcessed := 0
	totalSkipped := 0
	for {
		stories, err := h.listMayaAssignmentCandidates(ctx, cursor, mayaAssignmentBatchPageSize)
		if err != nil {
			return err
		}
		if len(stories) == 0 {
			break
		}
		nextCursor := stories[len(stories)-1].ID
		if nextCursor == uuid.Nil || nextCursor == cursor {
			return fmt.Errorf("list Maya assignment candidates: cursor did not advance")
		}
		cursor = nextCursor

		mayaAssigned := make([]mayaAssignmentCandidateStory, 0, len(stories))
		for _, story := range stories {
			if story.AssigneeID == h.systemUserID {
				mayaAssigned = append(mayaAssigned, story)
				continue
			}
			workspaceID := story.WorkspaceID
			storyID := story.ID
			if err := h.mayaService.ReconcileSchedule(ctx, maya.ReconcileScheduleInput{
				WorkspaceID: &workspaceID,
				StoryID:     &storyID,
			}); err != nil {
				h.log.Error(ctx, "failed to recover enabled human-owned Maya schedule", "workspace_id", story.WorkspaceID, "story_id", story.ID, "error", err)
				totalSkipped++
				continue
			}
			totalProcessed++
		}

		groups := groupMayaAssignmentCandidates(mayaAssigned)
		windowStart := time.Now().UTC()
		for _, key := range sortedMayaAssignmentGroupKeys(groups) {
			storyIDs := groups[key]
			result, err := h.mayaService.ProcessAssignmentBatch(ctx, maya.ProcessAssignmentBatchInput{
				WorkspaceID: key.WorkspaceID,
				TeamID:      key.TeamID,
				StoryIDs:    storyIDs,
				TriggeredBy: h.systemUserID,
				WindowStart: windowStart,
				WindowEnd:   windowStart.Add(14 * 24 * time.Hour),
				AutoApply:   true,
			})
			if err != nil {
				h.log.Error(ctx, "failed to process Maya assignment batch", "workspace_id", key.WorkspaceID, "team_id", key.TeamID, "error", err)
				totalSkipped += len(storyIDs)
				continue
			}
			totalProcessed += result.Processed
			totalSkipped += result.Skipped
		}
		if len(stories) < mayaAssignmentBatchPageSize {
			break
		}
	}

	h.log.Info(ctx, "HANDLER: Successfully processed MayaBatchAssignment task", "task_id", t.ResultWriter().TaskID(), "processed", totalProcessed, "skipped", totalSkipped)
	return nil
}

func (h *handlers) HandleMayaScheduleRecovery(ctx context.Context, t *asynq.Task) error {
	if h.mayaService == nil {
		return fmt.Errorf("maya schedule recovery worker is not configured")
	}
	processed, err := h.mayaService.RecoverScheduleOwnerships(ctx, mayaScheduleRecoveryBatchSize)
	if err != nil {
		return fmt.Errorf("recover Maya-owned schedules: %w", err)
	}
	if h.log != nil {
		h.log.Info(ctx, "Recovered Maya-owned schedules", "task_id", t.ResultWriter().TaskID(), "processed", processed)
	}
	return nil
}

func (h *handlers) listMayaAssignmentCandidates(ctx context.Context, cursor uuid.UUID, limit int) ([]mayaAssignmentCandidateStory, error) {
	if h.mayaAssignments == nil {
		return nil, fmt.Errorf("list Maya assignment candidates: assignment store is not configured")
	}
	stories, err := h.mayaAssignments.ListMayaAssignmentCandidates(ctx, h.systemUserID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("list Maya assignment candidates: %w", err)
	}
	return stories, nil
}

func (h *handlers) listWorkspaceScheduleCandidates(
	ctx context.Context,
	workspaceID uuid.UUID,
	cursor uuid.UUID,
	limit int,
) ([]mayaAssignmentCandidateStory, error) {
	if h.mayaAssignments == nil {
		return nil, fmt.Errorf("list workspace schedule candidates: assignment store is not configured")
	}
	stories, err := h.mayaAssignments.ListWorkspaceScheduleCandidates(ctx, workspaceID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("list workspace schedule candidates: %w", err)
	}
	return stories, nil
}

func groupMayaAssignmentCandidates(stories []mayaAssignmentCandidateStory) map[mayaAssignmentGroupKey][]uuid.UUID {
	groups := make(map[mayaAssignmentGroupKey][]uuid.UUID)
	for _, story := range stories {
		key := mayaAssignmentGroupKey{
			WorkspaceID: story.WorkspaceID,
			TeamID:      story.TeamID,
		}
		groups[key] = append(groups[key], story.ID)
	}
	return groups
}

func sortedMayaAssignmentGroupKeys(groups map[mayaAssignmentGroupKey][]uuid.UUID) []mayaAssignmentGroupKey {
	keys := make([]mayaAssignmentGroupKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].WorkspaceID != keys[j].WorkspaceID {
			return keys[i].WorkspaceID.String() < keys[j].WorkspaceID.String()
		}
		return keys[i].TeamID.String() < keys[j].TeamID.String()
	})
	return keys
}
