package developercredentialshttp

import (
	"context"
	"net/http"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const maximumOverlapSeconds int64 = 24 * 60 * 60

func (handlers *Handlers) CreateServiceAccount(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	var body createServiceAccountRequest
	if err := web.Decode(request, &body); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	account, err := handlers.service.CreateServiceAccount(ctx, access, developercredentials.CreateServiceAccountInput{
		Name: body.Name, WorkspaceRole: body.WorkspaceRole, RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, serviceAccountModel(account), http.StatusCreated)
}

func (handlers *Handlers) ListServiceAccounts(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	accounts, err := handlers.service.ListServiceAccounts(ctx, access)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, serviceAccountModels(accounts), http.StatusOK)
}

func (handlers *Handlers) DisableServiceAccount(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, principalID, err := serviceAccountRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	err = handlers.service.DisableServiceAccount(ctx, access, principalID, developercredentials.RevokeCredentialInput{
		Reason: "admin_disabled", RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (handlers *Handlers) CreateServiceAccountKey(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, principalID, err := serviceAccountRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	var body createCredentialRequest
	if err := web.Decode(request, &body); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	issued, err := handlers.service.CreateServiceAccountKey(ctx, access, principalID, serviceAccountKeyInput(body, requestID(ctx)))
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, issuedCredentialModel(issued), http.StatusCreated)
}

func (handlers *Handlers) ListServiceAccountKeys(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, principalID, err := serviceAccountRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	credentials, err := handlers.service.ListServiceAccountKeys(ctx, access, principalID)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, credentialModels(credentials), http.StatusOK)
}

func (handlers *Handlers) RotateServiceAccountKey(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, principalID, credentialID, err := serviceAccountKeyRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	var body rotateCredentialRequest
	if err := web.Decode(request, &body); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	if body.OverlapSeconds < 0 || body.OverlapSeconds > maximumOverlapSeconds {
		return web.RespondError(ctx, writer, developercredentialsdomain.ErrInvalidRotationOverlap, http.StatusBadRequest)
	}
	issued, err := handlers.service.RotateServiceAccountKey(ctx, access, principalID, credentialID, developercredentials.RotateServiceAccountKeyInput{
		ExpiresAt: body.ExpiresAt, Overlap: time.Duration(body.OverlapSeconds) * time.Second,
		RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, issuedCredentialModel(issued), http.StatusCreated)
}

func (handlers *Handlers) RevokeServiceAccountKey(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, principalID, credentialID, err := serviceAccountKeyRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	err = handlers.service.RevokeServiceAccountKey(ctx, access, principalID, credentialID, developercredentials.RevokeCredentialInput{
		Reason: "admin_revoked", RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func serviceAccountRouteAccess(ctx context.Context, request *http.Request) (developercredentialsdomain.Access, uuid.UUID, error) {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return developercredentialsdomain.Access{}, uuid.Nil, err
	}
	principalID, err := uuid.Parse(web.Params(request, "principalId"))
	return access, principalID, err
}

func serviceAccountKeyRouteAccess(
	ctx context.Context,
	request *http.Request,
) (developercredentialsdomain.Access, uuid.UUID, uuid.UUID, error) {
	access, principalID, err := serviceAccountRouteAccess(ctx, request)
	if err != nil {
		return developercredentialsdomain.Access{}, uuid.Nil, uuid.Nil, err
	}
	credentialID, err := uuid.Parse(web.Params(request, "credentialId"))
	if err != nil {
		return developercredentialsdomain.Access{}, uuid.Nil, uuid.Nil, err
	}
	return access, principalID, credentialID, nil
}
