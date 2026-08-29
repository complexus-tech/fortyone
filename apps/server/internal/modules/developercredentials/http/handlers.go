package developercredentialshttp

import (
	"context"
	"net/http"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Service interface {
	CreatePersonalToken(context.Context, developercredentialsdomain.Access, developercredentials.CreatePersonalTokenInput) (developercredentialsdomain.IssuedCredential, error)
	ListPersonalTokens(context.Context, developercredentialsdomain.Access) ([]developercredentialsdomain.Credential, error)
	RotatePersonalToken(context.Context, developercredentialsdomain.Access, uuid.UUID, developercredentials.RotatePersonalTokenInput) (developercredentialsdomain.IssuedCredential, error)
	RevokePersonalToken(context.Context, developercredentialsdomain.Access, uuid.UUID, developercredentials.RevokeCredentialInput) error
	CreateServiceAccount(context.Context, developercredentialsdomain.Access, developercredentials.CreateServiceAccountInput) (developercredentialsdomain.ServiceAccount, error)
	ListServiceAccounts(context.Context, developercredentialsdomain.Access) ([]developercredentialsdomain.ServiceAccount, error)
	DisableServiceAccount(context.Context, developercredentialsdomain.Access, uuid.UUID, developercredentials.RevokeCredentialInput) error
	CreateServiceAccountKey(context.Context, developercredentialsdomain.Access, uuid.UUID, developercredentials.CreateServiceAccountKeyInput) (developercredentialsdomain.IssuedCredential, error)
	ListServiceAccountKeys(context.Context, developercredentialsdomain.Access, uuid.UUID) ([]developercredentialsdomain.Credential, error)
	RotateServiceAccountKey(context.Context, developercredentialsdomain.Access, uuid.UUID, uuid.UUID, developercredentials.RotateServiceAccountKeyInput) (developercredentialsdomain.IssuedCredential, error)
	RevokeServiceAccountKey(context.Context, developercredentialsdomain.Access, uuid.UUID, uuid.UUID, developercredentials.RevokeCredentialInput) error
}

type Handlers struct{ service Service }

func New(service Service) *Handlers { return &Handlers{service: service} }

func (handlers *Handlers) CreatePersonalToken(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	var body createCredentialRequest
	if err := web.Decode(request, &body); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	issued, err := handlers.service.CreatePersonalToken(ctx, access, personalTokenInput(body, requestID(ctx)))
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	setShowOnceCredentialHeaders(writer)
	return web.Respond(ctx, writer, issuedCredentialModel(issued), http.StatusCreated)
}

func (handlers *Handlers) ListPersonalTokens(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	credentials, err := handlers.service.ListPersonalTokens(ctx, access)
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, credentialModels(credentials), http.StatusOK)
}

func (handlers *Handlers) RotatePersonalToken(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, credentialID, err := credentialRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	var body rotateCredentialRequest
	if err := web.Decode(request, &body); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	issued, err := handlers.service.RotatePersonalToken(ctx, access, credentialID, developercredentials.RotatePersonalTokenInput{
		ExpiresAt: body.ExpiresAt, RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	setShowOnceCredentialHeaders(writer)
	return web.Respond(ctx, writer, issuedCredentialModel(issued), http.StatusCreated)
}

func setShowOnceCredentialHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Pragma", "no-cache")
}

func (handlers *Handlers) RevokePersonalToken(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	access, credentialID, err := credentialRouteAccess(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, requestErrorStatus(err))
	}
	err = handlers.service.RevokePersonalToken(ctx, access, credentialID, developercredentials.RevokeCredentialInput{
		Reason: "user_revoked", RequestID: requestID(ctx),
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, statusForError(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func credentialRouteAccess(ctx context.Context, request *http.Request) (developercredentialsdomain.Access, uuid.UUID, error) {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return developercredentialsdomain.Access{}, uuid.Nil, err
	}
	credentialID, err := uuid.Parse(web.Params(request, "credentialId"))
	return access, credentialID, err
}

func requestErrorStatus(err error) int {
	status := statusForError(err)
	if status == http.StatusInternalServerError {
		return http.StatusBadRequest
	}
	return status
}
