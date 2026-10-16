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
			CreationID: command.CreationID,
			Content:    command.Content, ActorID: command.Actor.PrincipalID,
			SystemActor: command.Actor.Kind == platformauth.PrincipalSystem,
			ParentID:    command.ParentID, StoryID: command.StoryID, WorkspaceID: command.WorkspaceID,
			TeamAccessUnrestricted: command.Actor.TeamAccess.IsUnrestricted(),
			AllowedTeamIds:         command.Actor.TeamAccess.RestrictedTeamIDs(),
		})
		if errors.Is(err, pgx.ErrNoRows) && command.CreationID != nil {
			return replayCommentCreation(ctx, queries, command, &created)
		}
		if err != nil {
			return mapCommentWriteError("create comment", err)
		}
		created = commentFromValues(
			row.CommentID, row.StoryID, row.ParentID, row.CommenterID,
			row.Content, row.CreatedAt, row.UpdatedAt,
		)
		if err := syncCommentAttachments(ctx, queries, created, command.WorkspaceID); err != nil {
			return err
		}
		scope := commentsdomain.ActorScope{
			CommentID: created.ID, WorkspaceID: command.WorkspaceID, Actor: command.Actor,
		}
		// A newly inserted comment has no mentions to clear. In particular, a
		// provider comment must not enter the human-author mention mutation path.
		if len(command.MentionedUserIDs) > 0 {
			if _, err := replaceMentions(ctx, queries, scope, command.MentionedUserIDs); err != nil {
				return err
			}
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

func replayCommentCreation(ctx context.Context, queries commentsql.Querier, command commentsdomain.CreateCommand, created *commentsdomain.Comment) error {
	row, err := queries.GetCommentCreationForActor(ctx, commentsql.GetCommentCreationForActorParams{
		CreationID: *command.CreationID, StoryID: command.StoryID,
		ActorID: command.Actor.PrincipalID, WorkspaceID: command.WorkspaceID,
		SystemActor:            command.Actor.Kind == platformauth.PrincipalSystem,
		TeamAccessUnrestricted: command.Actor.TeamAccess.IsUnrestricted(),
		AllowedTeamIds:         command.Actor.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return mapCommentWriteError("replay comment creation", err)
	}
	if row.Content != command.Content || !sameOptionalUUID(row.ParentID, command.ParentID) {
		return commentsdomain.ErrCreationConflict
	}
	mentions, err := queries.GetCommentMentionIDs(ctx, commentsql.GetCommentMentionIDsParams{CommentID: row.CommentID})
	if err != nil {
		return fmt.Errorf("read replayed comment mentions: %w", err)
	}
	if !sameMentionIDs(mentions, command.MentionedUserIDs) {
		return commentsdomain.ErrCreationConflict
	}
	*created = commentFromValues(row.CommentID, row.StoryID, row.ParentID, row.CommenterID, row.Content, row.CreatedAt, row.UpdatedAt)
	created.Replayed = true
	return nil
}

func sameMentionIDs(left, right []uuid.UUID) bool {
	leftSet := make(map[uuid.UUID]struct{}, len(left))
	for _, id := range left {
		leftSet[id] = struct{}{}
	}
	rightSet := make(map[uuid.UUID]struct{}, len(right))
	for _, id := range right {
		rightSet[id] = struct{}{}
	}
	if len(leftSet) != len(rightSet) {
		return false
	}
	for id := range leftSet {
		if _, found := rightSet[id]; !found {
			return false
		}
	}
	return true
}

func sameOptionalUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func (r *Repository) UpdateComment(
	ctx context.Context,
	command commentsdomain.UpdateCommand,
	eventID uuid.UUID,
) (commentsdomain.UpdateResult, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.comments.UpdateComment")
	defer span.End()

	var result commentsdomain.UpdateResult
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
		if err := syncCommentAttachments(ctx, queries, updated, command.Scope.WorkspaceID); err != nil {
			return err
		}
		addedMentions, err := replaceMentions(ctx, queries, command.Scope, command.MentionedUserIDs)
		if err != nil {
			return err
		}
		result = commentsdomain.UpdateResult{Comment: updated, StoryTitle: row.StoryTitle, AddedMentionIDs: addedMentions}
		return appendMutationEvent(
			ctx, queries, eventID, commentsdomain.EventUpdated,
			command.Scope.WorkspaceID, command.Scope.Actor, updated,
		)
	})
	if err != nil {
		r.log.Error(ctx, "error updating comment", "error", err, "comment_id", command.Scope.CommentID)
		return commentsdomain.UpdateResult{}, err
	}

	r.log.Info(ctx, "comment updated successfully", "comment_id", command.Scope.CommentID)
	return result, nil
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
) ([]uuid.UUID, error) {
	params := commentsql.DeleteCommentMentionsForAuthorParams{
		CommentID: scope.CommentID, ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID,
		TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(),
		AllowedTeamIds:         scope.Actor.TeamAccess.RestrictedTeamIDs(),
	}
	cleared, err := queries.DeleteCommentMentionsForAuthor(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("delete comment mentions: %w", err)
	}
	if !cleared.CommentFound {
		return nil, commentsdomain.ErrNotFound
	}
	if len(mentionedUserIDs) == 0 {
		return nil, nil
	}

	inserted, err := queries.InsertCommentMentionsForAuthor(ctx, commentsql.InsertCommentMentionsForAuthorParams{
		CommentID: scope.CommentID, ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID,
		TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(),
		AllowedTeamIds:         scope.Actor.TeamAccess.RestrictedTeamIDs(),
		MentionedUserIds:       mentionedUserIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("insert comment mentions: %w", err)
	}
	if inserted != int64(len(mentionedUserIDs)) {
		return nil, commentsdomain.ErrInvalidMention
	}
	previous := make(map[uuid.UUID]struct{}, len(cleared.MentionedUserIds))
	for _, userID := range cleared.MentionedUserIds {
		previous[userID] = struct{}{}
	}
	added := make([]uuid.UUID, 0, len(mentionedUserIDs))
	for _, userID := range mentionedUserIDs {
		if _, exists := previous[userID]; !exists {
			added = append(added, userID)
			previous[userID] = struct{}{}
		}
	}
	return added, nil
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
