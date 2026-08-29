package workspaces

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.list")
	ctx, span := startSpan(ctx, "business.core.workspaces.List")
	defer span.End()
	workspaces, err := s.repo.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("workspaces retrieved.", trace.WithAttributes(attribute.String("user_id", userID.String())))
	return workspaces, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, userID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.get")
	ctx, span := startSpan(ctx, "business.core.workspaces.Get")
	defer span.End()
	workspace, err := s.repo.Get(ctx, workspaceID, userID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}
	span.AddEvent("workspace retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("user_id", userID.String()),
	))
	return workspace, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string, userID uuid.UUID) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.getBySlug")
	ctx, span := startSpan(ctx, "business.core.workspaces.GetBySlug")
	defer span.End()
	workspace, err := s.repo.GetBySlug(ctx, slug, userID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}
	span.AddEvent("workspace retrieved by slug.", trace.WithAttributes(
		attribute.String("slug", slug), attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
	))
	return workspace, nil
}

func (s *Service) GetPublicBySlug(ctx context.Context, slug string) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.getPublicBySlug")
	ctx, span := startSpan(ctx, "business.core.workspaces.GetPublicBySlug")
	defer span.End()
	workspace, err := s.repo.GetPublicBySlug(ctx, slug)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}
	span.AddEvent("public workspace branding retrieved.", trace.WithAttributes(
		attribute.String("slug", slug), attribute.String("workspace_id", workspace.ID.String()),
	))
	return workspace, nil
}

func (s *Service) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	s.log.Info(ctx, "business.core.workspaces.checkSlugAvailability")
	ctx, span := startSpan(ctx, "business.core.workspaces.CheckSlugAvailability")
	defer span.End()
	slug = strings.ToLower(strings.TrimSpace(slug))
	if _, restricted := restrictedSlugs[slug]; restricted {
		return false, nil
	}
	available, err := s.repo.CheckSlugAvailability(ctx, slug)
	if err != nil {
		span.RecordError(err)
		return false, err
	}
	span.AddEvent("slug availability checked.", trace.WithAttributes(
		attribute.String("slug", slug), attribute.Bool("available", available),
	))
	return available, nil
}

func (s *Service) Update(ctx context.Context, workspaceID uuid.UUID, updates CoreWorkspace) (CoreWorkspace, error) {
	s.log.Info(ctx, "business.core.workspaces.update")
	ctx, span := startSpan(ctx, "business.core.workspaces.Update")
	defer span.End()
	workspace, err := s.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspace{}, err
	}
	span.AddEvent("workspace updated.", trace.WithAttributes(attribute.String("workspace_id", workspaceID.String())))
	return workspace, nil
}
