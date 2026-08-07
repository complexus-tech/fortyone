package documents

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid document input")
	ErrForbidden    = errors.New("document access denied")
	ErrNotFound     = sql.ErrNoRows
)

const maxListLimit = 100

// Repository provides persistence for workspace documents and their relationships.
type Repository interface {
	List(ctx context.Context, input CoreListInput) ([]CoreDocumentSummary, error)
	Get(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (CoreDocument, error)
	Create(ctx context.Context, input CoreCreateInput) (CoreDocument, error)
	Duplicate(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (CoreDocument, error)
	Update(ctx context.Context, input CoreUpdateInput) (CoreDocument, error)
	Archive(ctx context.Context, workspaceID, userID, documentID uuid.UUID) error
	Delete(ctx context.Context, workspaceID, userID, documentID uuid.UUID) ([]uuid.UUID, error)
	SetAccess(ctx context.Context, input CoreAccessInput) (CoreDocument, error)
	AddRelationship(ctx context.Context, input CoreRelationshipInput) (CoreRelatedWork, error)
	RemoveRelationship(ctx context.Context, input CoreRelationshipInput) error
	ListRelatedDocuments(ctx context.Context, workspaceID, userID uuid.UUID, entityType RelationshipType, entityID uuid.UUID) ([]CoreDocumentSummary, error)
	LinkMedia(ctx context.Context, input CoreMediaInput) error
	UnlinkMedia(ctx context.Context, input CoreMediaInput) (bool, error)
	AuthorizeMedia(ctx context.Context, input CoreMediaInput) error
}

type Service struct {
	repo Repository
	log  *logger.Logger
}

func New(log *logger.Logger, repo Repository) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) List(ctx context.Context, input CoreListInput) ([]CoreDocumentSummary, error) {
	ctx, span := web.AddSpan(ctx, "business.core.documents.List")
	defer span.End()
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	input.Search = strings.TrimSpace(input.Search)
	if input.Scope != "" && input.Scope != "all" && input.Scope != "mine" && input.Scope != "shared" {
		return nil, ErrInvalidInput
	}
	if input.Limit != nil {
		if *input.Limit <= 0 {
			return nil, ErrInvalidInput
		}
		limit := min(*input.Limit, maxListLimit)
		input.Limit = &limit
	}
	return s.repo.List(ctx, input)
}

func (s *Service) Get(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (CoreDocument, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || documentID == uuid.Nil {
		return CoreDocument{}, ErrInvalidInput
	}
	return s.repo.Get(ctx, workspaceID, userID, documentID)
}

func (s *Service) Create(ctx context.Context, input CoreCreateInput) (CoreDocument, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreDocument{}, ErrInvalidInput
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		input.Title = "Untitled document"
	}
	if len([]rune(input.Title)) > 255 || !validVisibility(input.Visibility) {
		return CoreDocument{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Duplicate(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (CoreDocument, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || documentID == uuid.Nil {
		return CoreDocument{}, ErrInvalidInput
	}
	return s.repo.Duplicate(ctx, workspaceID, userID, documentID)
}

func (s *Service) Update(ctx context.Context, input CoreUpdateInput) (CoreDocument, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.DocumentID == uuid.Nil {
		return CoreDocument{}, ErrInvalidInput
	}
	if input.Title == nil && input.ContentHTML == nil && input.ContentText == nil {
		return CoreDocument{}, ErrInvalidInput
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			title = "Untitled document"
		}
		if len([]rune(title)) > 255 {
			return CoreDocument{}, ErrInvalidInput
		}
		input.Title = &title
	}
	return s.repo.Update(ctx, input)
}

func (s *Service) Archive(ctx context.Context, workspaceID, userID, documentID uuid.UUID) error {
	if workspaceID == uuid.Nil || userID == uuid.Nil || documentID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.Archive(ctx, workspaceID, userID, documentID)
}

func (s *Service) Delete(ctx context.Context, workspaceID, userID, documentID uuid.UUID) ([]uuid.UUID, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || documentID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return s.repo.Delete(ctx, workspaceID, userID, documentID)
}

func (s *Service) SetAccess(ctx context.Context, input CoreAccessInput) (CoreDocument, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.DocumentID == uuid.Nil || !validVisibility(input.Visibility) {
		return CoreDocument{}, ErrInvalidInput
	}
	seen := make(map[uuid.UUID]struct{}, len(input.Members))
	members := make([]CoreDocumentMember, 0, len(input.Members))
	for _, member := range input.Members {
		if member.UserID == uuid.Nil || member.UserID == input.UserID {
			continue
		}
		if member.Role != "viewer" && member.Role != "editor" {
			return CoreDocument{}, ErrInvalidInput
		}
		if _, exists := seen[member.UserID]; exists {
			continue
		}
		seen[member.UserID] = struct{}{}
		members = append(members, member)
	}
	input.Members = members
	return s.repo.SetAccess(ctx, input)
}

func (s *Service) AddRelationship(ctx context.Context, input CoreRelationshipInput) (CoreRelatedWork, error) {
	if !validRelationshipInput(input) {
		return CoreRelatedWork{}, ErrInvalidInput
	}
	return s.repo.AddRelationship(ctx, input)
}

func (s *Service) RemoveRelationship(ctx context.Context, input CoreRelationshipInput) error {
	if !validRelationshipInput(input) {
		return ErrInvalidInput
	}
	return s.repo.RemoveRelationship(ctx, input)
}

func (s *Service) ListRelatedDocuments(ctx context.Context, workspaceID, userID uuid.UUID, entityType RelationshipType, entityID uuid.UUID) ([]CoreDocumentSummary, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || entityID == uuid.Nil || !validRelationshipType(entityType) {
		return nil, ErrInvalidInput
	}
	return s.repo.ListRelatedDocuments(ctx, workspaceID, userID, entityType, entityID)
}

func (s *Service) LinkMedia(ctx context.Context, input CoreMediaInput) error {
	if !validMediaInput(input) {
		return ErrInvalidInput
	}
	return s.repo.LinkMedia(ctx, input)
}

func (s *Service) UnlinkMedia(ctx context.Context, input CoreMediaInput) (bool, error) {
	if !validMediaInput(input) {
		return false, ErrInvalidInput
	}
	return s.repo.UnlinkMedia(ctx, input)
}

func (s *Service) AuthorizeMedia(ctx context.Context, input CoreMediaInput) error {
	if !validMediaInput(input) {
		return ErrInvalidInput
	}
	return s.repo.AuthorizeMedia(ctx, input)
}

func validVisibility(visibility Visibility) bool {
	return visibility == VisibilityWorkspace || visibility == VisibilityRestricted || visibility == VisibilityPrivate
}

func validRelationshipInput(input CoreRelationshipInput) bool {
	return input.WorkspaceID != uuid.Nil && input.UserID != uuid.Nil && input.DocumentID != uuid.Nil && input.EntityID != uuid.Nil && validRelationshipType(input.EntityType)
}

func validRelationshipType(entityType RelationshipType) bool {
	return entityType == RelationshipStory || entityType == RelationshipObjective
}

func validMediaInput(input CoreMediaInput) bool {
	return input.WorkspaceID != uuid.Nil &&
		input.UserID != uuid.Nil &&
		input.DocumentID != uuid.Nil &&
		input.AttachmentID != uuid.Nil
}
