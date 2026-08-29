package okractivities

import (
	"context"
	"errors"
	"fmt"

	okractivitiesdomain "github.com/complexus-tech/projects-api/internal/modules/okractivities/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

var (
	ErrInvalid       = okractivitiesdomain.ErrInvalid
	ErrForbidden     = okractivitiesdomain.ErrForbidden
	ErrScopeMismatch = okractivitiesdomain.ErrScopeMismatch
)

type Repository interface {
	Create(context.Context, okractivitiesdomain.NewActivity) error
	CreateBatch(context.Context, []okractivitiesdomain.NewActivity) error
	List(context.Context, okractivitiesdomain.ListQuery) ([]okractivitiesdomain.Activity, bool, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) Create(ctx context.Context, activity CoreNewActivity) error {
	return service.CreateBatch(ctx, []CoreNewActivity{activity})
}

func (service *Service) CreateBatch(ctx context.Context, activities []CoreNewActivity) error {
	command, err := (okractivitiesdomain.CreateBatchCommand{Activities: activities}).Normalize()
	if err != nil {
		return err
	}
	if len(command.Activities) == 0 {
		return nil
	}
	for _, activity := range command.Activities {
		actor, err := activityActor(ctx, activity.WorkspaceID, activity.UserID, platformauth.ScopeObjectivesWrite)
		if err != nil {
			return err
		}
		if actor.PrincipalID != activity.UserID {
			return ErrForbidden
		}
	}
	if err := service.repo.CreateBatch(ctx, command.Activities); err != nil {
		return fmt.Errorf("create OKR activities: %w", err)
	}
	return nil
}

func (service *Service) GetObjectiveActivities(
	ctx context.Context,
	objectiveID, workspaceID uuid.UUID,
	page, pageSize int,
) ([]CoreActivity, bool, error) {
	actor, err := activityActor(ctx, workspaceID, uuid.Nil, platformauth.ScopeObjectivesRead)
	if err != nil {
		return nil, false, err
	}
	query, err := (okractivitiesdomain.ListQuery{
		ObjectiveID: objectiveID, WorkspaceID: workspaceID, ActorID: actor.PrincipalID,
		Page: page, PageSize: pageSize,
	}).Normalize()
	if err != nil {
		return nil, false, err
	}
	activities, hasMore, err := service.repo.List(ctx, query)
	if err != nil {
		return nil, false, fmt.Errorf("get objective activities: %w", err)
	}
	return activities, hasMore, nil
}

func (service *Service) GetKeyResultActivities(
	ctx context.Context,
	keyResultID, workspaceID uuid.UUID,
	page, pageSize int,
) ([]CoreActivity, bool, error) {
	actor, err := activityActor(ctx, workspaceID, uuid.Nil, platformauth.ScopeObjectivesRead)
	if err != nil {
		return nil, false, err
	}
	query, err := (okractivitiesdomain.ListQuery{
		KeyResultID: &keyResultID, WorkspaceID: workspaceID, ActorID: actor.PrincipalID,
		Page: page, PageSize: pageSize,
	}).Normalize()
	if err != nil {
		return nil, false, err
	}
	activities, hasMore, err := service.repo.List(ctx, query)
	if err != nil {
		return nil, false, fmt.Errorf("get key result activities: %w", err)
	}
	return activities, hasMore, nil
}

func activityActor(
	ctx context.Context,
	workspaceID, fallbackUserID uuid.UUID,
	requiredScope platformauth.Scope,
) (platformauth.Actor, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		if !errors.Is(err, platformauth.ErrActorNotFound) || fallbackUserID == uuid.Nil {
			return platformauth.Actor{}, ErrForbidden
		}
		actor = platformauth.NewHumanActor(fallbackUserID)
	}
	if actor.WorkspaceID == uuid.Nil {
		actor, err = actor.WithWorkspace(workspaceID)
		if err != nil {
			return platformauth.Actor{}, ErrForbidden
		}
	}
	if actor.WorkspaceID != workspaceID || !actor.IsUserActor() || !actor.Scopes.Has(requiredScope) {
		return platformauth.Actor{}, ErrForbidden
	}
	if fallbackUserID != uuid.Nil && fallbackUserID != actor.PrincipalID {
		return platformauth.Actor{}, ErrForbidden
	}
	if !actor.TeamAccess.IsUnrestricted() {
		// This legacy activity API does not carry a credential team set into its
		// SQL contract. Fail closed rather than silently widening a restricted key.
		return platformauth.Actor{}, ErrForbidden
	}
	return actor, nil
}
