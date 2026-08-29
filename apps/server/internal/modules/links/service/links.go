package links

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
)

type Repository interface {
	CreateLink(ctx context.Context, actorID uuid.UUID, cnl CoreNewLink) (CoreLink, error)
	UpdateLink(ctx context.Context, actorID, linkID, workspaceID uuid.UUID, cul CoreUpdateLink) error
	DeleteLink(ctx context.Context, actorID, linkID, workspaceID uuid.UUID) error
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

func (s *Service) CreateLink(ctx context.Context, actorID uuid.UUID, cnl CoreNewLink) (CoreLink, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.links.CreateLink")
	defer span.End()

	link, err := s.repo.CreateLink(ctx, actorID, cnl)
	if err != nil {
		return CoreLink{}, err
	}

	return link, nil
}

func (s *Service) UpdateLink(ctx context.Context, actorID, linkID, workspaceID uuid.UUID, cul CoreUpdateLink) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.links.UpdateLink")
	defer span.End()

	err := s.repo.UpdateLink(ctx, actorID, linkID, workspaceID, cul)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteLink(ctx context.Context, actorID, linkID, workspaceID uuid.UUID) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.links.DeleteLink")
	defer span.End()

	err := s.repo.DeleteLink(ctx, actorID, linkID, workspaceID)
	if err != nil {
		return err
	}

	return nil
}
