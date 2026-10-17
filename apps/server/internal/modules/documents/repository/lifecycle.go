package documentsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) Duplicate(
	ctx context.Context,
	workspaceID, userID, documentID uuid.UUID,
) (documentdomain.Document, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.Document{}, err
	}
	var duplicate documentdomain.Document
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		sourceRow, err := queries.GetAccessibleDocumentForDuplicate(ctx, documentssql.GetAccessibleDocumentForDuplicateParams{
			ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return mapNotFound("get document for duplicate", err)
		}
		source := documentFromDuplicateSource(sourceRow)
		source.ContentHTML, err = filterDuplicateReferences(ctx, queries, workspaceID, userID, source.ContentHTML)
		if err != nil {
			return err
		}
		oldMediaPath := "/documents/" + source.ID.String() + "/media/"

		// Generate the identity before insertion so stable media URLs can be
		// rewritten without a second document update.
		newDocumentID := uuid.New()
		newMediaPath := "/documents/" + newDocumentID.String() + "/media/"
		created, err := queries.CreateDocumentWithID(ctx, documentssql.CreateDocumentWithIDParams{
			DocumentID: newDocumentID, Title: duplicateTitle(source.Title),
			ContentHtml: strings.ReplaceAll(source.ContentHTML, oldMediaPath, newMediaPath),
			ContentText: source.ContentText, Visibility: string(documentdomain.VisibilityPrivate),
			WorkspaceID: workspaceID, ActorID: userID,
		})
		if err != nil {
			return mapCreateError(err)
		}
		duplicate = documentFromCreateWithID(created)
		if err := queries.CopyWorkspaceDocumentMedia(ctx, documentssql.CopyWorkspaceDocumentMediaParams{
			ActorID: userID, TargetDocumentID: duplicate.ID,
			SourceDocumentID: documentID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("copy workspace document media: %w", err)
		}
		if err := syncDocumentBody(ctx, queries, workspaceID, userID, duplicate.ID, duplicate.ContentHTML); err != nil {
			return err
		}
		return hydrateDocument(ctx, queries, &duplicate, userID)
	})
	if err != nil {
		return documentdomain.Document{}, err
	}
	return duplicate, nil
}

func duplicateTitle(title string) string {
	runes := []rune("Copy of " + title)
	if len(runes) > 255 {
		runes = runes[:255]
	}
	return string(runes)
}

func (repository *Repository) Archive(
	ctx context.Context,
	workspaceID, userID, documentID uuid.UUID,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	_, err := repository.queries.ArchiveOwnedDocument(ctx, documentssql.ArchiveOwnedDocumentParams{
		ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return mapNotFound("archive owned document", err)
	}
	return nil
}

func (repository *Repository) Unarchive(
	ctx context.Context,
	workspaceID, userID, documentID uuid.UUID,
) (documentdomain.Document, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.Document{}, err
	}
	var document documentdomain.Document
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		if _, err := queries.LockOwnedDocumentForLifecycle(ctx, documentssql.LockOwnedDocumentForLifecycleParams{
			ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
		}); err != nil {
			return mapNotFound("lock owned document for unarchive", err)
		}
		if _, err := queries.UnarchiveOwnedDocument(ctx, documentssql.UnarchiveOwnedDocumentParams{
			ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return documentdomain.ErrConflict
			}
			return fmt.Errorf("unarchive owned document: %w", err)
		}
		var err error
		document, err = getDocument(ctx, queries, workspaceID, userID, documentID)
		return err
	})
	if err != nil {
		return documentdomain.Document{}, err
	}
	return document, nil
}

func (repository *Repository) Delete(
	ctx context.Context,
	workspaceID, userID, documentID uuid.UUID,
) ([]uuid.UUID, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	var orphanCandidates []uuid.UUID
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		params := documentssql.LockOwnedDocumentForLifecycleParams{
			ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
		}
		if _, err := queries.LockOwnedDocumentForLifecycle(ctx, params); err != nil {
			return mapNotFound("lock owned document for delete", err)
		}
		var err error
		orphanCandidates, err = queries.ListOrphanedDocumentMediaCandidates(
			ctx,
			documentssql.ListOrphanedDocumentMediaCandidatesParams{
				DocumentID: documentID, WorkspaceID: workspaceID, ActorID: userID,
			},
		)
		if err != nil {
			return fmt.Errorf("list orphaned document media candidates: %w", err)
		}
		if _, err := queries.DeleteOwnedDocument(ctx, documentssql.DeleteOwnedDocumentParams{
			DocumentID: documentID, WorkspaceID: workspaceID, ActorID: userID,
		}); err != nil {
			return mapNotFound("delete owned document", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return append([]uuid.UUID(nil), orphanCandidates...), nil
}
