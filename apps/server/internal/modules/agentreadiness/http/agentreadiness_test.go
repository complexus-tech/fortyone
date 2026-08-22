package agentreadinesshttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPIIsValidJSONWithUniqueOperationIDs(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	require.NoError(t, New().OpenAPI(context.Background(), recorder, httptest.NewRequest(http.MethodGet, "/openapi.json", nil)))
	require.Equal(t, http.StatusOK, recorder.Code)

	var document map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &document))
	require.Equal(t, "3.1.1", document["openapi"])

	seen := make(map[string]struct{})
	paths, ok := document["paths"].(map[string]any)
	require.True(t, ok)
	for _, rawPath := range paths {
		operations, ok := rawPath.(map[string]any)
		require.True(t, ok)
		for _, rawOperation := range operations {
			operation, ok := rawOperation.(map[string]any)
			require.True(t, ok)
			operationID, ok := operation["operationId"].(string)
			require.True(t, ok)
			require.NotEmpty(t, operation["description"])
			_, duplicate := seen[operationID]
			require.False(t, duplicate, "duplicate operationId %q", operationID)
			seen[operationID] = struct{}{}
		}
	}
}

func TestMCPInitializeAndResourceLifecycle(t *testing.T) {
	t.Parallel()
	handler := New()

	initialize := callMCP(t, handler, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`)
	require.Equal(t, http.StatusOK, initialize.Code)
	var initialized rpcResponse
	require.NoError(t, json.Unmarshal(initialize.Body.Bytes(), &initialized))
	require.Nil(t, initialized.Error)

	listed := callMCP(t, handler, `{"jsonrpc":"2.0","id":2,"method":"resources/list","params":{}}`)
	var listResponse struct {
		Result struct {
			Resources []resource `json:"resources"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(listed.Body.Bytes(), &listResponse))
	require.Len(t, listResponse.Result.Resources, 2)
	for _, listedResource := range listResponse.Result.Resources {
		require.NotEmpty(t, listedResource.MimeType)
		read := callMCP(t, handler, `{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"`+listedResource.URI+`"}}`)
		require.Equal(t, http.StatusOK, read.Code)
		require.Contains(t, read.Body.String(), `"text":`)
	}
}

func TestMCPListsAndCallsReadOnlyTools(t *testing.T) {
	t.Parallel()
	handler := New()

	listed := callMCP(t, handler, `{"jsonrpc":"2.0","id":"tools","method":"tools/list","params":{}}`)
	require.Contains(t, listed.Body.String(), "get_product_overview")
	require.Contains(t, listed.Body.String(), `"readOnlyHint":true`)

	called := callMCP(t, handler, `{"jsonrpc":"2.0","id":"call","method":"tools/call","params":{"name":"get_developer_resources","arguments":{}}}`)
	require.Contains(t, called.Body.String(), "openapi.json")
	require.Contains(t, called.Body.String(), `"isError":false`)
}

func TestMCPAcceptsNotificationsWithoutResponseBody(t *testing.T) {
	t.Parallel()

	recorder := callMCP(t, New(), `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Empty(t, recorder.Body.String())
}

func TestMCPAcceptsJSONContentTypeParameters(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	recorder := httptest.NewRecorder()
	require.NoError(t, New().MCPPost(context.Background(), recorder, request))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func callMCP(t *testing.T, handler *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.MCPPost(context.Background(), recorder, request))
	return recorder
}
