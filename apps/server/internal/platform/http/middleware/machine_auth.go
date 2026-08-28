package mid

import (
	"context"
	"errors"
	"net/http"
	"strings"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var errMachineAuthenticationRequired = errors.New("valid machine bearer credential required")

// HTTPErrorResponder lets a versioned transport preserve its own stable error
// envelope without duplicating bearer parsing or actor validation.
type HTTPErrorResponder func(context.Context, http.ResponseWriter, error, int) error

// MachineCredentialResolver is the narrow boundary between HTTP bearer
// extraction and credential verification. Implementations must return an actor
// already bound to the credential's authoritative workspace.
type MachineCredentialResolver interface {
	ResolveMachineCredential(context.Context, string) (platformauth.Actor, error)
}

// MachineAuth authenticates PATs and service-account keys only. It deliberately
// ignores cookies, legacy JWTs, and query parameters and must only be installed
// on the versioned external API, never on legacy first-party routes.
func MachineAuth(log *logger.Logger, resolver MachineCredentialResolver) web.Middleware {
	return MachineAuthWithErrorResponder(log, resolver, web.RespondError)
}

// MachineAuthWithErrorResponder is MachineAuth with a transport-owned error
// encoder. Authentication semantics remain identical and fail closed.
func MachineAuthWithErrorResponder(
	log *logger.Logger,
	resolver MachineCredentialResolver,
	respond HTTPErrorResponder,
) web.Middleware {
	if resolver == nil {
		panic("machine credential resolver is required")
	}
	if respond == nil {
		panic("machine authentication error responder is required")
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
			rawToken := resolveMachineBearer(request)
			if rawToken == "" {
				return respond(ctx, writer, errMachineAuthenticationRequired, http.StatusUnauthorized)
			}
			actor, err := resolver.ResolveMachineCredential(ctx, rawToken)
			if err != nil || !validMachineActor(actor) {
				log.Warn(ctx, "machine credential authentication rejected")
				return respond(ctx, writer, errMachineAuthenticationRequired, http.StatusUnauthorized)
			}
			ctx, err = platformauth.SetActor(ctx, actor)
			if err != nil {
				log.Error(ctx, "machine credential resolver returned an invalid actor")
				return respond(ctx, writer, errMachineAuthenticationRequired, http.StatusUnauthorized)
			}
			return next(ctx, writer, request)
		}
	}
}

func validMachineActor(actor platformauth.Actor) bool {
	return (actor.Kind == platformauth.PrincipalPersonalToken || actor.Kind == platformauth.PrincipalServiceAccount) &&
		actor.CredentialID != uuid.Nil && actor.WorkspaceID != uuid.Nil
}

func resolveMachineBearer(request *http.Request) string {
	values := request.Header.Values("Authorization")
	if len(values) != 1 {
		return ""
	}
	parts := strings.Fields(strings.TrimSpace(values[0]))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
