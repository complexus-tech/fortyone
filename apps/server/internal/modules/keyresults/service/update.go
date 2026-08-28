package keyresults

import (
	"context"
	"errors"
	"slices"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) UpdateIntent(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	patch KeyResultPatch,
	comment string,
) error {
	return service.updateIntent(ctx, id, workspaceID, userID, patch, comment, nil)
}

func (service *Service) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	expectedUpdatedAt time.Time,
	patch KeyResultPatch,
	comment string,
) error {
	if expectedUpdatedAt.IsZero() || patch.Contributors.Set {
		return ErrInvalid
	}
	expected := expectedUpdatedAt.UTC()
	err := service.updateIntent(ctx, id, workspaceID, userID, patch, comment, &expected)
	if !errors.Is(err, ErrVersionConflict) {
		return err
	}
	access, accessErr := service.accessFor(ctx, workspaceID, userID, platformauth.ScopeObjectivesRead)
	if accessErr != nil {
		return accessErr
	}
	current, getErr := service.repo.Get(ctx, keyresultsdomain.GetQuery{
		Access: access, KeyResultID: id,
	})
	if getErr == nil && patchAlreadyApplied(current, patch) {
		return nil
	}
	return ErrVersionConflict
}

func (service *Service) updateIntent(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	patch KeyResultPatch,
	comment string,
	expectedUpdatedAt *time.Time,
) error {
	access, err := service.accessFor(ctx, workspaceID, userID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return err
	}
	command, err := (keyresultsdomain.UpdateCommand{
		Access: access, KeyResultID: id, Patch: patch,
		Comment: comment, ExpectedUpdatedAt: expectedUpdatedAt,
	}).Normalize()
	if err != nil {
		return err
	}
	result, err := service.repo.Update(ctx, command)
	if err != nil {
		return err
	}
	service.publishUpdate(ctx, result, access.ActorID, workspaceID)
	return nil
}

func patchAlreadyApplied(current CoreKeyResult, patch KeyResultPatch) bool {
	if patch.Name.Set && patch.Name.Value != current.Name {
		return false
	}
	if patch.MeasurementType.Set && patch.MeasurementType.Value != current.MeasurementType {
		return false
	}
	if patch.StartValue.Set && patch.StartValue.Value != current.StartValue {
		return false
	}
	if patch.CurrentValue.Set && patch.CurrentValue.Value != current.CurrentValue {
		return false
	}
	if patch.TargetValue.Set && patch.TargetValue.Value != current.TargetValue {
		return false
	}
	if patch.Lead.Set && !sameUUIDPointer(patch.Lead.Value, current.Lead) {
		return false
	}
	if patch.Contributors.Set && !slices.Equal(patch.Contributors.Value, current.Contributors) {
		return false
	}
	if patch.StartDate.Set && !sameTimePointer(patch.StartDate.Value, current.StartDate) {
		return false
	}
	if patch.EndDate.Set && !sameTimePointer(patch.EndDate.Value, current.EndDate) {
		return false
	}
	return true
}

func sameUUIDPointer(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func sameTimePointer(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}
