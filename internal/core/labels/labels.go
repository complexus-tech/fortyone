package labels

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Repository interface {
	GetLabels(ctx context.Context, workspaceID uuid.UUID, filters map[string]any) ([]CoreLabel, error)
	CreateLabel(ctx context.Context, input CoreNewLabel) (CoreLabel, error)
	GetLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) (CoreLabel, error)
	UpdateLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID, name string, color string) (CoreLabel, error)
	DeleteLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) error
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
	s.log.Info(ctx, "Getting labels for workspace", "workspaceID", workspaceID)
	ctx, span := web.AddSpan(ctx, "business.service.labels.GetLabels")
	defer span.End()

	labels, err := s.repo.GetLabels(ctx, workspaceID, filters)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (s *Service) CreateLabel(ctx context.Context, cnl CoreNewLabel) (CoreLabel, error) {
	s.log.Info(ctx, "Creating label", "name", cnl.Name, "color", cnl.Color)
	ctx, span := web.AddSpan(ctx, "business.service.labels.CreateLabel")
	defer span.End()

	label, err := s.repo.CreateLabel(ctx, cnl)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) GetLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) (CoreLabel, error) {
	s.log.Info(ctx, "Getting label", "labelID", labelID, "workspaceID", workspaceID)
	ctx, span := web.AddSpan(ctx, "business.service.labels.GetLabel")
	defer span.End()

	label, err := s.repo.GetLabel(ctx, labelID, workspaceID)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) UpdateLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID, name string, color string) (CoreLabel, error) {
	s.log.Info(ctx, "Updating label", "labelID", labelID, "workspaceID", workspaceID, "name", name, "color", color)
	ctx, span := web.AddSpan(ctx, "business.service.labels.UpdateLabel")
	defer span.End()

	label, err := s.repo.UpdateLabel(ctx, labelID, workspaceID, name, color)
	if err != nil {
		return CoreLabel{}, err
	}

	return label, nil
}

func (s *Service) DeleteLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "Deleting label", "labelID", labelID, "workspaceID", workspaceID)
	ctx, span := web.AddSpan(ctx, "business.service.labels.DeleteLabel")
	defer span.End()

	err := s.repo.DeleteLabel(ctx, labelID, workspaceID)
	if err != nil {
		return err
	}

	return nil
}
