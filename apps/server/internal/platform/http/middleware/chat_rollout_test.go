package mid

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatRolloutNeedsAuthorizedWorkspaceContext(t *testing.T) {
	workspace := uuid.New()
	for _, test := range []struct {
		name      string
		enabled   bool
		workspace uuid.UUID
		status    int
	}{
		{"no membership context", true, uuid.Nil, http.StatusUnauthorized},
		{"default disabled", false, workspace, http.StatusForbidden},
		{"other workspace", false, uuid.New(), http.StatusForbidden},
		{"pilot workspace", true, workspace, http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.workspace != uuid.Nil {
				ctx = context.WithValue(ctx, workspaceKey, WorkspaceInfo{ID: test.workspace, ChatEnabled: test.enabled})
			}
			called := false
			handler := RequireChatEnabled()(func(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
				called = true
				w.WriteHeader(http.StatusNoContent)
				return nil
			})
			response := httptest.NewRecorder()
			require.NoError(t, handler(ctx, response, httptest.NewRequest(http.MethodPost, "/chat/conversations", nil)))
			require.Equal(t, test.status, response.Code)
			require.Equal(t, test.status == http.StatusNoContent, called)
		})
	}
}
