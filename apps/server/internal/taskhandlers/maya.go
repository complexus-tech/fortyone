package taskhandlers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/internal/platform/billing"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const mayaAssignmentBatchPageSize = 25

const mayaScheduleRecoveryBatchSize = 100

const mayaWorkspaceAssignmentBatchSize = 25

type mayaAssignmentCandidateStory struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	TeamID      uuid.UUID `db:"team_id"`
	AssigneeID  uuid.UUID `db:"assignee_id"`
}

type mayaAssignmentGroupKey struct {
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
}

func (h *handlers) processWorkspaceScheduleBatch(ctx context.Context, workspaceID uuid.UUID) error {
	stories, err := h.listWorkspaceScheduleCandidates(ctx, workspaceID)
	if err != nil {
		return err
	}

	humanOwnerIDs := make(map[uuid.UUID]struct{})
	mayaAssigned := make([]mayaAssignmentCandidateStory, 0, len(stories))
	for _, story := range stories {
		if story.AssigneeID == h.systemUserID {
			mayaAssigned = append(mayaAssigned, story)
			continue
		}
		humanOwnerIDs[story.AssigneeID] = struct{}{}
	}

	ownerIDs := make([]uuid.UUID, 0, len(humanOwnerIDs))
	for ownerID := range humanOwnerIDs {
		ownerIDs = append(ownerIDs, ownerID)
	}
	sort.Slice(ownerIDs, func(i, j int) bool { return ownerIDs[i].String() < ownerIDs[j].String() })

	var batchErr error
	for _, ownerID := range ownerIDs {
		ownerID := ownerID
		if err := h.mayaService.ReconcileSchedule(ctx, maya.ReconcileScheduleInput{
			WorkspaceID: &workspaceID,
			UserID:      &ownerID,
		}); err != nil {
			batchErr = errors.Join(batchErr, fmt.Errorf("reconcile owner %s: %w", ownerID, err))
		}
	}

	windowStart := time.Now().UTC()
	for key, storyIDs := range groupMayaAssignmentCandidates(mayaAssigned) {
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

	if batchErr != nil {
		return fmt.Errorf("process workspace %s schedule batch: %w", workspaceID, batchErr)
	}
	return nil
}

func (h *handlers) HandleMayaBatchAssignment(ctx context.Context, t *asynq.Task) error {
	h.log.Info(ctx, "HANDLER: Processing MayaBatchAssignment task", "task_id", t.ResultWriter().TaskID())

	if h.mayaService == nil || h.systemUserID == uuid.Nil {
		return fmt.Errorf("Maya batch assignment worker is not configured")
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
		cursor = stories[len(stories)-1].ID

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
		for key, storyIDs := range groups {
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
		return fmt.Errorf("Maya schedule recovery worker is not configured")
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
	query := `
			SELECT
				s.id,
				s.workspace_id,
				s.team_id,
				s.assignee_id
			FROM stories s
			INNER JOIN workspaces w ON w.workspace_id = s.workspace_id
			WHERE s.auto_scheduling_enabled = TRUE
				AND s.assignee_id IS NOT NULL
				AND (
					s.assignee_id = $1
					OR NOT EXISTS (
						SELECT 1
						FROM calendar_maya_schedule_ownerships ownership
						WHERE ownership.workspace_id = s.workspace_id
							AND ownership.story_id = s.id
					)
				)
				AND s.id > $2
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.is_draft = FALSE
			AND ` + billing.WorkspaceMayaAccessSQL("w") + `
			AND NOT EXISTS (
				SELECT 1
				FROM statuses stat
				WHERE stat.status_id = s.status_id
					AND stat.category IN ('completed', 'cancelled')
			)
		ORDER BY s.id
		LIMIT $3
	`

	var stories []mayaAssignmentCandidateStory
	if err := h.db.SelectContext(ctx, &stories, query, h.systemUserID, cursor, limit); err != nil {
		return nil, fmt.Errorf("list Maya assignment candidates: %w", err)
	}
	return stories, nil
}

func (h *handlers) listWorkspaceScheduleCandidates(ctx context.Context, workspaceID uuid.UUID) ([]mayaAssignmentCandidateStory, error) {
	query := `
		SELECT
			s.id,
			s.workspace_id,
			s.team_id,
			s.assignee_id
		FROM stories s
		INNER JOIN workspaces w ON w.workspace_id = s.workspace_id
		WHERE s.workspace_id = $1
			AND s.auto_scheduling_enabled = TRUE
			AND s.assignee_id IS NOT NULL
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.is_draft = FALSE
			AND ` + billing.WorkspaceMayaAccessSQL("w") + `
			AND NOT EXISTS (
				SELECT 1
				FROM statuses stat
				WHERE stat.status_id = s.status_id
					AND stat.category IN ('completed', 'cancelled')
			)
		ORDER BY s.team_id, s.id
	`

	var stories []mayaAssignmentCandidateStory
	if err := h.db.SelectContext(ctx, &stories, query, workspaceID); err != nil {
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
