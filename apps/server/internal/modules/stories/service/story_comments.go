package stories

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type commentOptions struct {
	actorID uuid.UUID
}

func (s *Service) ConfigureCommentCreator(creator CommentCreator) {
	s.commentCreator = creator
}

// CreateComment creates a comment for a story.
func (s *Service) CreateComment(
	ctx context.Context,
	workspaceID uuid.UUID,
	comment CoreNewComment,
) (CoreComment, error) {
	actorID, err := auth.GetUserID(ctx)
	if err != nil {
		return CoreComment{}, err
	}
	return s.createCommentWithOptions(ctx, workspaceID, comment, commentOptions{actorID: actorID})
}

func (s *Service) CreateCommentExternal(
	ctx context.Context,
	actorID, workspaceID uuid.UUID,
	comment CoreNewComment,
) (CoreComment, error) {
	return s.createCommentWithOptions(ctx, workspaceID, comment, commentOptions{actorID: actorID})
}

func (s *Service) createCommentWithOptions(
	ctx context.Context,
	workspaceID uuid.UUID,
	commentInput CoreNewComment,
	options commentOptions,
) (CoreComment, error) {
	s.log.Info(ctx, "business.core.stories.CreateComment")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.CreateComment")
	defer span.End()

	story, err := s.getVisibleStory(ctx, commentInput.StoryID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreComment{}, err
	}

	var parentComment *CoreComment
	if commentInput.Parent != nil {
		parent, err := s.getVisibleComment(ctx, *commentInput.Parent, commentInput.StoryID, workspaceID)
		if err != nil {
			span.RecordError(err)
			return CoreComment{}, err
		}
		parentComment = &parent
	}

	if s.commentCreator == nil {
		return CoreComment{}, ErrCommentWriterUnavailable
	}
	actor := auth.NewHumanActor(options.actorID)
	if contextualActor, actorErr := auth.GetActor(ctx); actorErr == nil && contextualActor.PrincipalID == options.actorID {
		actor = contextualActor
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		return CoreComment{}, err
	}
	created, err := s.commentCreator.CreateComment(ctx, CreateCommentCommand{
		WorkspaceID: workspaceID, StoryID: commentInput.StoryID, ParentID: commentInput.Parent,
		Actor: actor, Content: commentInput.Comment, MentionedUserIDs: commentInput.Mentions,
	})
	if err != nil {
		span.RecordError(err)
		return CoreComment{}, err
	}

	audienceIDs, audienceErr := s.GetNotificationAudience(ctx, commentInput.StoryID, workspaceID)
	audienceResolved := audienceErr == nil
	if audienceErr != nil {
		s.log.Error(ctx, "failed to load story notification audience", "error", audienceErr, "story_id", commentInput.StoryID)
	}

	if s.publisher != nil {
		s.publishCommentNotifications(
			ctx, story, created, parentComment, commentInput, options.actorID,
			audienceIDs, audienceResolved,
		)
	}

	span.AddEvent("comment created.", trace.WithAttributes(
		attribute.String("comment.id", created.ID.String()),
		attribute.Int("mentions.count", len(commentInput.Mentions)),
	))
	return created, nil
}

func (s *Service) publishCommentNotifications(
	ctx context.Context,
	story CoreSingleStory,
	comment CoreComment,
	parentComment *CoreComment,
	input CoreNewComment,
	actorID uuid.UUID,
	audienceIDs []uuid.UUID,
	audienceResolved bool,
) {
	if input.Parent != nil && parentComment != nil {
		event := events.Event{
			Type: events.CommentReplied,
			Payload: events.CommentRepliedPayload{
				CommentID: comment.ID, ParentCommentID: *input.Parent,
				ParentAuthorID: parentComment.UserID, StoryID: input.StoryID,
				StoryTitle: story.Title, WorkspaceID: story.Workspace,
				Content: input.Comment, Mentions: input.Mentions,
				AudienceIDs: audienceIDs, AudienceResolved: audienceResolved,
			},
			Timestamp: time.Now(), ActorID: actorID,
		}
		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish comment replied event", "error", err)
		}
	} else if input.Parent == nil {
		event := events.Event{
			Type: events.CommentCreated,
			Payload: events.CommentCreatedPayload{
				CommentID: comment.ID, StoryID: input.StoryID, StoryTitle: story.Title,
				AssigneeID: story.Assignee, WorkspaceID: story.Workspace,
				Content: input.Comment, Mentions: input.Mentions,
				AudienceIDs: audienceIDs, AudienceResolved: audienceResolved,
			},
			Timestamp: time.Now(), ActorID: actorID,
		}
		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish comment created event", "error", err)
		}
	}

	for _, mentionedUserID := range input.Mentions {
		event := events.Event{
			Type: events.UserMentioned,
			Payload: events.UserMentionedPayload{
				CommentID: comment.ID, StoryID: input.StoryID, StoryTitle: story.Title,
				WorkspaceID: story.Workspace, MentionedUser: mentionedUserID, Content: input.Comment,
			},
			Timestamp: time.Now(), ActorID: actorID,
		}
		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish user mentioned event", "error", err, "mentioned_user", mentionedUserID)
		}
	}
}

// GetComments returns the comments for a story with pagination.
func (s *Service) GetComments(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	page, pageSize int,
) ([]CoreComment, bool, error) {
	s.log.Info(ctx, "business.core.stories.GetComments")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.GetComments")
	defer span.End()

	if _, err := s.getVisibleStory(ctx, storyID, workspaceID); err != nil {
		span.RecordError(err)
		return nil, false, err
	}
	comments, hasMore, err := s.listVisibleComments(ctx, storyID, workspaceID, page, pageSize)
	if err != nil {
		s.log.Error(ctx, "failed to get comments", "error", err)
		span.RecordError(err)
		return nil, false, err
	}
	span.AddEvent("comments retrieved.", trace.WithAttributes(
		attribute.Int("comment.count", len(comments)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))
	return comments, hasMore, nil
}

// GetComment resolves a single comment so external callers can validate reply
// ownership before creating a nested comment.
func (s *Service) GetComment(
	ctx context.Context,
	commentID, storyID, workspaceID uuid.UUID,
) (CoreComment, error) {
	return s.getVisibleComment(ctx, commentID, storyID, workspaceID)
}
