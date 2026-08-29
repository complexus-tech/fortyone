package objectives

import (
	"context"
	"errors"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) Create(
	ctx context.Context,
	newObjective CoreNewObjective,
	workspaceID uuid.UUID,
	keyResults []CoreNewKeyResult,
) (CoreObjective, []CoreKeyResult, error) {
	actor, err := actorFor(ctx, workspaceID, newObjective.CreatedBy, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return CoreObjective{}, nil, err
	}
	newObjective.CreatedBy = actor.PrincipalID
	if newObjective.Color == "" {
		newObjective.Color = DefaultObjectiveColor
	}
	result, err := service.repo.Create(ctx, objectivesdomain.CreateCommand{
		WorkspaceID: workspaceID, Objective: newObjective, KeyResults: keyResults,
	})
	if err != nil {
		return CoreObjective{}, nil, err
	}
	return result.Objective, result.KeyResults, nil
}

func (service *Service) Update(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	comment string,
	updates map[string]any,
) error {
	patch, err := objectivePatchFromCompatibilityMap(updates)
	if err != nil {
		return err
	}
	return service.UpdateIntent(ctx, id, workspaceID, userID, comment, patch)
}

func (service *Service) UpdateIntent(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	comment string,
	patch objectivesdomain.ObjectivePatch,
) error {
	return service.UpdateIntentIfUnchanged(ctx, id, workspaceID, userID, comment, nil, patch)
}

// UpdateIntentIfUnchanged applies a typed objective patch. When an expected
// timestamp is supplied, the repository rejects a stale edit after locking the
// objective row, preventing a browser or integration from overwriting a newer
// change. A nil timestamp preserves the historical browser API behaviour.
func (service *Service) UpdateIntentIfUnchanged(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	comment string,
	expectedUpdatedAt *time.Time,
	patch objectivesdomain.ObjectivePatch,
) error {
	if expectedUpdatedAt != nil {
		expected := expectedUpdatedAt.UTC()
		expectedUpdatedAt = &expected
	}
	_, err := service.update(ctx, objectivesdomain.UpdateCommand{
		ObjectiveID: id, WorkspaceID: workspaceID, ActorID: userID,
		Comment: comment, Patch: patch, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	return err
}

func (service *Service) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	expectedUpdatedAt time.Time,
	comment string,
	updates map[string]any,
) error {
	if expectedUpdatedAt.IsZero() {
		return ErrInvalid
	}
	patch, err := objectivePatchFromCompatibilityMap(updates)
	if err != nil {
		return err
	}
	expected := expectedUpdatedAt.UTC()
	_, err = service.update(ctx, objectivesdomain.UpdateCommand{
		ObjectiveID: id, WorkspaceID: workspaceID, ActorID: userID,
		Comment: comment, Patch: patch, ExpectedUpdatedAt: &expected,
	})
	if !errors.Is(err, ErrVersionConflict) {
		return err
	}
	current, getErr := service.repo.Get(ctx, objectivesdomain.GetQuery{
		ObjectiveID: id, WorkspaceID: workspaceID, ActorID: userID,
	})
	if getErr == nil && objectivePatchAlreadyApplied(current, patch) {
		return nil
	}
	return ErrVersionConflict
}

func (service *Service) update(
	ctx context.Context,
	command objectivesdomain.UpdateCommand,
) (CoreObjective, error) {
	actor, err := actorFor(ctx, command.WorkspaceID, command.ActorID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return CoreObjective{}, err
	}
	command.ActorID = actor.PrincipalID
	updated, err := service.repo.Update(ctx, command)
	if err != nil {
		return CoreObjective{}, err
	}
	service.publishUpdate(ctx, updated, actor.PrincipalID, command.Patch)
	return updated, nil
}

func (service *Service) Delete(ctx context.Context, id, workspaceID uuid.UUID) error {
	actor, err := actorFor(ctx, workspaceID, uuid.Nil, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return err
	}
	return service.repo.Delete(ctx, objectivesdomain.DeleteCommand{
		ObjectiveID: id, WorkspaceID: workspaceID, ActorID: actor.PrincipalID,
	})
}
