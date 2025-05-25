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

// MentionsRepository provides access to comment mentions storage.
type MentionsRepository interface {
	SaveMentions(ctx context.Context, commentID uuid.UUID, userIDs []uuid.UUID) error
	DeleteMentions(ctx context.Context, commentID uuid.UUID) error
	GetMentions(ctx context.Context, commentID uuid.UUID) ([]uuid.UUID, error)
}

type Service struct {
	repo         Repository
	mentionsRepo MentionsRepository
	log          *logger.Logger
}

func New(log *logger.Logger, repo Repository, mentionsRepo MentionsRepository) *Service {
	return &Service{
		log:          log,
		repo:         repo,
		mentionsRepo: mentionsRepo,
	}
}

func (s *Service) UpdateComment(ctx context.Context, commentID uuid.UUID, comment string, mentions []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.service.comments.UpdateComment")
	defer span.End()

	// Update the comment content
	if err := s.repo.UpdateComment(ctx, commentID, comment); err != nil {
		return err
	}

	if err := s.mentionsRepo.SaveMentions(ctx, commentID, mentions); err != nil {
		s.log.Error(ctx, "failed to save mentions", "error", err, "commentId", commentID)
		// Note: We don't return error here to avoid failing comment update if mentions fail
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
