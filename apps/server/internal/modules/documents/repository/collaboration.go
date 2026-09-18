package documentsrepository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) ListRevisions(ctx context.Context, workspaceID, userID, documentID uuid.UUID, before int64) ([]documentdomain.Revision, error) {
	result := []documentdomain.Revision{}
	err := r.withinSerializable(ctx, func(q *documentssql.Queries) error {
		if _, err := getDocument(ctx, q, workspaceID, userID, documentID); err != nil {
			return err
		}
		rows, err := q.ListDocumentRevisions(ctx, documentssql.ListDocumentRevisionsParams{DocumentID: documentID, BeforeRevision: before})
		if err != nil {
			return err
		}
		result = make([]documentdomain.Revision, 0, len(rows))
		for _, row := range rows {
			result = append(result, documentdomain.Revision{Revision: row.Revision, Title: row.Title, EditedBy: row.EditedBy, CreatedAt: row.CreatedAt})
		}
		return nil
	})
	return result, err
}

func (r *Repository) GetRevision(ctx context.Context, workspaceID, userID, documentID uuid.UUID, revision int64) (documentdomain.Revision, error) {
	var result documentdomain.Revision
	err := r.withinSerializable(ctx, func(q *documentssql.Queries) error {
		if _, err := getDocument(ctx, q, workspaceID, userID, documentID); err != nil {
			return err
		}
		row, err := q.GetDocumentRevision(ctx, documentssql.GetDocumentRevisionParams{DocumentID: documentID, Revision: revision})
		if err != nil {
			return mapNotFound("get document revision", err)
		}
		result = documentdomain.Revision{Revision: row.Revision, Title: row.Title, ContentHTML: row.ContentHtml, ContentText: row.ContentText, EditedBy: row.EditedBy, CreatedAt: row.CreatedAt}
		return nil
	})
	return result, err
}

func (r *Repository) RestoreRevision(ctx context.Context, workspaceID, userID, documentID uuid.UUID, revision, expected int64) (documentdomain.Document, error) {
	var result documentdomain.Document
	err := r.withinSerializable(ctx, func(q *documentssql.Queries) error {
		document, err := getDocument(ctx, q, workspaceID, userID, documentID)
		if err != nil {
			return err
		}
		if !document.CanEdit {
			return documentdomain.ErrForbidden
		}
		if document.Revision != expected {
			return documentdomain.ErrConflict
		}
		if _, err := q.RestoreDocumentRevision(ctx, documentssql.RestoreDocumentRevisionParams{DocumentID: documentID, ActorID: userID, Revision: revision, ExpectedRevision: expected}); err != nil {
			return mapNotFound("restore document revision", err)
		}
		result, err = getDocument(ctx, q, workspaceID, userID, documentID)
		return err
	})
	return result, err
}

func (r *Repository) SetPublicLink(ctx context.Context, workspaceID, userID, documentID uuid.UUID, enabled bool) (documentdomain.Document, error) {
	var result documentdomain.Document
	err := r.withinSerializable(ctx, func(q *documentssql.Queries) error {
		document, err := getDocument(ctx, q, workspaceID, userID, documentID)
		if err != nil {
			return err
		}
		if !document.CanEdit || document.CreatedBy != userID {
			return documentdomain.ErrForbidden
		}
		var token *string
		if enabled {
			token = document.PublicToken
			if token == nil {
				value, err := randomToken()
				if err != nil {
					return err
				}
				token = &value
			}
		}
		if err := q.SetDocumentPublicToken(ctx, documentssql.SetDocumentPublicTokenParams{DocumentID: documentID, PublicToken: token}); err != nil {
			return err
		}
		result, err = getDocument(ctx, q, workspaceID, userID, documentID)
		return err
	})
	return result, err
}

func (r *Repository) GetPublicDocument(ctx context.Context, token string) (documentdomain.PublicDocument, error) {
	if err := r.configured(); err != nil {
		return documentdomain.PublicDocument{}, err
	}
	row, err := r.queries.GetPublicDocument(ctx, documentssql.GetPublicDocumentParams{PublicToken: &token})
	if err != nil {
		return documentdomain.PublicDocument{}, mapNotFound("get public document", err)
	}
	return documentdomain.PublicDocument{ID: row.DocumentID, WorkspaceID: row.WorkspaceID, Title: row.Title, ContentHTML: row.ContentHtml, ContentText: row.ContentText, UpdatedAt: row.UpdatedAt}, nil
}

func (r *Repository) AuthorizePublicMedia(ctx context.Context, token string, attachmentID uuid.UUID) (documentdomain.PublicDocument, error) {
	document, err := r.GetPublicDocument(ctx, token)
	if err != nil {
		return documentdomain.PublicDocument{}, err
	}
	_, err = r.queries.AuthorizePublicDocumentMedia(ctx, documentssql.AuthorizePublicDocumentMediaParams{PublicToken: &token, AttachmentID: attachmentID})
	if err != nil {
		return documentdomain.PublicDocument{}, mapNotFound("authorize public document media", err)
	}
	return document, nil
}

func (r *Repository) CreateCollaborationSession(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (documentdomain.CollaborationSession, error) {
	token, err := randomToken()
	if err != nil {
		return documentdomain.CollaborationSession{}, err
	}
	digest := sha256.Sum256([]byte(token))
	var result documentdomain.CollaborationSession
	err = r.withinSerializable(ctx, func(q *documentssql.Queries) error {
		document, err := getDocument(ctx, q, workspaceID, userID, documentID)
		if err != nil {
			return err
		}
		if err := q.DeleteExpiredDocumentCollaborationSessions(ctx); err != nil {
			return err
		}
		if err := q.CreateDocumentCollaborationSession(ctx, documentssql.CreateDocumentCollaborationSessionParams{TokenHash: hex.EncodeToString(digest[:]), DocumentID: documentID, ActorID: userID}); err != nil {
			return err
		}
		result = documentdomain.CollaborationSession{Token: token, Name: fmt.Sprintf("%s:%d", documentID, document.CollaborationEpoch)}
		return nil
	})
	return result, err
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate document token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
