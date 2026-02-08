package links

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Repository interface {
	CreateLink(ctx context.Context, cnl CoreNewLink) (CoreLink, error)
	UpdateLink(ctx context.Context, linkID uuid.UUID, cul CoreUpdateLink) error
	DeleteLink(ctx context.Context, linkID uuid.UUID) error
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

func (s *Service) CreateLink(ctx context.Context, cnl CoreNewLink) (CoreLink, error) {
	ctx, span := web.AddSpan(ctx, "business.service.links.CreateLink")
	defer span.End()

	link, err := s.repo.CreateLink(ctx, cnl)
	if err != nil {
		return CoreLink{}, err
	}

	return link, nil
}

func (s *Service) UpdateLink(ctx context.Context, linkID uuid.UUID, cul CoreUpdateLink) error {
	ctx, span := web.AddSpan(ctx, "business.service.links.UpdateLink")
	defer span.End()

	err := s.repo.UpdateLink(ctx, linkID, cul)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteLink(ctx context.Context, linkID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.service.links.DeleteLink")
	defer span.End()

	err := s.repo.DeleteLink(ctx, linkID)
	if err != nil {
		return err
	}

	return nil
}
