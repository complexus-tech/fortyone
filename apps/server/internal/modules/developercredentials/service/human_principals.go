package developercredentials

import (
	"context"
	"errors"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const ensureHumanPrincipalAttempts = 4

type EnsureHumanPrincipalInput struct {
	RequestID string
}

// EnsureHumanPrincipal provides other modules a narrow capability for
// referencing the shared principal registry without table write access. A
// first-party human may provision a missing row; a PAT may only resolve the
// row that necessarily backs that token. The subject always comes from Actor.
func (service *Service) EnsureHumanPrincipal(
	ctx context.Context,
	access developercredentialsdomain.Access,
	input EnsureHumanPrincipalInput,
) (uuid.UUID, error) {
	if err := authorizeHumanPrincipalResolution(access); err != nil {
		return uuid.Nil, err
	}
	userID, err := access.Actor.UserID()
	if err != nil {
		return uuid.Nil, err
	}
	if access.Actor.Kind == platformauth.PrincipalPersonalToken {
		return service.repository.ResolveHumanPrincipal(ctx, access.WorkspaceID, userID)
	}
	principalCandidateID, err := service.nextID()
	if err != nil {
		return uuid.Nil, err
	}
	auditID, err := service.nextID()
	if err != nil {
		return uuid.Nil, err
	}
	now := service.clock.Now().UTC()
	command := developercredentialsdomain.EnsureHumanPrincipal{
		PrincipalCandidateID: principalCandidateID,
		WorkspaceID:          access.WorkspaceID,
		UserID:               userID,
		CreatedAt:            now,
		Audit: auditEvent(
			auditID, access, "human_principal.ensured", "principal", principalCandidateID,
			input.RequestID, now, 0, 0,
		),
	}
	for attempt := 1; attempt <= ensureHumanPrincipalAttempts; attempt++ {
		principalID, ensureErr := service.repository.EnsureHumanPrincipal(ctx, command)
		if !errors.Is(ensureErr, developercredentialsdomain.ErrConcurrentUpdate) || attempt == ensureHumanPrincipalAttempts {
			return principalID, ensureErr
		}
	}
	return uuid.Nil, developercredentialsdomain.ErrConcurrentUpdate
}
