package comments

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Repository interface {
	UpdateComment(ctx context.Context, commentID uuid.UUID, comment string) error
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
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

func (s *Service) UpdateComment(ctx context.Context, commentID uuid.UUID, comment string) error {
	ctx, span := web.AddSpan(ctx, "business.service.comments.UpdateComment")
	defer span.End()

	err := s.repo.UpdateComment(ctx, commentID, comment)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.service.comments.DeleteComment")
	defer span.End()

	err := s.repo.DeleteComment(ctx, commentID)
	if err != nil {
		return err
	}

	return nil
}
