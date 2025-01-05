package labels

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type Repository interface {
	GetLabels(ctx context.Context, workspaceID uuid.UUID, filters map[string]any) ([]CoreLabel, error)
	CreateLabel(ctx context.Context, input CoreNewLabel) (CoreLabel, error)
	GetLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) (CoreLabel, error)
}

type Service struct {
	repo Repository
	log  *logger.Logger
}

func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		log:  log,
		repo: repo,
	}
}

func (s *Service) GetLabels(ctx context.Context, workspaceID uuid.UUID, filters map[string]any) ([]CoreLabel, error) {
	labels, err := s.repo.GetLabels(ctx, workspaceID, filters)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (s *Service) CreateLabel(ctx context.Context, cnl CoreNewLabel) (CoreLabel, error) {

	label, err := s.repo.CreateLabel(ctx, cnl)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) GetLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) (CoreLabel, error) {
	label, err := s.repo.GetLabel(ctx, labelID, workspaceID)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}
