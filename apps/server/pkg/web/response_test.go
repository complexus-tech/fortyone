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
	require.Equal(t, "Check the path and resource identifier.", response.Error.Hint)
	require.NotContains(t, response.Error.Hint, "openapi")
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

func TestRespondErrorPreservesPlatformRequestStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "unsupported content type",
			err:        ErrInvalidJSONContentType,
			wantStatus: http.StatusUnsupportedMediaType,
			wantCode:   "unsupported_media_type",
		},
		{
			name:       "oversized body",
			err:        ErrRequestBodyTooLarge,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantCode:   "request_too_large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			err := RespondError(context.Background(), recorder, tt.err, http.StatusBadRequest)
			require.NoError(t, err)
			require.Equal(t, tt.wantStatus, recorder.Code)

			var response Response
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.Equal(t, tt.wantCode, response.Error.Code)
		})
	}
}

func TestRespondErrorIncludesValueFreeFieldViolations(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	validationError := &ValidationError{Violations: []FieldViolation{{
		Field:   "invitations[0].email",
		Rule:    "email",
		Message: "invitations[0].email must be a valid email address",
	}}}
	require.NoError(t, RespondError(context.Background(), recorder, validationError, http.StatusBadRequest))

	var response Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "request validation failed", response.Error.Message)
	require.Equal(t, validationError.Violations, response.Error.Fields)
}
