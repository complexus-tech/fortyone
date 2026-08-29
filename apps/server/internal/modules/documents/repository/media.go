package documentsrepository

import (
	"context"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
)

func (repository *Repository) LinkMedia(ctx context.Context, input documentdomain.MediaInput) error {
	if err := repository.configured(); err != nil {
		return err
	}
	_, err := repository.queries.LinkEditableDocumentMedia(ctx, documentssql.LinkEditableDocumentMediaParams{
		ActorID: input.UserID, AttachmentID: input.AttachmentID,
		DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
	})
	if err != nil {
		return mapNotFound("link editable document media", err)
	}
	return nil
}

func (repository *Repository) UnlinkMedia(
	ctx context.Context,
	input documentdomain.MediaInput,
) (bool, error) {
	if err := repository.configured(); err != nil {
		return false, err
	}
	isOrphaned := false
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		attachmentID, err := queries.UnlinkEditableDocumentMedia(ctx, documentssql.UnlinkEditableDocumentMediaParams{
			DocumentID: input.DocumentID, AttachmentID: input.AttachmentID,
			WorkspaceID: input.WorkspaceID, ActorID: input.UserID,
		})
		if err != nil {
			return mapNotFound("unlink editable document media", err)
		}
		unreferenced, err := queries.IsWorkspaceAttachmentUnreferenced(
			ctx,
			documentssql.IsWorkspaceAttachmentUnreferencedParams{
				ActorID: input.UserID, AttachmentID: attachmentID, WorkspaceID: input.WorkspaceID,
			},
		)
		if err != nil {
			return mapNotFound("check workspace attachment references", err)
		}
		isOrphaned = boolValue(unreferenced)
		return nil
	})
	if err != nil {
		return false, err
	}
	return isOrphaned, nil
}

func (repository *Repository) AuthorizeMedia(ctx context.Context, input documentdomain.MediaInput) error {
	if err := repository.configured(); err != nil {
		return err
	}
	_, err := repository.queries.AuthorizeAccessibleDocumentMedia(
		ctx,
		documentssql.AuthorizeAccessibleDocumentMediaParams{
			ActorID: input.UserID, DocumentID: input.DocumentID,
			AttachmentID: input.AttachmentID, WorkspaceID: input.WorkspaceID,
		},
	)
	if err != nil {
		return mapNotFound("authorize accessible document media", err)
	}
	return nil
}
