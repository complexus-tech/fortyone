package developeroauthhttp

import (
	"context"
	"net/http"

	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (handlers *Handlers) InstallApplication(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}

	var body installApplicationRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementJSONBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}

	installation, err := handlers.service.InstallApplication(ctx, access, developeroauth.InstallApplicationInput{
		ClientID: body.ClientID, Resource: body.Resource, Scopes: body.Scopes,
		RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, applicationInstallationModel(installation), http.StatusCreated)
}

func (handlers *Handlers) ListApplicationInstallations(
	ctx context.Context,
	writer http.ResponseWriter,
	_ *http.Request,
) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}

	installations, err := handlers.service.ListApplicationInstallations(ctx, access)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, applicationInstallationModels(installations), http.StatusOK)
}

func (handlers *Handlers) UpdateApplicationInstallation(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, installationID, err := installationRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}

	var body updateApplicationInstallationRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementJSONBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}

	installation, err := handlers.service.UpdateApplicationInstallation(ctx, access, installationID, developeroauth.UpdateApplicationInstallationInput{
		Resource: body.Resource, Scopes: body.Scopes, RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, applicationInstallationModel(installation), http.StatusOK)
}

func (handlers *Handlers) RevokeApplicationInstallation(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, installationID, err := installationRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}

	err = handlers.service.RevokeApplicationInstallation(ctx, access, installationID, developeroauth.RevokeApplicationInput{
		Reason: administratorRevokedReason, RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}
