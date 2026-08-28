package apiv1http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	publicAPIRateLimit = int64(600)
	publicAPIWindow    = time.Minute
	maximumJSONBytes   = int64(64 << 10)
)

type requestBodyContextKey struct{}

func credentialRateLimit(log *logger.Logger, store mid.RateLimitStore) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
			actor, err := platformauth.GetActor(ctx)
			if err != nil || actor.CredentialID == uuid.Nil {
				return writeAPIError(ctx, writer, err, http.StatusUnauthorized)
			}
			if store == nil {
				log.Error(ctx, "public API rate limit store is not configured")
				writer.Header().Set("Retry-After", "60")
				return writeAPIError(ctx, writer, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
			}
			key := "rate-limit:api-v1:credential:" + actor.CredentialID.String()
			count, err := store.IncrementWithTTL(ctx, key, publicAPIWindow)
			if err != nil {
				log.Error(ctx, "failed to enforce public API credential rate limit")
				writer.Header().Set("Retry-After", "60")
				return writeAPIError(ctx, writer, err, http.StatusServiceUnavailable)
			}
			web.SetRateLimitHeaders(writer, "api-v1-credential", publicAPIRateLimit, publicAPIRateLimit-count, publicAPIWindow)
			if count > publicAPIRateLimit {
				writer.Header().Set("Retry-After", strconv.FormatInt(int64(publicAPIWindow/time.Second), 10))
				return writeAPIError(ctx, writer, errors.New("rate limit exceeded"), http.StatusTooManyRequests)
			}
			return next(ctx, writer, request)
		}
	}
}

func workspaceBoundary(workspaces WorkspaceReader) web.Middleware {
	if workspaces == nil {
		panic("public API workspace boundary requires a workspace reader")
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
			actor, err := platformauth.GetActor(ctx)
			if err != nil {
				return writeAPIError(ctx, writer, err, http.StatusUnauthorized)
			}
			workspaceID, err := uuid.Parse(web.Params(request, "workspaceId"))
			if err != nil {
				return writeAPIError(ctx, writer, err, http.StatusBadRequest)
			}

			switch actor.Kind {
			case platformauth.PrincipalPersonalToken, platformauth.PrincipalServiceAccount,
				platformauth.PrincipalOAuthApplication:
				if actor.WorkspaceID != workspaceID {
					return writeAPIError(ctx, writer, errors.New("credential workspace mismatch"), http.StatusForbidden)
				}
			case platformauth.PrincipalOAuthUser:
				if actor.WorkspaceID != uuid.Nil {
					return writeAPIError(ctx, writer, errors.New("OAuth credential workspace was selected before authorization"), http.StatusForbidden)
				}
				// An OAuth grant delegates one user, not one workspace. Rechecking the
				// selected workspace here prevents a valid grant from becoming a
				// cross-workspace capability after membership is removed.
				if _, err := workspaces.Get(ctx, workspaceID, actor.PrincipalID); err != nil {
					return writeAPIError(ctx, writer, errors.New("workspace membership required"), http.StatusForbidden)
				}
				ctx, err = platformauth.BindWorkspace(ctx, workspaceID)
				if err != nil {
					return writeAPIError(ctx, writer, err, http.StatusUnauthorized)
				}
			default:
				return writeAPIError(ctx, writer, errors.New("developer credential principal is unsupported"), http.StatusForbidden)
			}
			return next(ctx, writer, request)
		}
	}
}

// RequireScopes rejects a request before any service is invoked. Services
// independently repeat authorization because scopes never replace resource
// ownership, membership, or workspace-role checks.
func RequireScopes(required ...platformauth.Scope) web.Middleware {
	if len(required) == 0 {
		panic("at least one public API scope is required")
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
			actor, err := platformauth.GetActor(ctx)
			if err != nil {
				return writeAPIError(ctx, writer, err, http.StatusUnauthorized)
			}
			if !actor.Scopes.ContainsAll(required...) {
				return writeAPIError(ctx, writer, fmt.Errorf("required scope unavailable"), http.StatusForbidden)
			}
			return next(ctx, writer, request)
		}
	}
}

func boundedJSONBody(next web.Handler) web.Handler {
	return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
		request.Body = http.MaxBytesReader(writer, request.Body, maximumJSONBytes)
		return next(ctx, writer, request)
	}
}

// captureJSONBody retains the exact bounded bytes used by an idempotent
// mutation. The generated OpenAPI handler still receives an independent
// reader, while receipt hashing remains sensitive to every caller byte.
func captureJSONBody(next web.Handler) web.Handler {
	return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				return writeAPIError(ctx, writer, err, http.StatusRequestEntityTooLarge)
			}
			return writeAPIError(ctx, writer, err, http.StatusBadRequest)
		}
		request.Body = io.NopCloser(bytes.NewReader(body))
		ctx = context.WithValue(ctx, requestBodyContextKey{}, append([]byte(nil), body...))
		return next(ctx, writer, request)
	}
}

func exactRequestBody(ctx context.Context) ([]byte, bool) {
	body, ok := ctx.Value(requestBodyContextKey{}).([]byte)
	if !ok {
		return nil, false
	}
	return append([]byte(nil), body...), true
}
