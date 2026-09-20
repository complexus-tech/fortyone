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
	Revision           int64               `json:"revision"`
	CollaborationEpoch int64               `json:"collaborationEpoch"`
	Collaborative      bool                `json:"collaborative"`
	PublicToken        *string             `json:"publicToken"`
	ID                 uuid.UUID           `json:"id"`
	WorkspaceID        uuid.UUID           `json:"workspaceId"`
	Title              string              `json:"title"`
	ContentHTML        string              `json:"contentHtml"`
	ContentText        string              `json:"contentText"`
	Visibility         string              `json:"visibility"`
	CreatedBy          uuid.UUID           `json:"createdBy"`
	UpdatedBy          uuid.UUID           `json:"updatedBy"`
	CreatedAt          time.Time           `json:"createdAt"`
	UpdatedAt          time.Time           `json:"updatedAt"`
	CanEdit            bool                `json:"canEdit"`
	SharedWith         []AppDocumentMember `json:"sharedWith"`
	RelatedWork        []AppRelatedWork    `json:"relatedWork"`
	RelatedWorkCount   int                 `json:"relatedWorkCount"`
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
	ExpectedRevision int64   `json:"expectedRevision"`
	Title            *string `json:"title"`
	ContentHTML      *string `json:"contentHtml"`
	ContentText      *string `json:"contentText"`
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

type AppDocumentComment struct {
	ID           uuid.UUID `json:"id"`
	Body         string    `json:"body"`
	CreatedBy    uuid.UUID `json:"createdBy"`
	AuthorName   string    `json:"authorName"`
	AuthorAvatar *string   `json:"authorAvatar"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AppDocumentCommentThread struct {
	ID          uuid.UUID            `json:"id"`
	DocumentID  uuid.UUID            `json:"documentId"`
	Quote       string               `json:"quote"`
	AnchorStart int32                `json:"anchorStart"`
	AnchorEnd   int32                `json:"anchorEnd"`
	CreatedBy   uuid.UUID            `json:"createdBy"`
	ResolvedAt  *time.Time           `json:"resolvedAt"`
	ResolvedBy  *uuid.UUID           `json:"resolvedBy"`
	CreatedAt   time.Time            `json:"createdAt"`
	Comments    []AppDocumentComment `json:"comments"`
}

type AppCreateDocumentComment struct {
	Body        string `json:"body"`
	Quote       string `json:"quote"`
	AnchorStart int32  `json:"anchorStart"`
	AnchorEnd   int32  `json:"anchorEnd"`
}

type AppDocumentCommentReply struct {
	Body string `json:"body"`
}

type AppResolveDocumentComment struct {
	Resolved bool `json:"resolved"`
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
		Revision: document.Revision, CollaborationEpoch: document.CollaborationEpoch, Collaborative: document.Collaborative, PublicToken: document.PublicToken,
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

func toAppDocumentComment(comment documents.CoreComment) AppDocumentComment {
	return AppDocumentComment{
		ID: comment.ID, Body: comment.Body, CreatedBy: comment.CreatedBy,
		AuthorName: comment.AuthorName, AuthorAvatar: comment.AuthorAvatar,
		CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt,
	}
}

func toAppDocumentCommentThread(thread documents.CoreCommentThread) AppDocumentCommentThread {
	comments := make([]AppDocumentComment, len(thread.Comments))
	for index, comment := range thread.Comments {
		comments[index] = toAppDocumentComment(comment)
	}
	return AppDocumentCommentThread{
		ID: thread.ID, DocumentID: thread.DocumentID, Quote: thread.Quote,
		AnchorStart: thread.AnchorStart, AnchorEnd: thread.AnchorEnd,
		CreatedBy: thread.CreatedBy, ResolvedAt: thread.ResolvedAt,
		ResolvedBy: thread.ResolvedBy, CreatedAt: thread.CreatedAt, Comments: comments,
	}
}

func toAppDocumentCommentThreads(threads []documents.CoreCommentThread) []AppDocumentCommentThread {
	result := make([]AppDocumentCommentThread, len(threads))
	for index, thread := range threads {
		result[index] = toAppDocumentCommentThread(thread)
	}
	return result
}
