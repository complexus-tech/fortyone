package comments

import (
	"context"
	"fmt"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type AgentReplyScope = commentsdomain.AgentReplyScope
type AgentReplyCommand = commentsdomain.AgentReplyCommand
type MentionTarget = commentsdomain.MentionTarget
type AgentReply = commentsdomain.AgentReply

type agentReplyStore interface {
	ResolveAgentMentions(context.Context, AgentReplyScope, []uuid.UUID) ([]MentionTarget, error)
	CreateAgentReply(context.Context, AgentReplyCommand, uuid.UUID) (CoreComment, error)
	GetAgentReply(context.Context, AgentReplyScope, uuid.UUID) (AgentReply, error)
}

func (s *Service) ResolveAgentMentions(ctx context.Context, scope AgentReplyScope, ids []uuid.UUID) ([]MentionTarget, error) {
	if err := validateAgentReplyScope(scope); err != nil {
		return nil, err
	}
	if len(ids) > commentsdomain.MaximumMentions {
		return nil, ErrInvalidMention
	}
	repo, err := s.agentReplyStore()
	if err != nil {
		return nil, err
	}
	return repo.ResolveAgentMentions(ctx, scope, uniqueIDs(ids))
}

// CreateAgentReply is deliberately separate from ordinary/provider creation:
// system identity alone never grants permission to synthesize human mentions.
func (s *Service) CreateAgentReply(ctx context.Context, command AgentReplyCommand) (CoreComment, error) {
	if err := validateAgentReplyScope(command.Scope); err != nil {
		return CoreComment{}, err
	}
	if command.CreationID == uuid.Nil {
		return CoreComment{}, ErrInvalidComment
	}
	if err := validateContentAndMentions(command.Content, command.MentionedUserIDs); err != nil {
		return CoreComment{}, err
	}
	command.MentionedUserIDs = uniqueIDs(command.MentionedUserIDs)
	repo, err := s.agentReplyStore()
	if err != nil {
		return CoreComment{}, err
	}
	return repo.CreateAgentReply(ctx, command, s.eventIDs.New())
}

func (s *Service) GetAgentReply(ctx context.Context, scope AgentReplyScope, id uuid.UUID) (AgentReply, error) {
	if err := validateAgentReplyScope(scope); err != nil {
		return AgentReply{}, err
	}
	if id == uuid.Nil {
		return AgentReply{}, ErrInvalidComment
	}
	repo, err := s.agentReplyStore()
	if err != nil {
		return AgentReply{}, err
	}
	return repo.GetAgentReply(ctx, scope, id)
}

func (s *Service) agentReplyStore() (agentReplyStore, error) {
	repo, ok := s.repo.(agentReplyStore)
	if !ok || s.eventIDs == nil {
		return nil, fmt.Errorf("%w: agent replies are unavailable", ErrInvalidComment)
	}
	return repo, nil
}

func validateAgentReplyScope(scope AgentReplyScope) error {
	if scope.WorkspaceID == uuid.Nil || scope.StoryID == uuid.Nil || scope.ParentID == uuid.Nil {
		return ErrInvalidComment
	}
	if scope.Requester.Kind != platformauth.PrincipalHumanUser || validateMutationActor(scope.Requester, scope.WorkspaceID) != nil {
		return ErrForbidden
	}
	if scope.Author.Kind != platformauth.PrincipalSystem || scope.Author.Validate() != nil ||
		scope.Author.WorkspaceID != scope.WorkspaceID || !scope.Author.Scopes.Has(platformauth.ScopeCommentsWrite) {
		return ErrForbidden
	}
	return nil
}
