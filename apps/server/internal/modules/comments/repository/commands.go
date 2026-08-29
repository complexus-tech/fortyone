package commentsrepository

import (
	"context"
	"errors"
	"fmt"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *Repository) CreateComment(
	ctx context.Context,
	command commentsdomain.CreateCommand,
	eventID uuid.UUID,
) (commentsdomain.Comment, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.comments.CreateComment")
	defer span.End()

	var created commentsdomain.Comment
	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		row, err := queries.CreateCommentForActor(ctx, commentsql.CreateCommentForActorParams{
			Content: command.Content, ActorID: command.Actor.PrincipalID,
			ParentID: command.ParentID, StoryID: command.StoryID, WorkspaceID: command.WorkspaceID,
			TeamAccessUnrestricted: command.Actor.TeamAccess.IsUnrestricted(),
			AllowedTeamIds:         command.Actor.TeamAccess.RestrictedTeamIDs(),
		})
		if err != nil {
			return mapCommentWriteError("create comment", err)
		}
		created = commentFromValues(
			row.CommentID, row.StoryID, row.ParentID, row.CommenterID,
			row.Content, row.CreatedAt, row.UpdatedAt,
		)
		scope := commentsdomain.ActorScope{
			CommentID: created.ID, WorkspaceID: command.WorkspaceID, Actor: command.Actor,
		}
		if err := replaceMentions(ctx, queries, scope, command.MentionedUserIDs); err != nil {
			return err
		}
		return appendMutationEvent(ctx, queries, eventID, commentsdomain.EventCreated, command.WorkspaceID, command.Actor, created)
	})
	if err != nil {
		return commentsdomain.Comment{}, err
	}

	span.SetAttributes(
		attribute.String("comment.id", created.ID.String()),
		attribute.String("story.id", created.StoryID.String()),
		attribute.String("workspace.id", command.WorkspaceID.String()),
	)
	return created, nil
}

func (r *Repository) UpdateComment(
	ctx context.Context,
	command commentsdomain.UpdateCommand,
	eventID uuid.UUID,
) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.comments.UpdateComment")
	defer span.End()

	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		row, err := queries.UpdateCommentForAuthor(ctx, commentsql.UpdateCommentForAuthorParams{
			Content: command.Content, CommentID: command.Scope.CommentID,
			ActorID: command.Scope.Actor.PrincipalID, WorkspaceID: command.Scope.WorkspaceID,
			TeamAccessUnrestricted: command.Scope.Actor.TeamAccess.IsUnrestricted(),
			AllowedTeamIds:         command.Scope.Actor.TeamAccess.RestrictedTeamIDs(),
		})
		if err != nil {
			return mapCommentWriteError("update comment", err)
		}
		updated := commentFromValues(
			row.CommentID, row.StoryID, row.ParentID, row.CommenterID,
			row.Content, row.CreatedAt, row.UpdatedAt,
		)
		if err := replaceMentions(ctx, queries, command.Scope, command.MentionedUserIDs); err != nil {
			return err
		}
		return appendMutationEvent(
			ctx, queries, eventID, commentsdomain.EventUpdated,
			command.Scope.WorkspaceID, command.Scope.Actor, updated,
		)
	})
	if err != nil {
		r.log.Error(ctx, "error updating comment", "error", err, "comment_id", command.Scope.CommentID)
		return err
	}

	r.log.Info(ctx, "comment updated successfully", "comment_id", command.Scope.CommentID)
	return nil
}

func (r *Repository) DeleteComment(
	ctx context.Context,
	scope commentsdomain.ActorScope,
	eventID uuid.UUID,
) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.comments.DeleteComment")
	defer span.End()

	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		row, err := queries.DeleteCommentForAuthor(ctx, commentsql.DeleteCommentForAuthorParams{
			CommentID: scope.CommentID, ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID,
			TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(),
			AllowedTeamIds:         scope.Actor.TeamAccess.RestrictedTeamIDs(),
		})
		if err != nil {
			return mapCommentWriteError("delete comment", err)
		}
		deleted := commentFromValues(
			row.CommentID, row.StoryID, row.ParentID, row.CommenterID,
			row.Content, row.CreatedAt, row.UpdatedAt,
		)
		return appendMutationEvent(
			ctx, queries, eventID, commentsdomain.EventDeleted,
			scope.WorkspaceID, scope.Actor, deleted,
		)
	})
	if err != nil {
		r.log.Error(ctx, "failed to delete comment", "error", err, "comment_id", scope.CommentID)
		return err
	}

	r.log.Info(ctx, "comment deleted successfully", "comment_id", scope.CommentID)
	span.AddEvent("Comment deleted.", trace.WithAttributes(attribute.String("comment.id", scope.CommentID.String())))
	return nil
}

func replaceMentions(
	ctx context.Context,
	queries commentsql.Querier,
	scope commentsdomain.ActorScope,
	mentionedUserIDs []uuid.UUID,
) error {
	params := commentsql.DeleteCommentMentionsForAuthorParams{
		CommentID: scope.CommentID, ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID,
		TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(),
		AllowedTeamIds:         scope.Actor.TeamAccess.RestrictedTeamIDs(),
	}
	cleared, err := queries.DeleteCommentMentionsForAuthor(ctx, params)
	if err != nil {
		return fmt.Errorf("delete comment mentions: %w", err)
	}
	if !cleared.CommentFound {
		return commentsdomain.ErrNotFound
	}
	if len(mentionedUserIDs) == 0 {
		return nil
	}

	inserted, err := queries.InsertCommentMentionsForAuthor(ctx, commentsql.InsertCommentMentionsForAuthorParams{
		CommentID: scope.CommentID, ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID,
		TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(),
		AllowedTeamIds:         scope.Actor.TeamAccess.RestrictedTeamIDs(),
		MentionedUserIds:       mentionedUserIDs,
	})
	if err != nil {
		return fmt.Errorf("insert comment mentions: %w", err)
	}
	if inserted != int64(len(mentionedUserIDs)) {
		return commentsdomain.ErrInvalidMention
	}
	return nil
}

func appendMutationEvent(
	ctx context.Context,
	queries commentsql.Querier,
	eventID uuid.UUID,
	eventType commentsdomain.MutationEventType,
	workspaceID uuid.UUID,
	actor platformauth.Actor,
	comment commentsdomain.Comment,
) error {
	event, err := commentsdomain.NewMutationEvent(eventID, eventType, workspaceID, actor, comment)
	if err != nil {
		return err
	}
	_, err = queries.AppendCommentMutationEvent(ctx, commentsql.AppendCommentMutationEventParams{
		PayloadBody: event.DeliveryBody, EventType: string(event.Type), EventID: event.ID,
		WorkspaceID: event.WorkspaceID, CommentID: event.CommentID,
		ActorKind: string(event.Actor.Kind), ActorID: event.Actor.PrincipalID,
		ActorCredentialID: optionalUUID(event.Actor.CredentialID),
		Payload:           event.Payload, OccurredAt: event.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("append comment mutation event: %w", err)
	}
	return nil
}

func optionalUUID(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	copy := value
	return &copy
}

func mapCommentWriteError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return commentsdomain.ErrNotFound
	}
	if platformdatabase.Classify(err) == platformdatabase.ErrorClassForeignKeyViolation {
		return commentsdomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
