package developercredentials

import (
	"context"
	"errors"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
)

const lastUsedWriteInterval = 15 * time.Minute

// ResolveMachineCredential verifies a PAT or service-account key and returns a
// fully attributed actor. It deliberately has no legacy session/JWT fallback.
func (service *Service) ResolveMachineCredential(ctx context.Context, rawToken string) (platformauth.Actor, error) {
	parsed, err := parseToken(rawToken)
	if err != nil {
		return platformauth.Actor{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	now := service.clock.Now().UTC()
	record, err := service.repository.LookupCredential(ctx, parsed.Prefix, parsed.Kind, parsed.Version, now)
	if err != nil {
		return platformauth.Actor{}, authenticationError(err)
	}
	if err := service.tokens.verify(rawToken, record); err != nil {
		return platformauth.Actor{}, authenticationError(err)
	}

	scopes, err := platformauth.NewScopeSet(record.Scopes...)
	if err != nil || scopes.Has(platformauth.ScopeFirstParty) {
		return platformauth.Actor{}, authenticationError(err)
	}
	teamAccess := platformauth.UnrestrictedTeamAccess()
	if len(record.TeamIDs) > 0 {
		teamAccess, err = platformauth.RestrictedTeamAccess(record.TeamIDs...)
		if err != nil {
			return platformauth.Actor{}, authenticationError(err)
		}
	}

	principalID := record.PrincipalRecord
	principalKind := platformauth.PrincipalServiceAccount
	if record.CredentialKind == developercredentialsdomain.CredentialPersonalAccessToken {
		if record.SubjectUserID == nil {
			return platformauth.Actor{}, developercredentialsdomain.ErrAuthenticationFailed
		}
		principalID = *record.SubjectUserID
		principalKind = platformauth.PrincipalPersonalToken
	}
	actor, err := platformauth.NewActor(principalID, principalKind, record.CredentialID, scopes, teamAccess)
	if err != nil {
		return platformauth.Actor{}, authenticationError(err)
	}
	actor, err = actor.WithWorkspace(record.WorkspaceID)
	if err != nil {
		return platformauth.Actor{}, authenticationError(err)
	}
	if err := service.repository.ConfirmCredentialActiveAndTouch(ctx, record.CredentialID, now, now.Add(-lastUsedWriteInterval)); err != nil {
		return platformauth.Actor{}, authenticationError(err)
	}
	return actor, nil
}

func authenticationError(cause error) error {
	if cause == nil {
		cause = errors.New("invalid credential state")
	}
	return errors.Join(developercredentialsdomain.ErrAuthenticationFailed, cause)
}
