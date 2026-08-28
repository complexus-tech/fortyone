package apiv1http

import (
	"context"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (s *server) webhookAccess(ctx context.Context, workspaceID openapiv1.ComponentsCommonWorkspaceId) (outboundwebhooksservice.Access, *failure) {
	actor, problem := actorFor(ctx, workspaceID, platformauth.ScopeWebhooksManage)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	if problem != nil {
		return outboundwebhooksservice.Access{}, problem
	}
	workspace, err := s.workspaces.Get(ctx, workspaceID, actor.PrincipalID)
	if err != nil {
		return outboundwebhooksservice.Access{}, classifyFailure(err)
	}
	role := authorization.WorkspaceRole(workspace.UserRole)
	if err := authorization.ValidateWorkspaceRole(role); err != nil {
		return outboundwebhooksservice.Access{}, classifyFailure(err)
	}
	return outboundwebhooksservice.Access{Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: role}, nil
}

func (s *server) ListWebhookEndpoints(ctx context.Context, request openapiv1.ListWebhookEndpointsRequestObject) (openapiv1.ListWebhookEndpointsResponseObject, error) {
	access, problem := s.webhookAccess(ctx, request.WorkspaceId)
	if problem != nil {
		return listWebhookEndpointsFailure(ctx, problem), nil
	}
	limit := defaultPageLimit
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	if limit < 1 || limit > maximumPageLimit {
		return listWebhookEndpointsFailure(ctx, invalidCursor()), nil
	}
	var domainCursor *outboundwebhooksdomain.EndpointCursor
	if request.Params.Cursor != nil {
		decoded, err := s.webhookCursors.Decode(*request.Params.Cursor)
		if err != nil || decoded.Version != 1 || decoded.WorkspaceID != access.WorkspaceID ||
			decoded.PrincipalID != access.Actor.PrincipalID || decoded.Limit != limit ||
			decoded.CreatedAt.IsZero() || decoded.EndpointID == uuid.Nil {
			return listWebhookEndpointsFailure(ctx, invalidCursor()), nil
		}
		domainCursor = &outboundwebhooksdomain.EndpointCursor{CreatedAt: decoded.CreatedAt, ID: decoded.EndpointID}
	}
	page, err := s.webhooks.ListEndpoints(ctx, access, domainCursor, limit)
	if err != nil {
		return listWebhookEndpointsFailure(ctx, classifyFailure(err)), nil
	}
	data := make([]openapiv1.ComponentsWebhooksWebhookEndpoint, len(page.Items))
	for index, endpoint := range page.Items {
		data[index] = webhookModel(endpoint)
	}
	var nextCursor *string
	if page.NextCursor != nil {
		encoded, err := s.webhookCursors.Encode(webhookCursor{
			Version: 1, WorkspaceID: access.WorkspaceID, PrincipalID: access.Actor.PrincipalID,
			CreatedAt: page.NextCursor.CreatedAt.UTC(), EndpointID: page.NextCursor.ID, Limit: limit,
		})
		if err != nil {
			s.log.Error(ctx, "failed to encode public API webhook cursor")
			return listWebhookEndpointsFailure(ctx, statusFailure(http.StatusInternalServerError)), nil
		}
		nextCursor = &encoded
	}
	return openapiv1.ListWebhookEndpoints200JSONResponse{Body: openapiv1.ComponentsWebhooksWebhookEndpointPageResponse{
		Data: data, Meta: openapiv1.ComponentsCommonPageMeta{HasMore: nextCursor != nil, NextCursor: nextCursor},
	}}, nil
}

func (s *server) CreateWebhookEndpoint(ctx context.Context, request openapiv1.CreateWebhookEndpointRequestObject) (openapiv1.CreateWebhookEndpointResponseObject, error) {
	access, problem := s.webhookAccess(ctx, request.WorkspaceId)
	if problem != nil {
		return createWebhookEndpointFailure(ctx, problem), nil
	}
	if request.Body == nil {
		return createWebhookEndpointFailure(ctx, statusFailure(http.StatusBadRequest)), nil
	}
	created, err := s.webhooks.CreateEndpoint(ctx, access, outboundwebhooksservice.CreateEndpointInput{
		Name: request.Body.Name, URL: request.Body.Url,
		Subscriptions: webhookEventTypes(request.Body.Subscriptions), RequestID: web.GetRequestID(ctx),
	})
	if err != nil {
		return createWebhookEndpointFailure(ctx, classifyFailure(err)), nil
	}
	secret := created.Secret.Reveal()
	return openapiv1.CreateWebhookEndpoint201JSONResponse{Body: openapiv1.ComponentsWebhooksCreatedWebhookEndpointResponse{
		Data: webhookModel(created.Endpoint), SigningSecret: &secret,
	}}, nil
}

func (s *server) GetWebhookEndpoint(ctx context.Context, request openapiv1.GetWebhookEndpointRequestObject) (openapiv1.GetWebhookEndpointResponseObject, error) {
	access, problem := s.webhookAccess(ctx, request.WorkspaceId)
	if problem != nil {
		return getWebhookEndpointFailure(ctx, problem), nil
	}
	endpoint, err := s.webhooks.GetEndpoint(ctx, access, request.EndpointId)
	if err != nil {
		return getWebhookEndpointFailure(ctx, classifyFailure(err)), nil
	}
	return openapiv1.GetWebhookEndpoint200JSONResponse{Body: openapiv1.ComponentsWebhooksWebhookEndpointResponse{Data: webhookModel(endpoint)}}, nil
}

func (s *server) ReplaceWebhookSubscriptions(ctx context.Context, request openapiv1.ReplaceWebhookSubscriptionsRequestObject) (openapiv1.ReplaceWebhookSubscriptionsResponseObject, error) {
	access, problem := s.webhookAccess(ctx, request.WorkspaceId)
	if problem != nil {
		return replaceWebhookSubscriptionsFailure(ctx, problem), nil
	}
	if request.Body == nil {
		return replaceWebhookSubscriptionsFailure(ctx, statusFailure(http.StatusBadRequest)), nil
	}
	err := s.webhooks.ReplaceSubscriptions(ctx, access, request.EndpointId, webhookEventTypes(request.Body.Subscriptions), web.GetRequestID(ctx))
	if err != nil {
		return replaceWebhookSubscriptionsFailure(ctx, classifyFailure(err)), nil
	}
	return openapiv1.ReplaceWebhookSubscriptions204Response{}, nil
}

func (s *server) DisableWebhookEndpoint(ctx context.Context, request openapiv1.DisableWebhookEndpointRequestObject) (openapiv1.DisableWebhookEndpointResponseObject, error) {
	access, problem := s.webhookAccess(ctx, request.WorkspaceId)
	if problem != nil {
		return disableWebhookEndpointFailure(ctx, problem), nil
	}
	if request.Body == nil {
		return disableWebhookEndpointFailure(ctx, statusFailure(http.StatusBadRequest)), nil
	}
	err := s.webhooks.DisableEndpoint(ctx, access, request.EndpointId, request.Body.Reason, web.GetRequestID(ctx))
	if err != nil {
		return disableWebhookEndpointFailure(ctx, classifyFailure(err)), nil
	}
	return openapiv1.DisableWebhookEndpoint204Response{}, nil
}

func (s *server) RotateWebhookSecret(ctx context.Context, request openapiv1.RotateWebhookSecretRequestObject) (openapiv1.RotateWebhookSecretResponseObject, error) {
	access, problem := s.webhookAccess(ctx, request.WorkspaceId)
	if problem != nil {
		return rotateWebhookSecretFailure(ctx, problem), nil
	}
	secret, generation, expiresAt, err := s.webhooks.RotateEndpointSecret(ctx, access, request.EndpointId, web.GetRequestID(ctx))
	if err != nil {
		return rotateWebhookSecretFailure(ctx, classifyFailure(err)), nil
	}
	revealed := secret.Reveal()
	return openapiv1.RotateWebhookSecret200JSONResponse{Body: openapiv1.ComponentsWebhooksRotateWebhookSecretResponse{
		Generation: generation, PreviousSecretExpiresAt: expiresAt.UTC(), SigningSecret: &revealed,
	}}, nil
}

func webhookEventTypes(values []openapiv1.ComponentsWebhooksWebhookEventType) []outboundwebhooksdomain.EventType {
	result := make([]outboundwebhooksdomain.EventType, len(values))
	for index, value := range values {
		result[index] = outboundwebhooksdomain.EventType(value)
	}
	return result
}

func listWebhookEndpointsFailure(ctx context.Context, problem *failure) openapiv1.ListWebhookEndpointsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListWebhookEndpoints400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.ListWebhookEndpoints403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	default:
		return openapiv1.ListWebhookEndpoints500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func createWebhookEndpointFailure(ctx context.Context, problem *failure) openapiv1.CreateWebhookEndpointResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.CreateWebhookEndpoint400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.CreateWebhookEndpoint403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusConflict:
		return openapiv1.CreateWebhookEndpoint409JSONResponse{ComponentsCommonConflictJSONResponse: openapiv1.ComponentsCommonConflictJSONResponse{Body: body, Headers: conflictResponseHeaders(headers)}}
	default:
		return openapiv1.CreateWebhookEndpoint500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func getWebhookEndpointFailure(ctx context.Context, problem *failure) openapiv1.GetWebhookEndpointResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.GetWebhookEndpoint400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.GetWebhookEndpoint403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusNotFound:
		return openapiv1.GetWebhookEndpoint404JSONResponse{ComponentsCommonNotFoundJSONResponse: openapiv1.ComponentsCommonNotFoundJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers)}}
	default:
		return openapiv1.GetWebhookEndpoint500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func replaceWebhookSubscriptionsFailure(ctx context.Context, problem *failure) openapiv1.ReplaceWebhookSubscriptionsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ReplaceWebhookSubscriptions400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.ReplaceWebhookSubscriptions403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusNotFound:
		return openapiv1.ReplaceWebhookSubscriptions404JSONResponse{ComponentsCommonNotFoundJSONResponse: openapiv1.ComponentsCommonNotFoundJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers)}}
	case http.StatusConflict:
		return openapiv1.ReplaceWebhookSubscriptions409JSONResponse{ComponentsCommonConflictJSONResponse: openapiv1.ComponentsCommonConflictJSONResponse{Body: body, Headers: conflictResponseHeaders(headers)}}
	default:
		return openapiv1.ReplaceWebhookSubscriptions500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func disableWebhookEndpointFailure(ctx context.Context, problem *failure) openapiv1.DisableWebhookEndpointResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.DisableWebhookEndpoint400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.DisableWebhookEndpoint403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusNotFound:
		return openapiv1.DisableWebhookEndpoint404JSONResponse{ComponentsCommonNotFoundJSONResponse: openapiv1.ComponentsCommonNotFoundJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers)}}
	case http.StatusConflict:
		return openapiv1.DisableWebhookEndpoint409JSONResponse{ComponentsCommonConflictJSONResponse: openapiv1.ComponentsCommonConflictJSONResponse{Body: body, Headers: conflictResponseHeaders(headers)}}
	default:
		return openapiv1.DisableWebhookEndpoint500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func rotateWebhookSecretFailure(ctx context.Context, problem *failure) openapiv1.RotateWebhookSecretResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.RotateWebhookSecret400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.RotateWebhookSecret403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusNotFound:
		return openapiv1.RotateWebhookSecret404JSONResponse{ComponentsCommonNotFoundJSONResponse: openapiv1.ComponentsCommonNotFoundJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers)}}
	case http.StatusConflict:
		return openapiv1.RotateWebhookSecret409JSONResponse{ComponentsCommonConflictJSONResponse: openapiv1.ComponentsCommonConflictJSONResponse{Body: body, Headers: conflictResponseHeaders(headers)}}
	default:
		return openapiv1.RotateWebhookSecret500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func conflictResponseHeaders(headers openapiv1.ComponentsCommonBadRequestResponseHeaders) openapiv1.ComponentsCommonConflictResponseHeaders {
	return openapiv1.ComponentsCommonConflictResponseHeaders{XRequestID: headers.XRequestID}
}
