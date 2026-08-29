package developeraccess

import (
	"context"
	"errors"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const maximumBearerBytes = 8 << 10

var ErrAuthenticationFailed = errors.New("developer bearer authentication failed")

// MachineCredentialResolver validates PATs and service-account keys against
// their authoritative, revocation-aware credential store.
type MachineCredentialResolver interface {
	ResolveMachineCredential(context.Context, string) (platformauth.Actor, error)
}

// OAuthAccessTokenVerifier validates one exact OAuth audience and rechecks the
// active application, grant, user, and scope set in durable storage.
type OAuthAccessTokenVerifier interface {
	VerifyAccessToken(context.Context, string) (developeroauthdomain.AccessIdentity, error)
}

// Resolver authenticates the bearer credential families published by the
// external API. It does not select a workspace or authorize a resource; those
// decisions remain request and use-case boundaries.
type Resolver struct {
	machine MachineCredentialResolver
	oauth   OAuthAccessTokenVerifier
}

func NewResolver(machine MachineCredentialResolver, oauth OAuthAccessTokenVerifier) (*Resolver, error) {
	if machine == nil || oauth == nil {
		return nil, errors.New("machine credential and OAuth verifiers are required")
	}
	return &Resolver{machine: machine, oauth: oauth}, nil
}

func (resolver *Resolver) ResolveDeveloperCredential(
	ctx context.Context,
	raw string,
) (platformauth.Actor, error) {
	if resolver == nil || len(raw) == 0 || len(raw) > maximumBearerBytes || strings.TrimSpace(raw) != raw {
		return platformauth.Actor{}, ErrAuthenticationFailed
	}
	if actor, err := resolver.machine.ResolveMachineCredential(ctx, raw); err == nil {
		if validMachineActor(actor) {
			return actor, nil
		}
		return platformauth.Actor{}, ErrAuthenticationFailed
	}

	identity, err := resolver.oauth.VerifyAccessToken(ctx, raw)
	if err != nil {
		return platformauth.Actor{}, ErrAuthenticationFailed
	}
	actor, err := oauthActor(identity)
	if err != nil {
		return platformauth.Actor{}, ErrAuthenticationFailed
	}
	return actor, nil
}

func validMachineActor(actor platformauth.Actor) bool {
	if actor.Kind != platformauth.PrincipalPersonalToken && actor.Kind != platformauth.PrincipalServiceAccount {
		return false
	}
	return actor.PrincipalID != uuid.Nil && actor.CredentialID != uuid.Nil && actor.WorkspaceID != uuid.Nil && actor.Validate() == nil
}

func oauthActor(identity developeroauthdomain.AccessIdentity) (platformauth.Actor, error) {
	if identity.ApplicationID == uuid.Nil || strings.TrimSpace(identity.Resource) == "" {
		return platformauth.Actor{}, ErrAuthenticationFailed
	}

	scopes := make([]platformauth.Scope, 0, len(identity.Scopes))
	for _, raw := range identity.Scopes {
		if raw == developeroauthdomain.ScopeOfflineAccess {
			if identity.ActorKind == platformauth.PrincipalOAuthApplication {
				return platformauth.Actor{}, ErrAuthenticationFailed
			}
			continue
		}
		scopes = append(scopes, platformauth.Scope(raw))
	}
	scopeSet, err := platformauth.NewScopeSet(scopes...)
	if err != nil || len(scopeSet.Values()) == 0 {
		return platformauth.Actor{}, ErrAuthenticationFailed
	}

	switch identity.ActorKind {
	case platformauth.PrincipalOAuthUser:
		principalID := identity.PrincipalID
		if principalID == uuid.Nil {
			principalID = identity.UserID
		}
		if principalID == uuid.Nil || identity.UserID != principalID || identity.GrantID == uuid.Nil ||
			identity.InstallationID != uuid.Nil || identity.WorkspaceID != uuid.Nil {
			return platformauth.Actor{}, ErrAuthenticationFailed
		}
		// GrantID is stable across access-token refresh and therefore provides
		// the user credential identity for rate limits and idempotency. The JWT
		// ID remains audit-only on AccessIdentity.
		return platformauth.NewActor(
			principalID,
			platformauth.PrincipalOAuthUser,
			identity.GrantID,
			scopeSet,
			platformauth.UnrestrictedTeamAccess(),
		)
	case platformauth.PrincipalOAuthApplication:
		if identity.PrincipalID == uuid.Nil || identity.InstallationID == uuid.Nil || identity.WorkspaceID == uuid.Nil ||
			identity.UserID != uuid.Nil || identity.GrantID != uuid.Nil {
			return platformauth.Actor{}, ErrAuthenticationFailed
		}
		// InstallationID, not the rotating client secret or access-token JWT ID,
		// is the stable credential identity used by rate limits and idempotency.
		actor, err := platformauth.NewActor(
			identity.PrincipalID,
			platformauth.PrincipalOAuthApplication,
			identity.InstallationID,
			scopeSet,
			platformauth.UnrestrictedTeamAccess(),
		)
		if err != nil {
			return platformauth.Actor{}, err
		}
		return actor.WithWorkspace(identity.WorkspaceID)
	default:
		return platformauth.Actor{}, ErrAuthenticationFailed
	}
}
