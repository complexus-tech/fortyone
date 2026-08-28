package objectives

import (
	"context"
	"errors"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func actorFor(
	ctx context.Context,
	workspaceID, fallbackUserID uuid.UUID,
	requiredScope platformauth.Scope,
) (platformauth.Actor, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		if !errors.Is(err, platformauth.ErrActorNotFound) || fallbackUserID == uuid.Nil {
			return platformauth.Actor{}, ErrForbidden
		}
		actor = platformauth.NewHumanActor(fallbackUserID)
	}
	if actor.WorkspaceID == uuid.Nil {
		actor, err = actor.WithWorkspace(workspaceID)
		if err != nil {
			return platformauth.Actor{}, ErrForbidden
		}
	}
	if actor.WorkspaceID != workspaceID || !actor.IsUserActor() || !actor.Scopes.Has(requiredScope) {
		return platformauth.Actor{}, ErrForbidden
	}
	if fallbackUserID != uuid.Nil && actor.PrincipalID != fallbackUserID {
		return platformauth.Actor{}, ErrForbidden
	}
	if !actor.TeamAccess.IsUnrestricted() {
		// Objective SQL re-checks product team membership. Credential team
		// restrictions are not yet exposed on legacy objective routes, so fail
		// closed rather than silently widening a restricted credential.
		return platformauth.Actor{}, ErrForbidden
	}
	return actor, nil
}
