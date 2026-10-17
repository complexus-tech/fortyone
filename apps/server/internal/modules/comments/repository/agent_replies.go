package commentsrepository

import (
	"context"
	"fmt"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) CreateAgentReply(ctx context.Context, command commentsdomain.AgentReplyCommand, eventID uuid.UUID) (commentsdomain.Comment, error) {
	return r.createComment(ctx, commentsdomain.CreateCommand{
		CreationID: &command.CreationID, WorkspaceID: command.Scope.WorkspaceID, StoryID: command.Scope.StoryID,
		ParentID: &command.Scope.ParentID, Actor: command.Scope.Author,
		Content: command.Content, MentionedUserIDs: command.MentionedUserIDs,
	}, eventID, &command.Scope)
}

func (r *Repository) ResolveAgentMentions(ctx context.Context, scope commentsdomain.AgentReplyScope, ids []uuid.UUID) ([]commentsdomain.MentionTarget, error) {
	var targets []commentsdomain.MentionTarget
	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		if err := authorizeAgentReply(ctx, queries, scope); err != nil {
			return err
		}
		var err error
		targets, err = resolveAgentMentions(ctx, queries, scope, ids)
		return err
	})
	return targets, err
}

func (r *Repository) GetAgentReply(ctx context.Context, scope commentsdomain.AgentReplyScope, id uuid.UUID) (commentsdomain.AgentReply, error) {
	var reply commentsdomain.AgentReply
	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		if err := authorizeAgentReply(ctx, queries, scope); err != nil {
			return err
		}
		row, err := queries.GetCommentCreationForActor(ctx, commentsql.GetCommentCreationForActorParams{
			CreationID: id, StoryID: scope.StoryID, ActorID: scope.Author.PrincipalID, WorkspaceID: scope.WorkspaceID,
			SystemActor: true, TeamAccessUnrestricted: scope.Author.TeamAccess.IsUnrestricted(), AllowedTeamIds: scope.Author.TeamAccess.RestrictedTeamIDs(),
		})
		if err != nil {
			return mapCommentWriteError("read agent reply", err)
		}
		if row.ParentID == nil || *row.ParentID != scope.ParentID {
			return commentsdomain.ErrNotFound
		}
		ids, err := queries.GetCommentMentionIDs(ctx, commentsql.GetCommentMentionIDsParams{CommentID: id})
		if err != nil {
			return fmt.Errorf("read agent reply mentions: %w", err)
		}
		targets, err := resolveAgentMentions(ctx, queries, scope, ids)
		if err != nil {
			return err
		}
		reply.Comment = commentFromValues(row.CommentID, row.StoryID, row.ParentID, row.CommenterID, row.Content, row.CreatedAt, row.UpdatedAt)
		for _, target := range targets {
			reply.MentionedUserIDs = append(reply.MentionedUserIDs, target.ID)
		}
		return nil
	})
	return reply, err
}

func authorizeAgentReply(ctx context.Context, queries commentsql.Querier, scope commentsdomain.AgentReplyScope) error {
	_, err := queries.AuthorizeAgentReply(ctx, commentsql.AuthorizeAgentReplyParams{
		WorkspaceID: scope.WorkspaceID, StoryID: scope.StoryID, ParentID: scope.ParentID,
		RequesterID: scope.Requester.PrincipalID, AuthorID: scope.Author.PrincipalID,
		TeamAccessUnrestricted: scope.Requester.TeamAccess.IsUnrestricted(), AllowedTeamIds: scope.Requester.TeamAccess.RestrictedTeamIDs(),
		AuthorTeamAccessUnrestricted: scope.Author.TeamAccess.IsUnrestricted(), AuthorAllowedTeamIds: scope.Author.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return mapCommentWriteError("authorize agent reply", err)
	}
	return nil
}

func resolveAgentMentions(ctx context.Context, queries commentsql.Querier, scope commentsdomain.AgentReplyScope, ids []uuid.UUID) ([]commentsdomain.MentionTarget, error) {
	rows, err := queries.ResolveAgentMentionTargets(ctx, commentsql.ResolveAgentMentionTargetsParams{
		WorkspaceID: scope.WorkspaceID, StoryID: scope.StoryID, MentionedUserIds: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve agent mentions: %w", err)
	}
	targets := make([]commentsdomain.MentionTarget, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, commentsdomain.MentionTarget{ID: row.UserID, Name: row.Name})
	}
	return targets, nil
}

func insertAgentMentions(ctx context.Context, queries commentsql.Querier, scope commentsdomain.AgentReplyScope, id uuid.UUID, ids []uuid.UUID) error {
	count, err := queries.InsertAgentCommentMentions(ctx, commentsql.InsertAgentCommentMentionsParams{
		CommentID: id, WorkspaceID: scope.WorkspaceID, StoryID: scope.StoryID, ParentID: scope.ParentID,
		RequesterID: scope.Requester.PrincipalID, AuthorID: scope.Author.PrincipalID, MentionedUserIds: ids,
		TeamAccessUnrestricted: scope.Requester.TeamAccess.IsUnrestricted(), AllowedTeamIds: scope.Requester.TeamAccess.RestrictedTeamIDs(),
		AuthorTeamAccessUnrestricted: scope.Author.TeamAccess.IsUnrestricted(), AuthorAllowedTeamIds: scope.Author.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return fmt.Errorf("insert agent mentions: %w", err)
	}
	if count != int64(len(ids)) {
		return commentsdomain.ErrInvalidMention
	}
	return nil
}
