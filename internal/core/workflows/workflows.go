package workflows

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Repository interface {
	Create(ctx context.Context, workspaceId uuid.UUID, workflow CoreNewWorkflow) (CoreWorkflow, error)
	Update(ctx context.Context, workspaceId, workflowId uuid.UUID, workflow CoreUpdateWorkflow) (CoreWorkflow, error)
	Delete(ctx context.Context, workspaceId, workflowId uuid.UUID) error
	List(ctx context.Context, workspaceId uuid.UUID) ([]CoreWorkflow, error)
	ListByTeam(ctx context.Context, workspaceId, teamId uuid.UUID) ([]CoreWorkflow, error)
}

type Service struct {
	repo Repository
	log  *logger.Logger
}

func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) Create(ctx context.Context, workspaceId uuid.UUID, nw CoreNewWorkflow) (CoreWorkflow, error) {
	s.log.Info(ctx, "business.core.workflows.Create")
	ctx, span := web.AddSpan(ctx, "business.core.workflows.Create")
	defer span.End()

	workflow, err := s.repo.Create(ctx, workspaceId, nw)
	if err != nil {
		span.RecordError(err)
		return CoreWorkflow{}, err
	}

	span.AddEvent("workflow created.", trace.WithAttributes(
		attribute.String("workflow.name", workflow.Name),
	))
	return workflow, nil
}

func (s *Service) Update(ctx context.Context, workspaceId, workflowId uuid.UUID, uw CoreUpdateWorkflow) (CoreWorkflow, error) {
	s.log.Info(ctx, "business.core.workflows.Update")
	ctx, span := web.AddSpan(ctx, "business.core.workflows.Update")
	defer span.End()

	workflow, err := s.repo.Update(ctx, workspaceId, workflowId, uw)
	if err != nil {
		span.RecordError(err)
		return CoreWorkflow{}, err
	}

	span.AddEvent("workflow updated.", trace.WithAttributes(
		attribute.String("workflow.id", workflowId.String()),
	))
	return workflow, nil
}

func (s *Service) Delete(ctx context.Context, workspaceId, workflowId uuid.UUID) error {
	s.log.Info(ctx, "business.core.workflows.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.workflows.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, workspaceId, workflowId); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("workflow deleted.", trace.WithAttributes(
		attribute.String("workflow.id", workflowId.String()),
	))
	return nil
}

func (s *Service) List(ctx context.Context, workspaceId uuid.UUID) ([]CoreWorkflow, error) {
	s.log.Info(ctx, "business.core.workflows.List")
	ctx, span := web.AddSpan(ctx, "business.core.workflows.List")
	defer span.End()

	workflows, err := s.repo.List(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("workflows retrieved.", trace.WithAttributes(
		attribute.Int("workflows.count", len(workflows)),
	))
	return workflows, nil
}

func (s *Service) ListByTeam(ctx context.Context, workspaceId, teamId uuid.UUID) ([]CoreWorkflow, error) {
	s.log.Info(ctx, "business.core.workflows.ListByTeam")
	ctx, span := web.AddSpan(ctx, "business.core.workflows.ListByTeam")
	defer span.End()

	workflows, err := s.repo.ListByTeam(ctx, workspaceId, teamId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("team workflows retrieved.", trace.WithAttributes(
		attribute.Int("workflows.count", len(workflows)),
	))
	return workflows, nil
}
