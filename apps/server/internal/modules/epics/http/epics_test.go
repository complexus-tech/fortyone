package epicshttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListUsesStandardNotImplementedEnvelope(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	handler := New(epics.New())

	require.NoError(t, handler.listForWorkspace(context.Background(), recorder, uuid.New()))
	require.Equal(t, http.StatusNotImplemented, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response web.Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Nil(t, response.Data)
	require.NotNil(t, response.Error)
	require.Equal(t, "internal_error", response.Error.Code)
	require.Equal(t, "internal server error", response.Error.Message)
	require.NotEmpty(t, response.Error.Hint)
}
