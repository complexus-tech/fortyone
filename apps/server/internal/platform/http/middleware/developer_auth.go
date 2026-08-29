package mid

import (
	"context"
	"errors"
	"net/http"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var errDeveloperAuthenticationRequired = errors.New("valid developer bearer credential required")

// DeveloperCredentialResolver authenticates every bearer family published for
// the versioned external API. Implementations must validate expiry and current
// revocation state before returning an attributed actor.
type DeveloperCredentialResolver interface {
	ResolveDeveloperCredential(context.Context, string) (platformauth.Actor, error)
}

func DeveloperAuthWithErrorResponder(
	log *logger.Logger,
	resolver DeveloperCredentialResolver,
	respond HTTPErrorResponder,
) web.Middleware {
	if resolver == nil {
		panic("developer credential resolver is required")
	}
	if respond == nil {
		panic("developer authentication error responder is required")
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
			rawToken := resolveMachineBearer(request)
			if rawToken == "" {
				return respond(ctx, writer, errDeveloperAuthenticationRequired, http.StatusUnauthorized)
			}
			actor, err := resolver.ResolveDeveloperCredential(ctx, rawToken)
			if err != nil || !validDeveloperActor(actor) {
				log.Warn(ctx, "developer credential authentication rejected")
				return respond(ctx, writer, errDeveloperAuthenticationRequired, http.StatusUnauthorized)
			}
			ctx, err = platformauth.SetActor(ctx, actor)
			if err != nil {
				log.Error(ctx, "developer credential resolver returned an invalid actor")
				return respond(ctx, writer, errDeveloperAuthenticationRequired, http.StatusUnauthorized)
			}
			return next(ctx, writer, request)
		}
	}
}

func validDeveloperActor(actor platformauth.Actor) bool {
	if actor.CredentialID == uuid.Nil || actor.PrincipalID == uuid.Nil || actor.Validate() != nil {
		return false
	}
	switch actor.Kind {
	case platformauth.PrincipalPersonalToken, platformauth.PrincipalServiceAccount,
		platformauth.PrincipalOAuthApplication:
		return actor.WorkspaceID != uuid.Nil
	case platformauth.PrincipalOAuthUser:
		// OAuth user grants are audience- and user-bound, not workspace-bound.
		// The request selects a workspace and every use case rechecks current
		// membership and resource visibility before returning data or mutating.
		return actor.WorkspaceID == uuid.Nil
	default:
		return false
	}
}
