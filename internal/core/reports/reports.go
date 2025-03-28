package reports

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the reports storage.
type Repository interface {
	GetStoryStats(ctx context.Context, workspaceID uuid.UUID) (CoreStoryStats, error)
	GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, days int) ([]CoreContributionStats, error)
	GetUserStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (CoreUserStats, error)
	GetStatusStats(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) ([]CoreStatusStats, error)
	GetPriorityStats(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) ([]CorePriorityStats, error)
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

// GetStatusStats returns status statistics for stories
func (s *Service) GetStatusStats(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) ([]CoreStatusStats, error) {
	s.log.Info(ctx, "business.core.reports.GetStatusStats")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetStatusStats")
	defer span.End()

	stats, err := s.repo.GetStatusStats(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("status stats retrieved.", trace.WithAttributes(
		attribute.Int("stats.count", len(stats)),
	))
	return stats, nil
}

// GetPriorityStats returns priority statistics for stories
func (s *Service) GetPriorityStats(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) ([]CorePriorityStats, error) {
	s.log.Info(ctx, "business.core.reports.GetPriorityStats")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetPriorityStats")
	defer span.End()

	stats, err := s.repo.GetPriorityStats(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("priority stats retrieved.", trace.WithAttributes(
		attribute.Int("stats.count", len(stats)),
	))
	return stats, nil
}
