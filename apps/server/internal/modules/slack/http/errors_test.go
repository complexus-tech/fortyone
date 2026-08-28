package slackhttp

import (
	"errors"
	"net/http"
	"testing"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/stretchr/testify/require"
)

func TestStatusForSlackErrorMapsSafeDomainErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "forbidden", err: slack.ErrForbidden, status: http.StatusForbidden},
		{name: "not found", err: slack.ErrNotFound, status: http.StatusNotFound},
		{name: "conflict", err: slack.ErrConflict, status: http.StatusConflict},
		{name: "invalid", err: slack.ErrInvalidInput, status: http.StatusBadRequest},
		{name: "wrapped", err: errors.Join(errors.New("operation failed"), slack.ErrForbidden), status: http.StatusForbidden},
		{name: "unknown", err: errors.New("database unavailable"), status: http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.status, statusForSlackError(test.err, http.StatusServiceUnavailable))
		})
	}
}
