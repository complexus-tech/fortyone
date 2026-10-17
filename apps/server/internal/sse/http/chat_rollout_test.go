package ssehttp

import (
	"context"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSharedStreamSuppressesDisabledChatButKeepsOtherUpdates(t *testing.T) {
	enabled := Handler{}
	disabled := Handler{}
	for _, event := range []string{`{"type":"chat.updated"}`, `{"type":"chat.typing"}`, `{"type":"chat.future"}`, `{"type":"chat_message"}`, `{"entityType":"chat_message"}`, `{"entity_type":"chat_message"}`} {
		require.False(t, disabled.allowsEvent([]byte(event), false), event)
		require.False(t, enabled.allowsEvent([]byte(event), false), event)
		require.True(t, enabled.allowsEvent([]byte(event), true), event)
	}
	require.True(t, disabled.allowsEvent([]byte(`{"type":"calendar.updated"}`), false))
	require.True(t, disabled.allowsEvent([]byte(`{"type":"story_update"}`), false))
	require.False(t, disabled.allowsEvent([]byte(`invalid`), false))
}

func TestLiveWorkspaceSettingChangesFilterExistingStreamWithoutExtraLookups(t *testing.T) {
	user, workspace := uuid.New(), uuid.New()
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{{userID: user, ok: true}, {userID: user, ok: true}, {userID: user, ok: true}, {userID: user, ok: true}}}
	handler := newTestStreamHandler(t, resolver, workspace, make(chan time.Time))
	lookups := 0
	handler.WorkspaceAccess = workspaceResolverFunc(func(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
		lookups++
		return mid.WorkspaceInfo{ID: workspace, ChatEnabled: lookups == 1 || lookups == 4}, nil
	})
	messages := make(chan []byte, 4)
	messages <- []byte(`{"type":"chat.updated","phase":"before"}`)
	messages <- []byte(`{"type":"chat.updated","phase":"hidden"}`)
	messages <- []byte(`{"type":"story_update","phase":"other"}`)
	messages <- []byte(`{"type":"chat.updated","phase":"restored"}`)
	close(messages)
	response := httptest.NewRecorder()
	require.NoError(t, handler.serveStream(t.Context(), response, newStreamRequest(t.Context()), user, workspace, testWorkspaceSlug, t.Context(), messages))
	require.Contains(t, response.Body.String(), `"phase":"before"`)
	require.NotContains(t, response.Body.String(), `"phase":"hidden"`)
	require.Contains(t, response.Body.String(), `"phase":"other"`)
	require.Contains(t, response.Body.String(), `"phase":"restored"`)
	require.Equal(t, 4, lookups, "each delivered event reuses its existing authoritative membership lookup")
}
