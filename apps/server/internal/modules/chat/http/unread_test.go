package chathttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type unreadRepositoryStub struct {
	chatservice.Repository
	err error
}

func (r unreadRepositoryStub) UnreadThreadReadTargets(context.Context, domain.ActorScope) ([]domain.ThreadReadTarget, error) {
	return nil, r.err
}

func TestMarkAllThreadsReadReturnsNoContentAndSurfacesFailure(t *testing.T) {
	workspace := uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspace)
	require.NoError(t, err)
	scope := domain.ActorScope{WorkspaceID: workspace, Actor: actor}
	for _, test := range []struct {
		name string
		err  error
		code int
	}{
		{name: "success", code: http.StatusNoContent},
		{name: "forbidden", err: domain.ErrForbidden, code: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := handlers{service: chatservice.New(unreadRepositoryStub{err: test.err})}
			response := httptest.NewRecorder()
			require.NoError(t, h.markAllThreadsRead(t.Context(), response, httptest.NewRequest(http.MethodPost, "/chat/threads/read", nil), scope))
			require.Equal(t, test.code, response.Code)
			if test.err == nil {
				require.Empty(t, response.Body.String())
			}
		})
	}
}
