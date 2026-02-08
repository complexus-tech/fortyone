package mid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func Logger(log *logger.Logger) web.Middleware {

	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Debug(ctx, "request started", "method", r.Method, "path", path, "remote", r.RemoteAddr)

			err := next(ctx, w, r)

			log.Debug(ctx, "request completed", "method", r.Method, "path", path,
				"remote", r.RemoteAddr, "statusCode", v.StatusCode, "duration", time.Since(v.Now))

			return err
		}
		return h
	}
	return m
}
