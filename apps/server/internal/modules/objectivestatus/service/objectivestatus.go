package objectivestatus

import (
	"context"

	objectivestatusdomain "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/domain"
	"github.com/google/uuid"
)

type CoreObjectiveStatus = objectivestatusdomain.Status
type CoreNewObjectiveStatus = objectivestatusdomain.NewStatus
type CoreUpdateObjectiveStatus = objectivestatusdomain.UpdateStatus

var (
	ErrNotFound            = objectivestatusdomain.ErrNotFound
	ErrNoFields            = objectivestatusdomain.ErrNoFields
	ErrStatusHasObjectives = objectivestatusdomain.ErrStatusHasObjectives
	ErrLastInCategory      = objectivestatusdomain.ErrLastInCategory
	ErrInvalidOrder        = objectivestatusdomain.ErrInvalidOrder
)

type Repository interface {
	Create(context.Context, uuid.UUID, uuid.UUID, CoreNewObjectiveStatus) (CoreObjectiveStatus, error)
	Update(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, CoreUpdateObjectiveStatus) (CoreObjectiveStatus, error)
	Delete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	List(context.Context, uuid.UUID) ([]CoreObjectiveStatus, error)
	ListForMember(context.Context, uuid.UUID, uuid.UUID) ([]CoreObjectiveStatus, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) Create(ctx context.Context, actorID, workspaceID uuid.UUID, input CoreNewObjectiveStatus) (CoreObjectiveStatus, error) {
	return service.repo.Create(ctx, actorID, workspaceID, input)
}

func (service *Service) Update(ctx context.Context, actorID, workspaceID, statusID uuid.UUID, input CoreUpdateObjectiveStatus) (CoreObjectiveStatus, error) {
	return service.repo.Update(ctx, actorID, workspaceID, statusID, input)
}

func (service *Service) Delete(ctx context.Context, actorID, workspaceID, statusID uuid.UUID) error {
	return service.repo.Delete(ctx, actorID, workspaceID, statusID)
}

func (service *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]CoreObjectiveStatus, error) {
	return service.repo.List(ctx, workspaceID)
}

func (service *Service) ListForMember(ctx context.Context, actorID, workspaceID uuid.UUID) ([]CoreObjectiveStatus, error) {
	return service.repo.ListForMember(ctx, actorID, workspaceID)
}
