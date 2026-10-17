package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type ChatMutationTools interface {
	InspectChatMutation(context.Context, messaging.ToolScope, string) (messaging.StoryMutationInspection, error)
	ConfirmStoryMutation(context.Context, messaging.ToolScope, string) (messaging.StoryMutationResult, error)
}

type ChatRuntime struct {
	Rollout chatrollout.Policy
	Chats   ChatMessages
	Shared  *CommentRuntime
	Tools   ChatMutationTools
	Log     *logger.Logger
}

func chatHumanContext(ctx context.Context, run agentruns.Run) (context.Context, chatdomain.ActorScope, error) {
	access, err := auth.RestrictedTeamAccess(run.TeamID)
	if err != nil {
		return ctx, chatdomain.ActorScope{}, err
	}
	actor, err := auth.NewActor(run.RequesterID, auth.PrincipalHumanUser, uuid.Nil,
		auth.MustScopeSet(auth.ScopeStoriesRead, auth.ScopeStoriesWrite, auth.ScopeChatRead, auth.ScopeChatWrite), access)
	if err != nil {
		return ctx, chatdomain.ActorScope{}, err
	}
	actor, err = actor.WithWorkspace(run.WorkspaceID)
	if err != nil {
		return ctx, chatdomain.ActorScope{}, err
	}
	ctx, err = auth.SetActor(ctx, actor)
	return ctx, chatdomain.ActorScope{WorkspaceID: run.WorkspaceID, Actor: actor}, err
}

func (r *ChatRuntime) Authorize(ctx context.Context, run agentruns.Run) error {
	if err := r.requireEnabled(ctx, run.WorkspaceID); err != nil {
		return err
	}
	if run.Surface != ChatMessageSurface || run.TeamID == uuid.Nil || r.Chats == nil || r.Shared == nil || !r.Shared.Enabled || r.Shared.Access == nil || r.Shared.Memberships == nil {
		return agentruns.ErrPermissionDenied
	}
	workspace, err := r.Shared.Memberships.Get(ctx, run.WorkspaceID, run.RequesterID)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return agentruns.ErrPermissionDenied
		}
		return err
	}
	if workspace.DeletedAt != nil || authorization.RequireMinimumWorkspaceRole(authorization.WorkspaceRole(workspace.UserRole), authorization.WorkspaceRoleMember) != nil {
		return agentruns.ErrPermissionDenied
	}
	allowed, err := r.Shared.Access.WorkspaceCanUseMaya(ctx, run.WorkspaceID)
	if err != nil {
		return err
	}
	if !allowed {
		return agentruns.ErrPermissionDenied
	}
	ctx, scope, err := chatHumanContext(ctx, run)
	if err != nil {
		return agentruns.ErrPermissionDenied
	}
	conversation, err := r.Chats.GetConversation(ctx, scope, run.EntityID)
	if err != nil {
		return chatAccessError(err)
	}
	if conversation.Kind != "team" || conversation.TeamID == nil || *conversation.TeamID != run.TeamID {
		return agentruns.ErrPermissionDenied
	}
	source, err := r.Chats.GetMessage(ctx, scope, run.SourceID)
	if err != nil {
		return chatAccessError(err)
	}
	rootID := source.ID
	if source.ThreadRootID != nil {
		rootID = *source.ThreadRootID
	}
	if source.DeletedAt != nil || source.Author.ID != run.RequesterID || source.Author.IsSystem || source.ConversationID != run.EntityID || rootID != run.ConversationID {
		return agentruns.ErrSourceChanged
	}
	committed := run.Outcome != nil && len(run.Outcome.Receipt) > 0
	if !committed && (strconv.Itoa(source.Version) != run.SourceRevision || chatSourceHash(source) != run.SourceHash) {
		return agentruns.ErrSourceChanged
	}
	return nil
}

func (r *ChatRuntime) Prepare(ctx context.Context, run agentruns.Run) (proposal agentruns.Proposal, prepareErr error) {
	addressedFollowup := false
	defer func() {
		if addressedFollowup && prepareErr != nil {
			prepareErr = &addressedChatFailure{cause: prepareErr}
		}
	}()
	if err := r.requireEnabled(ctx, run.WorkspaceID); err != nil {
		return agentruns.Proposal{}, err
	}
	if r.Shared == nil || r.Shared.Assistant == nil || r.Shared.Context == nil || r.Tools == nil {
		return agentruns.Proposal{}, agentruns.ErrUnavailable
	}
	ctx, scope, err := chatHumanContext(ctx, run)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	source, err := r.Chats.GetMessage(ctx, scope, run.SourceID)
	if err != nil {
		return agentruns.Proposal{}, chatAccessError(err)
	}
	conversation, err := r.Chats.GetConversation(ctx, scope, run.EntityID)
	if err != nil {
		return agentruns.Proposal{}, chatAccessError(err)
	}
	history, err := loadChatHistory(ctx, r.Chats, scope, source)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	var payload chatPayload
	if len(run.Payload) > 0 && json.Unmarshal(run.Payload, &payload) != nil {
		return agentruns.Proposal{}, agentruns.ErrInvalidInput
	}
	if payload.Followup && !chatFollowupCandidate(history, source, r.Shared.MayaID) {
		return agentruns.Proposal{Ignore: true}, nil
	}
	if err := r.Shared.admit(ctx, run); err != nil {
		if payload.Followup {
			// Until the addressing decision succeeds, quota/rate limits must not
			// cause Maya to interrupt a human conversation with an error message.
			if r.Log != nil {
				r.Log.Warn(ctx, "Maya chat follow-up admission unavailable; no reply published", "run_id", run.ID)
			}
			return agentruns.Proposal{Ignore: true}, nil
		}
		return agentruns.Proposal{}, err
	}
	turns := chatConversation(history, source, r.Shared.MayaID)
	classificationUsage := messaging.Usage{}
	if payload.Followup {
		classifier, ok := r.Shared.Assistant.(messaging.ChatFollowupClassifier)
		if !ok {
			return agentruns.Proposal{Ignore: true}, nil
		}
		decision, err := classifier.ClassifyChatFollowup(ctx, messaging.ChatFollowupRequest{
			WorkspaceID: run.WorkspaceID, UserID: run.RequesterID, SourceScope: chatSourceScope(run),
			Conversation: turns, Prompt: source.Text, AuthorDisplayName: source.Author.FullName,
		})
		classificationUsage = decision.Usage
		if err != nil || !decision.ShouldRespond {
			if err != nil && r.Log != nil {
				r.Log.Warn(ctx, "Maya chat follow-up classification unavailable; no reply published", "run_id", run.ID)
			}
			// Ambiguous or unavailable classification must never invoke business
			// tools or post an unsolicited answer. Save the decision and usage
			// together before accounting so its retry cannot reclassify or speak.
			return ignoredChatProposal(run, classificationUsage)
		}
		addressedFollowup = true
	}
	kind := messaging.RuntimeSurfaceChannel
	if source.ThreadRootID != nil {
		kind = messaging.RuntimeSurfaceThread
	}
	runtime, err := r.Shared.Context.Load(ctx, run.WorkspaceID, run.RequesterID, []uuid.UUID{run.TeamID}, messaging.RuntimeSurfaceContext{Provider: "FortyOne", Kind: kind, Location: conversation.Name + " chat"}, time.Now())
	if err != nil {
		return agentruns.Proposal{}, errors.Join(err, r.recordUsage(ctx, run, classificationUsage))
	}
	response, responseErr := r.Shared.Assistant.Respond(ctx, messaging.Request{
		WorkspaceID: run.WorkspaceID, UserID: run.RequesterID, AllowedTeamIDs: []uuid.UUID{run.TeamID}, SharedTeamIDs: []uuid.UUID{run.TeamID},
		RuntimeContext: runtime, AllowMutations: true, WebsiteURL: r.Shared.WebsiteURL, SourceURL: r.sourceURL(runtime.Workspace.Slug, run),
		SourceScope: chatSourceScope(run), Conversation: turns, Prompt: source.Text,
	})
	response.Usage.InputTokens += classificationUsage.InputTokens
	response.Usage.OutputTokens += classificationUsage.OutputTokens
	response.Usage.TotalTokens += classificationUsage.TotalTokens
	if usageErr := r.recordUsage(ctx, run, response.Usage); usageErr != nil {
		return agentruns.Proposal{}, errors.Join(responseErr, usageErr)
	}
	if responseErr != nil {
		return agentruns.Proposal{}, responseErr
	}
	if response.Confirmation == nil {
		return agentruns.Proposal{Reply: response.Text, NeedsInput: response.NeedsInput}, nil
	}
	inspection, err := r.Tools.InspectChatMutation(ctx, r.toolScope(run), response.Confirmation.Token)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	action, err := json.Marshal(commentAction{Token: response.Confirmation.Token})
	if err != nil {
		return agentruns.Proposal{}, err
	}
	return agentruns.Proposal{Summary: "Review the proposed tasks", Review: inspection.Review, Action: action, RequiresApproval: true, Reply: "I’ve prepared these tasks for your review. Approve the proposal to create them."}, nil
}

func (r *ChatRuntime) recordUsage(ctx context.Context, run agentruns.Run, usage messaging.Usage) error {
	if r.Shared.Usage == nil {
		return nil
	}
	_, err := r.Shared.Usage.Record(ctx, messagingrepository.DailyUsageRecordInput{AgentInvocationID: run.ID, WorkspaceID: run.WorkspaceID, Provider: "native", ExternalWorkspaceID: run.WorkspaceID.String(), ExternalEventID: run.ID.String(), AttemptCount: int(run.AttemptCount), Usage: usage}, r.Shared.DailyTokenLimit)
	return err
}

func (r *ChatRuntime) Apply(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) (agentruns.Outcome, error) {
	if err := r.requireEnabled(ctx, run.WorkspaceID); err != nil {
		return agentruns.Outcome{}, err
	}
	if run.ApprovedAt == nil {
		return agentruns.Outcome{}, agentruns.ErrPermissionDenied
	}
	ctx, _, err := chatHumanContext(ctx, run)
	if err != nil {
		return agentruns.Outcome{}, err
	}
	var action commentAction
	if json.Unmarshal(proposal.Action, &action) != nil {
		return agentruns.Outcome{}, agentruns.ErrInvalidInput
	}
	if _, err = r.Tools.InspectChatMutation(ctx, r.toolScope(run), action.Token); err != nil {
		return agentruns.Outcome{}, mutationFailure(err)
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
	}{result, run.SourceID, run.RequesterID, r.Shared.MayaID})
	if err != nil {
		return agentruns.Outcome{}, err
	}
	var payload chatPayload
	if err := json.Unmarshal(run.Payload, &payload); err != nil {
		return agentruns.Outcome{}, err
	}
	var reply strings.Builder
	fmt.Fprintf(&reply, "Created %d tasks:", len(result.Items))
	for _, item := range result.Items {
		link := r.taskURL(payload.WorkspaceSlug, item.Reference)
		if link != "" {
			fmt.Fprintf(&reply, "\n- [%s](%s) — %s", item.Reference, link, item.Title)
		} else {
			fmt.Fprintf(&reply, "\n- %s — %s", item.Reference, item.Title)
		}
	}
	return agentruns.Outcome{Reply: reply.String(), Receipt: receipt}, nil
}

func (r *ChatRuntime) Publish(ctx context.Context, run agentruns.Run, outcome agentruns.Outcome) (uuid.UUID, error) {
	return r.publish(ctx, run, "reply", outcome.Reply)
}
func (r *ChatRuntime) PublishProposal(ctx context.Context, run agentruns.Run, proposal agentruns.Proposal) (uuid.UUID, error) {
	return r.publish(ctx, run, "proposal", proposal.Reply)
}
func (r *ChatRuntime) PublishFailure(ctx context.Context, run agentruns.Run, summary string) (uuid.UUID, error) {
	return r.publish(ctx, run, "failure", summary)
}

func (r *ChatRuntime) publish(ctx context.Context, run agentruns.Run, kind, text string) (uuid.UUID, error) {
	if err := r.requireEnabled(ctx, run.WorkspaceID); err != nil {
		return uuid.Nil, err
	}
	ctx, scope, err := chatHumanContext(ctx, run)
	if err != nil {
		return uuid.Nil, err
	}
	author, err := auth.NewActor(r.Shared.MayaID, auth.PrincipalSystem, uuid.Nil, auth.MustScopeSet(auth.ScopeChatWrite), scope.Actor.TeamAccess)
	if err != nil {
		return uuid.Nil, err
	}
	author, err = author.WithWorkspace(run.WorkspaceID)
	if err != nil {
		return uuid.Nil, err
	}
	if runes := []rune(text); len(runes) > 3500 {
		text = string(runes[:3500]) + "…"
	}
	plain, rich, err := chatReplyContent(text)
	if err != nil {
		return uuid.Nil, err
	}
	reply, err := r.Chats.CreateAgentReply(ctx, chatdomain.AgentReplyCommand{WorkspaceID: run.WorkspaceID, ConversationID: run.EntityID, ThreadRootID: run.ConversationID, CreationID: uuid.NewSHA1(run.ID, []byte(kind)), Requester: scope.Actor, Author: author, Text: plain, RichContent: rich})
	return reply.ID, err
}

func chatSourceScope(run agentruns.Run) *messaging.SourceScope {
	return &messaging.SourceScope{Surface: messaging.SourceSurfaceChatMessage, TeamID: run.TeamID, ConversationID: run.EntityID, MessageID: run.SourceID}
}
func (r *ChatRuntime) toolScope(run agentruns.Run) messaging.ToolScope {
	return messaging.ToolScope{WorkspaceID: run.WorkspaceID, UserID: run.RequesterID, AllowedTeamIDs: []uuid.UUID{run.TeamID}, SharedTeamIDs: []uuid.UUID{run.TeamID}, AllowMutations: true, WebsiteURL: r.Shared.WebsiteURL, SourceScope: chatSourceScope(run)}
}

func chatAccessError(err error) error {
	if errors.Is(err, chatdomain.ErrForbidden) || errors.Is(err, chatdomain.ErrNotFound) {
		return errors.Join(agentruns.ErrPermissionDenied, err)
	}
	return err
}

func chatConversation(history []chatdomain.Message, source chatdomain.Message, mayaID uuid.UUID) []messaging.ConversationTurn {
	turns := make([]messaging.ConversationTurn, 0, len(history))
	seen := make(map[uuid.UUID]bool, len(history))
	for _, message := range history {
		if seen[message.ID] || message.ID == source.ID || message.ConversationID != source.ConversationID || message.Sequence >= source.Sequence || message.DeletedAt != nil || strings.TrimSpace(message.Text) == "" {
			continue
		}
		seen[message.ID] = true
		text := message.Text
		if runes := []rune(text); len(runes) > 1600 {
			text = string(runes[:1600]) + " [message truncated]"
		}
		role := messaging.RoleUser
		if message.Author.ID == mayaID {
			role = messaging.RoleAssistant
		}
		turns = append(turns, messaging.ConversationTurn{Role: role, Text: fmt.Sprintf("Earlier message by %s at %s (context only): %s", message.Author.FullName, message.CreatedAt.UTC().Format(time.RFC3339), text)})
	}
	if len(turns) > messaging.MaximumConversationTurns {
		turns = turns[len(turns)-messaging.MaximumConversationTurns:]
	}
	return turns
}

func (r *ChatRuntime) baseURL(slug string) *url.URL {
	base, err := url.Parse(r.Shared.WebsiteURL)
	if err != nil || base.Scheme != "https" || base.Host == "" || slug == "" {
		return nil
	}
	base.RawQuery = ""
	base.Fragment = ""
	base.Path = "/" + slug
	return base
}
func (r *ChatRuntime) sourceURL(slug string, run agentruns.Run) string {
	base := r.baseURL(slug)
	if base == nil {
		return ""
	}
	base.Path += "/chat"
	query := url.Values{"conversation": []string{run.EntityID.String()}, "message": []string{run.SourceID.String()}}
	base.RawQuery = query.Encode()
	return base.String()
}
func (r *ChatRuntime) taskURL(slug, reference string) string {
	base := r.baseURL(slug)
	if base == nil {
		return ""
	}
	base.Path += "/work/" + reference
	return base.String()
}

func (r *ChatRuntime) requireEnabled(ctx context.Context, id uuid.UUID) error {
	enabled, err := r.Rollout.Enabled(ctx, id)
	if err != nil {
		return err
	}
	if !enabled {
		return agentruns.ErrPermissionDenied
	}
	return nil
}
