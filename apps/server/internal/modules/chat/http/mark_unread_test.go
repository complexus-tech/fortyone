package chathttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type markUnreadRepositoryStub struct {
	chatservice.Repository
	command domain.MarkUnreadCommand
	result  domain.MarkUnreadState
	err     error
}

func (r *markUnreadRepositoryStub) WithinUnit(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}

func (r *markUnreadRepositoryStub) MarkUnread(_ context.Context, command domain.MarkUnreadCommand) (domain.MarkUnreadState, error) {
	r.command = command
	return r.result, r.err
}

func (r *markUnreadRepositoryStub) Unread(context.Context, domain.ActorScope) (domain.UnreadSummary, error) {
	return r.result.Summary, nil
}

func TestMarkUnreadReturnsAuthoritativeCursorAndSummary(t *testing.T) {
	workspace, conversation, root, target := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspace)
	require.NoError(t, err)
	scope := domain.ActorScope{WorkspaceID: workspace, Actor: actor}
	repo := &markUnreadRepositoryStub{result: domain.MarkUnreadState{
		ReadState: domain.ReadState{ConversationID: conversation, ThreadRootID: &root, LastReadSequence: 8, UnreadCount: 2},
		Summary: domain.UnreadSummary{TotalUnread: 3, UnreadThreads: 1, Conversations: []domain.ConversationUnread{
			{ConversationID: conversation, UnreadCount: 3, UnreadRootCount: 1, UnreadReplyCount: 2},
		}},
		FirstUnreadReplyID: &target,
	}}
	h := handlers{service: chatservice.New(repo)}
	request := httptest.NewRequest(http.MethodPost, "/chat/conversations/"+conversation.String()+"/unread", strings.NewReader(`{"fromMessageId":"`+target.String()+`","threadRootId":"`+root.String()+`"}`))
	request.SetPathValue("id", conversation.String())
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	require.NoError(t, h.markUnread(t.Context(), response, request, scope))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
	require.Equal(t, domain.MarkUnreadCommand{Scope: scope, ConversationID: conversation, FromMessageID: target, ThreadRootID: &root}, repo.command)
	var body struct {
		Data domain.MarkUnreadState `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, repo.result, body.Data)
}

func TestMarkUnreadRejectsMalformedAndUnauthorizedRequests(t *testing.T) {
	workspace, conversation, target := uuid.New(), uuid.New(), uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspace)
	require.NoError(t, err)
	scope := domain.ActorScope{WorkspaceID: workspace, Actor: actor}
	for _, test := range []struct {
		name string
		body string
		err  error
		code int
	}{
		{"missing target", `{}`, nil, http.StatusBadRequest},
		{"invalid target", `{"fromMessageId":"invalid"}`, nil, http.StatusBadRequest},
		{"malformed JSON", `{`, nil, http.StatusBadRequest},
		{"invisible target", `{"fromMessageId":"` + target.String() + `"}`, domain.ErrNotFound, http.StatusNotFound},
		{"forbidden", `{"fromMessageId":"` + target.String() + `"}`, domain.ErrForbidden, http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := handlers{service: chatservice.New(&markUnreadRepositoryStub{err: test.err})}
			request := httptest.NewRequest(http.MethodPost, "/chat/conversations/"+conversation.String()+"/unread", strings.NewReader(test.body))
			request.SetPathValue("id", conversation.String())
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			require.NoError(t, h.markUnread(t.Context(), response, request, scope))
			require.Equal(t, test.code, response.Code)
		})
	}
}
