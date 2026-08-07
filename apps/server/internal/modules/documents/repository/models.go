package documentsrepository

import (
	"time"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
)

type dbDocument struct {
	ID               uuid.UUID  `db:"document_id"`
	WorkspaceID      uuid.UUID  `db:"workspace_id"`
	Title            string     `db:"title"`
	ContentHTML      string     `db:"content_html"`
	ContentText      string     `db:"content_text"`
	Visibility       string     `db:"visibility"`
	CreatedBy        uuid.UUID  `db:"created_by"`
	UpdatedBy        uuid.UUID  `db:"updated_by"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	ArchivedAt       *time.Time `db:"archived_at"`
	CanEdit          bool       `db:"can_edit"`
	RelatedWorkCount int        `db:"related_work_count"`
}

type dbDocumentSummary struct {
	ID               uuid.UUID `db:"document_id"`
	WorkspaceID      uuid.UUID `db:"workspace_id"`
	Title            string    `db:"title"`
	Visibility       string    `db:"visibility"`
	CreatedBy        uuid.UUID `db:"created_by"`
	UpdatedBy        uuid.UUID `db:"updated_by"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	CanEdit          bool      `db:"can_edit"`
	RelatedWorkCount int       `db:"related_work_count"`
}

type dbDocumentMember struct {
	UserID uuid.UUID `db:"user_id"`
	Role   string    `db:"role"`
}

type dbRelatedWork struct {
	EntityID   uuid.UUID  `db:"entity_id"`
	EntityType string     `db:"entity_type"`
	Title      string     `db:"title"`
	Reference  *string    `db:"reference"`
	TeamID     *uuid.UUID `db:"team_id"`
}

func toCoreDocument(document dbDocument) documents.CoreDocument {
	return documents.CoreDocument{
		ID:               document.ID,
		WorkspaceID:      document.WorkspaceID,
		Title:            document.Title,
		ContentHTML:      document.ContentHTML,
		ContentText:      document.ContentText,
		Visibility:       documents.Visibility(document.Visibility),
		CreatedBy:        document.CreatedBy,
		UpdatedBy:        document.UpdatedBy,
		CreatedAt:        document.CreatedAt,
		UpdatedAt:        document.UpdatedAt,
		ArchivedAt:       document.ArchivedAt,
		CanEdit:          document.CanEdit,
		RelatedWorkCount: document.RelatedWorkCount,
		SharedWith:       []documents.CoreDocumentMember{},
		RelatedWork:      []documents.CoreRelatedWork{},
	}
}

func toCoreDocuments(rows []dbDocument) []documents.CoreDocument {
	result := make([]documents.CoreDocument, len(rows))
	for i, row := range rows {
		result[i] = toCoreDocument(row)
	}
	return result
}

func toCoreDocumentSummaries(rows []dbDocumentSummary) []documents.CoreDocumentSummary {
	result := make([]documents.CoreDocumentSummary, len(rows))
	for i, row := range rows {
		result[i] = documents.CoreDocumentSummary{
			ID:               row.ID,
			WorkspaceID:      row.WorkspaceID,
			Title:            row.Title,
			Visibility:       documents.Visibility(row.Visibility),
			CreatedBy:        row.CreatedBy,
			UpdatedBy:        row.UpdatedBy,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			CanEdit:          row.CanEdit,
			RelatedWorkCount: row.RelatedWorkCount,
		}
	}
	return result
}

func toCoreRelatedWork(row dbRelatedWork) documents.CoreRelatedWork {
	return documents.CoreRelatedWork{
		EntityID:   row.EntityID,
		EntityType: documents.RelationshipType(row.EntityType),
		Title:      row.Title,
		Reference:  row.Reference,
		TeamID:     row.TeamID,
	}
}
