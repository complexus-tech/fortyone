package states

import (
	"context"

	statesdomain "github.com/complexus-tech/projects-api/internal/modules/states/domain"
	"github.com/google/uuid"
)

type CoreState = statesdomain.State
type CoreNewState = statesdomain.NewState
type CoreUpdateState = statesdomain.UpdateState

var (
	ErrNotFound         = statesdomain.ErrNotFound
	ErrNoFields         = statesdomain.ErrNoFields
	ErrStatusHasStories = statesdomain.ErrStatusHasStories
	ErrLastInCategory   = statesdomain.ErrLastInCategory
	ErrInvalidOrder     = statesdomain.ErrInvalidOrder
)

type Repository interface {
	Create(context.Context, uuid.UUID, uuid.UUID, CoreNewState) (CoreState, error)
	Update(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, CoreUpdateState) (CoreState, error)
	Delete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	List(context.Context, uuid.UUID, uuid.UUID) ([]CoreState, error)
	TeamList(context.Context, uuid.UUID, uuid.UUID) ([]CoreState, error)
	TeamListForMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]CoreState, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (CoreState, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) Create(ctx context.Context, actorID, workspaceID uuid.UUID, input CoreNewState) (CoreState, error) {
	return service.repo.Create(ctx, actorID, workspaceID, input)
}

func (service *Service) Update(ctx context.Context, actorID, workspaceID, stateID uuid.UUID, input CoreUpdateState) (CoreState, error) {
	return service.repo.Update(ctx, actorID, workspaceID, stateID, input)
}

func (service *Service) Delete(ctx context.Context, actorID, workspaceID, stateID uuid.UUID) error {
	return service.repo.Delete(ctx, actorID, workspaceID, stateID)
}

func (service *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreState, error) {
	return service.repo.List(ctx, workspaceID, userID)
}

// TeamList is reserved for trusted workflows that already authorized a team.
func (service *Service) TeamList(ctx context.Context, workspaceID, teamID uuid.UUID) ([]CoreState, error) {
	return service.repo.TeamList(ctx, workspaceID, teamID)
}

func (service *Service) TeamListForMember(ctx context.Context, workspaceID, teamID, userID uuid.UUID) ([]CoreState, error) {
	return service.repo.TeamListForMember(ctx, workspaceID, teamID, userID)
}

func (service *Service) Get(ctx context.Context, workspaceID, stateID uuid.UUID) (CoreState, error) {
	return service.repo.Get(ctx, workspaceID, stateID)
}
