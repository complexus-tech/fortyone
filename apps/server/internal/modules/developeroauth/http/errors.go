package developeroauthhttp

import (
	"errors"
	"net/http"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
)

func statusForError(err error) int {
	switch {
	case errors.Is(err, platformauth.ErrActorNotFound),
		errors.Is(err, platformauth.ErrInvalidActor):
		return http.StatusUnauthorized
	case errors.Is(err, developeroauthdomain.ErrAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, developeroauthdomain.ErrApplicationNotFound),
		errors.Is(err, developeroauthdomain.ErrInstallationNotFound),
		errors.Is(err, developeroauthdomain.ErrSecretNotFound):
		return http.StatusNotFound
	case errors.Is(err, developeroauthdomain.ErrConcurrentUpdate),
		errors.Is(err, developeroauthdomain.ErrInstallationRevoked):
		return http.StatusConflict
	case errors.Is(err, developeroauthdomain.ErrInvalidClient),
		errors.Is(err, developeroauthdomain.ErrInvalidExpiry),
		errors.Is(err, developeroauthdomain.ErrInvalidName),
		errors.Is(err, developeroauthdomain.ErrInvalidReason),
		errors.Is(err, developeroauthdomain.ErrInvalidRedirectURI),
		errors.Is(err, developeroauthdomain.ErrInvalidResource),
		errors.Is(err, developeroauthdomain.ErrInvalidRotationOverlap),
		errors.Is(err, developeroauthdomain.ErrInvalidScope):
		return http.StatusBadRequest
	case errors.Is(err, developeroauthdomain.ErrApplicationActorUnavailable),
		errors.Is(err, developeroauthdomain.ErrTokenKeyUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func requestErrorStatus(err error) int {
	if errors.Is(err, errInvalidRouteIdentifier) {
		return http.StatusBadRequest
	}
	return statusForError(err)
}
