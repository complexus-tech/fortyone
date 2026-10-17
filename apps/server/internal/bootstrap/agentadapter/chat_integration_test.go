//go:build integration

package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	agentrepository "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	messagingcontext "github.com/complexus-tech/projects-api/internal/modules/messaging/context"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	searchrepository "github.com/complexus-tech/projects-api/internal/modules/search/repository"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// The provider is deterministic; the source, membership checks, invocation,
// proposal, task mutation, receipts, reply and outbox all use real PostgreSQL.
type chatIntegrationAssistant struct {
	tools               *messaging.FortyOneToolExecutor
	createTasks         bool
	calls               int
	request             messaging.Request
	followupDecision    bool
	followupError       error
	followupUsage       messaging.Usage
	classificationCalls int
	classification      messaging.ChatFollowupRequest
}

func (a *chatIntegrationAssistant) ClassifyChatFollowup(_ context.Context, request messaging.ChatFollowupRequest) (messaging.ChatFollowupDecision, error) {
	a.classificationCalls++
	a.classification = request
	return messaging.ChatFollowupDecision{ShouldRespond: a.followupDecision, Usage: a.followupUsage}, a.followupError
}

func (a *chatIntegrationAssistant) Respond(ctx context.Context, request messaging.Request) (messaging.Response, error) {
	a.calls++
	a.request = request
	if !a.createTasks {
		return messaging.Response{Text: "The team agreed to improve onboarding."}, nil
	}
	scope := messaging.ToolScope{WorkspaceID: request.WorkspaceID, UserID: request.UserID, AllowedTeamIDs: request.AllowedTeamIDs, SharedTeamIDs: request.SharedTeamIDs, SourceScope: request.SourceScope, SourceURL: request.SourceURL, AllowMutations: request.AllowMutations, WebsiteURL: request.WebsiteURL}
	result, err := a.tools.Execute(ctx, scope, messaging.ToolCall{Name: "create_stories", Arguments: json.RawMessage(`{"stories":[{"title":"Ship onboarding checklist","description":"Add the checklist agreed in this chat.","priority":null,"assignee_id":null}]}`)})
	if err != nil {
		return messaging.Response{}, err
	}
	var proposed struct {
		Confirmation messaging.StoryMutationConfirmation `json:"confirmation"`
	}
	if err := json.Unmarshal(result, &proposed); err != nil {
		return messaging.Response{}, err
	}
	return messaging.Response{Confirmation: &proposed.Confirmation}, nil
}

type chatAgentFixture struct {
	pool                                             *pgxpool.Pool
	ctx                                              context.Context
	workspace, user, other, maya, team, conversation uuid.UUID
	scope                                            chatdomain.ActorScope
	chats                                            *chatservice.Service
	runs                                             *agentruns.Service
	creator                                          *ChatMessageCreator
	runtime                                          *ChatRuntime
	assistant                                        *chatIntegrationAssistant
}

func newChatAgentFixture(t *testing.T) chatAgentFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	f := chatAgentFixture{pool: postgres.Pool, ctx: t.Context(), workspace: uuid.New(), user: uuid.New(), other: uuid.New(), team: uuid.New(), maya: uuid.MustParse("00000000-0000-0000-0000-000000000001")}
	_, err := f.pool.Exec(f.ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Chat Maya test',$2)`, f.workspace, "chat-"+f.workspace.String())
	require.NoError(t, err)
	for _, id := range []uuid.UUID{f.user, f.other} {
		_, err = f.pool.Exec(f.ctx, `INSERT INTO users(user_id,username,email,full_name) VALUES($1,$2,$3,'Chat Person')`, id, id.String(), id.String()+"@example.test")
		require.NoError(t, err)
		_, err = f.pool.Exec(f.ctx, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, f.workspace, id)
		require.NoError(t, err)
	}
	_, err = f.pool.Exec(f.ctx, `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'Product','PROD','#111111')`, f.team, f.workspace)
	require.NoError(t, err)
	for _, id := range []uuid.UUID{f.user, f.other} {
		_, err = f.pool.Exec(f.ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, f.team, id)
		require.NoError(t, err)
	}
	_, err = f.pool.Exec(f.ctx, `INSERT INTO statuses(workspace_id,team_id,name,category,order_index,is_default) VALUES($1,$2,'Todo','unstarted',0,true)`, f.workspace, f.team)
	require.NoError(t, err)
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id FROM chat_conversations WHERE workspace_id=$1 AND team_id=$2`, f.workspace, f.team).Scan(&f.conversation))
	actor, err := auth.NewHumanActor(f.user).WithWorkspace(f.workspace)
	require.NoError(t, err)
	f.ctx, err = auth.SetActor(f.ctx, actor)
	require.NoError(t, err)
	f.scope = chatdomain.ActorScope{WorkspaceID: f.workspace, Actor: actor}
	f.runs = agentruns.New(agentrepository.New(f.pool), nil)
	f.creator = &ChatMessageCreator{Runs: f.runs, MayaID: f.maya, Tasks: &unavailableWakeup{}}
	f.chats = chatservice.New(chatrepository.New(nil, f.pool), chatservice.WithMayaActorID(f.maya), chatservice.WithMessageCreatedHook(f.creator.MessageCreated))
	f.creator.Chats = f.chats
	log := logger.NewWithText(io.Discard, slog.LevelError, "chat-test")
	storyService := stories.New(log, storiesrepository.New(log, f.pool), nil, nil)
	teamService := teams.New(log, teamsrepository.New(f.pool))
	userService := users.New(log, usersrepository.New(f.pool), nil)
	stateService := states.New(statesrepository.New(f.pool))
	tools, err := messaging.NewFortyOneToolExecutor(teamService, storyService, search.New(log, searchrepository.New(f.pool)), objectives.New(log, objectivesrepository.New(f.pool)),
		messaging.WithOperationalTools(messaging.OperationalToolServices{States: stateService, Users: userService}), messaging.WithStoryMutations("a deterministic test-only chat mutation secret"), messaging.WithStoryMutationConfirmationStore(messagingrepository.New(f.pool)))
	require.NoError(t, err)
	provider, err := messagingcontext.New(userService, workspacesrepository.New(f.pool), teamService)
	require.NoError(t, err)
	f.assistant = &chatIntegrationAssistant{tools: tools}
	f.runtime = &ChatRuntime{Rollout: chatrollout.New(workspacesrepository.New(f.pool)), Chats: f.chats, Tools: tools, Shared: &CommentRuntime{Enabled: true, MayaID: f.maya, Access: runtimeAccess(true), Memberships: workspacesrepository.New(f.pool), Assistant: f.assistant, Context: provider, WebsiteURL: "https://fortyone.app"}}
	return f
}

func (f chatAgentFixture) command(t *testing.T) chatdomain.CreateMessageCommand {
	t.Helper()
	rich, err := json.Marshal(chatContentNode{Type: "doc", Content: []chatContentNode{{Type: "paragraph", Content: []chatContentNode{{Type: "mention", Attrs: map[string]any{"id": f.maya.String(), "label": "Maya", "mentionKind": "agent", "agentKey": "maya"}}, {Type: "text", Text: " create a task for our onboarding checklist"}}}}})
	require.NoError(t, err)
	return chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), Text: "@Maya create a task for our onboarding checklist", RichContent: rich, AgentMentions: []string{"maya"}}
}

func TestChatMayaAcceptanceAndPublicationAreAtomicAndIdempotent(t *testing.T) {
	f := newChatAgentFixture(t)
	command := f.command(t)
	rollback := errors.New("abort source creation")
	aborting := chatservice.New(chatrepository.New(nil, f.pool), chatservice.WithMayaActorID(f.maya), chatservice.WithMessageCreatedHook(func(ctx context.Context, c chatdomain.CreateMessageCommand, m chatdomain.Message) error {
		if err := f.creator.MessageCreated(ctx, c, m); err != nil {
			return err
		}
		return rollback
	}))
	_, err := aborting.CreateMessage(f.ctx, command)
	require.ErrorIs(t, err, rollback)
	for _, table := range []string{"chat_conversation_messages", "agent_invocations", "chat_outbox"} {
		var count int
		require.NoError(t, f.pool.QueryRow(f.ctx, "SELECT count(*) FROM "+table+" WHERE workspace_id=$1", f.workspace).Scan(&count))
		require.Zero(t, count, table+" must roll back with its source")
	}
	message, err := f.chats.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	replay, err := f.chats.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	require.Equal(t, message.ID, replay.ID)
	runs, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	worker := agentruns.New(agentrepository.New(f.pool), &RuntimeRouter{Chat: f.runtime})
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))
	finished, err := f.runs.Get(f.ctx, f.workspace, runs[0].ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, finished.Status)
	require.NotNil(t, finished.ReplyCommentID)
	reply, err := f.chats.GetMessage(f.ctx, f.scope, *finished.ReplyCommentID)
	require.NoError(t, err)
	require.Equal(t, f.maya, reply.Author.ID)
	require.Equal(t, &message.ID, reply.ThreadRootID)
	require.Contains(t, reply.Text, "onboarding")
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))
	require.Equal(t, 1, f.assistant.calls)
	var replies int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_messages WHERE conversation_id=$1 AND author_id=$2`, f.conversation, f.maya).Scan(&replies))
	require.Equal(t, 1, replies)
	require.Equal(t, []uuid.UUID{f.team}, f.assistant.request.AllowedTeamIDs)
	require.Equal(t, []uuid.UUID{f.team}, f.assistant.request.SharedTeamIDs)
}

func TestChatMayaTaskApprovalCreatesRealReceiptOnlyForRequester(t *testing.T) {
	f := newChatAgentFixture(t)
	f.assistant.createTasks = true
	message, err := f.chats.CreateMessage(f.ctx, f.command(t))
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	worker := agentruns.New(agentrepository.New(f.pool), &RuntimeRouter{Chat: f.runtime})
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))
	pending, err := f.runs.Get(f.ctx, f.workspace, runs[0].ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusAwaitingApproval, pending.Status)
	require.Contains(t, pending.Proposal.Review, "Ship onboarding checklist")
	var count int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM stories WHERE workspace_id=$1`, f.workspace).Scan(&count))
	require.Zero(t, count)
	_, err = f.runs.Approve(f.ctx, agentruns.Control{WorkspaceID: f.workspace, RequesterID: f.other, RunID: pending.ID, ExpectedVersion: pending.Version})
	require.ErrorIs(t, err, agentruns.ErrConflict)
	_, err = f.runs.Approve(f.ctx, agentruns.Control{WorkspaceID: f.workspace, RequesterID: f.user, RunID: pending.ID, ExpectedVersion: pending.Version})
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, pending.ID))
	finished, err := f.runs.Get(f.ctx, f.workspace, pending.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, finished.Status)
	require.NotNil(t, finished.Outcome)
	var receipt struct {
		Result messaging.StoryMutationResult `json:"result"`
	}
	require.NoError(t, json.Unmarshal(finished.Outcome.Receipt, &receipt))
	require.Len(t, receipt.Result.Items, 1)
	item := receipt.Result.Items[0]
	require.Equal(t, f.team, item.TeamID)
	require.NotEmpty(t, item.Reference)
	var title string
	var reporter uuid.UUID
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT title,reporter_id FROM stories WHERE id=$1 AND workspace_id=$2`, item.StoryID, f.workspace).Scan(&title, &reporter))
	require.Equal(t, "Ship onboarding checklist", title)
	require.Equal(t, f.user, reporter)
	reply, err := f.chats.GetMessage(f.ctx, f.scope, *finished.ReplyCommentID)
	require.NoError(t, err)
	require.Equal(t, &message.ID, reply.ThreadRootID)
	require.Contains(t, reply.Text, item.Reference)
	require.NoError(t, worker.Process(f.ctx, pending.ID))
	require.Equal(t, 1, f.assistant.calls, "approval must reuse the persisted proposal")
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM stories WHERE workspace_id=$1`, f.workspace).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_messages WHERE conversation_id=$1 AND author_id=$2`, f.conversation, f.maya).Scan(&count))
	require.Equal(t, 2, count, "one proposal and one committed receipt")
}
