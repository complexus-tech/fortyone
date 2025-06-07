package teamsettings

import (
	"context"
	"errors"

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
	CreateDefaultStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamStoryAutomationSettings, error)
}

// Validation errors
var (
	ErrInvalidSprintStartDay = errors.New("sprint start day must be Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, or Sunday")
	ErrInvalidSprintDuration = errors.New("sprint duration must be between 1 and 8 weeks")
	ErrInvalidUpcomingCount  = errors.New("upcoming sprints count must be between 0 and 10")
	ErrInvalidCloseMonths    = errors.New("auto-close inactive months must be between 1 and 24")
	ErrInvalidArchiveMonths  = errors.New("auto-archive months must be between 1 and 24")
)

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

// CreateDefaultStoryAutomationSettings creates default story automation settings for a team.
func (s *Service) CreateDefaultStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamStoryAutomationSettings, error) {
	s.log.Info(ctx, "business.core.teamsettings.createDefaultStoryAutomationSettings")
	ctx, span := web.AddSpan(ctx, "business.core.teamsettings.CreateDefaultStoryAutomationSettings")
	defer span.End()

	settings, err := s.repo.CreateDefaultStoryAutomationSettings(ctx, teamID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreTeamStoryAutomationSettings{}, err
	}

	span.AddEvent("default story automation settings created.", trace.WithAttributes(
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

	// Validate the updates
	if err := s.validateSprintSettingsUpdate(updates); err != nil {
		span.RecordError(err)
		return CoreTeamSprintSettings{}, err
	}

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

	// Validate the updates
	if err := s.validateStoryAutomationSettingsUpdate(updates); err != nil {
		span.RecordError(err)
		return CoreTeamStoryAutomationSettings{}, err
	}

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

// validateSprintSettingsUpdate validates sprint settings updates
func (s *Service) validateSprintSettingsUpdate(updates CoreUpdateTeamSprintSettings) error {
	validDays := map[string]bool{
		"Monday":    true,
		"Tuesday":   true,
		"Wednesday": true,
		"Thursday":  true,
		"Friday":    true,
		"Saturday":  true,
		"Sunday":    true,
	}

	if updates.SprintStartDay != nil && !validDays[*updates.SprintStartDay] {
		return ErrInvalidSprintStartDay
	}

	if updates.SprintDurationWeeks != nil && (*updates.SprintDurationWeeks < 1 || *updates.SprintDurationWeeks > 8) {
		return ErrInvalidSprintDuration
	}

	if updates.UpcomingSprintsCount != nil && (*updates.UpcomingSprintsCount < 0 || *updates.UpcomingSprintsCount > 10) {
		return ErrInvalidUpcomingCount
	}

	return nil
}

// validateStoryAutomationSettingsUpdate validates story automation settings updates
func (s *Service) validateStoryAutomationSettingsUpdate(updates CoreUpdateTeamStoryAutomationSettings) error {
	if updates.AutoCloseInactiveMonths != nil && (*updates.AutoCloseInactiveMonths < 1 || *updates.AutoCloseInactiveMonths > 24) {
		return ErrInvalidCloseMonths
	}

	if updates.AutoArchiveMonths != nil && (*updates.AutoArchiveMonths < 1 || *updates.AutoArchiveMonths > 24) {
		return ErrInvalidArchiveMonths
	}

	return nil
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
