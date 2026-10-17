package agentadapter

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type chatRuntimeMessages struct {
	conversation chatdomain.Conversation
	source       chatdomain.Message
	scope        chatdomain.ActorScope
	err          error
	replies      []chatdomain.AgentReplyCommand
}

func (s *chatRuntimeMessages) GetConversation(_ context.Context, scope chatdomain.ActorScope, _ uuid.UUID) (chatdomain.Conversation, error) {
	s.scope = scope
	return s.conversation, s.err
}
func (s *chatRuntimeMessages) GetMessage(context.Context, chatdomain.ActorScope, uuid.UUID) (chatdomain.Message, error) {
	return s.source, s.err
}
func (s *chatRuntimeMessages) BoundedContext(context.Context, chatdomain.ContextQuery) ([]chatdomain.Message, error) {
	return nil, s.err
}
func (s *chatRuntimeMessages) CreateAgentReply(_ context.Context, command chatdomain.AgentReplyCommand) (chatdomain.Message, error) {
	s.replies = append(s.replies, command)
	return chatdomain.Message{ID: command.CreationID}, s.err
}

func chatRuntimeFixture() (*ChatRuntime, agentruns.Run, *chatRuntimeMessages) {
	run := agentruns.Run{ID: uuid.New(), WorkspaceID: uuid.New(), RequesterID: uuid.New(), TeamID: uuid.New(), Surface: ChatMessageSurface, SourceID: uuid.New(), EntityID: uuid.New(), SourceRevision: "1"}
	run.ConversationID = run.SourceID
	store := &chatRuntimeMessages{conversation: chatdomain.Conversation{ID: run.EntityID, Kind: "team", TeamID: &run.TeamID}, source: chatdomain.Message{ID: run.SourceID, ConversationID: run.EntityID, Version: 1, Text: "@Maya summarize", Author: chatdomain.Person{ID: run.RequesterID}}}
	run.SourceHash = chatSourceHash(store.source)
	runtime := &ChatRuntime{Rollout: chatrollout.New(chatrollout.ReaderFunc(func(_ context.Context, id uuid.UUID) (bool, error) { return true && id == run.WorkspaceID, nil })), Chats: store, Shared: &CommentRuntime{Enabled: true, MayaID: uuid.New(), Access: runtimeAccess(true), Memberships: runtimeMembership{workspace: workspaces.CoreWorkspace{UserRole: "member"}}}}
	return runtime, run, store
}

func TestChatRuntimeReauthorizesTeamAndSourceBeforeWork(t *testing.T) {
	t.Parallel()
	runtime, run, store := chatRuntimeFixture()
	require.NoError(t, runtime.Authorize(t.Context(), run))
	require.True(t, store.scope.Actor.TeamAccess.Allows(run.TeamID))
	require.False(t, store.scope.Actor.TeamAccess.Allows(uuid.New()))
	require.True(t, store.scope.Actor.Scopes.Has(auth.ScopeChatWrite))
	require.Equal(t, run.RequesterID, store.scope.Actor.PrincipalID)
	store.source.Version++
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrSourceChanged)
	run.Outcome = &agentruns.Outcome{Receipt: json.RawMessage(`{"committed":true}`)}
	require.NoError(t, runtime.Authorize(t.Context(), run), "an already committed task still has a durable receipt to publish")
	deleted := time.Now()
	store.source.DeletedAt = &deleted
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrSourceChanged)
	store.source.DeletedAt = nil
	store.conversation.Kind = "dm"
	store.conversation.TeamID = nil
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
	store.err = chatdomain.ErrForbidden
	require.ErrorIs(t, runtime.Authorize(t.Context(), run), agentruns.ErrPermissionDenied)
}

func TestChatRuntimePublicationUsesServerIdentityAndStableKey(t *testing.T) {
	t.Parallel()
	runtime, run, store := chatRuntimeFixture()
	_, err := runtime.Apply(t.Context(), run, agentruns.Proposal{Action: json.RawMessage(`{"token":"never-executed"}`)})
	require.ErrorIs(t, err, agentruns.ErrPermissionDenied)
	first, err := runtime.Publish(t.Context(), run, agentruns.Outcome{Reply: "**Done.**"})
	require.NoError(t, err)
	second, err := runtime.Publish(t.Context(), run, agentruns.Outcome{Reply: "**Done.**"})
	require.NoError(t, err)
	require.Equal(t, first, second)
	proposal, err := runtime.PublishProposal(t.Context(), run, agentruns.Proposal{Reply: "Review these tasks."})
	require.NoError(t, err)
	require.NotEqual(t, first, proposal)
	command := store.replies[0]
	require.Equal(t, runtime.Shared.MayaID, command.Author.PrincipalID)
	require.Equal(t, auth.PrincipalSystem, command.Author.Kind)
	require.Equal(t, run.RequesterID, command.Requester.PrincipalID)
	require.Equal(t, run.EntityID, command.ConversationID)
	require.Equal(t, run.ConversationID, command.ThreadRootID)
	require.Empty(t, command.MentionedUserIDs)
}

func TestChatContextExcludesFutureDeletedAndForeignMessages(t *testing.T) {
	t.Parallel()
	mayaID := uuid.New()
	source := chatdomain.Message{ID: uuid.New(), ConversationID: uuid.New(), Sequence: 3}
	deleted := time.Now()
	history := []chatdomain.Message{
		{ID: uuid.New(), ConversationID: source.ConversationID, Sequence: 1, Text: "Decision: ship", Author: chatdomain.Person{FullName: "Person"}},
		{ID: uuid.New(), ConversationID: source.ConversationID, Sequence: 2, Text: "Acknowledged", Author: chatdomain.Person{ID: mayaID, FullName: "Maya"}},
		{ID: uuid.New(), ConversationID: source.ConversationID, Sequence: 4, Text: "Future context"},
		{ID: uuid.New(), ConversationID: uuid.New(), Sequence: 1, Text: "Another team's secret"},
		{ID: uuid.New(), ConversationID: source.ConversationID, Sequence: 1, Text: "Deleted context", DeletedAt: &deleted},
	}
	turns := chatConversation(history, source, mayaID)
	require.Len(t, turns, 2)
	require.Equal(t, messaging.RoleUser, turns[0].Role)
	require.Equal(t, messaging.RoleAssistant, turns[1].Role)
	require.Contains(t, turns[0].Text, "context only")
}
