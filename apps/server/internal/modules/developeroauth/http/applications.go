package developeroauthhttp

import (
	"context"
	"net/http"
	"time"

	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const administratorRevokedReason = "admin_revoked"

func (handlers *Handlers) CreateManagedApplication(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}

	var body createManagedApplicationRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementJSONBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}

	issued, err := handlers.service.CreateManagedApplication(ctx, access, developeroauth.CreateManagedApplicationInput{
		Name: body.Name, RedirectURIs: body.RedirectURIs, ExpiresAt: body.ExpiresAt,
		SecretExpiresAt: body.SecretExpiresAt, RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	setShowOnceSecretHeaders(writer)
	return web.Respond(ctx, writer, issuedManagedApplicationModel(issued), http.StatusCreated)
}

func (handlers *Handlers) ListManagedApplications(
	ctx context.Context,
	writer http.ResponseWriter,
	_ *http.Request,
) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}

	applications, err := handlers.service.ListManagedApplications(ctx, access)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, managedApplicationModels(applications), http.StatusOK)
}

func (handlers *Handlers) ListClientSecrets(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, applicationID, err := applicationRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}

	secrets, err := handlers.service.ListClientSecrets(ctx, access, applicationID)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, clientSecretModels(secrets), http.StatusOK)
}

func (handlers *Handlers) RotateClientSecret(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, applicationID, err := applicationRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}

	var body rotateClientSecretRequest
	if err := web.DecodeWithLimit(request, &body, maximumManagementJSONBytes); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}

	issued, err := handlers.service.RotateClientSecret(ctx, access, applicationID, developeroauth.RotateClientSecretInput{
		ExpiresAt: body.ExpiresAt,
		Overlap:   time.Duration(*body.OverlapSeconds) * time.Second,
		RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	setShowOnceSecretHeaders(writer)
	return web.Respond(ctx, writer, issuedClientSecretModel(issued), http.StatusCreated)
}

func setShowOnceSecretHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Pragma", "no-cache")
}

func (handlers *Handlers) RevokeClientSecret(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
) error {
	access, applicationID, secretID, err := clientSecretRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}

	err = handlers.service.RevokeClientSecret(ctx, access, applicationID, secretID, developeroauth.RevokeApplicationInput{
		Reason: administratorRevokedReason, RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}
