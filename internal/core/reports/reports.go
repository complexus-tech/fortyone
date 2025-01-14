package reports

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// Repository provides access to the reports storage.
type Repository interface {
	GetStoryStats(ctx context.Context, workspaceID uuid.UUID) (CoreStoryStats, error)
	GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, days int) ([]CoreContributionStats, error)
	GetUserStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (CoreUserStats, error)
}

// Service manages the reports operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a reports service for api access.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// GetStoryStats retrieves story statistics for a workspace.
func (s *Service) GetStoryStats(ctx context.Context, workspaceID uuid.UUID) (CoreStoryStats, error) {
	s.log.Info(ctx, "getting story stats")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetStoryStats")
	defer span.End()

	stats, err := s.repo.GetStoryStats(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreStoryStats{}, fmt.Errorf("getting story stats: %w", err)
	}

	return stats, nil
}

// GetContributionStats retrieves contribution statistics for a user.
func (s *Service) GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, days int) ([]CoreContributionStats, error) {
	s.log.Info(ctx, "getting contribution stats")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetContributionStats")
	defer span.End()

	stats, err := s.repo.GetContributionStats(ctx, userID, workspaceID, days)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("getting contribution stats: %w", err)
	}

	return stats, nil
}

// GetUserStats retrieves user-specific statistics.
func (s *Service) GetUserStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (CoreUserStats, error) {
	s.log.Info(ctx, "getting user stats")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetUserStats")
	defer span.End()

	stats, err := s.repo.GetUserStats(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreUserStats{}, fmt.Errorf("getting user stats: %w", err)
	}

	return stats, nil
}
