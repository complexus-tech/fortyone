package outboundwebhookshttp

import (
	"context"
	"net/http"
	"strconv"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	defaultManagementPageSize = 25
	maximumManagementPageSize = 100
	maximumManagementBytes    = 32 << 10
)

type Service interface {
	CreateEndpoint(context.Context, outboundwebhooksservice.Access, outboundwebhooksservice.CreateEndpointInput) (outboundwebhooksdomain.CreatedEndpoint, error)
	ListEndpoints(context.Context, outboundwebhooksservice.Access, *outboundwebhooksdomain.EndpointCursor, int) (outboundwebhooksdomain.EndpointPage, error)
	ReplaceSubscriptions(context.Context, outboundwebhooksservice.Access, uuid.UUID, []outboundwebhooksdomain.EventType, string) error
	DisableEndpoint(context.Context, outboundwebhooksservice.Access, uuid.UUID, string, string) error
	RotateEndpointSecret(context.Context, outboundwebhooksservice.Access, uuid.UUID, string) (outboundwebhooksdomain.SigningSecret, int, time.Time, error)
}

type Handlers struct {
	service Service
	cursors pagination.CursorCodec[endpointCursor]
}

func New(service Service, cursors pagination.CursorCodec[endpointCursor]) *Handlers {
	if service == nil {
		panic("outbound webhook management service is required")
	}
	return &Handlers{service: service, cursors: cursors}
}

func (handlers *Handlers) ListEndpoints(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	limit := defaultManagementPageSize
	if rawLimit := request.URL.Query().Get("limit"); rawLimit != "" {
		limit, err = strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > maximumManagementPageSize {
			return web.RespondError(ctx, writer, outboundwebhooksdomain.ErrInvalidEndpoint, http.StatusBadRequest)
		}
	}
	var cursor *outboundwebhooksdomain.EndpointCursor
	if rawCursor := request.URL.Query().Get("cursor"); rawCursor != "" {
		decoded, decodeErr := handlers.cursors.Decode(rawCursor)
		if decodeErr != nil || decoded.Version != 1 || decoded.WorkspaceID != access.WorkspaceID ||
			decoded.PrincipalID != access.Actor.PrincipalID || decoded.Limit != limit ||
			decoded.CreatedAt.IsZero() || decoded.EndpointID == uuid.Nil {
			return web.RespondError(ctx, writer, outboundwebhooksdomain.ErrInvalidEndpoint, http.StatusBadRequest)
		}
		cursor = &outboundwebhooksdomain.EndpointCursor{CreatedAt: decoded.CreatedAt, ID: decoded.EndpointID}
	}
	page, err := handlers.service.ListEndpoints(ctx, access, cursor, limit)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	var nextCursor *string
	if page.NextCursor != nil {
		encoded, encodeErr := handlers.cursors.Encode(endpointCursor{
			Version: 1, WorkspaceID: access.WorkspaceID, PrincipalID: access.Actor.PrincipalID,
			CreatedAt: page.NextCursor.CreatedAt.UTC(), EndpointID: page.NextCursor.ID, Limit: limit,
		})
		if encodeErr != nil {
			return web.RespondError(ctx, writer, encodeErr, http.StatusInternalServerError)
		}
		nextCursor = &encoded
	}
	return web.Respond(ctx, writer, endpointPageResponse{
		Items: endpointModels(page.Items), NextCursor: nextCursor,
	}, http.StatusOK)
}

func (handlers *Handlers) CreateEndpoint(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	var body createEndpointRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	created, err := handlers.service.CreateEndpoint(ctx, access, outboundwebhooksservice.CreateEndpointInput{
		Name: body.Name, URL: body.URL, Subscriptions: body.Subscriptions, RequestID: web.GetRequestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	setShowOnceHeaders(writer)
	return web.Respond(ctx, writer, createdEndpointResponse{
		Endpoint: endpointModel(created.Endpoint), SigningSecret: created.Secret.Reveal(),
	}, http.StatusCreated)
}

func (handlers *Handlers) ReplaceSubscriptions(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, endpointID, err := endpointRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	var body replaceSubscriptionsRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	if err := handlers.service.ReplaceSubscriptions(ctx, access, endpointID, body.Subscriptions, web.GetRequestID(ctx)); err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (handlers *Handlers) RotateSecret(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, endpointID, err := endpointRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	secret, generation, previousExpiresAt, err := handlers.service.RotateEndpointSecret(ctx, access, endpointID, web.GetRequestID(ctx))
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	setShowOnceHeaders(writer)
	return web.Respond(ctx, writer, rotatedSecretResponse{
		SigningSecret: secret.Reveal(), Generation: generation,
		PreviousSecretExpiresAt: previousExpiresAt.UTC(),
	}, http.StatusOK)
}

func (handlers *Handlers) DisableEndpoint(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, endpointID, err := endpointRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	var body disableEndpointRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	if err := handlers.service.DisableEndpoint(ctx, access, endpointID, body.Reason, web.GetRequestID(ctx)); err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func endpointRouteAccess(ctx context.Context, request *http.Request) (outboundwebhooksservice.Access, uuid.UUID, error) {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return outboundwebhooksservice.Access{}, uuid.Nil, err
	}
	endpointID, err := uuid.Parse(web.Params(request, "endpointId"))
	return access, endpointID, err
}

func requestErrorStatus(err error) int {
	status := statusForError(err)
	if status == http.StatusInternalServerError {
		return http.StatusBadRequest
	}
	return status
}

func setShowOnceHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Pragma", "no-cache")
}
