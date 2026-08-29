package feedback

import (
	"context"
	"errors"
	"fmt"

	feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

var ErrForbidden = feedbackdomain.ErrForbidden

// CoreAccessScope is the finite authorization projection passed to feedback
// persistence queries. SQL still re-checks current workspace/team membership;
// these credential restrictions only narrow that product authorization.
type CoreAccessScope = feedbackdomain.CoreAccessScope

func accessScopeFromContext(ctx context.Context, workspaceID, fallbackUserID uuid.UUID) (CoreAccessScope, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		if !errors.Is(err, platformauth.ErrActorNotFound) || fallbackUserID == uuid.Nil {
			return CoreAccessScope{}, ErrForbidden
		}
		actor = platformauth.NewHumanActor(fallbackUserID)
	}
	if actor.WorkspaceID == uuid.Nil {
		actor, err = actor.WithWorkspace(workspaceID)
		if err != nil {
			return CoreAccessScope{}, ErrForbidden
		}
	}
	// Feedback's authenticated HTTP surface is currently first-party only.
	// Fail closed for personal tokens, OAuth applications, service accounts, and
	// system actors until a dedicated, schema-backed feedback scope is added.
	if actor.Kind != platformauth.PrincipalHumanUser ||
		actor.WorkspaceID != workspaceID ||
		!actor.Scopes.Has(platformauth.ScopeFirstParty) {
		return CoreAccessScope{}, ErrForbidden
	}
	if fallbackUserID != uuid.Nil && actor.PrincipalID != fallbackUserID {
		return CoreAccessScope{}, ErrForbidden
	}
	scope := CoreAccessScope{
		WorkspaceID:       workspaceID,
		ActorID:           actor.PrincipalID,
		AllTeams:          actor.TeamAccess.IsUnrestricted(),
		CredentialTeamIDs: actor.TeamAccess.RestrictedTeamIDs(),
	}
	if err := scope.Validate(); err != nil {
		return CoreAccessScope{}, fmt.Errorf("%w: %v", ErrForbidden, err)
	}
	return scope, nil
}
