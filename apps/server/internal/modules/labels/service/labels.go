package labels

import (
	"context"

	labelsdomain "github.com/complexus-tech/projects-api/internal/modules/labels/domain"
	"github.com/google/uuid"
)

type CoreLabel = labelsdomain.Label
type CoreNewLabel = labelsdomain.NewLabel
type LabelFilters = labelsdomain.Filters

var (
	ErrNotFound          = labelsdomain.ErrNotFound
	ErrInvalidPagination = labelsdomain.ErrInvalidPagination
)

type Repository interface {
	GetLabels(ctx context.Context, actorID, workspaceID uuid.UUID, filters LabelFilters) ([]CoreLabel, error)
	CreateLabel(ctx context.Context, actorID uuid.UUID, input CoreNewLabel) (CoreLabel, error)
	GetLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID) (CoreLabel, error)
	UpdateLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID, name string, color string) (CoreLabel, error)
	DeleteLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetLabels(ctx context.Context, actorID, workspaceID uuid.UUID, filters LabelFilters) ([]CoreLabel, error) {
	labels, err := s.repo.GetLabels(ctx, actorID, workspaceID, filters)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (s *Service) CreateLabel(ctx context.Context, actorID uuid.UUID, cnl CoreNewLabel) (CoreLabel, error) {
	label, err := s.repo.CreateLabel(ctx, actorID, cnl)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) GetLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID) (CoreLabel, error) {
	label, err := s.repo.GetLabel(ctx, actorID, labelID, workspaceID)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) UpdateLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID, name string, color string) (CoreLabel, error) {
	label, err := s.repo.UpdateLabel(ctx, actorID, labelID, workspaceID, name, color)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) DeleteLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID) error {
	err := s.repo.DeleteLabel(ctx, actorID, labelID, workspaceID)
	if err != nil {
		return err
	}

	return nil
}
