package workspacesecurityhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type sessionMetadataService struct {
	Service
	rows domain.SessionList
}

func (s sessionMetadataService) Sessions(context.Context, domain.Scope, *uuid.UUID, bool) (domain.SessionList, error) {
	return s.rows, nil
}

func TestSessionsExposeUsernameAndNullableBrowserName(t *testing.T) {
	t.Parallel()
	chrome := "Chrome"
	service := sessionMetadataService{rows: domain.SessionList{Items: []domain.Session{
		{ID: uuid.New(), UserID: uuid.New(), Name: "Person", Email: "person@example.com", Username: "maya", BrowserName: &chrome},
		{ID: uuid.New(), UserID: uuid.New(), Username: "older-user"},
	}}}
	handler, err := New(service, "test-secret")
	require.NoError(t, err)
	scope := domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	response := request(t, handler.Sessions, scope, "/workspaces/test/security/sessions")
	require.Equal(t, http.StatusOK, response.Code)
	var body struct {
		Data struct {
			Items []map[string]any `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Data.Items, 2)
	require.Equal(t, "maya", body.Data.Items[0]["username"])
	require.Equal(t, "Chrome", body.Data.Items[0]["browserName"])
	require.Equal(t, "Person", body.Data.Items[0]["name"], "existing clients retain the display name field")
	require.Contains(t, body.Data.Items[1], "browserName")
	require.Nil(t, body.Data.Items[1]["browserName"], "unrecorded metadata must remain unknown")
}
