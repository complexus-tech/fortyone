package developercredentialshttp

import (
	"errors"
	"net/http"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
)

func statusForError(err error) int {
	switch {
	case errors.Is(err, developercredentialsdomain.ErrAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, developercredentialsdomain.ErrAuthenticationFailed):
		return http.StatusUnauthorized
	case errors.Is(err, developercredentialsdomain.ErrCredentialNotFound),
		errors.Is(err, developercredentialsdomain.ErrPrincipalNotFound):
		return http.StatusNotFound
	case errors.Is(err, developercredentialsdomain.ErrCredentialRotationConflict),
		errors.Is(err, developercredentialsdomain.ErrConcurrentUpdate),
		errors.Is(err, developercredentialsdomain.ErrTokenPrefixCollision):
		return http.StatusConflict
	case errors.Is(err, developercredentialsdomain.ErrInvalidCredentialKind),
		errors.Is(err, developercredentialsdomain.ErrExpiryRequired),
		errors.Is(err, developercredentialsdomain.ErrExpiryTooSoon),
		errors.Is(err, developercredentialsdomain.ErrExpiryTooLong),
		errors.Is(err, developercredentialsdomain.ErrNoScopes),
		errors.Is(err, developercredentialsdomain.ErrInvalidName),
		errors.Is(err, developercredentialsdomain.ErrInvalidReason),
		errors.Is(err, developercredentialsdomain.ErrInvalidScope),
		errors.Is(err, developercredentialsdomain.ErrInvalidServiceAccountRole),
		errors.Is(err, developercredentialsdomain.ErrInvalidTeamRestriction),
		errors.Is(err, developercredentialsdomain.ErrTeamRestrictionNotAllowed),
		errors.Is(err, developercredentialsdomain.ErrInvalidRotationOverlap):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
