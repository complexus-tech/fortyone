package documents

import (
	"context"
	"encoding/json"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/complexus-tech/projects-api/internal/platform/documentmentions"
	"github.com/google/uuid"
)

type CommentWriter interface {
	CreateComment(context.Context, CoreCreateCommentInput) (CoreCommentThread, error)
	ReplyToComment(context.Context, CoreReplyCommentInput) (CoreComment, error)
}
type commentAgentRepository interface {
	GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (CoreCommentContext, error)
	ListCommentMentions(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]CoreCommentMention, error)
	CreateAgentReply(context.Context, CoreAgentReplyInput) (CoreComment, error)
	EnqueueAgentEdit(context.Context, CoreAgentEditInput) (CoreAgentEdit, error)
	GetAgentEdit(context.Context, uuid.UUID, uuid.UUID) (CoreAgentEdit, error)
	AgentEditsAvailable(context.Context) (bool, error)
}

func HasMayaMention(body string) bool { return documentmentions.Candidates(body).Maya }
func (s *Service) GetComment(ctx context.Context, w, user, document, id uuid.UUID) (CoreCommentContext, error) {
	repo, ok := s.repo.(commentAgentRepository)
	if !ok {
		return CoreCommentContext{}, domain.ErrNotConfigured
	}
	if w == uuid.Nil || user == uuid.Nil || document == uuid.Nil || id == uuid.Nil {
		return CoreCommentContext{}, ErrInvalidInput
	}
	return repo.GetComment(ctx, w, user, document, id)
}
func (s *Service) ListCommentMentions(ctx context.Context, w, user, document uuid.UUID) ([]CoreCommentMention, error) {
	repo, ok := s.repo.(commentAgentRepository)
	if !ok {
		return nil, domain.ErrNotConfigured
	}
	if w == uuid.Nil || user == uuid.Nil || document == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return repo.ListCommentMentions(ctx, w, user, document)
}
func (s *Service) CreateAgentReply(ctx context.Context, input CoreAgentReplyInput) (CoreComment, error) {
	repo, ok := s.repo.(commentAgentRepository)
	if !ok {
		return CoreComment{}, domain.ErrNotConfigured
	}
	if input.WorkspaceID == uuid.Nil || input.DocumentID == uuid.Nil || input.RequesterID == uuid.Nil || input.AuthorID == uuid.Nil || input.ThreadID == uuid.Nil || input.CreationID == uuid.Nil || input.AuthorID == input.RequesterID || strings.TrimSpace(input.Body) == "" || len([]rune(input.Body)) > 10000 {
		return CoreComment{}, ErrInvalidInput
	}
	return repo.CreateAgentReply(ctx, input)
}
func (s *Service) EnqueueAgentEdit(ctx context.Context, input CoreAgentEditInput) (CoreAgentEdit, error) {
	repo, ok := s.repo.(commentAgentRepository)
	if !ok {
		return CoreAgentEdit{}, domain.ErrNotConfigured
	}
	if input.WorkspaceID == uuid.Nil || input.DocumentID == uuid.Nil || input.RequesterID == uuid.Nil || input.InvocationID == uuid.Nil || input.ExpectedEpoch < 1 || input.ExpectedRevision < 1 || len(input.Operations) > 65536 || !json.Valid(input.Operations) {
		return CoreAgentEdit{}, ErrInvalidInput
	}
	return repo.EnqueueAgentEdit(ctx, input)
}
func (s *Service) GetAgentEdit(ctx context.Context, w, invocation uuid.UUID) (CoreAgentEdit, error) {
	repo, ok := s.repo.(commentAgentRepository)
	if !ok {
		return CoreAgentEdit{}, domain.ErrNotConfigured
	}
	return repo.GetAgentEdit(ctx, w, invocation)
}
func (s *Service) AgentEditsAvailable(ctx context.Context) (bool, error) {
	repo, ok := s.repo.(commentAgentRepository)
	if !ok {
		return false, nil
	}
	return repo.AgentEditsAvailable(ctx)
}
