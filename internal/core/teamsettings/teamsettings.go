package teamsettings

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the team settings storage.
type Repository interface {
	GetSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamSprintSettings, error)
	UpdateSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamSprintSettings) (CoreTeamSprintSettings, error)
	GetStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamStoryAutomationSettings, error)
	UpdateStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamStoryAutomationSettings) (CoreTeamStoryAutomationSettings, error)
	GetTeamsWithAutoSprintCreation(ctx context.Context) ([]CoreTeamSprintSettings, error)
	IncrementAutoSprintNumber(ctx context.Context, teamID, workspaceID uuid.UUID) error
}

// Service provides team settings-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new team settings service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// GetSettings returns the complete team settings.
func (s *Service) GetSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.getSettings")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.GetSettings")
	defer span.End()

	sprintSettings, err := s.repo.GetSprintSettings(ctx, teamID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreTeamSettings{}, err
	}

	storySettings, err := s.repo.GetStoryAutomationSettings(ctx, teamID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreTeamSettings{}, err
	}

	result := CoreTeamSettings{
		SprintSettings:          sprintSettings,
		StoryAutomationSettings: storySettings,
	}

	span.AddEvent("team settings retrieved.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))
	return result, nil
}

// GetSprintSettings returns the sprint settings for a team.
func (s *Service) GetSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamSprintSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.getSprintSettings")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.GetSprintSettings")
	defer span.End()

	settings, err := s.repo.GetSprintSettings(ctx, teamID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreTeamSprintSettings{}, err
	}

	span.AddEvent("sprint settings retrieved.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))
	return settings, nil
}

// UpdateSprintSettings updates the sprint settings for a team.
func (s *Service) UpdateSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamSprintSettings) (CoreTeamSprintSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.updateSprintSettings")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.UpdateSprintSettings")
	defer span.End()

	result, err := s.repo.UpdateSprintSettings(ctx, teamID, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreTeamSprintSettings{}, err
	}

	span.AddEvent("sprint settings updated.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))
	return result, nil
}

// GetStoryAutomationSettings returns the story automation settings for a team.
func (s *Service) GetStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamStoryAutomationSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.getStoryAutomationSettings")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.GetStoryAutomationSettings")
	defer span.End()

	settings, err := s.repo.GetStoryAutomationSettings(ctx, teamID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreTeamStoryAutomationSettings{}, err
	}

	span.AddEvent("story automation settings retrieved.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))
	return settings, nil
}

// UpdateStoryAutomationSettings updates the story automation settings for a team.
func (s *Service) UpdateStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamStoryAutomationSettings) (CoreTeamStoryAutomationSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.updateStoryAutomationSettings")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.UpdateStoryAutomationSettings")
	defer span.End()

	result, err := s.repo.UpdateStoryAutomationSettings(ctx, teamID, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreTeamStoryAutomationSettings{}, err
	}

	span.AddEvent("story automation settings updated.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))
	return result, nil
}

// GetTeamsWithAutoSprintCreation returns teams that have auto sprint creation enabled.
func (s *Service) GetTeamsWithAutoSprintCreation(ctx context.Context) ([]CoreTeamSprintSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.getTeamsWithAutoSprintCreation")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.GetTeamsWithAutoSprintCreation")
	defer span.End()

	teams, err := s.repo.GetTeamsWithAutoSprintCreation(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("teams with auto sprint creation retrieved.", trace.WithAttributes(
		attribute.Int("teams.count", len(teams)),
	))
	return teams, nil
}
