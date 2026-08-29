package documentsrepository

import (
	"context"
	"fmt"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) List(
	ctx context.Context,
	input documentdomain.ListInput,
) ([]documentdomain.Summary, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	var rowLimit *int32
	if input.Limit != nil {
		converted, err := safecast.Int32(*input.Limit)
		if err != nil {
			return nil, fmt.Errorf("validate document list limit: %w", err)
		}
		rowLimit = &converted
	}
	rows, err := repository.queries.ListAccessibleDocuments(ctx, documentssql.ListAccessibleDocumentsParams{
		ActorID: input.UserID, WorkspaceID: input.WorkspaceID, SearchText: input.Search,
		ListScope: input.Scope, RowLimit: rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list accessible documents: %w", err)
	}
	result := make([]documentdomain.Summary, 0, len(rows))
	for _, row := range rows {
		result = append(result, summaryFromList(row))
	}
	return result, nil
}

func (repository *Repository) Get(
	ctx context.Context,
	workspaceID, userID, documentID uuid.UUID,
) (documentdomain.Document, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.Document{}, err
	}
	var document documentdomain.Document
	err := repository.transactor.WithinTransaction(
		ctx,
		pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly},
		func(tx pgx.Tx) error {
			var err error
			document, err = getDocument(ctx, documentssql.New(tx), workspaceID, userID, documentID)
			return err
		},
	)
	if err != nil {
		return documentdomain.Document{}, err
	}
	return document, nil
}

func getDocument(
	ctx context.Context,
	queries *documentssql.Queries,
	workspaceID, userID, documentID uuid.UUID,
) (documentdomain.Document, error) {
	row, err := queries.GetAccessibleDocument(ctx, documentssql.GetAccessibleDocumentParams{
		ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return documentdomain.Document{}, mapNotFound("get accessible document", err)
	}
	document := documentFromGet(row)
	if err := hydrateDocument(ctx, queries, &document, userID); err != nil {
		return documentdomain.Document{}, err
	}
	return document, nil
}

func hydrateDocument(
	ctx context.Context,
	queries *documentssql.Queries,
	document *documentdomain.Document,
	actorID uuid.UUID,
) error {
	members, err := queries.ListAccessibleDocumentMembers(ctx, documentssql.ListAccessibleDocumentMembersParams{
		ActorID: actorID, DocumentID: document.ID, WorkspaceID: document.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("list accessible document members: %w", err)
	}
	document.SharedWith = make([]documentdomain.Member, 0, len(members))
	for _, member := range members {
		document.SharedWith = append(document.SharedWith, documentdomain.Member{
			UserID: member.UserID, Role: member.Role,
		})
	}

	relatedWork, err := listRelationships(ctx, queries, document.WorkspaceID, document.ID, actorID)
	if err != nil {
		return err
	}
	document.RelatedWork = relatedWork
	document.RelatedWorkCount = len(relatedWork)
	return nil
}

func (repository *Repository) Create(
	ctx context.Context,
	input documentdomain.CreateInput,
) (documentdomain.Document, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.Document{}, err
	}
	row, err := repository.queries.CreateDocument(ctx, documentssql.CreateDocumentParams{
		Title: input.Title, ContentHtml: input.ContentHTML, ContentText: input.ContentText,
		Visibility: string(input.Visibility), WorkspaceID: input.WorkspaceID, ActorID: input.UserID,
	})
	if err != nil {
		return documentdomain.Document{}, mapCreateError(err)
	}
	return documentFromCreate(row), nil
}

func (repository *Repository) Update(
	ctx context.Context,
	input documentdomain.UpdateInput,
) (documentdomain.Document, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.Document{}, err
	}
	var document documentdomain.Document
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		row, err := queries.UpdateEditableDocument(ctx, documentssql.UpdateEditableDocumentParams{
			Title: input.Title, ContentHtml: input.ContentHTML, ContentText: input.ContentText,
			ActorID: input.UserID, DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
		})
		if err != nil {
			return mapNotFound("update editable document", err)
		}
		document = documentFromUpdate(row)
		return hydrateDocument(ctx, queries, &document, input.UserID)
	})
	if err != nil {
		return documentdomain.Document{}, err
	}
	return document, nil
}
