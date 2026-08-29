package workspaces

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) UploadWorkspaceLogo(ctx context.Context, workspaceID uuid.UUID, file multipart.File, header *multipart.FileHeader, attachments AttachmentsService) error {
	s.log.Info(ctx, "business.core.workspaces.uploadWorkspaceLogo")
	ctx, span := startSpan(ctx, "business.core.workspaces.UploadWorkspaceLogo")
	defer span.End()
	workspace, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	blobName, err := attachments.UploadWorkspaceLogo(ctx, file, header, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if workspace.AvatarURL != nil && *workspace.AvatarURL != "" {
		if err := attachments.DeleteWorkspaceLogo(ctx, *workspace.AvatarURL); err != nil {
			s.log.Error(ctx, "failed to delete previous workspace logo", "error", err)
		}
	}
	if _, err = s.repo.Update(ctx, workspaceID, CoreWorkspace{AvatarURL: &blobName}); err != nil {
		span.RecordError(err)
		if cleanupErr := attachments.DeleteWorkspaceLogo(ctx, blobName); cleanupErr != nil {
			s.log.Error(ctx, "failed to clean up unreferenced workspace logo", "error", cleanupErr)
		}
		return err
	}
	span.AddEvent("workspace logo updated", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("blob_name", blobName),
	))
	return nil
}

func (s *Service) DeleteWorkspaceLogo(ctx context.Context, workspaceID uuid.UUID, attachments AttachmentsService) error {
	s.log.Info(ctx, "business.core.workspaces.deleteWorkspaceLogo")
	ctx, span := startSpan(ctx, "business.core.workspaces.DeleteWorkspaceLogo")
	defer span.End()
	workspace, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if workspace.AvatarURL != nil && *workspace.AvatarURL != "" {
		if err := attachments.DeleteWorkspaceLogo(ctx, *workspace.AvatarURL); err != nil {
			s.log.Error(ctx, "failed to delete workspace logo", "error", err)
		}
	}
	empty := ""
	if _, err = s.repo.Update(ctx, workspaceID, CoreWorkspace{AvatarURL: &empty}); err != nil {
		span.RecordError(err)
		return err
	}
	span.AddEvent("workspace logo deleted", trace.WithAttributes(attribute.String("workspace_id", workspaceID.String())))
	return nil
}
