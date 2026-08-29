package slackhttp

import (
	"errors"
	"net/http"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
)

func statusForSlackError(err error, fallback int) int {
	switch {
	case errors.Is(err, slack.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, slack.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, slack.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, slack.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return fallback
	}
}
