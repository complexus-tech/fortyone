package commentsrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) GetStoryState(ctx context.Context, scope commentsdomain.StoryScope) (commentsdomain.StoryState, error) {
	return readStoryState(ctx, r.queries, scope)
}

func readStoryState(ctx context.Context, queries commentsql.Querier, scope commentsdomain.StoryScope) (commentsdomain.StoryState, error) {
	row, err := queries.GetStoryCommentReadState(ctx, commentsql.GetStoryCommentReadStateParams{
		ActorID: scope.Actor.PrincipalID, StoryID: scope.StoryID, WorkspaceID: scope.WorkspaceID,
		TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return commentsdomain.StoryState{}, mapCommentWriteError("get comment read state", err)
	}
	comments, err := readCollaboration(ctx, queries, scope, nil)
	if err != nil {
		return commentsdomain.StoryState{}, err
	}
	threadReads := make(map[uuid.UUID]time.Time)
	if err := json.Unmarshal(row.ThreadReads, &threadReads); err != nil {
		return commentsdomain.StoryState{}, fmt.Errorf("decode comment thread read markers: %w", err)
	}
	return commentsdomain.StoryState{
		LastReadAt: row.LastReadAt, ObservedAt: row.ObservedAt, UnreadCount: row.UnreadCount, Comments: comments,
		LatestCommentID: row.LatestCommentID,
		ThreadReads:     threadReads,
	}, nil
}

func (r *Repository) SetReaction(ctx context.Context, command commentsdomain.ReactionCommand) (commentsdomain.CommentState, error) {
	var result commentsdomain.CommentState
	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		storyID, err := queries.SetCommentReaction(ctx, commentsql.SetCommentReactionParams{
			ActorID: command.Scope.Actor.PrincipalID, WorkspaceID: command.Scope.WorkspaceID, CommentID: command.Scope.CommentID,
			TeamAccessUnrestricted: command.Scope.Actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: command.Scope.Actor.TeamAccess.RestrictedTeamIDs(),
			Emoji: command.Emoji, Active: command.Active,
		})
		if err != nil {
			return mapCommentWriteError("set comment reaction", err)
		}
		result, err = readCommentState(ctx, queries, command.Scope, storyID)
		return err
	})
	return result, err
}

func (r *Repository) SetResolution(ctx context.Context, command commentsdomain.ResolutionCommand) (commentsdomain.CommentState, error) {
	var result commentsdomain.CommentState
	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		storyID, err := queries.SetCommentResolution(ctx, commentsql.SetCommentResolutionParams{
			ActorID: command.Scope.Actor.PrincipalID, WorkspaceID: command.Scope.WorkspaceID, CommentID: command.Scope.CommentID,
			TeamAccessUnrestricted: command.Scope.Actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: command.Scope.Actor.TeamAccess.RestrictedTeamIDs(),
			Resolved: command.Resolved,
		})
		if err != nil {
			return mapCommentWriteError("set comment resolution", err)
		}
		result, err = readCommentState(ctx, queries, command.Scope, storyID)
		return err
	})
	return result, err
}

func (r *Repository) MarkRead(ctx context.Context, command commentsdomain.ReadCommand) (commentsdomain.StoryState, error) {
	var result commentsdomain.StoryState
	err := r.withinTransaction(ctx, func(queries commentsql.Querier) error {
		var err error
		if command.ThreadID == nil {
			_, err = queries.MarkStoryCommentsRead(ctx, commentsql.MarkStoryCommentsReadParams{
				ActorID: command.Scope.Actor.PrincipalID, WorkspaceID: command.Scope.WorkspaceID, StoryID: command.Scope.StoryID,
				TeamAccessUnrestricted: command.Scope.Actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: command.Scope.Actor.TeamAccess.RestrictedTeamIDs(),
				ObservedAt: command.ObservedAt,
			})
		} else {
			_, err = queries.MarkStoryCommentThreadRead(ctx, commentsql.MarkStoryCommentThreadReadParams{
				ActorID: command.Scope.Actor.PrincipalID, WorkspaceID: command.Scope.WorkspaceID, StoryID: command.Scope.StoryID,
				TeamAccessUnrestricted: command.Scope.Actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: command.Scope.Actor.TeamAccess.RestrictedTeamIDs(),
				ThreadID: *command.ThreadID, ObservedAt: command.ObservedAt,
			})
		}
		if err != nil {
			return mapCommentWriteError("mark story comments read", err)
		}
		result, err = readStoryState(ctx, queries, command.Scope)
		return err
	})
	return result, err
}

func readCommentState(ctx context.Context, queries commentsql.Querier, scope commentsdomain.ActorScope, storyID uuid.UUID) (commentsdomain.CommentState, error) {
	states, err := readCollaboration(ctx, queries, commentsdomain.StoryScope{
		StoryID: storyID, WorkspaceID: scope.WorkspaceID, Actor: scope.Actor,
	}, &scope.CommentID)
	if err != nil {
		return commentsdomain.CommentState{}, err
	}
	if len(states) != 1 {
		return commentsdomain.CommentState{}, commentsdomain.ErrNotFound
	}
	return states[0], nil
}

func readCollaboration(ctx context.Context, queries commentsql.Querier, scope commentsdomain.StoryScope, commentID *uuid.UUID) ([]commentsdomain.CommentState, error) {
	rows, err := queries.ListCommentCollaboration(ctx, commentsql.ListCommentCollaborationParams{
		ActorID: scope.Actor.PrincipalID, StoryID: scope.StoryID, WorkspaceID: scope.WorkspaceID, CommentID: commentID,
		TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return nil, fmt.Errorf("read comment collaboration: %w", err)
	}
	states := make([]commentsdomain.CommentState, 0)
	for _, row := range rows {
		if len(states) == 0 || states[len(states)-1].CommentID != row.CommentID {
			states = append(states, commentsdomain.CommentState{
				CommentID: row.CommentID, ResolvedAt: row.ResolvedAt, ResolvedBy: row.ResolvedBy, Reactions: []commentsdomain.Reaction{},
			})
		}
		if row.Emoji != nil {
			last := &states[len(states)-1]
			last.Reactions = append(last.Reactions, commentsdomain.Reaction{Emoji: *row.Emoji, Count: row.ReactionCount, Reacted: row.Reacted})
		}
	}
	return states, nil
}
