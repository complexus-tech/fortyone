package mid

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func Logger(log *logger.Logger) web.Middleware {

	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			route := strings.TrimSpace(r.Pattern)
			if route == "" {
				route = "unmatched"
			}

			log.Debug(ctx, "request started", "method", r.Method, "route", route)

			err := next(ctx, w, r)

			log.Debug(ctx, "request completed", "method", r.Method, "route", route,
				"statusCode", v.StatusCode, "duration", time.Since(v.Now))

			return err
		}
		return h
	}
	return m
}
