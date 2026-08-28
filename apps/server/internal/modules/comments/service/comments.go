package comments

import (
	"context"
	"fmt"
	"strings"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
)

type Store interface {
	CreateComment(context.Context, commentsdomain.CreateCommand, uuid.UUID) (commentsdomain.Comment, error)
	UpdateComment(context.Context, commentsdomain.UpdateCommand, uuid.UUID) error
	DeleteComment(context.Context, commentsdomain.ActorScope, uuid.UUID) error
	GetComment(context.Context, commentsdomain.GetQuery) (commentsdomain.Comment, error)
}

type Repository interface {
	Store
}

type eventIDGenerator interface {
	New() uuid.UUID
}

type randomEventIDGenerator struct{}

func (randomEventIDGenerator) New() uuid.UUID { return uuid.New() }

type Service struct {
	repo     Repository
	eventIDs eventIDGenerator
}

func New(repo Repository) *Service {
	return newService(repo, randomEventIDGenerator{})
}

func newService(repo Repository, eventIDs eventIDGenerator) *Service {
	return &Service{repo: repo, eventIDs: eventIDs}
}

func (s *Service) CreateComment(
	ctx context.Context,
	command CreateCommentCommand,
) (CoreComment, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.comments.CreateComment")
	defer span.End()

	if err := validateCreateCommand(command); err != nil {
		return CoreComment{}, err
	}
	command.MentionedUserIDs = uniqueIDs(command.MentionedUserIDs)

	if s.repo == nil || s.eventIDs == nil {
		return CoreComment{}, fmt.Errorf("%w: comment service is unavailable", ErrInvalidComment)
	}
	return s.repo.CreateComment(ctx, command, s.eventIDs.New())
}

func (s *Service) UpdateComment(
	ctx context.Context,
	command UpdateCommentCommand,
) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.comments.UpdateComment")
	defer span.End()

	if err := validateActorScope(command.Scope); err != nil {
		return err
	}
	if err := validateContentAndMentions(command.Content, command.MentionedUserIDs); err != nil {
		return err
	}
	command.MentionedUserIDs = uniqueIDs(command.MentionedUserIDs)
	if s.repo == nil || s.eventIDs == nil {
		return fmt.Errorf("%w: comment service is unavailable", ErrInvalidComment)
	}
	return s.repo.UpdateComment(ctx, command, s.eventIDs.New())
}

func (s *Service) DeleteComment(ctx context.Context, scope AuthorScope) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.comments.DeleteComment")
	defer span.End()

	if err := validateActorScope(scope); err != nil {
		return err
	}
	if s.repo == nil || s.eventIDs == nil {
		return fmt.Errorf("%w: comment service is unavailable", ErrInvalidComment)
	}
	return s.repo.DeleteComment(ctx, scope, s.eventIDs.New())
}

func (s *Service) GetComment(ctx context.Context, query GetCommentQuery) (CoreComment, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.service.comments.GetComment")
	defer span.End()

	if query.CommentID == uuid.Nil || query.WorkspaceID == uuid.Nil {
		return CoreComment{}, fmt.Errorf("%w: comment and workspace are required", ErrInvalidComment)
	}
	if s.repo == nil {
		return CoreComment{}, fmt.Errorf("%w: comment service is unavailable", ErrInvalidComment)
	}
	return s.repo.GetComment(ctx, query)
}

func validateCreateCommand(command commentsdomain.CreateCommand) error {
	if command.StoryID == uuid.Nil || command.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: story and workspace are required", ErrInvalidComment)
	}
	if command.ParentID != nil && *command.ParentID == uuid.Nil {
		return fmt.Errorf("%w: parent id is invalid", ErrInvalidComment)
	}
	if err := validateMutationActor(command.Actor, command.WorkspaceID); err != nil {
		return err
	}
	return validateContentAndMentions(command.Content, command.MentionedUserIDs)
}

func validateActorScope(scope commentsdomain.ActorScope) error {
	if scope.CommentID == uuid.Nil || scope.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: comment and workspace are required", ErrInvalidComment)
	}
	return validateMutationActor(scope.Actor, scope.WorkspaceID)
}

func validateMutationActor(actor platformauth.Actor, workspaceID uuid.UUID) error {
	if err := actor.Validate(); err != nil || actor.WorkspaceID != workspaceID || !actor.IsUserActor() {
		return ErrForbidden
	}
	if !actor.Scopes.Has(platformauth.ScopeCommentsWrite) {
		return ErrForbidden
	}
	return nil
}

func validateContentAndMentions(content string, mentionedUserIDs []uuid.UUID) error {
	if strings.TrimSpace(content) == "" || len([]rune(content)) > commentsdomain.MaximumContentRunes {
		return fmt.Errorf(
			"%w: content must contain between 1 and %d characters",
			ErrInvalidComment,
			commentsdomain.MaximumContentRunes,
		)
	}
	if len(mentionedUserIDs) > commentsdomain.MaximumMentions {
		return fmt.Errorf("%w: at most %d mentions are allowed", ErrInvalidComment, commentsdomain.MaximumMentions)
	}
	for _, userID := range mentionedUserIDs {
		if userID == uuid.Nil {
			return fmt.Errorf("%w: mention ids must be non-zero", ErrInvalidComment)
		}
	}
	return nil
}

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) < 2 {
		return append([]uuid.UUID(nil), ids...)
	}

	unique := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
