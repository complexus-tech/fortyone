package agentadapter

import (
	"context"
	"encoding/json"
	"errors"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (r *CommentRuntime) replyScope(ctx context.Context, run agentruns.Run) (comments.AgentReplyScope, error) {
	ctx, err := humanContext(ctx, run)
	if err != nil {
		return comments.AgentReplyScope{}, err
	}
	requester, err := auth.GetActor(ctx)
	if err != nil {
		return comments.AgentReplyScope{}, err
	}
	author, err := auth.NewActor(r.MayaID, auth.PrincipalSystem, uuid.Nil, auth.MustScopeSet(auth.ScopeCommentsWrite), requester.TeamAccess)
	if err != nil {
		return comments.AgentReplyScope{}, err
	}
	author, err = author.WithWorkspace(run.WorkspaceID)
	if err != nil {
		return comments.AgentReplyScope{}, err
	}
	return comments.AgentReplyScope{WorkspaceID: run.WorkspaceID, StoryID: run.EntityID, ParentID: run.SourceID, Requester: requester, Author: author}, nil
}

func (r *CommentRuntime) Deliver(ctx context.Context, run agentruns.Run, _ agentruns.Outcome, replyID uuid.UUID) error {
	if r.NotificationRules == nil || r.Notifications == nil || r.Comments == nil {
		return errors.New("Maya notification delivery is unavailable")
	}
	ctx, err := humanContext(ctx, run)
	if err != nil {
		return err
	}
	scope, err := r.replyScope(ctx, run)
	if err != nil {
		return err
	}
	// Use the committed comment and recipients. Re-rendering model output here
	// could notify a truncated mention or change content after a member rename.
	reply, err := r.Comments.GetAgentReply(ctx, scope, replyID)
	if err != nil {
		return err
	}
	audience, err := r.Stories.GetNotificationAudience(ctx, run.EntityID, run.WorkspaceID)
	if err != nil {
		return err
	}
	var payload commentPayload
	if err := json.Unmarshal(run.Payload, &payload); err != nil {
		return err
	}
	inputs, err := r.NotificationRules.ProcessCommentReplied(ctx, events.CommentRepliedPayload{
		CommentID: replyID, ParentCommentID: run.SourceID, ParentAuthorID: run.RequesterID,
		StoryID: run.EntityID, StoryTitle: payload.StoryTitle, WorkspaceID: run.WorkspaceID, Content: reply.Comment.Comment,
		AudienceIDs: audience, AudienceResolved: true, Mentions: reply.MentionedUserIDs,
	}, r.MayaID)
	if err != nil {
		return err
	}
	for _, recipient := range reply.MentionedUserIDs {
		mentioned, err := r.NotificationRules.ProcessUserMentioned(ctx, events.UserMentionedPayload{
			CommentID: replyID, StoryID: run.EntityID, StoryTitle: payload.StoryTitle, WorkspaceID: run.WorkspaceID,
			MentionedUser: recipient, Content: reply.Comment.Comment,
		}, r.MayaID)
		if err != nil {
			return err
		}
		inputs = append(inputs, mentioned...)
	}
	for _, input := range inputs {
		prefix := "maya-reply:"
		if input.Type == "mention" {
			prefix = "maya-mention:"
		}
		input.DedupeKey = prefix + replyID.String() + ":" + input.RecipientID.String()
		actor := input.Message.Variables["actor"]
		actor.Value = "Maya"
		input.Message.Variables["actor"] = actor
		input.OccurredAt = run.CreatedAt
		if input.Type == "comment_reply" {
			if err := r.Notifications.DeferMayaReply(ctx, input, replyID); err != nil {
				return err
			}
			continue
		}
		if _, err := r.Notifications.Create(ctx, input); err != nil {
			return err
		}
	}
	return nil
}
