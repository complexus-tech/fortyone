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
	GetStoryStats(ctx context.Context, workspaceID uuid.UUID, filters StoryStatsFilters) (CoreStoryStats, error)
	GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, days int) ([]CoreContributionStats, error)
	GetUserStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (CoreUserStats, error)
	GetStatusStats(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) ([]CoreStatusStats, error)
	GetPriorityStats(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) ([]CorePriorityStats, error)

	// Workspace Reports Methods
	GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceOverview, error)
	GetStoryAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreStoryAnalytics, error)
	GetObjectiveProgress(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreObjectiveProgress, error)
	GetTeamPerformance(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTeamPerformance, error)
	GetSprintAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreSprintAnalyticsWorkspace, error)
	GetTimelineTrends(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTimelineTrends, error)
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
func (s *Service) GetStoryStats(ctx context.Context, workspaceID uuid.UUID, filters StoryStatsFilters) (CoreStoryStats, error) {
	s.log.Info(ctx, "getting story stats")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetStoryStats")
	defer span.End()

	stats, err := s.repo.GetStoryStats(ctx, workspaceID, filters)
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

// Workspace Reports Service Methods

// GetWorkspaceOverview retrieves workspace overview with key metrics and trends.
func (s *Service) GetWorkspaceOverview(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreWorkspaceOverview, error) {
	s.log.Info(ctx, "business.core.reports.GetWorkspaceOverview")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetWorkspaceOverview")
	defer span.End()

	overview, err := s.repo.GetWorkspaceOverview(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceOverview{}, fmt.Errorf("getting workspace overview: %w", err)
	}

	span.AddEvent("workspace overview retrieved.")
	return overview, nil
}

// GetStoryAnalytics retrieves story analytics including status breakdown and burndown.
func (s *Service) GetStoryAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreStoryAnalytics, error) {
	s.log.Info(ctx, "business.core.reports.GetStoryAnalytics")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetStoryAnalytics")
	defer span.End()

	analytics, err := s.repo.GetStoryAnalytics(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CoreStoryAnalytics{}, fmt.Errorf("getting story analytics: %w", err)
	}

	span.AddEvent("story analytics retrieved.")
	return analytics, nil
}

// GetObjectiveProgress retrieves objective progress including health and key results.
func (s *Service) GetObjectiveProgress(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreObjectiveProgress, error) {
	s.log.Info(ctx, "business.core.reports.GetObjectiveProgress")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetObjectiveProgress")
	defer span.End()

	progress, err := s.repo.GetObjectiveProgress(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CoreObjectiveProgress{}, fmt.Errorf("getting objective progress: %w", err)
	}

	span.AddEvent("objective progress retrieved.")
	return progress, nil
}

// GetTeamPerformance retrieves team performance including workload and velocity.
func (s *Service) GetTeamPerformance(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTeamPerformance, error) {
	s.log.Info(ctx, "business.core.reports.GetTeamPerformance")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetTeamPerformance")
	defer span.End()

	performance, err := s.repo.GetTeamPerformance(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CoreTeamPerformance{}, fmt.Errorf("getting team performance: %w", err)
	}

	span.AddEvent("team performance retrieved.")
	return performance, nil
}

// GetSprintAnalytics retrieves sprint analytics including progress and burndown.
func (s *Service) GetSprintAnalytics(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreSprintAnalyticsWorkspace, error) {
	s.log.Info(ctx, "business.core.reports.GetSprintAnalytics")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetSprintAnalytics")
	defer span.End()

	analytics, err := s.repo.GetSprintAnalytics(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CoreSprintAnalyticsWorkspace{}, fmt.Errorf("getting sprint analytics: %w", err)
	}

	span.AddEvent("sprint analytics retrieved.")
	return analytics, nil
}

// GetTimelineTrends retrieves timeline trends for all key metrics.
func (s *Service) GetTimelineTrends(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters) (CoreTimelineTrends, error) {
	s.log.Info(ctx, "business.core.reports.GetTimelineTrends")
	ctx, span := web.AddSpan(ctx, "business.core.reports.GetTimelineTrends")
	defer span.End()

	trends, err := s.repo.GetTimelineTrends(ctx, workspaceID, filters)
	if err != nil {
		span.RecordError(err)
		return CoreTimelineTrends{}, fmt.Errorf("getting timeline trends: %w", err)
	}

	span.AddEvent("timeline trends retrieved.")
	return trends, nil
}
