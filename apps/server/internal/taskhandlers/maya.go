package taskhandlers

import (
	"context"
	"fmt"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const mayaAssignmentBatchPageSize = 25

type mayaAssignmentCandidateStory struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	TeamID      uuid.UUID `db:"team_id"`
}

type mayaAssignmentGroupKey struct {
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
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

		groups := groupMayaAssignmentCandidates(stories)
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

func (h *handlers) listMayaAssignmentCandidates(ctx context.Context, cursor uuid.UUID, limit int) ([]mayaAssignmentCandidateStory, error) {
	query := `
		SELECT
			s.id,
			s.workspace_id,
			s.team_id
		FROM stories s
		WHERE s.assignee_id = $1
			AND s.id > $2
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.is_draft = FALSE
			AND EXISTS (
				SELECT 1
				FROM workspace_subscriptions ws
				WHERE ws.workspace_id = s.workspace_id
					AND ws.subscription_tier <> 'free'
					AND ws.subscription_status IN ('active', 'trialing', 'past_due')
				LIMIT 1
			)
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
