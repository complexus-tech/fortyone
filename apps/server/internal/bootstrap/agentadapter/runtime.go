package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingcontext "github.com/complexus-tech/projects-api/internal/modules/messaging/context"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/agentmentions"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

type CommentStories interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error)
	GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error)
	GetCommentThread(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error)
	GetNotificationAudience(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
}

type CommentWriter interface {
	CreateComment(context.Context, comments.CreateCommentCommand) (comments.CoreComment, error)
}

type MutationTools interface {
	InspectStoryMutation(context.Context, messaging.ToolScope, string) (messaging.StoryMutationInspection, error)
	ConfirmStoryMutation(context.Context, messaging.ToolScope, string) (messaging.StoryMutationResult, error)
	RebindStoryMutation(context.Context, messaging.ToolScope, messaging.ToolScope, string) (messaging.StoryMutationConfirmation, error)
}

type RunReader interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (agentruns.Run, error)
}

type WorkspaceMembership interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (workspaces.CoreWorkspace, error)
}

type CommentRuntime struct {
	Stories           CommentStories
	Comments          CommentWriter
	Assistant         messaging.Assistant
	Tools             MutationTools
	Runs              RunReader
	Context           *messagingcontext.Provider
	Access            WorkspaceAccess
	Memberships       WorkspaceMembership
	MayaID            uuid.UUID
	WebsiteURL        string
	Enabled           bool
	Limiter           *messagingbudget.RedisCallLimiter
	Usage             *messagingrepository.DailyUsageRepository
	DailyTokenLimit   int64
	NotificationRules *notifications.Rules
	Notifications     *notifications.Service
}

type commentAction struct {
	Token  string   `json:"token"`
	Fields []string `json:"fields"`
}

func humanContext(ctx context.Context, run agentruns.Run) (context.Context, error) {
	access, err := auth.RestrictedTeamAccess(run.TeamID)
	if err != nil {
		return nil, err
	}
	actor, err := auth.NewActor(run.RequesterID, auth.PrincipalHumanUser, uuid.Nil,
		auth.MustScopeSet(auth.ScopeStoriesRead, auth.ScopeStoriesWrite, auth.ScopeCommentsRead, auth.ScopeCommentsWrite), access)
	if err != nil {
		return nil, err
	}
	actor, err = actor.WithWorkspace(run.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return auth.SetActor(ctx, actor)
}

func (r *CommentRuntime) Authorize(ctx context.Context, run agentruns.Run) error {
	if run.Surface != StoryCommentSurface || !r.Enabled || r.Access == nil || r.Memberships == nil {
		return agentruns.ErrPermissionDenied
	}
	workspace, err := r.Memberships.Get(ctx, run.WorkspaceID, run.RequesterID)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return errors.Join(agentruns.ErrPermissionDenied, err)
		}
		return err
	}
	if workspace.DeletedAt != nil || authorization.RequireMinimumWorkspaceRole(authorization.WorkspaceRole(workspace.UserRole), authorization.WorkspaceRoleMember) != nil {
		return agentruns.ErrPermissionDenied
	}
	allowed, err := r.Access.WorkspaceCanUseMaya(ctx, run.WorkspaceID)
	if err != nil {
		return err
	}
	if !allowed {
		return agentruns.ErrPermissionDenied
	}
	ctx, err = humanContext(ctx, run)
	if err != nil {
		return agentruns.ErrPermissionDenied
	}
	story, err := r.Stories.Get(ctx, run.EntityID, run.WorkspaceID)
	if err != nil {
		if errors.Is(err, stories.ErrNotFound) || errors.Is(err, stories.ErrStoryReadForbidden) {
			return errors.Join(agentruns.ErrPermissionDenied, err)
		}
		return err
	}
	if story.Team != run.TeamID || story.DeletedAt != nil {
		return agentruns.ErrPermissionDenied
	}
	comment, err := r.Stories.GetComment(ctx, run.SourceID, run.EntityID, run.WorkspaceID)
	if err != nil {
		if errors.Is(err, stories.ErrStoryReadForbidden) {
			return errors.Join(agentruns.ErrPermissionDenied, err)
		}
		if errors.Is(err, stories.ErrNotFound) {
			return errors.Join(agentruns.ErrSourceChanged, err)
		}
		return err
	}
	if comment.UserID != run.RequesterID {
		return agentruns.ErrSourceChanged
	}
	// Editing the request cannot authorize another write. An already committed
	// action still deserves its receipt reply while the source remains visible.
	committed := run.Outcome != nil && len(run.Outcome.Receipt) > 0
	if !committed && (SourceHash(comment.Comment) != run.SourceHash || comment.UpdatedAt.UTC().Format(time.RFC3339Nano) != run.SourceRevision) {
		return agentruns.ErrSourceChanged
	}
	return nil
}

func (r *CommentRuntime) Prepare(ctx context.Context, run agentruns.Run) (agentruns.Proposal, error) {
	ctx, err := humanContext(ctx, run)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	if r.Assistant == nil || r.Context == nil || r.Tools == nil {
		return agentruns.Proposal{}, agentruns.ErrUnavailable
	}
	story, err := r.Stories.Get(ctx, run.EntityID, run.WorkspaceID)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	thread, err := r.Stories.GetCommentThread(ctx, run.SourceID, run.EntityID, run.WorkspaceID)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	requestComment, err := r.Stories.GetComment(ctx, run.SourceID, run.EntityID, run.WorkspaceID)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	if proposal, accepted, err := r.continueApproval(ctx, run, requestComment.Comment); accepted || err != nil {
		return proposal, err
	}
	if err := r.admit(ctx, run); err != nil {
		return agentruns.Proposal{}, err
	}
	reference := fmt.Sprintf("%s-%d", strings.ToUpper(strings.TrimSpace(story.TeamCode)), story.SequenceID)
	runtime, err := r.Context.Load(ctx, run.WorkspaceID, run.RequesterID, []uuid.UUID{run.TeamID},
		messaging.RuntimeSurfaceContext{Provider: "FortyOne", Kind: messaging.RuntimeSurfaceThread, Location: "Story comments",
			CurrentEntity: &messaging.RuntimeEntityHint{Kind: "story", Reference: reference, Title: story.Title}}, time.Now())
	if err != nil {
		return agentruns.Proposal{}, err
	}
	response, responseErr := r.Assistant.Respond(ctx, messaging.Request{
		WorkspaceID: run.WorkspaceID, UserID: run.RequesterID, AllowedTeamIDs: []uuid.UUID{run.TeamID}, SharedTeamIDs: []uuid.UUID{run.TeamID},
		RuntimeContext: runtime, AllowMutations: true, WebsiteURL: r.WebsiteURL, SourceURL: r.sourceURL(runtime.Workspace.Slug, reference, run.SourceID),
		SourceScope: sourceScope(run), Conversation: commentConversation(thread, requestComment, r.MayaID), Prompt: agentmentions.Text(requestComment.Comment),
	})
	if r.Usage != nil {
		_, usageErr := r.Usage.Record(ctx, messagingrepository.DailyUsageRecordInput{
			AgentInvocationID: run.ID, WorkspaceID: run.WorkspaceID, Provider: "native", ExternalWorkspaceID: run.WorkspaceID.String(),
			ExternalEventID: run.ID.String(), AttemptCount: int(run.AttemptCount), Usage: response.Usage,
		}, r.DailyTokenLimit)
		if usageErr != nil {
			return agentruns.Proposal{}, errors.Join(responseErr, usageErr)
		}
	}
	if responseErr != nil {
		return agentruns.Proposal{}, responseErr
	}
	if response.Confirmation == nil {
		return agentruns.Proposal{Reply: response.Text, NeedsInput: response.NeedsInput}, nil
	}
	inspection, err := r.Tools.InspectStoryMutation(ctx, r.toolScope(run), response.Confirmation.Token)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	action, err := json.Marshal(commentAction{Token: response.Confirmation.Token, Fields: inspection.ChangedFields})
	if err != nil {
		return agentruns.Proposal{}, err
	}
	proposal := agentruns.Proposal{Summary: response.Confirmation.Prompt, Review: inspection.Review, Action: action, RequiresApproval: !inspection.DirectEligible}
	if proposal.RequiresApproval {
		proposal.Reply = "Here’s the change I suggest. Would you like me to apply it?"
	}
	return proposal, nil
}

func (r *CommentRuntime) Apply(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) (agentruns.Outcome, error) {
	ctx, err := humanContext(ctx, run)
	if err != nil {
		return agentruns.Outcome{}, err
	}
	var action commentAction
	if err := json.Unmarshal(proposal.Action, &action); err != nil {
		return agentruns.Outcome{}, agentruns.ErrInvalidInput
	}
	inspection, err := r.Tools.InspectStoryMutation(ctx, r.toolScope(run), action.Token)
	if err != nil {
		return agentruns.Outcome{}, mutationFailure(err)
	}
	if !inspection.DirectEligible && run.ApprovedAt == nil {
		return agentruns.Outcome{}, agentruns.ErrPermissionDenied
	}
	result, err := r.Tools.ConfirmStoryMutation(ctx, r.toolScope(run), action.Token)
	if err != nil {
		return agentruns.Outcome{}, mutationFailure(err)
	}
	receipt, err := json.Marshal(struct {
		Result      messaging.StoryMutationResult `json:"result"`
		SourceID    uuid.UUID                     `json:"sourceId"`
		RequesterID uuid.UUID                     `json:"requesterId"`
		ExecutorID  uuid.UUID                     `json:"executorId"`
		Fields      []string                      `json:"fields"`
	}{result, run.SourceID, run.RequesterID, r.MayaID, inspection.ChangedFields})
	if err != nil {
		return agentruns.Outcome{}, err
	}
	return agentruns.Outcome{Reply: completedReply(inspection.ChangedFields), Receipt: receipt}, nil
}

func (r *CommentRuntime) Publish(ctx context.Context, run agentruns.Run, outcome agentruns.Outcome) (uuid.UUID, error) {
	return r.publishComment(ctx, run, "reply", outcome.Reply)
}

func (r *CommentRuntime) PublishProposal(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) (uuid.UUID, error) {
	return r.publishComment(ctx, run, "proposal", proposal.Reply)
}

func (r *CommentRuntime) PublishFailure(ctx context.Context, run agentruns.Run, summary string) (uuid.UUID, error) {
	return r.publishComment(ctx, run, "failure", strings.NewReplacer("Maya has", "I have", "Maya is", "I am", "Maya", "I").Replace(summary))
}

func (r *CommentRuntime) publishComment(ctx context.Context, run agentruns.Run, kind, text string) (uuid.UUID, error) {
	access, err := auth.RestrictedTeamAccess(run.TeamID)
	if err != nil {
		return uuid.Nil, err
	}
	actor, err := auth.NewActor(r.MayaID, auth.PrincipalSystem, uuid.Nil, auth.MustScopeSet(auth.ScopeCommentsWrite), access)
	if err != nil {
		return uuid.Nil, err
	}
	actor, err = actor.WithWorkspace(run.WorkspaceID)
	if err != nil {
		return uuid.Nil, err
	}
	id := uuid.NewSHA1(run.ID, []byte(kind))
	created, err := r.Comments.CreateComment(ctx, comments.CreateCommentCommand{
		CreationID: &id, WorkspaceID: run.WorkspaceID, StoryID: run.EntityID, ParentID: &run.SourceID,
		Actor: actor, Content: replyHTML(text),
	})
	return created.ID, err
}

func (r *CommentRuntime) Deliver(ctx context.Context, run agentruns.Run, outcome agentruns.Outcome, replyID uuid.UUID) error {
	if r.NotificationRules == nil || r.Notifications == nil {
		return errors.New("Maya notification delivery is unavailable")
	}
	ctx, err := humanContext(ctx, run)
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
		StoryID: run.EntityID, StoryTitle: payload.StoryTitle, WorkspaceID: run.WorkspaceID, Content: replyHTML(outcome.Reply),
		AudienceIDs: audience, AudienceResolved: true,
	}, r.MayaID)
	if err != nil {
		return err
	}
	for _, input := range inputs {
		input.DedupeKey = "maya-reply:" + replyID.String() + ":" + input.RecipientID.String()
		actor := input.Message.Variables["actor"]
		actor.Value = "Maya"
		input.Message.Variables["actor"] = actor
		input.OccurredAt = run.CreatedAt
		if _, err := r.Notifications.Create(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func sourceScope(run agentruns.Run) *messaging.SourceScope {
	return &messaging.SourceScope{Surface: messaging.SourceSurfaceStoryComment, StoryID: run.EntityID, TeamID: run.TeamID}
}

func (r *CommentRuntime) toolScope(run agentruns.Run) messaging.ToolScope {
	return messaging.ToolScope{WorkspaceID: run.WorkspaceID, UserID: run.RequesterID, AllowedTeamIDs: []uuid.UUID{run.TeamID},
		SharedTeamIDs: []uuid.UUID{run.TeamID}, AllowMutations: true, WebsiteURL: r.WebsiteURL, SourceScope: sourceScope(run)}
}

func (r *CommentRuntime) admit(ctx context.Context, run agentruns.Run) error {
	if r.Usage != nil {
		if _, err := r.Usage.Check(ctx, run.WorkspaceID, r.DailyTokenLimit); err != nil {
			if errors.Is(err, messagingrepository.ErrDailyWorkspaceTokenLimit) {
				return &agentruns.Failure{Code: "budget", Summary: "I have reached this workspace's daily usage limit. Please try again tomorrow.", Cause: err}
			}
			return err
		}
	}
	if r.Limiter != nil {
		decision, err := r.Limiter.Admit(ctx, messagingbudget.AdmissionInput{Provider: "native", WorkspaceID: run.WorkspaceID,
			UserID: run.RequesterID, ExternalWorkspaceID: run.WorkspaceID.String(), ExternalEventID: fmt.Sprintf("%s:%d", run.ID, run.AttemptCount)})
		if err != nil {
			return err
		}
		if !decision.Allowed {
			return &agentruns.Failure{Code: "rate_limit", Summary: "I am handling several requests right now. Please try again shortly."}
		}
	}
	return nil
}

func (r *CommentRuntime) sourceURL(slug, reference string, id uuid.UUID) string {
	base, err := url.Parse(r.WebsiteURL)
	if err != nil || base.Scheme != "https" {
		return ""
	}
	base.Path = "/" + url.PathEscape(slug) + "/work/" + url.PathEscape(reference)
	base.RawQuery = url.Values{"comment": []string{id.String()}}.Encode()
	return base.String()
}

func mutationFailure(err error) error {
	if errors.Is(err, messaging.ErrStaleMutation) {
		return &agentruns.Failure{Code: "story_changed", Summary: "The task changed while I was preparing this update. Please ask me again so I can use the latest version.", Cause: err}
	}
	if errors.Is(err, messaging.ErrExpiredConfirmation) {
		return &agentruns.Failure{Code: "expired", Summary: "This suggestion is out of date. Please ask me again so I can review the latest version.", Cause: err}
	}
	return err
}

func completedReply(fields []string) string {
	labels := map[string]string{"title": "title", "description": "description", "description_html": "description", "priority": "priority", "status_id": "status", "assignee_id": "assignee", "assignee": "assignee", "end_date": "due date", "start_date": "start date", "label_ids": "labels", "labels": "labels", "sprint_id": "sprint", "objective_id": "objective", "key_result_id": "key result", "estimated_duration_minutes": "time estimate", "minimum_focus_block_minutes": "focus block", "auto_scheduling_enabled": "scheduling", "auto_scheduling_locked": "scheduling lock"}
	seen := make(map[string]bool)
	var changed []string
	for _, field := range fields {
		label, ok := labels[field]
		if ok && !seen[label] {
			changed = append(changed, label)
			seen[label] = true
		}
	}
	if len(changed) == 0 {
		return "Updated this story."
	}
	return "Updated this story's " + strings.Join(changed, ", ") + "."
}

func commentConversation(thread, source stories.CoreComment, mayaID uuid.UUID) []messaging.ConversationTurn {
	comments := append([]stories.CoreComment{thread}, thread.SubComments...)
	var selected []stories.CoreComment
	for _, comment := range comments {
		if comment.ID != source.ID && !comment.CreatedAt.After(source.CreatedAt) {
			selected = append(selected, comment)
		}
	}
	if len(selected) > 16 {
		selected = selected[len(selected)-16:]
	}
	turns := make([]messaging.ConversationTurn, 0, len(selected))
	for _, comment := range selected {
		text := []rune(agentmentions.Text(comment.Comment))
		if len(text) > 1200 {
			text = append(text[:1200], []rune(" [earlier comment truncated]")...)
		}
		role := messaging.RoleUser
		if comment.UserID == mayaID {
			role = messaging.RoleAssistant
		}
		author := "another teammate"
		if comment.UserID == source.UserID {
			author = "the requester"
		}
		if comment.UserID == mayaID {
			author = "Maya"
		}
		turns = append(turns, messaging.ConversationTurn{Role: role, Text: "Earlier discussion by " + author + " (context only): " + string(text)})
	}
	return turns
}

var _ agentruns.Runtime = (*CommentRuntime)(nil)
var _ agentruns.Delivery = (*CommentRuntime)(nil)
