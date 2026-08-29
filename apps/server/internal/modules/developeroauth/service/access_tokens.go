package developeroauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type accessTokenClaims struct {
	jwt.RegisteredClaims
	Scope          string `json:"scope"`
	ClientID       string `json:"client_id"`
	ApplicationID  string `json:"application_id"`
	GrantID        string `json:"grant_id,omitempty"`
	InstallationID string `json:"installation_id,omitempty"`
	WorkspaceID    string `json:"workspace_id,omitempty"`
	ActorKind      string `json:"actor_kind"`
}

func (service *Service) signAccessToken(
	grant developeroauthdomain.Grant,
	issuedAt time.Time,
) (developeroauthdomain.PlaintextSecret, error) {
	tokenID, err := service.nextID()
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, fmt.Errorf("generate OAuth access token ID: %w", err)
	}
	expiresAt := issuedAt.Add(service.accessTokenTTL)
	claims := accessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: service.issuer, Subject: grant.UserID.String(), Audience: jwt.ClaimStrings{grant.Resource},
			ExpiresAt: jwt.NewNumericDate(expiresAt), IssuedAt: jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt), ID: tokenID.String(),
		},
		Scope: ScopeString(grant.Scopes), ClientID: grant.ClientID,
		ApplicationID: grant.ApplicationID.String(), GrantID: grant.ID.String(), ActorKind: string(grant.ActorKind),
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.accessTokenSigningKey)
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, fmt.Errorf("sign OAuth access token: %w", err)
	}
	return developeroauthdomain.NewPlaintextSecret(raw), nil
}

func (service *Service) signApplicationAccessToken(
	installation developeroauthdomain.ApplicationInstallation,
	tokenID uuid.UUID,
	issuedAt time.Time,
) (developeroauthdomain.PlaintextSecret, error) {
	if tokenID == uuid.Nil {
		return developeroauthdomain.PlaintextSecret{}, errors.New("OAuth application access token ID is required")
	}
	expiresAt := issuedAt.Add(service.accessTokenTTL)
	claims := accessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: service.issuer, Subject: installation.PrincipalID.String(),
			Audience: jwt.ClaimStrings{installation.Resource}, ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt: jwt.NewNumericDate(issuedAt), NotBefore: jwt.NewNumericDate(issuedAt), ID: tokenID.String(),
		},
		Scope: ScopeString(installation.Scopes), ClientID: installation.ClientID,
		ApplicationID: installation.ApplicationID.String(), InstallationID: installation.ID.String(),
		WorkspaceID: installation.WorkspaceID.String(), ActorKind: string(platformauth.PrincipalOAuthApplication),
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.accessTokenSigningKey)
	if err != nil {
		return developeroauthdomain.PlaintextSecret{}, fmt.Errorf("sign OAuth application access token: %w", err)
	}
	return developeroauthdomain.NewPlaintextSecret(raw), nil
}

func (service *Service) VerifyAccessToken(
	ctx context.Context,
	raw string,
) (developeroauthdomain.AccessIdentity, error) {
	claims := &accessTokenClaims{}
	parsed, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected OAuth access token algorithm")
			}
			return service.accessTokenSigningKey, nil
		},
		jwt.WithAudience(service.resource),
		jwt.WithIssuer(service.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(30*time.Second),
		jwt.WithTimeFunc(func() time.Time { return service.clock.Now().UTC() }),
	)
	if err != nil || !parsed.Valid || claims.ExpiresAt == nil || claims.Subject == "" || claims.ID == "" {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	applicationID, applicationErr := uuid.Parse(claims.ApplicationID)
	credentialID, credentialErr := uuid.Parse(claims.ID)
	if errors.Join(applicationErr, credentialErr) != nil || strings.TrimSpace(claims.ClientID) == "" {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	now := service.clock.Now().UTC()
	claimScopes := strings.Fields(claims.Scope)
	switch platformauth.PrincipalKind(claims.ActorKind) {
	case platformauth.PrincipalOAuthUser:
		return service.verifyUserAccessToken(ctx, *claims, applicationID, credentialID, claimScopes, now)
	case platformauth.PrincipalOAuthApplication:
		return service.verifyApplicationAccessToken(ctx, *claims, applicationID, credentialID, claimScopes, now)
	default:
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
}

func (service *Service) verifyUserAccessToken(
	ctx context.Context,
	claims accessTokenClaims,
	applicationID uuid.UUID,
	credentialID uuid.UUID,
	claimScopes []string,
	now time.Time,
) (developeroauthdomain.AccessIdentity, error) {
	if claims.InstallationID != "" || claims.WorkspaceID != "" || claims.GrantID == "" {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	userID, userErr := uuid.Parse(claims.Subject)
	grantID, grantErr := uuid.Parse(claims.GrantID)
	if errors.Join(userErr, grantErr) != nil {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	grant, err := service.repository.GetActiveGrant(ctx, grantID, applicationID, service.resource, now)
	if err != nil {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	if grant.UserID != userID || grant.ClientID != claims.ClientID || grant.ActorKind != platformauth.PrincipalOAuthUser ||
		!scopesAreSubset(claimScopes, grant.Scopes) || !service.scopePolicy.accepts(claimScopes) {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	if err := service.repository.TouchGrant(ctx, grant.ID, now, now.Add(-grantTouchInterval)); err != nil {
		return developeroauthdomain.AccessIdentity{}, err
	}
	return developeroauthdomain.AccessIdentity{
		PrincipalID: grant.UserID, UserID: grant.UserID, ApplicationID: grant.ApplicationID, GrantID: grant.ID,
		ActorKind: grant.ActorKind, ClientID: grant.ClientID, Resource: grant.Resource,
		Scopes: claimScopes, ExpiresAt: claims.ExpiresAt.Time, OAuthCredential: credentialID,
	}, nil
}

func (service *Service) verifyApplicationAccessToken(
	ctx context.Context,
	claims accessTokenClaims,
	applicationID uuid.UUID,
	credentialID uuid.UUID,
	claimScopes []string,
	now time.Time,
) (developeroauthdomain.AccessIdentity, error) {
	if service.applicationActors == nil || claims.GrantID != "" || claims.InstallationID == "" || claims.WorkspaceID == "" {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	installationID, installationErr := uuid.Parse(claims.InstallationID)
	if installationErr != nil {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	installation, err := service.applicationActors.GetActiveApplicationInstallation(
		ctx,
		installationID,
		applicationID,
		service.resource,
		now,
	)
	if err != nil || validateApplicationInstallationIdentity(installation, claims) != nil ||
		installation.Resource != service.resource ||
		!scopesAreSubset(claimScopes, installation.Scopes) || !service.applicationActorScopes.accepts(claimScopes) {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	if err := service.applicationActors.TouchApplicationInstallation(
		ctx,
		installation.ID,
		now,
		now.Add(-grantTouchInterval),
	); err != nil {
		return developeroauthdomain.AccessIdentity{}, err
	}
	identity := installation.AccessIdentity(claims.ExpiresAt.Time, credentialID)
	identity.Scopes = append([]string(nil), claimScopes...)
	return identity, nil
}
