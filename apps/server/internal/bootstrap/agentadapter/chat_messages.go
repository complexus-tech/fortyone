package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

const ChatMessageSurface = "chat_message"

type ChatMessages interface {
	GetConversation(context.Context, chatdomain.ActorScope, uuid.UUID) (chatdomain.Conversation, error)
	GetMessage(context.Context, chatdomain.ActorScope, uuid.UUID) (chatdomain.Message, error)
	BoundedContext(context.Context, chatdomain.ContextQuery) ([]chatdomain.Message, error)
	CreateAgentReply(context.Context, chatdomain.AgentReplyCommand) (chatdomain.Message, error)
}

type ChatMessageCreator struct {
	Chats  ChatMessages
	Runs   *agentruns.Service
	MayaID uuid.UUID
	Tasks  Wakeup
	Log    *logger.Logger
}

type chatPayload struct {
	WorkspaceSlug string `json:"workspaceSlug"`
}

// MessageCreated is called inside Chat's source-message transaction. The run is
// the durable dispatch record, so a queue outage cannot lose an accepted ask.
func (c *ChatMessageCreator) MessageCreated(ctx context.Context, input chatdomain.CreateMessageCommand, source chatdomain.Message) error {
	if len(input.AgentMentions) == 0 {
		return nil
	}
	if input.Scope.Actor.Kind != auth.PrincipalHumanUser {
		return chatdomain.ErrForbidden
	}
	if err := validateChatMayaMention(source.RichContent, input.AgentMentions, c.MayaID); err != nil {
		return errors.Join(chatdomain.ErrInvalid, err)
	}
	if c.Chats == nil || c.Runs == nil {
		return agentruns.ErrUnavailable
	}
	if _, joined := database.UnitTransaction(ctx); !joined {
		return errors.New("Maya message acceptance requires the source transaction")
	}
	conversation, err := c.Chats.GetConversation(ctx, input.Scope, input.ConversationID)
	if err != nil {
		return err
	}
	if conversation.Kind != "team" || conversation.TeamID == nil || *conversation.TeamID == uuid.Nil {
		return chatdomain.ErrForbidden
	}
	rootID := source.ID
	if source.ThreadRootID != nil {
		rootID = *source.ThreadRootID
	}
	workspace, _ := mid.GetWorkspace(ctx)
	payload, err := json.Marshal(chatPayload{WorkspaceSlug: workspace.Slug})
	if err != nil {
		return err
	}
	run, accepted, err := c.Runs.Accept(ctx, agentruns.Input{
		WorkspaceID: input.Scope.WorkspaceID, RequesterID: input.Scope.Actor.PrincipalID, TeamID: *conversation.TeamID,
		Surface: ChatMessageSurface, SourceID: source.ID, ConversationID: rootID, EntityID: conversation.ID,
		SourceRevision: strconv.Itoa(source.Version), SourceHash: chatSourceHash(source),
		IdempotencyKey: "chat-message:" + source.ID.String(), Payload: payload,
	})
	if err != nil {
		return err
	}
	if accepted && c.Tasks != nil {
		database.AfterCommit(ctx, func() {
			if err := c.Tasks.EnqueueAgentInvocation(run.ID); err != nil && c.Log != nil {
				c.Log.Error(context.WithoutCancel(ctx), "Maya chat request awaits recovery dispatch", "run_id", run.ID, "error", err)
			}
		})
	}
	return nil
}

func chatSourceHash(source chatdomain.Message) string {
	var value any
	content := source.RichContent
	if json.Unmarshal(content, &value) == nil {
		if canonical, err := json.Marshal(value); err == nil {
			content = canonical
		}
	}
	return SourceHash(source.Text + "\x00" + string(content))
}

type ChatAccess struct {
	Rollout    chatrollout.Policy
	Chats      ChatMessages
	Workspace  WorkspaceAccess
	Configured bool
}

func (a ChatAccess) CanRead(ctx context.Context, conversationID, workspaceID uuid.UUID) error {
	if !a.Rollout.Allows(ctx, workspaceID) {
		return chatdomain.ErrForbidden
	}
	actor, err := auth.GetActor(ctx)
	if err != nil {
		return err
	}
	_, err = a.Chats.GetConversation(ctx, chatdomain.ActorScope{WorkspaceID: workspaceID, Actor: actor}, conversationID)
	return err
}

func (a ChatAccess) Enabled(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	if !a.Rollout.Allows(ctx, workspaceID) || !a.Configured || a.Workspace == nil {
		return false, nil
	}
	return a.Workspace.WorkspaceCanUseMaya(ctx, workspaceID)
}

func (a ChatAccess) EnabledForEntity(ctx context.Context, conversationID, workspaceID uuid.UUID) (bool, error) {
	actor, err := auth.GetActor(ctx)
	if err != nil {
		return false, err
	}
	conversation, err := a.Chats.GetConversation(ctx, chatdomain.ActorScope{WorkspaceID: workspaceID, Actor: actor}, conversationID)
	if err != nil {
		return false, err
	}
	if conversation.Kind != "team" || conversation.TeamID == nil {
		return false, nil
	}
	return a.Enabled(ctx, workspaceID)
}
