package documentsrepository

import (
	"time"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
)

func documentFromGet(row documentssql.GetAccessibleDocumentRow) documentdomain.Document {
	return newDocument(
		row.DocumentID, row.WorkspaceID, row.Title, row.ContentHtml, row.ContentText,
		row.Visibility, row.CreatedBy, row.UpdatedBy, row.CreatedAt, row.UpdatedAt,
		row.ArchivedAt, boolValue(row.CanEdit),
	)
}

func documentFromCreate(row documentssql.CreateDocumentRow) documentdomain.Document {
	return newDocument(
		row.DocumentID, row.WorkspaceID, row.Title, row.ContentHtml, row.ContentText,
		row.Visibility, row.CreatedBy, row.UpdatedBy, row.CreatedAt, row.UpdatedAt,
		row.ArchivedAt, row.CanEdit,
	)
}

func documentFromCreateWithID(row documentssql.CreateDocumentWithIDRow) documentdomain.Document {
	return newDocument(
		row.DocumentID, row.WorkspaceID, row.Title, row.ContentHtml, row.ContentText,
		row.Visibility, row.CreatedBy, row.UpdatedBy, row.CreatedAt, row.UpdatedAt,
		row.ArchivedAt, row.CanEdit,
	)
}

func documentFromUpdate(row documentssql.UpdateEditableDocumentRow) documentdomain.Document {
	return newDocument(
		row.DocumentID, row.WorkspaceID, row.Title, row.ContentHtml, row.ContentText,
		row.Visibility, row.CreatedBy, row.UpdatedBy, row.CreatedAt, row.UpdatedAt,
		row.ArchivedAt, row.CanEdit,
	)
}

func documentFromDuplicateSource(row documentssql.GetAccessibleDocumentForDuplicateRow) documentdomain.Document {
	return newDocument(
		row.DocumentID, row.WorkspaceID, row.Title, row.ContentHtml, row.ContentText,
		row.Visibility, row.CreatedBy, row.UpdatedBy, row.CreatedAt, row.UpdatedAt,
		row.ArchivedAt, row.CanEdit,
	)
}

func newDocument(
	id, workspaceID uuid.UUID,
	title, contentHTML, contentText, visibility string,
	createdBy, updatedBy uuid.UUID,
	createdAt, updatedAt time.Time,
	archivedAt *time.Time,
	canEdit bool,
) documentdomain.Document {
	return documentdomain.Document{
		ID: id, WorkspaceID: workspaceID, Title: title, ContentHTML: contentHTML,
		ContentText: contentText, Visibility: documentdomain.Visibility(visibility),
		CreatedBy: createdBy, UpdatedBy: updatedBy, CreatedAt: createdAt,
		UpdatedAt: updatedAt, ArchivedAt: archivedAt, CanEdit: canEdit,
		SharedWith: []documentdomain.Member{}, RelatedWork: []documentdomain.RelatedWork{},
	}
}

func summaryFromList(row documentssql.ListAccessibleDocumentsRow) documentdomain.Summary {
	return newSummary(
		row.DocumentID, row.WorkspaceID, row.Title, row.Visibility, row.CreatedBy,
		row.UpdatedBy, row.CreatedAt, row.UpdatedAt, boolValue(row.CanEdit), row.RelatedWorkCount,
	)
}

func summaryFromRelationship(row documentssql.ListAccessibleDocumentsForRelationshipRow) documentdomain.Summary {
	return newSummary(
		row.DocumentID, row.WorkspaceID, row.Title, row.Visibility, row.CreatedBy,
		row.UpdatedBy, row.CreatedAt, row.UpdatedAt, boolValue(row.CanEdit), row.RelatedWorkCount,
	)
}

func newSummary(
	id, workspaceID uuid.UUID,
	title, visibility string,
	createdBy, updatedBy uuid.UUID,
	createdAt, updatedAt time.Time,
	canEdit bool,
	relatedWorkCount int64,
) documentdomain.Summary {
	return documentdomain.Summary{
		ID: id, WorkspaceID: workspaceID, Title: title,
		Visibility: documentdomain.Visibility(visibility), CreatedBy: createdBy,
		UpdatedBy: updatedBy, CreatedAt: createdAt, UpdatedAt: updatedAt,
		CanEdit: canEdit, RelatedWorkCount: int(relatedWorkCount),
	}
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func stringPointer(value string) *string {
	copy := value
	return &copy
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	copy := value
	return &copy
}
