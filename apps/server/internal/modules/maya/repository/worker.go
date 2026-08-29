package mayarepository

import (
	"context"
	"fmt"
	"math"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func (r *Repo) ListMayaAssignmentCandidates(
	ctx context.Context,
	mayaActorID, afterStoryID uuid.UUID,
	limit int,
) ([]mayadomain.AssignmentCandidateStory, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rowLimit, err := validatedWorkerLimit(limit, 500)
	if err != nil {
		return nil, fmt.Errorf("validate Maya assignment candidate limit: %w", err)
	}
	rows, err := r.queries.ListMayaAssignmentCandidates(ctx, mayasql.ListMayaAssignmentCandidatesParams{
		MayaActorID:  &mayaActorID,
		AfterStoryID: afterStoryID,
		RowLimit:     rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list Maya assignment candidates: %w", err)
	}
	return mapAssignmentCandidates(rows)
}

func (r *Repo) ListWorkspaceScheduleCandidates(
	ctx context.Context,
	workspaceID, afterStoryID uuid.UUID,
	limit int,
) ([]mayadomain.AssignmentCandidateStory, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rowLimit, err := validatedWorkerLimit(limit, 500)
	if err != nil {
		return nil, fmt.Errorf("validate workspace schedule candidate limit: %w", err)
	}
	rows, err := r.queries.ListWorkspaceScheduleCandidates(ctx, mayasql.ListWorkspaceScheduleCandidatesParams{
		WorkspaceID:  workspaceID,
		AfterStoryID: afterStoryID,
		RowLimit:     rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list workspace schedule candidates: %w", err)
	}
	return mapWorkspaceScheduleCandidates(rows)
}

func (r *Repo) ListMayaWorkFocusCandidates(
	ctx context.Context,
	limit int,
) ([]mayadomain.WorkFocusMember, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rowLimit, err := validatedWorkerLimit(limit, 500)
	if err != nil {
		return nil, fmt.Errorf("validate Maya work-focus candidate limit: %w", err)
	}
	rows, err := r.queries.ListMayaWorkFocusCandidates(ctx, mayasql.ListMayaWorkFocusCandidatesParams{
		RowLimit: rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list Maya work-focus candidates: %w", err)
	}
	members := make([]mayadomain.WorkFocusMember, len(rows))
	for index, row := range rows {
		members[index] = mayadomain.WorkFocusMember{
			WorkspaceID:           row.WorkspaceID,
			TeamID:                row.TeamID,
			UserID:                row.UserID,
			ManualRoleTitle:       row.AiRoleTitle,
			ManualRoleDescription: row.AiRoleDescription,
		}
	}
	return members, nil
}

func (r *Repo) ListMayaWorkFocusEvidence(
	ctx context.Context,
	member mayadomain.WorkFocusMember,
	updatedAfter time.Time,
	limit int,
) ([]mayadomain.WorkFocusEvidence, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rowLimit, err := validatedWorkerLimit(limit, 100)
	if err != nil {
		return nil, fmt.Errorf("validate Maya work-focus evidence limit: %w", err)
	}
	rows, err := r.queries.ListMayaWorkFocusEvidence(ctx, mayasql.ListMayaWorkFocusEvidenceParams{
		WorkspaceID:  member.WorkspaceID,
		TeamID:       member.TeamID,
		UserID:       &member.UserID,
		UpdatedAfter: updatedAfter.UTC(),
		RowLimit:     rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list Maya work-focus evidence: %w", err)
	}

	evidence := make([]mayadomain.WorkFocusEvidence, 0, len(rows))
	storyIndexes := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		index, exists := storyIndexes[row.StoryID]
		if !exists {
			description := ""
			if row.Description != nil {
				description = *row.Description
			}
			index = len(evidence)
			storyIndexes[row.StoryID] = index
			evidence = append(evidence, mayadomain.WorkFocusEvidence{
				Title:       row.Title,
				Description: description,
			})
		}
		if row.Label != nil {
			evidence[index].Labels = append(evidence[index].Labels, *row.Label)
		}
	}
	return evidence, nil
}

func (r *Repo) SaveMayaInferredWorkFocus(
	ctx context.Context,
	member mayadomain.WorkFocusMember,
	result mayadomain.WorkFocusInferenceResult,
) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	storyCount, err := safecast.Int32(result.StoryCount)
	if err != nil {
		return false, fmt.Errorf("validate Maya work-focus story count: %w", err)
	}
	if math.IsNaN(result.Confidence) || math.IsInf(result.Confidence, 0) || result.Confidence < 0 || result.Confidence > 1 {
		return false, fmt.Errorf("validate Maya work-focus confidence: must be between zero and one")
	}
	rows, err := r.queries.SaveMayaInferredWorkFocus(ctx, mayasql.SaveMayaInferredWorkFocusParams{
		RoleTitle:       result.RoleTitle,
		RoleDescription: result.RoleDescription,
		StoryCount:      storyCount,
		Confidence:      float32(result.Confidence),
		TeamID:          member.TeamID,
		UserID:          member.UserID,
		WorkspaceID:     member.WorkspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("save Maya inferred work focus: %w", err)
	}
	if rows > 1 {
		return false, fmt.Errorf("save Maya inferred work focus: updated %d memberships, want at most one", rows)
	}
	return rows == 1, nil
}

func validatedWorkerLimit(limit, maximum int) (int32, error) {
	if limit <= 0 || limit > maximum {
		return 0, fmt.Errorf("limit must be between 1 and %d", maximum)
	}
	return safecast.Int32(limit)
}

func mapAssignmentCandidates(
	rows []mayasql.ListMayaAssignmentCandidatesRow,
) ([]mayadomain.AssignmentCandidateStory, error) {
	candidates := make([]mayadomain.AssignmentCandidateStory, len(rows))
	for index, row := range rows {
		if row.AssigneeID == nil {
			return nil, fmt.Errorf("maya assignment candidate %s has no assignee", row.ID)
		}
		candidates[index] = mayadomain.AssignmentCandidateStory{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			TeamID:      row.TeamID,
			AssigneeID:  *row.AssigneeID,
		}
	}
	return candidates, nil
}

func mapWorkspaceScheduleCandidates(
	rows []mayasql.ListWorkspaceScheduleCandidatesRow,
) ([]mayadomain.AssignmentCandidateStory, error) {
	candidates := make([]mayadomain.AssignmentCandidateStory, len(rows))
	for index, row := range rows {
		if row.AssigneeID == nil {
			return nil, fmt.Errorf("workspace schedule candidate %s has no assignee", row.ID)
		}
		candidates[index] = mayadomain.AssignmentCandidateStory{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			TeamID:      row.TeamID,
			AssigneeID:  *row.AssigneeID,
		}
	}
	return candidates, nil
}
