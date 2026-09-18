package documents

import (
	"context"
	"encoding/hex"
	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func (s *Service) ListRevisions(ctx context.Context, workspaceID, userID, documentID uuid.UUID, before int64) ([]documentdomain.Revision, error) {
	if before < 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.ListRevisions(ctx, workspaceID, userID, documentID, before)
}
func (s *Service) GetRevision(ctx context.Context, workspaceID, userID, documentID uuid.UUID, revision int64) (documentdomain.Revision, error) {
	if revision < 1 {
		return documentdomain.Revision{}, ErrInvalidInput
	}
	return s.repo.GetRevision(ctx, workspaceID, userID, documentID, revision)
}
func (s *Service) RestoreRevision(ctx context.Context, workspaceID, userID, documentID uuid.UUID, revision, expected int64) (CoreDocument, error) {
	if revision < 1 || expected < 1 {
		return CoreDocument{}, ErrInvalidInput
	}
	return s.repo.RestoreRevision(ctx, workspaceID, userID, documentID, revision, expected)
}
func (s *Service) SetPublicLink(ctx context.Context, workspaceID, userID, documentID uuid.UUID, enabled bool) (CoreDocument, error) {
	return s.repo.SetPublicLink(ctx, workspaceID, userID, documentID, enabled)
}
func validPublicToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}
func (s *Service) GetPublicDocument(ctx context.Context, token string) (documentdomain.PublicDocument, error) {
	if !validPublicToken(token) {
		return documentdomain.PublicDocument{}, ErrNotFound
	}
	return s.repo.GetPublicDocument(ctx, token)
}
func (s *Service) AuthorizePublicMedia(ctx context.Context, token string, attachmentID uuid.UUID) (documentdomain.PublicDocument, error) {
	if !validPublicToken(token) {
		return documentdomain.PublicDocument{}, ErrNotFound
	}
	return s.repo.AuthorizePublicMedia(ctx, token, attachmentID)
}
func (s *Service) CreateCollaborationSession(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (documentdomain.CollaborationSession, error) {
	return s.repo.CreateCollaborationSession(ctx, workspaceID, userID, documentID)
}
