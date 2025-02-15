package teams

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the teams storage.
type Repository interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]CoreTeam, error)
	Create(ctx context.Context, team CoreTeam) (CoreTeam, error)
	Update(ctx context.Context, teamID uuid.UUID, updates CoreTeam) (CoreTeam, error)
	Delete(ctx context.Context, teamID uuid.UUID, workspaceID uuid.UUID) error
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID, workspaceID uuid.UUID) error
}

// Service provides team-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new teams service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]CoreTeam, error) {
	s.log.Info(ctx, "business.core.teams.list")
	ctx, span := web.AddSpan(ctx, "business.core.teams.List")
	defer span.End()

	teams, err := s.repo.List(ctx, workspaceID, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("teams retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.Int("teams.count", len(teams)),
	))
	return teams, nil
}

func (s *Service) Create(ctx context.Context, team CoreTeam) (CoreTeam, error) {
	s.log.Info(ctx, "business.core.teams.create")
	ctx, span := web.AddSpan(ctx, "business.core.teams.Create")
	defer span.End()

	result, err := s.repo.Create(ctx, team)
	if err != nil {
		span.RecordError(err)
		return CoreTeam{}, err
	}

	span.AddEvent("team created.", trace.WithAttributes(
		attribute.String("team_id", result.ID.String()),
		attribute.String("workspace_id", result.Workspace.String()),
	))
	return result, nil
}

func (s *Service) Update(ctx context.Context, teamID uuid.UUID, updates CoreTeam) (CoreTeam, error) {
	s.log.Info(ctx, "business.core.teams.update")
	ctx, span := web.AddSpan(ctx, "business.core.teams.Update")
	defer span.End()

	result, err := s.repo.Update(ctx, teamID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreTeam{}, err
	}

	span.AddEvent("team updated.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", updates.Workspace.String()),
	))
	return result, nil
}

func (s *Service) Delete(ctx context.Context, teamID uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.teams.delete")
	ctx, span := web.AddSpan(ctx, "business.core.teams.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, teamID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("team deleted.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))
	return nil
}

func (s *Service) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	s.log.Info(ctx, "business.core.teams.addMember")
	ctx, span := web.AddSpan(ctx, "business.core.teams.AddMember")
	defer span.End()

	if role == "" {
		role = "member"
	}

	if err := s.repo.AddMember(ctx, teamID, userID, role); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("team member added.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, teamID, userID uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.teams.removeMember")
	ctx, span := web.AddSpan(ctx, "business.core.teams.RemoveMember")
	defer span.End()

	if err := s.repo.RemoveMember(ctx, teamID, userID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("team member removed.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))
	return nil
}
