package messaging

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPermanentOpenAIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		code      string
		permanent bool
	}{
		{name: "bad request", status: http.StatusBadRequest, code: "invalid_request_error", permanent: true},
		{name: "invalid authentication", status: http.StatusUnauthorized, code: "invalid_api_key", permanent: true},
		{name: "permission denied", status: http.StatusForbidden, permanent: true},
		{name: "model not found", status: http.StatusNotFound, code: "model_not_found", permanent: true},
		{name: "request timeout", status: http.StatusRequestTimeout, permanent: false},
		{name: "conflict", status: http.StatusConflict, permanent: false},
		{name: "unprocessable entity", status: http.StatusUnprocessableEntity, permanent: true},
		{name: "too early", status: http.StatusTooEarly, permanent: false},
		{name: "rate limit", status: http.StatusTooManyRequests, code: "rate_limit_exceeded", permanent: false},
		{name: "credit exhausted", status: http.StatusTooManyRequests, code: "credit_balance_exhausted", permanent: true},
		{name: "project spend limit", status: http.StatusTooManyRequests, code: "project_spend_limit_exceeded", permanent: true},
		{name: "legacy insufficient quota", status: http.StatusTooManyRequests, code: "insufficient_quota", permanent: true},
		{name: "recoverable previous response", status: http.StatusBadRequest, code: "previous_response_not_found", permanent: false},
		{name: "server failure", status: http.StatusInternalServerError, code: "server_error", permanent: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := fmt.Errorf("respond: %w", &APIError{StatusCode: test.status, Code: test.code})
			require.Equal(t, test.permanent, IsPermanentOpenAIError(err))
		})
	}
}

func TestIsPermanentOpenAIErrorRejectsUnrelatedErrors(t *testing.T) {
	t.Parallel()

	require.False(t, IsPermanentOpenAIError(errors.New("database unavailable")))
}
