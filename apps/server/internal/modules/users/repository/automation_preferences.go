package usersrepository

import (
	"context"
	"errors"
	"fmt"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) GetAutomationPreferences(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) (usersdomain.AutomationPreferences, error) {
	row, err := r.queries.GetOrCreateAutomationPreferencesForMember(
		ctx,
		usersql.GetOrCreateAutomationPreferencesForMemberParams{
			UserID:      userID,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usersdomain.AutomationPreferences{}, usersdomain.ErrNotFound
		}
		return usersdomain.AutomationPreferences{}, fmt.Errorf("get or create automation preferences: %w", err)
	}
	return mapAutomationPreferences(row), nil
}

func (r *repo) UpdateAutomationPreferences(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	updates usersdomain.UpdateAutomationPreferences,
) error {
	if !hasAutomationPreferenceUpdates(updates) {
		return nil
	}

	params := usersql.UpsertAutomationPreferencesForMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	}
	setOptionalBool(updates.AutoAssignSelf, &params.SetAutoAssignSelf, &params.AutoAssignSelf)
	setOptionalBool(updates.AutoScheduling, &params.SetAutoScheduling, &params.AutoScheduling)
	setOptionalBool(
		updates.AssignSelfOnBranchCopy,
		&params.SetAssignSelfOnBranchCopy,
		&params.AssignSelfOnBranchCopy,
	)
	setOptionalBool(
		updates.MoveStoryToStartedOnBranch,
		&params.SetMoveStoryToStartedOnBranch,
		&params.MoveStoryToStartedOnBranch,
	)
	setOptionalBool(
		updates.OpenStoryInDialog,
		&params.SetOpenStoryInDialog,
		&params.OpenStoryInDialog,
	)

	rows, err := r.queries.UpsertAutomationPreferencesForMember(ctx, params)
	if err != nil {
		return fmt.Errorf("upsert automation preferences: %w", err)
	}
	if rows == 0 {
		return usersdomain.ErrNotFound
	}
	return nil
}

func hasAutomationPreferenceUpdates(updates usersdomain.UpdateAutomationPreferences) bool {
	return updates.AutoAssignSelf != nil ||
		updates.AutoScheduling != nil ||
		updates.AssignSelfOnBranchCopy != nil ||
		updates.MoveStoryToStartedOnBranch != nil ||
		updates.OpenStoryInDialog != nil
}

func setOptionalBool(value *bool, present *bool, destination *bool) {
	if value == nil {
		return
	}
	*present = true
	*destination = *value
}
