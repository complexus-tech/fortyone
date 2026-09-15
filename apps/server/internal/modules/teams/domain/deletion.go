package teamsdomain

import (
	"errors"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

var ErrDeletionForbidden = errors.New("team deletion is not permitted")

type Deletion struct {
	TeamID      uuid.UUID
	WorkspaceID uuid.UUID
	Actor       platformauth.Actor
	DeletedAt   time.Time
}

func (deletion Deletion) Validate() error {
	if deletion.TeamID == uuid.Nil || deletion.WorkspaceID == uuid.Nil || deletion.DeletedAt.IsZero() ||
		deletion.Actor.WorkspaceID != deletion.WorkspaceID || deletion.Actor.Kind != platformauth.PrincipalHumanUser ||
		!deletion.Actor.Scopes.Has(platformauth.ScopeFirstParty) ||
		!deletion.Actor.TeamAccess.Allows(deletion.TeamID) || deletion.Actor.Validate() != nil {
		return ErrDeletionForbidden
	}
	return nil
}
