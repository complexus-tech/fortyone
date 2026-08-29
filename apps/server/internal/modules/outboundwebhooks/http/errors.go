package outboundwebhookshttp

import (
	"errors"
	"net/http"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
)

var errAccessDenied = errors.New("outbound webhook management access denied")

func statusForError(err error) int {
	switch {
	case errors.Is(err, platformauth.ErrActorNotFound),
		errors.Is(err, platformauth.ErrInvalidActor):
		return http.StatusUnauthorized
	case errors.Is(err, errAccessDenied),
		errors.Is(err, outboundwebhooksdomain.ErrEndpointOwnerInactive):
		return http.StatusForbidden
	case errors.Is(err, outboundwebhooksdomain.ErrEndpointNotFound):
		return http.StatusNotFound
	case errors.Is(err, outboundwebhooksdomain.ErrEndpointConflict),
		errors.Is(err, outboundwebhooksdomain.ErrEndpointDisabled):
		return http.StatusConflict
	case errors.Is(err, outboundwebhooksdomain.ErrInvalidEndpoint),
		errors.Is(err, outboundwebhooksdomain.ErrInvalidEventType),
		errors.Is(err, outboundwebhooksdomain.ErrInvalidSubscription):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
