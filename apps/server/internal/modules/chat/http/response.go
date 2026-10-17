package chathttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *handlers) respond(ctx context.Context, w http.ResponseWriter, value any, err error, success int) error {
	if err == nil {
		return web.Respond(ctx, w, h.resolveResponseAvatars(ctx, value), success)
	}
	code := status(err)
	if code >= http.StatusInternalServerError && h.log != nil {
		// RespondError writes a sanitized response and consumes the service error.
		// Retain a request-correlated classification here without logging message
		// content, database details, URLs or credentials from the underlying error.
		kind, state := "internal", ""
		var databaseError interface{ SQLState() string }
		switch {
		case errors.As(err, &databaseError):
			kind, state = "database", databaseError.SQLState()
		case errors.Is(err, context.DeadlineExceeded):
			kind = "deadline"
		case errors.Is(err, context.Canceled):
			kind = "canceled"
		}
		h.log.Error(ctx, "chat request failed", "status_code", code, "error_kind", kind, "sql_state", state)
	}
	return web.RespondError(ctx, w, err, code)
}
