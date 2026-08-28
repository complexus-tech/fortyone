package documentshttp

import (
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
)

type AppDocumentMember struct {
	UserID uuid.UUID `json:"userId"`
	Role   string    `json:"role"`
}

type AppRelatedWork struct {
	EntityID   uuid.UUID  `json:"entityId"`
	EntityType string     `json:"entityType"`
	Title      string     `json:"title"`
	Reference  *string    `json:"reference"`
	TeamID     *uuid.UUID `json:"teamId"`
}

type AppDocument struct {
	ID               uuid.UUID           `json:"id"`
	WorkspaceID      uuid.UUID           `json:"workspaceId"`
	Title            string              `json:"title"`
	ContentHTML      string              `json:"contentHtml"`
	ContentText      string              `json:"contentText"`
	Visibility       string              `json:"visibility"`
	CreatedBy        uuid.UUID           `json:"createdBy"`
	UpdatedBy        uuid.UUID           `json:"updatedBy"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
	CanEdit          bool                `json:"canEdit"`
	SharedWith       []AppDocumentMember `json:"sharedWith"`
	RelatedWork      []AppRelatedWork    `json:"relatedWork"`
	RelatedWorkCount int                 `json:"relatedWorkCount"`
}

type AppDocumentSummary struct {
	ID               uuid.UUID `json:"id"`
	WorkspaceID      uuid.UUID `json:"workspaceId"`
	Title            string    `json:"title"`
	Visibility       string    `json:"visibility"`
	CreatedBy        uuid.UUID `json:"createdBy"`
	UpdatedBy        uuid.UUID `json:"updatedBy"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	CanEdit          bool      `json:"canEdit"`
	RelatedWorkCount int       `json:"relatedWorkCount"`
}

type AppCreateDocument struct {
	Title       string `json:"title"`
	Visibility  string `json:"visibility"`
	ContentHTML string `json:"contentHtml"`
	ContentText string `json:"contentText"`
}

type AppUpdateDocument struct {
	Title       *string `json:"title"`
	ContentHTML *string `json:"contentHtml"`
	ContentText *string `json:"contentText"`
}

type AppDocumentAccess struct {
	Visibility string              `json:"visibility"`
	Members    []AppDocumentMember `json:"members"`
}

type AppDocumentRelationship struct {
	EntityType string    `json:"entityType"`
	EntityID   uuid.UUID `json:"entityId"`
}

type AppDocumentMedia struct {
	ID         uuid.UUID `json:"id"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mimeType"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"createdAt"`
	UploadedBy uuid.UUID `json:"uploadedBy"`
}

func toAppDocument(document documents.CoreDocument, canMutate bool) AppDocument {
	sharedWith := make([]AppDocumentMember, len(document.SharedWith))
	for i, member := range document.SharedWith {
		sharedWith[i] = AppDocumentMember{UserID: member.UserID, Role: member.Role}
	}
	relatedWork := make([]AppRelatedWork, len(document.RelatedWork))
	for i, related := range document.RelatedWork {
		relatedWork[i] = AppRelatedWork{
			EntityID: related.EntityID, EntityType: string(related.EntityType), Title: related.Title,
			Reference: related.Reference, TeamID: related.TeamID,
		}
	}
	return AppDocument{
		ID: document.ID, WorkspaceID: document.WorkspaceID, Title: document.Title,
		ContentHTML: document.ContentHTML, ContentText: document.ContentText,
		Visibility: string(document.Visibility), CreatedBy: document.CreatedBy,
		UpdatedBy: document.UpdatedBy, CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt, CanEdit: document.CanEdit && canMutate,
		SharedWith: sharedWith, RelatedWork: relatedWork,
		RelatedWorkCount: document.RelatedWorkCount,
	}
}

func toAppDocumentSummaries(coreDocuments []documents.CoreDocumentSummary, canMutate bool) []AppDocumentSummary {
	result := make([]AppDocumentSummary, len(coreDocuments))
	for i, document := range coreDocuments {
		result[i] = AppDocumentSummary{
			ID:               document.ID,
			WorkspaceID:      document.WorkspaceID,
			Title:            document.Title,
			Visibility:       string(document.Visibility),
			CreatedBy:        document.CreatedBy,
			UpdatedBy:        document.UpdatedBy,
			CreatedAt:        document.CreatedAt,
			UpdatedAt:        document.UpdatedAt,
			CanEdit:          document.CanEdit && canMutate,
			RelatedWorkCount: document.RelatedWorkCount,
		}
	}
	return result
}

func toAppRelatedWork(related documents.CoreRelatedWork) AppRelatedWork {
	return AppRelatedWork{
		EntityID: related.EntityID, EntityType: string(related.EntityType), Title: related.Title,
		Reference: related.Reference, TeamID: related.TeamID,
	}
}

func toAppDocumentMedia(file attachments.FileInfo, stableURL string) AppDocumentMedia {
	return AppDocumentMedia{
		ID:         file.ID,
		Filename:   file.Filename,
		Size:       file.Size,
		MimeType:   file.MimeType,
		URL:        stableURL,
		CreatedAt:  file.CreatedAt,
		UploadedBy: file.UploadedBy,
	}
}
