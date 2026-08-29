package admin

import (
	"context"
	"strings"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
)

func (service *Service) UpdateWorkspaceTrial(ctx context.Context, actorID, workspaceID uuid.UUID, input UpdateWorkspaceTrialInput) (WorkspaceOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.UpdateWorkspaceTrial")
	defer span.End()

	reason, err := admindomain.RequireReason(input.Reason)
	if err != nil {
		return WorkspaceOverview{}, err
	}
	now := service.clock.Now().UTC()
	trialEndsOn := input.TrialEndsOn.UTC()
	if !trialEndsOn.After(now) {
		return WorkspaceOverview{}, ErrInvalidTrialEndsOn
	}

	overview, err := service.repo.UpdateWorkspaceTrial(ctx, admindomain.UpdateWorkspaceTrialCommand{
		ActorID: actorID, WorkspaceID: workspaceID, TrialEndsOn: trialEndsOn,
		Reason: reason, Now: now,
	})
	if err != nil {
		return WorkspaceOverview{}, err
	}
	service.resolveWorkspaceLogo(ctx, &overview.Workspace)
	return overview, nil
}

func (service *Service) UpdateWorkspaceDeleted(ctx context.Context, actorID, workspaceID uuid.UUID, input UpdateWorkspaceDeletedInput) (WorkspaceOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.UpdateWorkspaceDeleted")
	defer span.End()

	reason, err := admindomain.RequireReason(input.Reason)
	if err != nil {
		return WorkspaceOverview{}, err
	}
	overview, err := service.repo.SetWorkspaceDeleted(ctx, admindomain.SetWorkspaceDeletedCommand{
		ActorID: actorID, WorkspaceID: workspaceID, Deleted: input.Deleted,
		Reason: reason, Now: service.clock.Now().UTC(),
	})
	if err != nil {
		return WorkspaceOverview{}, err
	}
	service.resolveWorkspaceLogo(ctx, &overview.Workspace)
	return overview, nil
}

func (service *Service) UpdateUserState(ctx context.Context, actorID, userID uuid.UUID, input UpdateUserStateInput) (UserOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.UpdateUserState")
	defer span.End()

	if input.Patch.Empty() {
		return UserOverview{}, ErrInvalidAdminAction
	}
	if err := validateBooleanPatch(input.Patch); err != nil {
		return UserOverview{}, err
	}
	reason, err := admindomain.RequireReason(input.Reason)
	if err != nil {
		return UserOverview{}, err
	}
	overview, err := service.repo.UpdateUserState(ctx, admindomain.UpdateUserStateCommand{
		ActorID: actorID, UserID: userID, Patch: input.Patch,
		Reason: reason, Now: service.clock.Now().UTC(),
	})
	if err != nil {
		return UserOverview{}, err
	}
	service.resolveUserAvatar(ctx, &overview.User)
	return overview, nil
}

func (service *Service) RequestUserSessionRevocation(ctx context.Context, actorID, userID uuid.UUID, input RequestUserSessionRevocationInput) (UserOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.RequestUserSessionRevocation")
	defer span.End()

	reason, err := admindomain.RequireReason(input.Reason)
	if err != nil {
		return UserOverview{}, err
	}
	overview, err := service.repo.RequestSessionRevocation(ctx, admindomain.RequestSessionRevocationCommand{
		ActorID: actorID, UserID: userID, Reason: reason,
	})
	if err != nil {
		return UserOverview{}, err
	}
	service.resolveUserAvatar(ctx, &overview.User)
	return overview, nil
}

func (service *Service) CreateAdminNote(ctx context.Context, actorID uuid.UUID, input CreateAdminNoteInput) (AdminNote, error) {
	ctx, span := adminTracer.Start(ctx, "admin.CreateAdminNote")
	defer span.End()

	body := strings.TrimSpace(input.Body)
	if body == "" {
		return AdminNote{}, ErrInvalidAdminNote
	}
	if input.TargetID == uuid.Nil {
		return AdminNote{}, ErrInvalidAdminAction
	}
	targetType, err := admindomain.ParseNoteTargetType(input.TargetType)
	if err != nil {
		return AdminNote{}, err
	}
	workspaceID := input.WorkspaceID
	if targetType == admindomain.TargetWorkspace {
		workspaceID = &input.TargetID
	}
	return service.repo.CreateAdminNote(ctx, admindomain.CreateAdminNoteCommand{
		ActorID: actorID, TargetType: targetType, TargetID: input.TargetID,
		WorkspaceID: workspaceID, Body: body,
	})
}

func validateBooleanPatch(patch admindomain.UserStatePatch) error {
	if value, specified := patch.IsActive.Value(); specified && value == nil {
		return ErrInvalidAdminAction
	}
	if value, specified := patch.IsInternal.Value(); specified && value == nil {
		return ErrInvalidAdminAction
	}
	return nil
}
