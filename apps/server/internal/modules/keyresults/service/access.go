package keyresults

import (
	"context"
	"errors"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) accessFor(
	ctx context.Context,
	workspaceID, fallbackUserID uuid.UUID,
	requiredScope platformauth.Scope,
) (keyresultsdomain.AccessScope, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		if !errors.Is(err, platformauth.ErrActorNotFound) || fallbackUserID == uuid.Nil {
			return keyresultsdomain.AccessScope{}, ErrForbidden
		}
		actor = platformauth.NewHumanActor(fallbackUserID)
	}
	if actor.WorkspaceID == uuid.Nil {
		actor, err = actor.WithWorkspace(workspaceID)
		if err != nil {
			return keyresultsdomain.AccessScope{}, ErrForbidden
		}
	}
	if actor.WorkspaceID != workspaceID || !actor.IsUserActor() || !actor.Scopes.Has(requiredScope) {
		return keyresultsdomain.AccessScope{}, ErrForbidden
	}
	if fallbackUserID != uuid.Nil && actor.PrincipalID != fallbackUserID {
		return keyresultsdomain.AccessScope{}, ErrForbidden
	}
	return keyresultsdomain.AccessScope{
		WorkspaceID: workspaceID, ActorID: actor.PrincipalID,
		AllTeams: actor.TeamAccess.IsUnrestricted(),
		TeamIDs:  actor.TeamAccess.RestrictedTeamIDs(),
	}, nil
}
