package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRespondErrorReturnsMachineReadableDetail(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := RespondError(context.Background(), recorder, errors.New("missing thing"), http.StatusNotFound)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Nil(t, response.Data)
	require.Equal(t, "not_found", response.Error.Code)
	require.Equal(t, "missing thing", response.Error.Message)
	require.Contains(t, response.Error.Hint, "openapi.json")
}

func TestRespondErrorSanitizesInternalErrors(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := RespondError(context.Background(), recorder, errors.New("database password leaked"), http.StatusInternalServerError)
	require.NoError(t, err)

	var response Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "internal_error", response.Error.Code)
	require.Equal(t, "internal server error", response.Error.Message)
}
