package workspaces

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the users storage.
type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]CoreWorkspace, error)
	Create(ctx context.Context, newWorkspace CoreWorkspace) (CoreWorkspace, error)
	Update(ctx context.Context, workspaceID uuid.UUID, updates CoreWorkspace) (CoreWorkspace, error)
	Delete(ctx context.Context, workspaceID uuid.UUID) error
	AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error
	Get(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspace, error)
	RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error
}

// Service provides user-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new users service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.list")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.List")
	defer span.End()

	workspaces, err := s.repo.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("workspaces retrieved.", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
	))
	return workspaces, nil
}

func (s *Service) Create(ctx context.Context, newWorkspace CoreWorkspace) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.create")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Create")
	defer span.End()

	workspace, err := s.repo.Create(ctx, newWorkspace)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	span.AddEvent("workspace created.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()),
	))
	return workspace, nil
}

func (s *Service) Update(ctx context.Context, workspaceID uuid.UUID, updates CoreWorkspace) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.update")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Update")
	defer span.End()

	workspace, err := s.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	span.AddEvent("workspace updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return workspace, nil
}

func (s *Service) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.delete")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace deleted.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return nil
}

func (s *Service) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	s.log.Info(ctx, "business.core.workspaces.addMember")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.AddMember")
	defer span.End()

	if role == "" {
		role = "member"
	}

	if err := s.repo.AddMember(ctx, workspaceID, userID, role); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

func (s *Service) Get(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.get")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.Get")
	defer span.End()

	workspace, err := s.repo.Get(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}

	span.AddEvent("workspace retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return workspace, nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.removeMember")
	ctx, span := web.AddSpan(ctx, "business.core.workspaces.RemoveMember")
	defer span.End()

	if err := s.repo.RemoveMember(ctx, workspaceID, userID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
	))
	return nil
}
