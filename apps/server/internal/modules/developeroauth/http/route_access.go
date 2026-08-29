package developeroauthhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var errInvalidRouteIdentifier = errors.New("invalid route identifier")

func applicationRouteAccess(
	ctx context.Context,
	request *http.Request,
) (developeroauthdomain.ManagementAccess, uuid.UUID, error) {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return developeroauthdomain.ManagementAccess{}, uuid.Nil, err
	}
	applicationID, err := parseRouteUUID(request, "applicationId")
	return access, applicationID, err
}

func clientSecretRouteAccess(
	ctx context.Context,
	request *http.Request,
) (developeroauthdomain.ManagementAccess, uuid.UUID, uuid.UUID, error) {
	access, applicationID, err := applicationRouteAccess(ctx, request)
	if err != nil {
		return developeroauthdomain.ManagementAccess{}, uuid.Nil, uuid.Nil, err
	}
	secretID, err := parseRouteUUID(request, "secretId")
	if err != nil {
		return developeroauthdomain.ManagementAccess{}, uuid.Nil, uuid.Nil, err
	}
	return access, applicationID, secretID, nil
}

func installationRouteAccess(
	ctx context.Context,
	request *http.Request,
) (developeroauthdomain.ManagementAccess, uuid.UUID, error) {
	access, err := humanAccessFromContext(ctx)
	if err != nil {
		return developeroauthdomain.ManagementAccess{}, uuid.Nil, err
	}
	installationID, err := parseRouteUUID(request, "installationId")
	return access, installationID, err
}

func parseRouteUUID(request *http.Request, name string) (uuid.UUID, error) {
	value, err := uuid.Parse(web.Params(request, name))
	if err != nil || value == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a non-zero UUID", errInvalidRouteIdentifier, name)
	}
	return value, nil
}
