package apiv1http

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type Config struct {
	Log                  *logger.Logger
	SecretKey            string
	Cache                mid.RateLimitStore
	DeveloperCredentials mid.DeveloperCredentialResolver
	Workspaces           WorkspaceReader
	Teams                TeamReader
	Stories              StoryService
	StoryComments        StoryCommentReader
	Labels               LabelReader
	States               WorkflowStateReader
	Sprints              SprintReader
	Objectives           ObjectiveReader
	KeyResults           KeyResultReader
	Idempotency          IdempotencyManager
	Webhooks             WebhookManager
}

func Routes(config Config, app *web.App) {
	if app == nil || config.Log == nil || config.DeveloperCredentials == nil {
		panic("public API transport dependencies are required")
	}
	server, err := newServer(serverConfig{
		Log: config.Log, SecretKey: config.SecretKey, Workspaces: config.Workspaces,
		Teams: config.Teams, Stories: config.Stories, StoryComments: config.StoryComments,
		Labels: config.Labels, States: config.States,
		Sprints: config.Sprints, Objectives: config.Objectives, KeyResults: config.KeyResults,
		Idempotency: config.Idempotency, Webhooks: config.Webhooks,
	})
	if err != nil {
		panic("initialize public API: " + err.Error())
	}

	strict := openapiv1.NewStrictHandlerWithOptions(server, nil, openapiv1.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(writer http.ResponseWriter, request *http.Request, requestErr error) {
			status := http.StatusBadRequest
			var maxBytesError *http.MaxBytesError
			if errors.As(requestErr, &maxBytesError) {
				status = http.StatusRequestEntityTooLarge
			}
			_ = writeAPIError(request.Context(), writer, requestErr, status)
		},
		ResponseErrorHandlerFunc: func(writer http.ResponseWriter, request *http.Request, responseErr error) {
			_ = writeAPIError(request.Context(), writer, responseErr, http.StatusInternalServerError)
		},
	})
	generated := openapiv1.HandlerWithOptions(strict, openapiv1.StdHTTPServerOptions{
		ErrorHandlerFunc: func(writer http.ResponseWriter, request *http.Request, pathErr error) {
			_ = writeAPIError(request.Context(), writer, pathErr, http.StatusBadRequest)
		},
	})
	spec, err := openapiv1.GetSpec()
	if err != nil {
		panic("load public API OpenAPI contract: " + err.Error())
	}
	validator := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		DoNotValidateServers: true,
		Options:              openapi3filter.Options{AuthenticationFunc: validateContractAuthentication},
		ErrorHandlerWithOpts: func(ctx context.Context, validationErr error, writer http.ResponseWriter, request *http.Request, options nethttpmiddleware.ErrorHandlerOpts) {
			status := canonicalValidationStatus(validationErr, options.StatusCode, request)
			_ = writeAPIError(ctx, writer, validationErr, status)
		},
	})
	validated := validator(generated)
	handler := func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
		validated.ServeHTTP(writer, request.WithContext(ctx))
		return nil
	}

	auth := mid.DeveloperAuthWithErrorResponder(config.Log, config.DeveloperCredentials, writeAPIError)
	rateLimit := credentialRateLimit(config.Log, config.Cache)
	workspace := workspaceBoundary(config.Workspaces)
	workspaceRead := RequireScopes(platformauth.ScopeWorkspacesRead)
	teamsRead := RequireScopes(platformauth.ScopeTeamsRead)
	storiesRead := RequireScopes(platformauth.ScopeStoriesRead)
	storiesWrite := RequireScopes(platformauth.ScopeStoriesWrite)
	commentsRead := RequireScopes(platformauth.ScopeCommentsRead, platformauth.ScopeStoriesRead)
	labelsRead := RequireScopes(platformauth.ScopeLabelsRead)
	statesRead := RequireScopes(platformauth.ScopeStoriesRead)
	sprintsRead := RequireScopes(platformauth.ScopeSprintsRead)
	objectivesRead := RequireScopes(platformauth.ScopeObjectivesRead)
	webhooksManage := RequireScopes(platformauth.ScopeWebhooksManage)

	app.Get("/api/v1/workspaces/{workspaceId}", handler, auth, rateLimit, workspace, workspaceRead)
	app.Get("/api/v1/workspaces/{workspaceId}/teams", handler, auth, rateLimit, workspace, teamsRead)
	app.Get("/api/v1/workspaces/{workspaceId}/stories", handler, auth, rateLimit, workspace, storiesRead)
	app.Post("/api/v1/workspaces/{workspaceId}/stories", handler, auth, rateLimit, workspace, storiesWrite, boundedJSONBody, captureJSONBody)
	app.Get("/api/v1/workspaces/{workspaceId}/stories/{storyId}", handler, auth, rateLimit, workspace, storiesRead)
	app.Get("/api/v1/workspaces/{workspaceId}/stories/{storyId}/comments", handler, auth, rateLimit, workspace, commentsRead)
	app.Get("/api/v1/workspaces/{workspaceId}/stories/{storyId}/comments/{commentId}", handler, auth, rateLimit, workspace, commentsRead)
	app.Get("/api/v1/workspaces/{workspaceId}/labels", handler, auth, rateLimit, workspace, labelsRead)
	app.Get("/api/v1/workspaces/{workspaceId}/states", handler, auth, rateLimit, workspace, statesRead)
	app.Get("/api/v1/workspaces/{workspaceId}/sprints", handler, auth, rateLimit, workspace, sprintsRead)
	app.Get("/api/v1/workspaces/{workspaceId}/objectives", handler, auth, rateLimit, workspace, objectivesRead)
	app.Get("/api/v1/workspaces/{workspaceId}/key-results", handler, auth, rateLimit, workspace, objectivesRead)
	app.Get("/api/v1/workspaces/{workspaceId}/webhook-endpoints", handler, auth, rateLimit, workspace, webhooksManage)
	app.Post("/api/v1/workspaces/{workspaceId}/webhook-endpoints", handler, auth, rateLimit, workspace, webhooksManage, boundedJSONBody)
	app.Get("/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}", handler, auth, rateLimit, workspace, webhooksManage)
	app.Put("/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}/subscriptions", handler, auth, rateLimit, workspace, webhooksManage, boundedJSONBody)
	app.Post("/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}/rotate-secret", handler, auth, rateLimit, workspace, webhooksManage)
	app.Post("/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}/disable", handler, auth, rateLimit, workspace, webhooksManage, boundedJSONBody)

	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/teams", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/stories", "GET, POST", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/stories/{storyId}", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/stories/{storyId}/comments", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/stories/{storyId}/comments/{commentId}", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/labels", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/states", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/sprints", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/objectives", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/key-results", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/webhook-endpoints", "GET, POST", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}", "GET", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}/subscriptions", "PUT", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}/rotate-secret", "POST", auth, rateLimit, workspace)
	registerMethodFallback(app, "/api/v1/workspaces/{workspaceId}/webhook-endpoints/{endpointId}/disable", "POST", auth, rateLimit, workspace)
	app.Handle("", "/api/v1/", func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
		return writeAPIError(ctx, writer, errors.New("public API route not found"), http.StatusNotFound)
	}, auth, rateLimit)
}

func registerMethodFallback(
	app *web.App,
	path string,
	allowed string,
	middleware ...web.Middleware,
) {
	app.Handle("", path, func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
		writer.Header().Set("Allow", allowed)
		return writeAPIError(ctx, writer, errors.New("public API method not allowed"), http.StatusMethodNotAllowed)
	}, middleware...)
}

func validateContractAuthentication(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	if input == nil || input.SecuritySchemeName != "machineBearer" {
		return errors.New("unsupported security scheme")
	}
	actor, err := platformauth.GetActor(ctx)
	if err != nil || (actor.Kind != platformauth.PrincipalPersonalToken && actor.Kind != platformauth.PrincipalServiceAccount &&
		actor.Kind != platformauth.PrincipalOAuthUser && actor.Kind != platformauth.PrincipalOAuthApplication) {
		return errors.New("developer authentication required")
	}
	return nil
}

func canonicalValidationStatus(err error, suggested int, request *http.Request) int {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return http.StatusRequestEntityTooLarge
	}
	if requiresJSONBody(request) {
		mediaType, _, parseErr := mime.ParseMediaType(request.Header.Get("Content-Type"))
		if parseErr != nil || (mediaType != "application/json" && !strings.HasSuffix(strings.ToLower(mediaType), "+json")) {
			return http.StatusUnsupportedMediaType
		}
	}
	switch suggested {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusUnsupportedMediaType:
		return suggested
	default:
		return http.StatusBadRequest
	}
}

func requiresJSONBody(request *http.Request) bool {
	if request == nil {
		return false
	}
	path := request.URL.Path
	return (request.Method == http.MethodPost && (strings.HasSuffix(path, "/stories") || strings.HasSuffix(path, "/webhook-endpoints") || strings.HasSuffix(path, "/disable"))) ||
		(request.Method == http.MethodPut && strings.HasSuffix(path, "/subscriptions"))
}
