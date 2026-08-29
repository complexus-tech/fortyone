package feedbackdomain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("feedback access is not permitted")

// CoreAccessScope is the finite authorization projection accepted by
// persistence. It is populated by the service from the authenticated actor;
// repositories never read ambient authentication context.
type CoreAccessScope struct {
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	AllTeams          bool
	CredentialTeamIDs []uuid.UUID
}

func (scope CoreAccessScope) Validate() error {
	if scope.WorkspaceID == uuid.Nil || scope.ActorID == uuid.Nil {
		return errors.New("workspace and actor ids are required")
	}
	if scope.AllTeams && len(scope.CredentialTeamIDs) > 0 {
		return errors.New("team access cannot be both unrestricted and restricted")
	}
	for _, teamID := range scope.CredentialTeamIDs {
		if teamID == uuid.Nil {
			return errors.New("credential team ids cannot contain a zero id")
		}
	}
	return nil
}
