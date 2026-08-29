package emailreply

import (
	"context"
	"time"

	feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/google/uuid"
)

type objectiveMutationBackend interface {
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		id, workspaceID, userID uuid.UUID,
		expectedUpdatedAt time.Time,
		comment string,
		updates map[string]any,
	) error
}

type keyResultMutationBackend interface {
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		id, workspaceID, userID uuid.UUID,
		expectedUpdatedAt time.Time,
		patch keyresultsdomain.Patch,
		comment string,
	) error
}

type storyMutationBackend interface {
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		actorID, storyID, workspaceID uuid.UUID,
		expectedUpdatedAt time.Time,
		updates map[string]any,
	) error
}

type feedbackMutationBackend interface {
	UpdateItemStatusIfUnchanged(
		ctx context.Context,
		workspaceID, itemID uuid.UUID,
		expectedUpdatedAt time.Time,
		input feedbackdomain.CoreUpdateItemStatusInput,
	) (feedbackdomain.CoreItem, error)
}

type objectiveHealthCommand struct {
	ObjectiveID       uuid.UUID
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	ExpectedUpdatedAt time.Time
	Health            string
	CheckIn           string
}

type keyResultValueCommand struct {
	KeyResultID       uuid.UUID
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	ExpectedUpdatedAt time.Time
	CurrentValue      float64
	CheckIn           string
}

type storyMutationChanges struct {
	DueDateSet  bool
	DueDate     *time.Time
	StatusSet   bool
	StatusID    uuid.UUID
	AssigneeSet bool
	AssigneeID  *uuid.UUID
}

func (changes storyMutationChanges) empty() bool {
	return !changes.DueDateSet && !changes.StatusSet && !changes.AssigneeSet
}

type storyMutationCommand struct {
	StoryID           uuid.UUID
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	ExpectedUpdatedAt time.Time
	Changes           storyMutationChanges
}

type feedbackStatusCommand struct {
	ItemID            uuid.UUID
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	ExpectedUpdatedAt time.Time
	Status            string
}

type objectiveMutationPort interface {
	ApplyObjectiveHealth(context.Context, objectiveHealthCommand) error
}

type keyResultMutationPort interface {
	ApplyKeyResultValue(context.Context, keyResultValueCommand) error
}

type storyMutationPort interface {
	ApplyStoryMutation(context.Context, storyMutationCommand) error
}

type feedbackMutationPort interface {
	ApplyFeedbackStatus(context.Context, feedbackStatusCommand) error
}

type objectiveMutationAdapter struct{ backend objectiveMutationBackend }

func (adapter objectiveMutationAdapter) ApplyObjectiveHealth(ctx context.Context, command objectiveHealthCommand) error {
	return normalizeMutationBackendError(adapter.backend.UpdateExternalUserActionIfUnchanged(
		ctx,
		command.ObjectiveID,
		command.WorkspaceID,
		command.ActorID,
		command.ExpectedUpdatedAt,
		command.CheckIn,
		map[string]any{"health": command.Health},
	))
}

type keyResultMutationAdapter struct{ backend keyResultMutationBackend }

func (adapter keyResultMutationAdapter) ApplyKeyResultValue(ctx context.Context, command keyResultValueCommand) error {
	return normalizeMutationBackendError(adapter.backend.UpdateExternalUserActionIfUnchanged(
		ctx,
		command.KeyResultID,
		command.WorkspaceID,
		command.ActorID,
		command.ExpectedUpdatedAt,
		keyresultsdomain.Patch{CurrentValue: keyresultsdomain.SetField(command.CurrentValue)},
		command.CheckIn,
	))
}

type storyMutationAdapter struct{ backend storyMutationBackend }

func (adapter storyMutationAdapter) ApplyStoryMutation(ctx context.Context, command storyMutationCommand) error {
	updates := make(map[string]any, 3)
	if command.Changes.DueDateSet {
		if command.Changes.DueDate == nil {
			updates["end_date"] = nil
		} else {
			updates["end_date"] = *command.Changes.DueDate
		}
	}
	if command.Changes.StatusSet {
		updates["status_id"] = command.Changes.StatusID
	}
	if command.Changes.AssigneeSet {
		if command.Changes.AssigneeID == nil {
			updates["assignee_id"] = nil
		} else {
			updates["assignee_id"] = *command.Changes.AssigneeID
		}
	}
	return normalizeMutationBackendError(adapter.backend.UpdateExternalUserActionIfUnchanged(
		ctx,
		command.ActorID,
		command.StoryID,
		command.WorkspaceID,
		command.ExpectedUpdatedAt,
		updates,
	))
}

type feedbackMutationAdapter struct{ backend feedbackMutationBackend }

func (adapter feedbackMutationAdapter) ApplyFeedbackStatus(ctx context.Context, command feedbackStatusCommand) error {
	_, err := adapter.backend.UpdateItemStatusIfUnchanged(
		ctx,
		command.WorkspaceID,
		command.ItemID,
		command.ExpectedUpdatedAt,
		feedbackdomain.CoreUpdateItemStatusInput{Status: command.Status, ActorID: command.ActorID},
	)
	return normalizeMutationBackendError(err)
}
